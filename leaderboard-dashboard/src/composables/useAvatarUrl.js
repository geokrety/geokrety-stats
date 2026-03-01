const STORAGE_BASE = 'https://minio.geokrety.org'
const FALLBACK_USER_AVATAR = 'https://cdn.geokrety.org/images/log-icons/2/icon.svg'
const FALLBACK_GK_AVATAR = 'https://cdn.geokrety.org/images/the-mole.svg'

const THUMBNAIL_BUCKETS = {
  'users-avatars': 'users-avatars-thumbnails',
  'gk-avatars': 'gk-avatars-thumbnails',
}

function hasProtocol(value) {
  return typeof value === 'string' && (value.startsWith('http://') || value.startsWith('https://'))
}

function resolveBucket(bucket) {
  return THUMBNAIL_BUCKETS[bucket] || bucket
}

function buildAvatarUrl(avatar, fallback) {
  if (!avatar || typeof avatar !== 'string' || avatar === 'null' || avatar === 'undefined') return fallback
  if (hasProtocol(avatar)) return avatar

  const parts = avatar.split('/')
  if (parts.length !== 2) return fallback

  const [bucket, key] = parts
  if (!bucket || !key) return fallback

  const targetBucket = resolveBucket(bucket)
  return `${STORAGE_BASE}/${targetBucket}/${encodeURIComponent(key)}`
}

export function userAvatarUrl(userId) {
  return buildAvatarUrl(userId, FALLBACK_USER_AVATAR)
}

export function gkAvatarUrl(gkId) {
  return buildAvatarUrl(gkId, FALLBACK_GK_AVATAR)
}
