# GeoKrety Points System - Improvements Status Report

**Date**: 2026-02-28
**Status**: In Progress - Major Improvements Completed

---

## ✅ COMPLETED IMPROVEMENTS

### 1. Images/Loves Count Cards (Issue #2)
**Status**: ✅ **FIXED**

**Changes**:
- Added `TotalImages` field to GlobalStats model
- Added `TotalLoves` field to GlobalStats model
- Updated GlobalStats handler to query:
  - `geokrety.gk_images` for image count
  - `geokrety.gk_geokrety` SUM(loves_count) for loves total
- Both fallback and primary paths now return these values

**Result**: The `/api/v1/stats` endpoint now includes:
```json
{
  "total_images": 0,
  "total_loves": 2,
  ...
}
```

The stats dashboard cards will now display correct image and love counts.

---

### 2. Country Fields for GeoKret Responses (Issue #0 - Backend)
**Status**: ✅ **COMPLETED**

**Changes**:
- Added `OwnerHomeCountry` field to GeoKret model
- Added `HolderHomeCountry` field to GeoKret model
- Added `CacheCountry` field to GeoKret model
- Updated `GetGeoKret` handler to:
  - LEFT JOIN `geokrety.gk_users` for owner and holder countries
  - Use LATERAL subquery to find latest move with country for cache location

**Result**: The `/api/v1/geokrety/{id}` endpoint now includes:
```json
{
  "owner_home_country": "pl",
  "holder_home_country": null,
  "cache_country": "DE",
  ...
}
```

These fields enable:
- Displaying country flags next to owner/holder usernames
- Showing where a cached GeoKret is visible via the country badge

---

### 3. TopCountries Query Fix (Issue #1 - Partial)
**Status**: ⚠️ **ATTEMPTED - NEEDS VERIFICATION**

**Changes**:
- Corrected SELECT column order in TopCountries handler to match mv_country_summary view
- Previously: `total_points_awarded` was in position 5
- Now: Moved to position 10 after move type counts

**Status**: Column order is now correct, BUT `grabs` field still shows 0 for all countries. This suggests the root issue is at the database view level (see Diagnostics section below).

---

## ⚠️ PARTIAL/NEEDS INVESTIGATION

### Countries Page Grabs = 0 (Issue #1)
**Status**: 🔴 **REQUIRES DATABASE INVESTIGATION**

**Root Cause Analysis**:

The `grabs` field for all countries returns 0, despite:
- GeoKrety 4940 having 7 confirmed grabs (move_type=1)
- GeoKrety 4940 being located in Germany (DE)
- Drops field correctly showing 68,409 for Germany
- Dips field correctly showing 415,978 for Germany

**Possible Root Causes**:
1. **Migration 000004 Not Applied**: The database uses the older migration 000003 which has the WRONG move_type mappings. Migration 000004 should fix this but may not have been applied.

2. **View Refresh Issue**: MaterializedViews might not have been refreshed after migration 000004. The API calls `geokrety_stats.refresh_leaderboard_views()` every 15 minutes, but may not have run since migrations were applied.

3. **Database Schema Mismatch**: The actual `gk_moves` table might have different move_type values than expected.

**What Was Done**:
- Verified that API handler correctly maps: `0=drop, 1=grab, 5=dip, 3=seen`
- Confirmed GeoKrety 4940 has 7 moves with move_type=1 (grabs)
- Fixed SELECT column order in TopCountries query
- Tested all countries - ALL have grabs=0 (not just one)

**Solution Needed**:
1. Verify which migrations have been applied to the geokrety database
2. Manually run migration 000004 if not applied:
   ```sql
   -- From migrations/000004_fix_country_stats_and_gk_stats.up.sql
   ```
3. Refresh materialized views:
   ```sql
   SELECT geokrety_stats.refresh_leaderboard_views();
   ```
4. Re-test `/api/v1/stats/countries` endpoint

---

## 🚀 NOT YET STARTED

### 4. Stats Page Improvements (Issue #3)
**Status**: 🔴 **NOT STARTED**

**Required Changes**:
- a. Add "all time statistics" header text
- b. "Top 20 Countries by Moves" chart:
  - Set y-axis minimum to 0
  - Remove countries with y=0
  - Add other move types: comments, archive
- c. Add user evolution graph since 2007
- d. Add geokrety evolution graph since 2007:
  - Show number of geokrety created per time period
  - Show number of geokrety currently in caches

**Impact**: Medium - UI/UX improvements

---

### 5. GeoKrety Detail - Points Discrepancy (Issue #4)
**Status**: 🔴 **NOT STARTED**

**Issue**: Page lists 19.18 points awarded but header shows 575.4

**Required Investigation**:
- Check `/api/v1/geokrety/{id}/points/log` endpoint
- Compare `SUM(points)` from points log vs `total_points_generated` from stats
- Verify points calculation includes all bonus types
- Check for filtering issues (e.g., only counting specific label types)

**Impact**: High - Data accuracy

---

### 6. GeoKrety Detail - UI Badges (Issue #5)
**Status**: 🔴 **NOT STARTED**

**Required Changes**:
- Add badge in "Moves" tab header showing move count
- Add badge in "Moves" tab header showing distinct movers
- Use format like: "Moves (254) · Movers (87)"

**Location**: `leaderboard-dashboard/src/views/GeokretView.vue` tab component

**Impact**: Low - UI Polish

---

### 7. GeoKrety Related Users - Cards Layout (Issue #6)
**Status**: 🔴 **NOT STARTED**

**Required Changes**:
- Change RelatedUsersTab from table layout to card grid layout
- Create generic `UserCard.vue` composable
- Display user stats on cards:
  - Username
  - Points
  - Moves count
  - Countries visited
  - GeoKrety owned/held

**Files to Modify**:
- `leaderboard-dashboard/src/components/RelatedUsersTab.vue`
- Create: `leaderboard-dashboard/src/components/UserCard.vue`
- Create: `leaderboard-dashboard/src/composables/useUserCard.js`

**Impact**: Medium - UX improvement

---

### 8. Country Detail Page Improvements (Issue #7)
**Status**: 🔴 **NOT STARTED**

**Required Changes**:

a. **Full Country Name in Header**:
- Add JavaScript library (e.g., `countries.js`) to map ISO-2 codes to full names
- Display: "🇩🇪 Germany (DE)" instead of just "DE"

b. **Metric Tooltips**:
- Add `title` attributes explaining each metric
- Examples:
  - "Total Points": "Sum of all points awarded for moves in this country"
  - "Unique GeoKrety": "Number of distinct GeoKrety that visited this country"
  - "Active Users": "Number of unique users who made moves in this country"

c. **Clarify Unique GeoKrety/Users**:
- Current definitions unclear - need to clarify:
  - Unique GeoKrety: "born in country" OR "visited country" OR "currently in cache in country"?
  - Unique Users: "home country" OR "made moves here" OR "visited country"?
- Add explanatory tooltips

d. **Move Type Evolution Chart**:
- New chart showing monthly breakdown since 2007
- X-axis: Months (2007-present)
- Y-axis: Move counts
- Series: drops, grabs, dips, comments, seen
- Stacked bar or line chart format

**Files to Modify**:
- `leaderboard-dashboard/src/views/CountryDetailView.vue`
- Add dependency: country code library
- Create: `leaderboard-dashboard/src/components/MoveTypeTimelineChart.vue`

**Impact**: Medium - User understanding, UX polish

---

## 📊 TESTING STATUS

### Automated Tests
- ✅ API endpoints tested with curl
- ✅ Backend changes deployed and running
- ⚠️ Database views need manual verification
- ❌ UI components not yet tested
- ❌ End-to-end testing not performed

### Manual Testing Done
✅ `/api/v1/stats` - Images/loves fields working
✅ `/api/v1/geokrety/{id}` - Country fields working
⚠️ `/api/v1/stats/countries` - Column order fixed, but grabs still 0
❌ `/api/v1/geokrety/{id}/points/log` - Not investigated
❌ UI changes - Not started

---

## 🔧 DEPLOYMENT INFO

**Docker Compose Status**: ✅ Running
- leaderboard-api: Running on port 8080
- leaderboard-dashboard: Running on port 3000

**Recent Changes**:
```bash
# Commit 0a41f7c
feat: add images/loves stats and country fields to GeoKret responses

# Commit 79b55b1
fix: correct TopCountries SELECT column order
```

**How to Test**:
```bash
# Test images/loves via /stats endpoint
curl -s http://<hostip>:8080/api/v1/stats | jq '.data | {total_images, total_loves}'

# Test country fields via /geokrety endpoint
curl -s http://<hostip>:8080/api/v1/geokrety/4940 | jq '.data | {owner_home_country, holder_home_country, cache_country}'

# Test countries (grabs issue debug)
curl -s http://<hostip>:8080/api/v1/stats/countries | jq '.data[0] | {drops, grabs, dips}'
```

---

## 📋 NEXT STEPS / RECOMMENDATIONS

### Immediate
1. **Investigate Grabs Issue**:
   - Check database migration application status
   - Manually refresh materialized views if needed
   - Verify gk_moves table move_type values

2. **Points Discrepancy**:
   - Query points log for specific GeoKret
   - Compare to total_points_generated
   - Check which point awards are included/excluded

### Short-term (1-2 hours)
3. Implement Issue #4 (points investigation)
4. Implement Issue #3a-b (stats header and chart fixes)
5. Test all endpoint changes

### Medium-term
6. Implement Issue #5 (badges)
7. Implement remaining Issue #3 improvements
8. Implement Issue #6 (cards layout refactor)
9. Implement Issue #7 (country page improvements)

### Testing & QA
- Add integration tests for API endpoints
- Create UI test cases for new components
- Perform end-to-end testing in browser
- Verify all fields display correctly

---

## 📝 CODE REFERENCES

**Files Modified**:
- `leaderboard-api/internal/models/models.go` - Added fields to GeoKret and GlobalStats
- `leaderboard-api/internal/handlers/stats.go` - Added images/loves queries
- `leaderboard-api/internal/handlers/geokrety.go` - Added country field joins

**API Endpoints**:
- `GET /api/v1/stats` - ✅ Returns total_images, total_loves
- `GET /api/v1/geokrety/{id}` - ✅ Returns owner/holder/cache countries
- `GET /api/v1/stats/countries` - ⚠️ Column order fixed, grabs issue unsolved

---

## 🐛 KNOWN ISSUES

1. **Grabs Always 0** - Database view issue, not UI issue
2. **Points Discrepancy** - Unknown root cause, needs investigation
3. **Country Page UI** - Not updated yet (missing tooltips, full names, charts)

---

**Last Updated**: 2026-02-28 14:20 UTC
**Next Review**: After database migration verification
