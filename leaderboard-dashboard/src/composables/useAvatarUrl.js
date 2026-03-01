const USER_AVATAR_BASE = 'https://minio.geokrety.org/users-avatars-thumbnails/'
const GK_AVATAR_BASE = 'https://minio.geokrety.org/gk-avatars-thumbnails/'

export function userAvatarUrl(avatar) {
  if (!avatar) return "https://cdn.geokrety.org/images/log-icons/2/icon.svg"
  return `${USER_AVATAR_BASE}US000001_6503ce0ca3b2c` // TODO temp fix
  // return `${USER_AVATAR_BASE}${avatar}`
}

export function gkAvatarUrl(avatar) {
  if (!avatar) return "https://cdn.geokrety.org/images/the-mole.svg"
  return `${GK_AVATAR_BASE}GK1A0C9_68c6951fce84e` // TODO temp fix
  // return `${GK_AVATAR_BASE}${avatar}`
}
