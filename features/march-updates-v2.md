# Feature Spec: March Updates V2

## Overview
This feature set includes bug fixes for the Chain Detail view, UI cleanup on the home page, and a significant enhancement to the real-time synchronization (WebSocket) user experience in the navbar. It also includes a major refactor of the user profile view for better maintainability.

## User Persona
- **System Monitors**: Want to see when the dashboard is live and syncing.
- **Power Users**: Looking for quick global stats in the navbar.
- **Developers**: Benefiting from modular code architecture.

## Requirements

### 1. Bug Fixes
- Fix `TypeError: i.userAvatarUrl is not a function` in `ChainDetailView.vue`.

### 2. Home Page UI (HomeView)
- Remove the redundant live/sync icon from above the main table.
- Maintain the one in the navbar.

### 3. Navbar Enhancements
- **Live Sync Icon**:
  - Add descriptive tooltip.
  - Implement toggle functionality to Enable/Disable live WebSocket synchronization.
- **Moves Counter**:
  - Display "X moves" in a user-friendly way.
  - Expand on hover to show quick site stats (updated dynamically via WebSocket).
  - Add a visual effect (e.g., flash or pulse) when a new move is received.

### 4. WebSocket Improvements
- **Client**: Add console logging for new messages.
- **Server**: Ensure efficiency for quick stats.

### 5. Code Architecture (Refactor)
- Move all `UserView.vue` tabs into dedicated files/components to follow the composable reuse and modularity standards.

## Technical Implementation Details

### Files Touched
- `leaderboard-dashboard/src/views/ChainDetailView.vue`
- `leaderboard-dashboard/src/views/HomeView.vue`
- `leaderboard-dashboard/src/components/AppNavbar.vue` (assumed name)
- `leaderboard-dashboard/src/views/UserView.vue`
- `leaderboard-dashboard/src/components/UserOverviewTab.vue` (new)
- `leaderboard-dashboard/src/components/UserMovesTab.vue` (new)
- ... (other user tabs)

### API Endpoints
- No new endpoints expected; utilizing existing WebSocket stream.

## Testing Procedures
1. **Manual Check**: Navigate to `/chains/[id]` and verify avatars load without error.
2. **UI Check**: Verify navbar tooltip and status toggle.
3. **WebSocket Check**: Open browser console, verify logs appear when moves arrive.
4. **Visual Check**: Hover over navbar moves counter, verify stats popup.
5. **Regression Check**: Verify User profile tabs still function identically after refactor.

## Known Limitations
- Hover-expand stats must be lightweight.
- Pulse effect should be subtle to avoid distraction.
