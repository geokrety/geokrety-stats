// Package store provides the PostgreSQL-backed data access layer for the scoring pipeline.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

// pgStore implements Store using a PostgreSQL pgxpool.
type pgStore struct {
	pool *pgxpool.Pool
}

// New creates a Store backed by the given connection pool.
func New(pool *pgxpool.Pool) Store {
	return &pgStore{pool: pool}
}

// ---- Event idempotency ----

func (s *pgStore) IsEventProcessed(ctx context.Context, moveID int64) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM geokrety_stats.processed_events WHERE move_id = $1)`,
		moveID,
	).Scan(&exists)
	return exists, err
}

func (s *pgStore) MarkEventProcessed(ctx context.Context, moveID int64, result string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO geokrety_stats.processed_events (move_id, processed_at, pipeline_result)
		 VALUES ($1, NOW(), $2)
		 ON CONFLICT (move_id) DO UPDATE SET pipeline_result = $2, processed_at = NOW()`,
		moveID, result,
	)
	return err
}

// ---- GeoKret read (geokrety schema, read-only) ----

func (s *pgStore) GetMove(ctx context.Context, moveID int64) (*GKMoveRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, geokret, author, move_type, waypoint, country, lat, lon, moved_on_datetime
		FROM geokrety.gk_moves
		WHERE id = $1
	`, moveID)

	r := &GKMoveRow{}
	err := row.Scan(
		&r.ID, &r.GeoKretID, &r.Author, &r.MoveType, &r.Waypoint,
		&r.Country, &r.Lat, &r.Lon, &r.MovedOnDatetime,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return r, err
}

func (s *pgStore) GetGK(ctx context.Context, gkID int64) (*GKRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, type, owner, created_on_datetime, holder
		FROM geokrety.gk_geokrety
		WHERE id = $1
	`, gkID)

	r := &GKRow{}
	err := row.Scan(&r.ID, &r.Type, &r.Owner, &r.CreatedAt, &r.Holder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return r, err
}

// TODO such query is really un-efficace. We should maintain a separate table with the previous holder to avoid doing this complex query on every move.
func (s *pgStore) GetGKPreviousHolder(ctx context.Context, gkID int64) (int64, error) {
	// The "previous holder" is the holder just before the current holder.
	// We look at the 2 most recent GRAB moves on this GK and take the second one.
	var userID *int64
	err := s.pool.QueryRow(ctx, `
		SELECT author
		FROM geokrety.gk_moves
		WHERE geokret = $1 AND move_type = 1 AND author IS NOT NULL
		ORDER BY moved_on_datetime DESC
		LIMIT 1 OFFSET 1
	`, gkID).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) || userID == nil {
		return 0, nil
	}
	return *userID, err
}

func (s *pgStore) GetGKHomeCountry(ctx context.Context, gkID int64) (string, error) {
	var country *string
	err := s.pool.QueryRow(ctx, `
		SELECT country
		FROM geokrety.gk_moves
		WHERE geokret = $1 AND country IS NOT NULL AND move_type IN (0, 3, 5)
		ORDER BY moved_on_datetime ASC
		LIMIT 1
	`, gkID).Scan(&country)

	if errors.Is(err, pgx.ErrNoRows) || country == nil {
		return "", nil
	}
	return *country, err
}

func (s *pgStore) GetGKLastDrop(ctx context.Context, gkID int64, before time.Time) (dropAt *time.Time, dropUser int64, err error) {
	var userID *int64
	var dt *time.Time

	err = s.pool.QueryRow(ctx, `
		SELECT moved_on_datetime, author
		FROM geokrety.gk_moves
		WHERE geokret = $1 AND move_type = 0 AND moved_on_datetime < $2 AND author IS NOT NULL
		ORDER BY moved_on_datetime DESC
		LIMIT 1
	`, gkID, before).Scan(&dt, &userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}

	if userID != nil {
		dropUser = *userID
	}
	return dt, dropUser, nil
}

func (s *pgStore) GetGKLastCacheEntry(ctx context.Context, gkID int64, before time.Time) (*time.Time, error) {
	var dt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT moved_on_datetime
		FROM geokrety.gk_moves
		WHERE geokret = $1 AND move_type IN (0, 3) AND moved_on_datetime < $2
		ORDER BY moved_on_datetime DESC
		LIMIT 1
	`, gkID, before).Scan(&dt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return dt, err
}

func (s *pgStore) GetGKDistinctUsers6M(ctx context.Context, gkID int64, before time.Time) ([]int64, error) {
	sixMonthsAgo := before.AddDate(0, -6, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT m.author
		FROM geokrety.gk_moves m
		INNER JOIN geokrety_stats.user_points_log upl ON upl.move_id = m.id
		WHERE m.geokret = $1
		  AND m.moved_on_datetime >= $2
		  AND m.moved_on_datetime < $3
		  AND m.author IS NOT NULL
		  AND m.author != (SELECT owner FROM geokrety.gk_geokrety WHERE id = $1)
		  AND upl.points > 0
	`, gkID, sixMonthsAgo, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		users = append(users, uid)
	}
	return users, rows.Err()
}

// ---- Multiplier state ----

func (s *pgStore) GetGKMultiplierState(ctx context.Context, gkID int64) (
	multiplier float64, lastUpdatedAt time.Time, holderID int64, holderAcquiredAt *time.Time, err error,
) {
	row := s.pool.QueryRow(ctx, `
		SELECT current_multiplier, last_updated_at, COALESCE(current_holder_id, 0), holder_acquired_at
		FROM geokrety_stats.gk_multiplier_state
		WHERE gk_id = $1
	`, gkID)

	err = row.Scan(&multiplier, &lastUpdatedAt, &holderID, &holderAcquiredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 1.0, time.Now(), 0, nil, nil
	}
	return
}

func (s *pgStore) SaveGKMultiplierState(ctx context.Context, gkID int64, multiplier float64, at time.Time, holderID int64, holderAcquiredAt *time.Time) error {
	var holderPtr *int64
	if holderID != 0 {
		holderPtr = &holderID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.gk_multiplier_state
			(gk_id, current_multiplier, last_updated_at, current_holder_id, holder_acquired_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (gk_id) DO UPDATE SET
			current_multiplier = $2,
			last_updated_at = $3,
			current_holder_id = $4,
			holder_acquired_at = $5
	`, gkID, multiplier, at, holderPtr, holderAcquiredAt)
	return err
}

func (s *pgStore) LogGKMultiplierChange(ctx context.Context, gkID int64, moveID int64, oldMult, newMult float64, reason, source string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.gk_points_log
			(gk_id, move_id, old_multiplier, new_multiplier, multiplier_delta, reason, module_source, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, gkID, moveID, oldMult, newMult, newMult-oldMult, reason, source)
	return err
}

// ---- Countries visited ----

func (s *pgStore) GetGKCountriesVisited(ctx context.Context, gkID int64) (map[string]bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT country_code FROM geokrety_stats.gk_countries_visited WHERE gk_id = $1
	`, gkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	countries := make(map[string]bool)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		countries[c] = true
	}
	return countries, rows.Err()
}

func (s *pgStore) AddGKCountryVisited(ctx context.Context, gkID int64, country string, at time.Time, moveID int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.gk_countries_visited (gk_id, country_code, first_visited_at, first_move_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (gk_id, country_code) DO NOTHING
	`, gkID, country, at, moveID)
	return err
}

// ---- User move history ----

func (s *pgStore) GetUserMoveHistoryOnGK(ctx context.Context, userID, gkID int64) (map[pipeline.LogType]bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT log_type FROM geokrety_stats.user_move_history WHERE user_id = $1 AND gk_id = $2
	`, userID, gkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make(map[pipeline.LogType]bool)
	for rows.Next() {
		var lt int
		if err := rows.Scan(&lt); err != nil {
			return nil, err
		}
		history[pipeline.LogType(lt)] = true
	}
	return history, rows.Err()
}

func (s *pgStore) AddUserMoveHistory(ctx context.Context, userID, gkID int64, logType pipeline.LogType, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_move_history (user_id, gk_id, log_type, first_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, gk_id, log_type) DO NOTHING
	`, userID, gkID, int(logType), at)
	return err
}

// ---- Owner GK limit ----

func (s *pgStore) GetActorGKsPerOwnerCount(ctx context.Context, actorID, ownerID int64) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM geokrety_stats.user_owner_gk_counts
		WHERE user_id = $1 AND owner_id = $2
	`, actorID, ownerID).Scan(&count)
	return count, err
}

func (s *pgStore) IsGKAlreadyCountedForOwner(ctx context.Context, actorID, ownerID, gkID int64) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM geokrety_stats.user_owner_gk_counts
			WHERE user_id = $1 AND owner_id = $2 AND gk_id = $3
		)
	`, actorID, ownerID, gkID).Scan(&exists)
	return exists, err
}

func (s *pgStore) RecordGKCountedForOwner(ctx context.Context, actorID, ownerID, gkID int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_owner_gk_counts (user_id, owner_id, gk_id, first_earned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, owner_id, gk_id) DO NOTHING
	`, actorID, ownerID, gkID, at)
	return err
}

// ---- Waypoint penalty ----

func (s *pgStore) GetActorGKsAtLocationThisMonth(ctx context.Context, actorID int64, locationID string, yearMonth string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM geokrety_stats.user_waypoint_monthly_counts
		WHERE user_id = $1 AND location_id = $2 AND year_month = $3
	`, actorID, locationID, yearMonth).Scan(&count)
	return count, err
}

func (s *pgStore) RecordActorGKAtLocation(ctx context.Context, actorID int64, locationID string, yearMonth string, gkID int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_waypoint_monthly_counts
			(user_id, location_id, year_month, gk_id, scored_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, location_id, year_month, gk_id) DO NOTHING
	`, actorID, locationID, yearMonth, gkID, at)
	return err
}

// ---- Monthly diversity ----

func (s *pgStore) GetActorMonthlyDiversity(ctx context.Context, actorID int64, yearMonth string) (
	dropsCount int, dropsBonusAwarded bool,
	ownersCount int, ownersBonusAwarded bool,
	err error,
) {
	err = s.pool.QueryRow(ctx, `
		SELECT gks_dropped_count, gks_dropped_bonus_awarded,
		       distinct_owners_count, distinct_owners_bonus_awarded
		FROM geokrety_stats.user_monthly_diversity
		WHERE user_id = $1 AND year_month = $2
	`, actorID, yearMonth).Scan(&dropsCount, &dropsBonusAwarded, &ownersCount, &ownersBonusAwarded)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, 0, false, nil
	}
	return
}

func (s *pgStore) GetActorMonthlyDiversityCountries(ctx context.Context, actorID int64, yearMonth string) (map[string]bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT country FROM geokrety_stats.user_monthly_diversity_countries
		WHERE user_id = $1 AND year_month = $2
	`, actorID, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		result[c] = true
	}
	return result, rows.Err()
}

func (s *pgStore) GetActorMonthlyDiversityOwners(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT owner_id FROM geokrety_stats.user_monthly_diversity_owners
		WHERE user_id = $1 AND year_month = $2
	`, actorID, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]bool)
	for rows.Next() {
		var ownerID int64
		if err := rows.Scan(&ownerID); err != nil {
			return nil, err
		}
		result[ownerID] = true
	}
	return result, rows.Err()
}

func (s *pgStore) GetActorMonthlyDiversityDrops(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT gk_id FROM geokrety_stats.user_monthly_diversity_drops
		WHERE user_id = $1 AND year_month = $2
	`, actorID, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]bool)
	for rows.Next() {
		var gkID int64
		if err := rows.Scan(&gkID); err != nil {
			return nil, err
		}
		result[gkID] = true
	}
	return result, rows.Err()
}

func (s *pgStore) IncrementActorMonthlyDrops(ctx context.Context, actorID int64, yearMonth string, gkID int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity_drops (user_id, year_month, gk_id, dropped_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, year_month, gk_id) DO NOTHING
	`, actorID, yearMonth, gkID, at)
	if err != nil {
		return err
	}
	// Update aggregate
	_, err = s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity (user_id, year_month, gks_dropped_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, year_month) DO UPDATE
		SET gks_dropped_count = (
			SELECT COUNT(*) FROM geokrety_stats.user_monthly_diversity_drops
			WHERE user_id = $1 AND year_month = $2
		)
	`, actorID, yearMonth)
	return err
}

func (s *pgStore) SetDropsBonusAwarded(ctx context.Context, actorID int64, yearMonth string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity (user_id, year_month, gks_dropped_bonus_awarded)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (user_id, year_month) DO UPDATE SET gks_dropped_bonus_awarded = TRUE
	`, actorID, yearMonth)
	return err
}

func (s *pgStore) IncrementActorMonthlyOwners(ctx context.Context, actorID int64, yearMonth string, ownerID int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity_owners (user_id, year_month, owner_id, counted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, year_month, owner_id) DO NOTHING
	`, actorID, yearMonth, ownerID, at)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity (user_id, year_month, distinct_owners_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, year_month) DO UPDATE
		SET distinct_owners_count = (
			SELECT COUNT(*) FROM geokrety_stats.user_monthly_diversity_owners
			WHERE user_id = $1 AND year_month = $2
		)
	`, actorID, yearMonth)
	return err
}

func (s *pgStore) SetOwnersBonusAwarded(ctx context.Context, actorID int64, yearMonth string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity (user_id, year_month, distinct_owners_bonus_awarded)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (user_id, year_month) DO UPDATE SET distinct_owners_bonus_awarded = TRUE
	`, actorID, yearMonth)
	return err
}

func (s *pgStore) RecordActorDiversityCountry(ctx context.Context, actorID int64, yearMonth, country string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.user_monthly_diversity_countries (user_id, year_month, country, awarded_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, year_month, country) DO NOTHING
	`, actorID, yearMonth, country, at)
	return err
}

// ---- Chain state ----

func (s *pgStore) GetActiveChain(ctx context.Context, gkID int64) (*ChainRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, gk_id, status, started_at, ended_at, chain_last_active, holder_acquired_at
		FROM geokrety_stats.gk_chains
		WHERE gk_id = $1 AND status = 'active'
		ORDER BY started_at DESC
		LIMIT 1
	`, gkID)

	r := &ChainRow{}
	err := row.Scan(&r.ID, &r.GKID, &r.Status, &r.StartedAt, &r.EndedAt, &r.ChainLastActive, &r.HolderAcquiredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return r, err
}

func (s *pgStore) GetChainMembers(ctx context.Context, chainID int64) ([]int64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT user_id FROM geokrety_stats.gk_chain_members
		WHERE chain_id = $1
		ORDER BY position ASC
	`, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		members = append(members, uid)
	}
	return members, rows.Err()
}

func (s *pgStore) CreateChain(ctx context.Context, gkID int64, at time.Time) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO geokrety_stats.gk_chains
			(gk_id, status, started_at, chain_last_active, holder_acquired_at)
		VALUES ($1, 'active', $2, $2, $2)
		RETURNING id
	`, gkID, at).Scan(&id)
	return id, err
}

func (s *pgStore) AddChainMember(ctx context.Context, chainID int64, userID int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.gk_chain_members (chain_id, user_id, position, joined_at)
		VALUES (
			$1, $2,
			(SELECT COALESCE(MAX(position), 0) + 1 FROM geokrety_stats.gk_chain_members WHERE chain_id = $1),
			$3
		)
		ON CONFLICT (chain_id, user_id) DO NOTHING
	`, chainID, userID, at)
	return err
}

func (s *pgStore) UpdateChainLastActive(ctx context.Context, chainID int64, at time.Time, holderAcquiredAt *time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE geokrety_stats.gk_chains
		SET chain_last_active = $2,
		    holder_acquired_at = COALESCE($3, holder_acquired_at)
		WHERE id = $1
	`, chainID, at, holderAcquiredAt)
	return err
}

func (s *pgStore) EndChain(ctx context.Context, chainID int64, at time.Time, reason string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE geokrety_stats.gk_chains
		SET status = 'ended', ended_at = $2, end_reason = $3
		WHERE id = $1
	`, chainID, at, reason)
	return err
}

func (s *pgStore) GetChainLastBonus(ctx context.Context, userID, gkID int64) (time.Time, error) {
	var t *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(completed_at) FROM geokrety_stats.gk_chain_completions
		WHERE user_id = $1 AND gk_id = $2
	`, userID, gkID).Scan(&t)
	if err != nil || t == nil {
		return time.Time{}, err
	}
	return *t, nil
}

func (s *pgStore) RecordChainCompletion(ctx context.Context, userID, gkID, chainID int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO geokrety_stats.gk_chain_completions (user_id, gk_id, chain_id, completed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, chain_id) DO NOTHING
	`, userID, gkID, chainID, at)
	return err
}

func (s *pgStore) GetExpiredChains(ctx context.Context, timeoutDays int, asOf time.Time) ([]*ChainRow, error) {
	cutoff := asOf.Add(-time.Duration(timeoutDays) * 24 * time.Hour)
	rows, err := s.pool.Query(ctx, `
		SELECT id, gk_id, status, started_at, ended_at, chain_last_active, holder_acquired_at
		FROM geokrety_stats.gk_chains
		WHERE status = 'active' AND chain_last_active <= $1
	`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []*ChainRow
	for rows.Next() {
		r := &ChainRow{}
		if err := rows.Scan(&r.ID, &r.GKID, &r.Status, &r.StartedAt, &r.EndedAt, &r.ChainLastActive, &r.HolderAcquiredAt); err != nil {
			return nil, err
		}
		chains = append(chains, r)
	}
	return chains, rows.Err()
}

// ---- Point recording ----

func (s *pgStore) SaveAwards(ctx context.Context, awards []pipeline.FinalAward) error {
	if len(awards) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, fa := range awards {
		for _, a := range fa.Awards {
			if a.Points == 0 {
				continue
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO geokrety_stats.user_points_log
					(user_id, points, reason, label, module_source, is_owner_reward,
					 move_id, gk_id, awarded_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, a.RecipientUserID, a.Points, a.Reason, a.Label, a.ModuleSource,
			a.IsOwnerReward, fa.EventLogID, fa.GKID, fa.AwardedAt)
			if err != nil {
				return fmt.Errorf("inserting award: %w", err)
			}
		}

		// Update the user's total
		if fa.TotalPoints > 0 {
			_, err := tx.Exec(ctx, `
				INSERT INTO geokrety_stats.user_points_totals (user_id, total_points, last_updated_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO UPDATE
			SET total_points = user_points_totals.total_points + $2,
			    last_updated_at = $3
		`, fa.RecipientUserID, fa.TotalPoints, fa.AwardedAt)
			if err != nil {
				return fmt.Errorf("updating user total: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// ---- Historical replay ----

// GetMoveIDsPage returns a page of gk_moves IDs in chronological order for replay.
// Uses keyset pagination on (moved_on_datetime, id) to ensure consistent ordering.
func (s *pgStore) GetMoveIDsPage(
	ctx context.Context,
	afterID, maxID int64,
	startDate, endDate *time.Time,
	limit int,
) ([]int64, error) {
	// Get the pagination cursor: datetime and id of the last move from previous page
	var afterDatetime *time.Time
	if afterID > 0 {
		err := s.pool.QueryRow(ctx,
			`SELECT moved_on_datetime FROM geokrety.gk_moves WHERE id = $1`,
			afterID,
		).Scan(&afterDatetime)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("fetching pagination cursor: %w", err)
		}
	}

	// Build dynamic WHERE clause using keyset pagination.
	// Pattern: (datetime > cursor_datetime) OR (datetime = cursor_datetime AND id > cursor_id)
	args := []interface{}{limit}
	where := "TRUE"
	argIdx := 2

	// Keyset pagination on (moved_on_datetime, id)
	if afterDatetime != nil {
		where += fmt.Sprintf(` AND (moved_on_datetime > $%d OR (moved_on_datetime = $%d AND id > $%d))`,
			argIdx, argIdx, argIdx+1)
		args = append(args, *afterDatetime, afterID)
		argIdx += 2
	}

	if maxID > 0 {
		where += fmt.Sprintf(" AND id <= $%d", argIdx)
		args = append(args, maxID)
		argIdx++
	}
	if startDate != nil {
		where += fmt.Sprintf(" AND moved_on_datetime >= $%d", argIdx)
		args = append(args, *startDate)
		argIdx++
	}
	if endDate != nil {
		where += fmt.Sprintf(" AND moved_on_datetime <= $%d", argIdx)
		args = append(args, *endDate)
		argIdx++
	}

	query := fmt.Sprintf(
		`SELECT id FROM geokrety.gk_moves WHERE %s ORDER BY moved_on_datetime ASC, id ASC LIMIT $1`,
		where,
	)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetMoveIDsPage query: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning move id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
