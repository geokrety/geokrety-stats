// Package maintenance runs periodic background jobs:
//   - expire chains that have been inactive for >= ChainTimeoutDays
//   - (future) other cleanup tasks
package maintenance

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// ChainAwarder awards timeout bonuses for expired chains.
// Implemented by engine.Engine (via ChainBonus helper).
type ChainAwarder interface {
	AwardTimeoutBonus(ctx context.Context, chainID int64, gkID int64, now time.Time) error
}

// Scheduler runs the periodic maintenance jobs using robfig/cron.
type Scheduler struct {
	store   store.Store
	awarder ChainAwarder
	cfg     config.Config
	cron    *cron.Cron
}

// New creates a Scheduler. Call Start() to activate jobs.
func New(s store.Store, awarder ChainAwarder, cfg config.Config) *Scheduler {
	return &Scheduler{
		store:   s,
		awarder: awarder,
		cfg:     cfg,
		cron:    cron.New(),
	}
}

// Start registers all maintenance jobs and starts the cron scheduler.
// It returns immediately; jobs run in background goroutines.
func (s *Scheduler) Start() {
	// Expire inactive chains every hour.
	if _, err := s.cron.AddFunc("@hourly", s.expireChains); err != nil {
		log.Error().Err(err).Msg("maintenance: failed to register chain expiry job")
	}

	// Daily cleanup and partition rotation.
	if _, err := s.cron.AddFunc("@daily", s.cleanupAndRotate); err != nil {
		log.Error().Err(err).Msg("maintenance: failed to register daily cleanup job")
	}

	s.cron.Start()
	log.Info().Msg("maintenance scheduler started")
}

// Stop gracefully shuts down the scheduler, waiting for running jobs to finish.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Info().Msg("maintenance scheduler stopped")
}

// expireChains finds all active chains that have been inactive for too long
// and awards (or declines to award) their chain bonuses.
func (s *Scheduler) expireChains() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	now := time.Now().UTC()
	if err := ExpireChainsAt(ctx, s.store, s.awarder, s.cfg.Stats.ChainTimeoutDays, now); err != nil {
		log.Error().Err(err).Msg("maintenance: failed to expire chains")
	}
}

// ExpireChainsAt expires all active chains whose inactivity is beyond timeoutDays at the given timestamp.
// This is used by both the real-time scheduler (asOf=now) and historical replay (asOf=replay time).
func ExpireChainsAt(ctx context.Context, st store.Store, awarder ChainAwarder, timeoutDays int, asOf time.Time) error {
	expired, err := st.GetExpiredChains(ctx, timeoutDays, asOf)
	if err != nil {
		return err
	}

	if len(expired) == 0 {
		return nil
	}

	log.Info().Int("count", len(expired)).Time("as_of", asOf).Msg("maintenance: expiring chains")

	for _, chain := range expired {
		if err := awarder.AwardTimeoutBonus(ctx, chain.ID, chain.GKID, asOf); err != nil {
			log.Error().
				Int64("chain_id", chain.ID).
				Int64("gk_id", chain.GKID).
				Err(err).
				Msg("maintenance: chain timeout bonus error")
		}
	}

	return nil
}

func (s *Scheduler) cleanupAndRotate() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	now := time.Now().UTC()

	if _, err := s.store.PruneEndedChainsOlderThan(ctx, now.AddDate(0, 0, -s.cfg.Maintenance.EndedChainRetentionDays)); err != nil {
		log.Error().Err(err).Msg("maintenance: pruning ended chains failed")
	}

	if _, err := s.store.PruneGKPointsLogOlderThan(ctx, now.AddDate(0, 0, -s.cfg.Maintenance.GKPointsLogRetentionDays)); err != nil {
		log.Error().Err(err).Msg("maintenance: pruning gk_points_log failed")
	}

	if err := s.store.RotateWaypointMonthlyPartitions(
		ctx,
		now,
		s.cfg.Maintenance.WaypointPartitionRetentionMonths,
		s.cfg.Maintenance.WaypointPartitionFutureMonths,
	); err != nil {
		log.Error().Err(err).Msg("maintenance: rotating waypoint month partitions failed")
	}
}
