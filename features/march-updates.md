# Feature Updates - March 2026

## Overview
A collection of UI improvements, bug fixes, and feature enhancements across the Geokrety Points System.

## Items

### 1. Geokrety Detail Page Cleanup
- **Changes**:
    - Remove placeholder counts on "Moves" tab (use header count).
    - Fix "Reach" average point calculation (`points / moves_count`).
- **Files**:
    - `leaderboard-dashboard/src/views/GkDetail.vue` (or similar)

### 2. Dropdown Filter Fixes
- **Changes**:
    - Fix "AllNone" text spacing (change to "All / None").
    - Fix dropdown closing behavior when clicking outside after interaction.
- **Files**:
    - Components/composables handling filter dropdowns.

### 3. Countries Page Enhancements
- **Changes**:
    - Use Bootstrap tooltips for table headers.
    - Implement URL hash navigation for table/cards view.
    - Improve card visual layout (center elements).
    - General visual polish.
- **Files**:
    - `leaderboard-dashboard/src/views/Countries.vue` (or equivalent)

### 4. Chain Enhancements
- **Geokrety List**: Add filter by chain.
- **Chain Detail Page**:
    - Visual improvement (header inspiration from user/gk headers).
    - Description text for "Members" and "Moves" panels.
    - Reuse `Points` and `MoveType` components.
- **User Profile**: Add filters to "Chains" tab.
- **Files**:
    - `leaderboard-dashboard/src/views/GkList.vue`
    - `leaderboard-dashboard/src/views/ChainDetail.vue`
    - `leaderboard-dashboard/src/views/UserDetail.vue`

### 5. Table Header Sorting
- **Changes**: Highlight/color active sorting header.
- **Files**:
    - `leaderboard-dashboard/src/assets/main.css` (or generic table component)

## Testing Plan
- **API**: Check that response data correctly populates new filters.
- **UI**: Use Playwright to verify:
    - Dropdown behavior (closing).
    - Tooltips presence.
    - URL hash navigation.
    - Visual consistency.
- **Docker**: Build and deploy to verify all services interact correctly.
