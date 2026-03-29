/**
 * v3 API type definitions — aligned with Go backend response structures.
 *
 * JSON field names match the Go struct `json:"..."` tags exactly.
 * All responses are wrapped in the Envelope format.
 */

// ── Response Envelope ───────────────────────────────────────────────────────

export interface PaginationMeta {
  type?: string
  limit: number
  offset: number
  count: number
  returned?: number
  cursor?: string
  nextCursor?: string
  hasMore?: boolean
  totalItems?: number
  totalPages?: number
}

export interface ResponseMeta {
  requestedAt: string
  queryMs: number
  pagination?: PaginationMeta
}

export interface ApiResponse<T> {
  data: T
  meta: ResponseMeta
}

// ── Global Stats (GET /api/v3/stats/kpis) ───────────────────────────────────

export interface GeokretyTypeStats {
  traditional: number
  book: number
  human: number
  coin: number
  kretypost: number
  pebble: number
  car: number
  playingcard: number
  dogtag: number
  jigsaw: number
  easteregg: number
}

export interface MoveTypeStats {
  dropped: number
  grabbed: number
  commented: number
  seen: number
  archived: number
  dipped: number
}

export interface PictureTypeStats {
  geokretAvatars: number
  geokretMoves: number
  userAvatars: number
}

export interface GlobalStats {
  totalGeokrety: number
  totalGeokretyHidden: number
  totalMoves: number
  movesLast30Days: number
  registeredUsers: number
  activeUsers: number
  activeUsersLast30d: number
  countriesReached: number
  picturesUploaded: number
  geokretyByType: GeokretyTypeStats
  movesByType: MoveTypeStats
  picturesByType: PictureTypeStats
}

// ── Country Stats (GET /api/v3/stats/countries) ─────────────────────────────

export interface CountryStats {
  code: string
  name: string
  flag: string
  movesCount: number
  usersHome: number
  activeUsers: number
  dropped: number
  dipped: number
  seen: number
  loves: number
  pictures: number
  pointsSum: number
  pointsSumMoves: number
  geokretyInCache: number
  geokretyLost: number
  avgPointsPerMove: number
}

// ── Leaderboard User (GET /api/v3/stats/leaderboard) ────────────────────────

export interface LeaderboardUser {
  rank: number
  userId: number
  username: string
  avatarId?: number | null
  avatarUrl?: string | null
  initials: string
  points: number
  movesCount: number
  avatarColor: string
}

// ── Recent Move (GET /api/v3/geokrety/recent-moves) ────────────────────────

export interface RecentMove {
  id: number
  geokretGkid?: number | null
  geokretName: string
  geokretType?: number | null
  geokretTypeIconUrl?: string | null
  geokretAvatarUrl?: string | null
  type: string
  userId?: number | null
  userAvatarUrl?: string | null
  username: string
  country: string
  countryFlag: string
  timestamp: string
}

// ── GeoKret List Item ───────────────────────────────────────────────────────

export interface GeoJSONPoint {
  type: string
  coordinates: [number, number]
}

export interface GeokretListItem {
  id: number
  gkid?: string | null
  name: string
  avatarId?: number | null
  avatarUrl?: string | null
  type: number
  typeName: string
  typeIconUrl: string
  missing: boolean
  missingAt?: string | null
  ownerId?: number | null
  ownerUsername?: string | null
  holderId?: number | null
  holderUsername?: string | null
  country?: string | null
  waypoint?: string | null
  lat?: number | null
  lon?: number | null
  lovesCount: number
  picturesCount: number
  cachesCount: number
  bornAt?: string | null
  lastMoveAt?: string | null
  lastMoveType?: number | null
  geojson?: GeoJSONPoint | null
}

// ── User List Item (GET /api/v3/users/) ─────────────────────────────────────

export interface UserListItem {
  id: number
  username: string
  joinedAt: string
  homeCountry?: string | null
  avatarId?: number | null
  avatarUrl?: string | null
  lastMoveAt?: string | null
}

// ── Move Record (GET /api/v3/geokrety/{gkid}/moves) ────────────────────────

export interface MoveRecord {
  id: number
  geokretId: number
  moveType: number
  moveTypeName: string
  authorId?: number | null
  authorAvatarId?: number | null
  authorAvatarUrl?: string | null
  username?: string | null
  country?: string | null
  waypoint?: string | null
  lat?: number | null
  lon?: number | null
  elevation?: number | null
  kmDistance?: number | null
  movedOn: string
  createdOn: string
  picturesCount: number
  commentsCount: number
  comment?: string | null
  commentHidden: boolean
  geojson?: GeoJSONPoint | null
}

// ── Error Response ──────────────────────────────────────────────────────────

export interface ApiErrorDetail {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface ApiErrorResponse {
  error: ApiErrorDetail
  status: number
  timestamp: string
  requestId?: string
}
