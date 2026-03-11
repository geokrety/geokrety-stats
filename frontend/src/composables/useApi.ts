// API composable — uses fetch to query the configured backend.
// Base URL can be overridden with Vite env `VITE_API_URL`.

const API_BASE = (import.meta as any).env?.VITE_API_URL || 'http://192.168.130.65:3001'

async function getJson<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { credentials: 'same-origin' })
  if (!res.ok) throw new Error(`API ${res.status}: ${res.statusText}`)
  return (await res.json()) as T
}

export interface GlobalStats {
  totalGeokrety: number
  totalGeokretyHidden: number
  geokretyByType: {
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
  totalMoves: number
  movesLast30Days: number
  movesByType: {
    dropped: number
    grabbed: number
    commented: number
    seen: number
    archived: number
    dipped: number
  }
  registeredUsers: number
  activeUsers: number
  activeUsersLast30d: number
  countriesReached: number
  picturesUploaded: number
  picturesByType: {
    geokretAvatars: number
    geokretMoves: number
    userAvatars: number
  }
}

export interface RecentMove {
  id: number
  geokretName: string
  type: 'grabbed' | 'dropped' | 'dipped' | 'seen' | 'commented' | 'archived'
  username: string
  country: string
  countryFlag: string
  timestamp: string
}

export interface LeaderboardUser {
  rank: number
  username: string
  initials: string
  points: number
  movesCount: number
  avatarColor: string
}

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

export function useStats() {
  const fetchStats = async (): Promise<GlobalStats> => {
    return getJson<GlobalStats>('/api/v1/stats/global')
  }
  return { fetchStats }
}

export function useRecentActivity() {
  const fetchRecentActivity = async (): Promise<RecentMove[]> => {
    return getJson<RecentMove[]>('/api/v1/stats/recent')
  }
  return { fetchRecentActivity }
}

export function useLeaderboard() {
  const fetchLeaderboard = async (): Promise<LeaderboardUser[]> => {
    return getJson<LeaderboardUser[]>('/api/v1/stats/leaderboard')
  }
  return { fetchLeaderboard }
}

// ─── Country Statistics ────────────────────────────────────────────────────

export interface CountryStats {
  /** ISO 3166-1 alpha-2 country code, e.g. "DE" */
  code: string
  /** Full English country name */
  name: string
  /** Emoji flag */
  flag: string
  /**
   * Total moves with location in this country:
   * logtype 0 (drop) + logtype 5 (dip/visiting) + logtype 3 (seen WITH location).
   * Note: seen moves may have no location — only counted when location is present.
   */
  movesCount: number
  /** Number of users whose home coordinates are in this country */
  usersHome: number
  /** Distinct users who have at least one qualifying move in the country */
  activeUsers: number
  /** Break-down of move types (subset that have location) */
  movesByType: {
    /** logtype 0 — GeoKret dropped in a cache */
    dropped: number
    /** logtype 5 — visiting/dipped (always carries location) */
    dipped: number
    /** logtype 3 — seen/met WITH confirmed location */
    seen: number
  }
  /** Total "love" (like/heart) reactions on GeoKrety actually spotted in this country */
  loves: number
  /** Pictures attached to moves that happened inside the country */
  pictures: number
  /** Sum of all points earned by users whose home is in this country */
  pointsSum: number
  /** Sum of all points earned by moves in this country */
  pointsSumMoves: number
  /** GeoKrety currently sitting in a cache whose last position is in this country (last logtype = drop or seen WITH location) */
  geokretyInCache: number
  /** GeoKrety marked as missing whose last known position was in this country */
  geokretyLost: number
  /** Average points per move (pointsSum / movesCount, rounded to 2 decimals) */
  avgPointsPerMove: number
}

const COUNTRY_DATA: CountryStats[] = [
  // ── Europe — top tier ────────────────────────────
  {
    code: 'PL',
    name: 'Poland',
    flag: '🇵🇱',
    movesCount: 312_000,
    usersHome: 14_200,
    activeUsers: 22_000,
    movesByType: { dropped: 74_000, dipped: 196_000, seen: 42_000 },
    loves: 58_000,
    pictures: 49_000,
    pointsSum: 5_900_000,
    pointsSumMoves: 7_100_000,
    geokretyInCache: 19_400,
    geokretyLost: 1_250,
    avgPointsPerMove: 5.12,
  },
  {
    code: 'DE',
    name: 'Germany',
    flag: '🇩🇪',
    movesCount: 210_000,
    usersHome: 9_100,
    activeUsers: 13_500,
    movesByType: { dropped: 38_000, dipped: 148_000, seen: 24_000 },
    loves: 39_000,
    pictures: 31_000,
    pointsSum: 3_800_000,
    pointsSumMoves: 5_250_000,
    geokretyInCache: 14_200,
    geokretyLost: 920,
    avgPointsPerMove: 4.42,
  },
  {
    code: 'CZ',
    name: 'Czech Republic',
    flag: '🇨🇿',
    movesCount: 178_000,
    usersHome: 7_800,
    activeUsers: 11_400,
    movesByType: { dropped: 42_000, dipped: 112_000, seen: 24_000 },
    loves: 33_500,
    pictures: 27_000,
    pointsSum: 3_300_000,
    pointsSumMoves: 3_850_000,
    geokretyInCache: 11_600,
    geokretyLost: 760,
    avgPointsPerMove: 4.31,
  },
  {
    code: 'RU',
    name: 'Russia',
    flag: '🇷🇺',
    movesCount: 132_000,
    usersHome: 6_100,
    activeUsers: 9_200,
    movesByType: { dropped: 29_000, dipped: 86_000, seen: 17_000 },
    loves: 25_500,
    pictures: 21_000,
    pointsSum: 2_450_000,
    pointsSumMoves: 2_900_000,
    geokretyInCache: 8_600,
    geokretyLost: 560,
    avgPointsPerMove: 4.05,
  },
  {
    code: 'IT',
    name: 'Italy',
    flag: '🇮🇹',
    movesCount: 118_000,
    usersHome: 5_400,
    activeUsers: 8_200,
    movesByType: { dropped: 28_000, dipped: 74_000, seen: 16_000 },
    loves: 22_500,
    pictures: 19_000,
    pointsSum: 2_120_000,
    pointsSumMoves: 2_680_000,
    geokretyInCache: 7_600,
    geokretyLost: 490,
    avgPointsPerMove: 4.02,
  },
  {
    code: 'HU',
    name: 'Hungary',
    flag: '🇭🇺',
    movesCount: 96_000,
    usersHome: 4_500,
    activeUsers: 6_700,
    movesByType: { dropped: 23_000, dipped: 60_000, seen: 13_000 },
    loves: 18_200,
    pictures: 15_000,
    pointsSum: 1_750_000,
    pointsSumMoves: 2_100_000,
    geokretyInCache: 6_200,
    geokretyLost: 400,
    avgPointsPerMove: 3.93,
  },
  {
    code: 'FR',
    name: 'France',
    flag: '🇫🇷',
    movesCount: 140_000,
    usersHome: 6_500,
    activeUsers: 10_000,
    movesByType: { dropped: 33_000, dipped: 88_000, seen: 19_000 },
    loves: 23_500,
    pictures: 19_500,
    pointsSum: 2_300_000,
    pointsSumMoves: 3_100_000,
    geokretyInCache: 9_000,
    geokretyLost: 570,
    avgPointsPerMove: 3.83,
  },
  {
    code: 'ES',
    name: 'Spain',
    flag: '🇪🇸',
    movesCount: 104_000,
    usersHome: 4_700,
    activeUsers: 7_200,
    movesByType: { dropped: 25_000, dipped: 65_000, seen: 14_000 },
    loves: 19_500,
    pictures: 16_200,
    pointsSum: 1_870_000,
    pointsSumMoves: 2_350_000,
    geokretyInCache: 6_700,
    geokretyLost: 440,
    avgPointsPerMove: 4.0,
  },
  {
    code: 'SK',
    name: 'Slovakia',
    flag: '🇸🇰',
    movesCount: 82_500,
    usersHome: 3_700,
    activeUsers: 5_600,
    movesByType: { dropped: 20_000, dipped: 52_000, seen: 10_500 },
    loves: 15_600,
    pictures: 12_800,
    pointsSum: 1_460_000,
    pointsSumMoves: 1_720_000,
    geokretyInCache: 5_300,
    geokretyLost: 335,
    avgPointsPerMove: 3.89,
  },
  {
    code: 'GB',
    name: 'United Kingdom',
    flag: '🇬🇧',
    movesCount: 89_000,
    usersHome: 4_100,
    activeUsers: 6_400,
    movesByType: { dropped: 21_000, dipped: 55_000, seen: 13_000 },
    loves: 16_800,
    pictures: 14_200,
    pointsSum: 1_620_000,
    pointsSumMoves: 2_050_000,
    geokretyInCache: 5_700,
    geokretyLost: 370,
    avgPointsPerMove: 3.97,
  },
  {
    code: 'AT',
    name: 'Austria',
    flag: '🇦🇹',
    movesCount: 74_000,
    usersHome: 3_300,
    activeUsers: 5_300,
    movesByType: { dropped: 17_500, dipped: 47_000, seen: 9_500 },
    loves: 13_800,
    pictures: 11_200,
    pointsSum: 1_310_000,
    pointsSumMoves: 1_590_000,
    geokretyInCache: 4_700,
    geokretyLost: 305,
    avgPointsPerMove: 3.91,
  },
  {
    code: 'NL',
    name: 'Netherlands',
    flag: '🇳🇱',
    movesCount: 86_000,
    usersHome: 3_950,
    activeUsers: 6_100,
    movesByType: { dropped: 21_000, dipped: 54_000, seen: 11_000 },
    loves: 16_200,
    pictures: 13_100,
    pointsSum: 1_560_000,
    pointsSumMoves: 1_980_000,
    geokretyInCache: 5_500,
    geokretyLost: 355,
    avgPointsPerMove: 3.95,
  },
  {
    code: 'BE',
    name: 'Belgium',
    flag: '🇧🇪',
    movesCount: 65_000,
    usersHome: 3_000,
    activeUsers: 4_800,
    movesByType: { dropped: 15_500, dipped: 41_000, seen: 8_500 },
    loves: 11_700,
    pictures: 9_600,
    pointsSum: 1_140_000,
    pointsSumMoves: 1_420_000,
    geokretyInCache: 4_200,
    geokretyLost: 275,
    avgPointsPerMove: 3.79,
  },
  {
    code: 'SE',
    name: 'Sweden',
    flag: '🇸🇪',
    movesCount: 63_000,
    usersHome: 2_900,
    activeUsers: 4_600,
    movesByType: { dropped: 15_000, dipped: 40_000, seen: 8_000 },
    loves: 11_900,
    pictures: 9_800,
    pointsSum: 1_080_000,
    pointsSumMoves: 1_320_000,
    geokretyInCache: 4_000,
    geokretyLost: 262,
    avgPointsPerMove: 3.75,
  },
  {
    code: 'CH',
    name: 'Switzerland',
    flag: '🇨🇭',
    movesCount: 54_000,
    usersHome: 2_500,
    activeUsers: 4_100,
    movesByType: { dropped: 13_000, dipped: 34_000, seen: 7_000 },
    loves: 10_200,
    pictures: 8_400,
    pointsSum: 940_000,
    pointsSumMoves: 1_160_000,
    geokretyInCache: 3_450,
    geokretyLost: 220,
    avgPointsPerMove: 3.75,
  },
  {
    code: 'NO',
    name: 'Norway',
    flag: '🇳🇴',
    movesCount: 42_000,
    usersHome: 1_950,
    activeUsers: 3_200,
    movesByType: { dropped: 10_000, dipped: 26_500, seen: 5_500 },
    loves: 7_900,
    pictures: 6_500,
    pointsSum: 715_000,
    pointsSumMoves: 860_000,
    geokretyInCache: 2_700,
    geokretyLost: 176,
    avgPointsPerMove: 3.68,
  },
  {
    code: 'FI',
    name: 'Finland',
    flag: '🇫🇮',
    movesCount: 48_000,
    usersHome: 2_200,
    activeUsers: 3_500,
    movesByType: { dropped: 11_400, dipped: 30_200, seen: 6_400 },
    loves: 8_900,
    pictures: 7_400,
    pointsSum: 810_000,
    pointsSumMoves: 980_000,
    geokretyInCache: 3_050,
    geokretyLost: 198,
    avgPointsPerMove: 3.71,
  },
  {
    code: 'DK',
    name: 'Denmark',
    flag: '🇩🇰',
    movesCount: 38_500,
    usersHome: 1_780,
    activeUsers: 2_900,
    movesByType: { dropped: 9_100, dipped: 24_200, seen: 5_200 },
    loves: 7_200,
    pictures: 5_900,
    pointsSum: 650_000,
    pointsSumMoves: 780_000,
    geokretyInCache: 2_450,
    geokretyLost: 161,
    avgPointsPerMove: 3.71,
  },
  {
    code: 'PT',
    name: 'Portugal',
    flag: '🇵🇹',
    movesCount: 29_000,
    usersHome: 1_300,
    activeUsers: 2_200,
    movesByType: { dropped: 6_800, dipped: 18_200, seen: 4_000 },
    loves: 5_400,
    pictures: 4_500,
    pointsSum: 490_000,
    pointsSumMoves: 620_000,
    geokretyInCache: 1_850,
    geokretyLost: 120,
    avgPointsPerMove: 3.62,
  },
  {
    code: 'RO',
    name: 'Romania',
    flag: '🇷🇴',
    movesCount: 34_000,
    usersHome: 1_550,
    activeUsers: 2_600,
    movesByType: { dropped: 8_100, dipped: 21_400, seen: 4_500 },
    loves: 6_400,
    pictures: 5_300,
    pointsSum: 570_000,
    pointsSumMoves: 720_000,
    geokretyInCache: 2_180,
    geokretyLost: 143,
    avgPointsPerMove: 3.66,
  },
  {
    code: 'HR',
    name: 'Croatia',
    flag: '🇭🇷',
    movesCount: 22_500,
    usersHome: 1_020,
    activeUsers: 1_750,
    movesByType: { dropped: 5_300, dipped: 14_200, seen: 3_000 },
    loves: 4_200,
    pictures: 3_500,
    pointsSum: 375_000,
    pointsSumMoves: 520_000,
    geokretyInCache: 1_440,
    geokretyLost: 93,
    avgPointsPerMove: 3.58,
  },
  // ── Americas ─────────────────────────────────
  {
    code: 'US',
    name: 'United States',
    flag: '🇺🇸',
    movesCount: 24_500,
    usersHome: 1_280,
    activeUsers: 2_700,
    movesByType: { dropped: 5_800, dipped: 15_400, seen: 3_300 },
    loves: 4_600,
    pictures: 3_800,
    pointsSum: 415_000,
    pointsSumMoves: 580_000,
    geokretyInCache: 1_560,
    geokretyLost: 98,
    avgPointsPerMove: 3.25,
  },
  {
    code: 'CA',
    name: 'Canada',
    flag: '🇨🇦',
    movesCount: 13_800,
    usersHome: 720,
    activeUsers: 1_500,
    movesByType: { dropped: 3_300, dipped: 8_700, seen: 1_800 },
    loves: 2_600,
    pictures: 2_150,
    pointsSum: 235_000,
    pointsSumMoves: 320_000,
    geokretyInCache: 880,
    geokretyLost: 58,
    avgPointsPerMove: 3.24,
  },
  {
    code: 'BR',
    name: 'Brazil',
    flag: '🇧🇷',
    movesCount: 8_900,
    usersHome: 480,
    activeUsers: 950,
    movesByType: { dropped: 2_100, dipped: 5_600, seen: 1_200 },
    loves: 1_650,
    pictures: 1_380,
    pointsSum: 155_000,
    pointsSumMoves: 215_000,
    geokretyInCache: 566,
    geokretyLost: 38,
    avgPointsPerMove: 3.22,
  },
  {
    code: 'AR',
    name: 'Argentina',
    flag: '🇦🇷',
    movesCount: 5_500,
    usersHome: 295,
    activeUsers: 610,
    movesByType: { dropped: 1_300, dipped: 3_460, seen: 740 },
    loves: 1_010,
    pictures: 845,
    pointsSum: 97_000,
    pointsSumMoves: 138_000,
    geokretyInCache: 348,
    geokretyLost: 24,
    avgPointsPerMove: 3.29,
  },
  {
    code: 'MX',
    name: 'Mexico',
    flag: '🇲🇽',
    movesCount: 4_200,
    usersHome: 225,
    activeUsers: 460,
    movesByType: { dropped: 1_000, dipped: 2_640, seen: 560 },
    loves: 770,
    pictures: 645,
    pointsSum: 72_000,
    pointsSumMoves: 103_000,
    geokretyInCache: 265,
    geokretyLost: 18,
    avgPointsPerMove: 3.24,
  },
  {
    code: 'CO',
    name: 'Colombia',
    flag: '🇨🇴',
    movesCount: 2_780,
    usersHome: 148,
    activeUsers: 308,
    movesByType: { dropped: 660, dipped: 1_750, seen: 370 },
    loves: 510,
    pictures: 428,
    pointsSum: 47_500,
    pointsSumMoves: 68_000,
    geokretyInCache: 175,
    geokretyLost: 12,
    avgPointsPerMove: 3.21,
  },
  {
    code: 'CL',
    name: 'Chile',
    flag: '🇨🇱',
    movesCount: 2_260,
    usersHome: 122,
    activeUsers: 250,
    movesByType: { dropped: 538, dipped: 1_420, seen: 302 },
    loves: 414,
    pictures: 347,
    pointsSum: 38_000,
    pointsSumMoves: 53_000,
    geokretyInCache: 143,
    geokretyLost: 10,
    avgPointsPerMove: 3.13,
  },
  {
    code: 'PE',
    name: 'Peru',
    flag: '🇵🇪',
    movesCount: 1_640,
    usersHome: 87,
    activeUsers: 182,
    movesByType: { dropped: 390, dipped: 1_030, seen: 220 },
    loves: 300,
    pictures: 251,
    pointsSum: 27_500,
    pointsSumMoves: 39_000,
    geokretyInCache: 104,
    geokretyLost: 7,
    avgPointsPerMove: 3.11,
  },
  // ── Asia ─────────────────────────────────────
  {
    code: 'JP',
    name: 'Japan',
    flag: '🇯🇵',
    movesCount: 11_400,
    usersHome: 605,
    activeUsers: 1_210,
    movesByType: { dropped: 2_710, dipped: 7_170, seen: 1_520 },
    loves: 2_090,
    pictures: 1_750,
    pointsSum: 193_000,
    pointsSumMoves: 276_000,
    geokretyInCache: 726,
    geokretyLost: 48,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'CN',
    name: 'China',
    flag: '🇨🇳',
    movesCount: 6_400,
    usersHome: 340,
    activeUsers: 690,
    movesByType: { dropped: 1_520, dipped: 4_020, seen: 860 },
    loves: 1_175,
    pictures: 985,
    pointsSum: 108_000,
    pointsSumMoves: 154_000,
    geokretyInCache: 406,
    geokretyLost: 27,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'KR',
    name: 'South Korea',
    flag: '🇰🇷',
    movesCount: 5_200,
    usersHome: 275,
    activeUsers: 560,
    movesByType: { dropped: 1_235, dipped: 3_270, seen: 695 },
    loves: 952,
    pictures: 798,
    pointsSum: 88_000,
    pointsSumMoves: 125_000,
    geokretyInCache: 330,
    geokretyLost: 22,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'IN',
    name: 'India',
    flag: '🇮🇳',
    movesCount: 7_600,
    usersHome: 405,
    activeUsers: 810,
    movesByType: { dropped: 1_810, dipped: 4_780, seen: 1_010 },
    loves: 1_395,
    pictures: 1_168,
    pointsSum: 127_000,
    pointsSumMoves: 181_000,
    geokretyInCache: 482,
    geokretyLost: 32,
    avgPointsPerMove: 3.16,
  },
  {
    code: 'ID',
    name: 'Indonesia',
    flag: '🇮🇩',
    movesCount: 3_280,
    usersHome: 174,
    activeUsers: 350,
    movesByType: { dropped: 780, dipped: 2_060, seen: 440 },
    loves: 602,
    pictures: 504,
    pointsSum: 55_000,
    pointsSumMoves: 78_000,
    geokretyInCache: 208,
    geokretyLost: 14,
    avgPointsPerMove: 3.15,
  },
  {
    code: 'TH',
    name: 'Thailand',
    flag: '🇹🇭',
    movesCount: 3_890,
    usersHome: 207,
    activeUsers: 415,
    movesByType: { dropped: 925, dipped: 2_445, seen: 520 },
    loves: 714,
    pictures: 598,
    pointsSum: 66_000,
    pointsSumMoves: 94_000,
    geokretyInCache: 248,
    geokretyLost: 16,
    avgPointsPerMove: 3.23,
  },
  {
    code: 'TR',
    name: 'Turkey',
    flag: '🇹🇷',
    movesCount: 6_100,
    usersHome: 306,
    activeUsers: 622,
    movesByType: { dropped: 1_450, dipped: 3_840, seen: 810 },
    loves: 1_119,
    pictures: 938,
    pointsSum: 107_000,
    pointsSumMoves: 152_000,
    geokretyInCache: 388,
    geokretyLost: 25,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'IL',
    name: 'Israel',
    flag: '🇮🇱',
    movesCount: 5_150,
    usersHome: 262,
    activeUsers: 535,
    movesByType: { dropped: 1_225, dipped: 3_240, seen: 685 },
    loves: 945,
    pictures: 792,
    pointsSum: 92_000,
    pointsSumMoves: 131_000,
    geokretyInCache: 326,
    geokretyLost: 22,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'MY',
    name: 'Malaysia',
    flag: '🇲🇾',
    movesCount: 2_400,
    usersHome: 128,
    activeUsers: 260,
    movesByType: { dropped: 570, dipped: 1_510, seen: 320 },
    loves: 440,
    pictures: 369,
    pointsSum: 40_000,
    pointsSumMoves: 57_000,
    geokretyInCache: 153,
    geokretyLost: 11,
    avgPointsPerMove: 3.17,
  },
  {
    code: 'TW',
    name: 'Taiwan',
    flag: '🇹🇼',
    movesCount: 1_950,
    usersHome: 103,
    activeUsers: 211,
    movesByType: { dropped: 463, dipped: 1_225, seen: 262 },
    loves: 357,
    pictures: 299,
    pointsSum: 32_500,
    pointsSumMoves: 46_000,
    geokretyInCache: 124,
    geokretyLost: 9,
    avgPointsPerMove: 3.16,
  },
  {
    code: 'PH',
    name: 'Philippines',
    flag: '🇵🇭',
    movesCount: 2_060,
    usersHome: 109,
    activeUsers: 222,
    movesByType: { dropped: 490, dipped: 1_295, seen: 275 },
    loves: 378,
    pictures: 317,
    pointsSum: 34_600,
    pointsSumMoves: 49_000,
    geokretyInCache: 131,
    geokretyLost: 9,
    avgPointsPerMove: 3.16,
  },
  // ── Africa ────────────────────────────────────
  {
    code: 'ZA',
    name: 'South Africa',
    flag: '🇿🇦',
    movesCount: 3_700,
    usersHome: 197,
    activeUsers: 393,
    movesByType: { dropped: 879, dipped: 2_326, seen: 495 },
    loves: 679,
    pictures: 569,
    pointsSum: 61_000,
    pointsSumMoves: 87_000,
    geokretyInCache: 235,
    geokretyLost: 16,
    avgPointsPerMove: 3.14,
  },
  {
    code: 'MA',
    name: 'Morocco',
    flag: '🇲🇦',
    movesCount: 1_890,
    usersHome: 100,
    activeUsers: 204,
    movesByType: { dropped: 449, dipped: 1_190, seen: 251 },
    loves: 347,
    pictures: 290,
    pointsSum: 31_500,
    pointsSumMoves: 44_500,
    geokretyInCache: 120,
    geokretyLost: 8,
    avgPointsPerMove: 3.16,
  },
  {
    code: 'EG',
    name: 'Egypt',
    flag: '🇪🇬',
    movesCount: 1_440,
    usersHome: 76,
    activeUsers: 156,
    movesByType: { dropped: 342, dipped: 906, seen: 192 },
    loves: 263,
    pictures: 221,
    pointsSum: 24_300,
    pointsSumMoves: 34_500,
    geokretyInCache: 92,
    geokretyLost: 7,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'NG',
    name: 'Nigeria',
    flag: '🇳🇬',
    movesCount: 958,
    usersHome: 51,
    activeUsers: 104,
    movesByType: { dropped: 228, dipped: 602, seen: 128 },
    loves: 176,
    pictures: 147,
    pointsSum: 15_800,
    pointsSumMoves: 22_500,
    geokretyInCache: 61,
    geokretyLost: 5,
    avgPointsPerMove: 3.13,
  },
  {
    code: 'KE',
    name: 'Kenya',
    flag: '🇰🇪',
    movesCount: 742,
    usersHome: 39,
    activeUsers: 80,
    movesByType: { dropped: 176, dipped: 466, seen: 100 },
    loves: 136,
    pictures: 114,
    pointsSum: 12_200,
    pointsSumMoves: 17_400,
    geokretyInCache: 47,
    geokretyLost: 4,
    avgPointsPerMove: 3.1,
  },
  // ── Oceania ───────────────────────────────────
  {
    code: 'AU',
    name: 'Australia',
    flag: '🇦🇺',
    movesCount: 10_400,
    usersHome: 551,
    activeUsers: 1_105,
    movesByType: { dropped: 2_470, dipped: 6_540, seen: 1_390 },
    loves: 1_908,
    pictures: 1_598,
    pointsSum: 175_000,
    pointsSumMoves: 249_000,
    geokretyInCache: 661,
    geokretyLost: 44,
    avgPointsPerMove: 3.19,
  },
  {
    code: 'NZ',
    name: 'New Zealand',
    flag: '🇳🇿',
    movesCount: 3_900,
    usersHome: 207,
    activeUsers: 418,
    movesByType: { dropped: 927, dipped: 2_451, seen: 522 },
    loves: 714,
    pictures: 599,
    pointsSum: 66_500,
    pointsSumMoves: 94_500,
    geokretyInCache: 248,
    geokretyLost: 17,
    avgPointsPerMove: 3.23,
  },
  {
    code: 'PG',
    name: 'Papua New Guinea',
    flag: '🇵🇬',
    movesCount: 558,
    usersHome: 30,
    activeUsers: 61,
    movesByType: { dropped: 133, dipped: 351, seen: 74 },
    loves: 102,
    pictures: 86,
    pointsSum: 9_300,
    pointsSumMoves: 13_200,
    geokretyInCache: 36,
    geokretyLost: 3,
    avgPointsPerMove: 3.14,
  },
  {
    code: 'FJ',
    name: 'Fiji',
    flag: '🇫🇯',
    movesCount: 441,
    usersHome: 23,
    activeUsers: 47,
    movesByType: { dropped: 105, dipped: 277, seen: 59 },
    loves: 81,
    pictures: 68,
    pointsSum: 7_400,
    pointsSumMoves: 10_500,
    geokretyInCache: 28,
    geokretyLost: 2,
    avgPointsPerMove: 3.18,
  },
]

export function useCountries() {
  const fetchCountries = async (): Promise<CountryStats[]> => {
    await delay(400)
    return COUNTRY_DATA
  }
  return { fetchCountries }
}
