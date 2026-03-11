variable "VERSION" {
  default = "latest"
}

variable "REGISTRY" {
  default = "geokrety"
}

variable "BUILD_DATE" {
  default = timestamp()
}

group "default" {
  targets = ["runtime"]
}

# ── Shared base ───────────────────────────────────────────────────────────────
target "_base" {
  dockerfile = "Dockerfile"
  args = {
    BUILD_VERSION = VERSION
    BUILD_DATE    = BUILD_DATE
  }
  labels = {
    "org.opencontainers.image.title"   = "geokrety-stats-frontend"
    "org.opencontainers.image.source"  = "https://github.com/geokrety/geokrety-stats-frontend"
    "org.opencontainers.image.version" = VERSION
    "org.opencontainers.image.created" = BUILD_DATE
  }
}

# ── Development build (single-platform, loads locally) ───────────────────────
target "dev" {
  inherits = ["_base"]
  target   = "builder"
  tags     = ["${REGISTRY}/geokrety-stats-frontend:dev"]
  platforms = ["linux/amd64"]
  cache-from = ["type=local,src=.docker-cache"]
  cache-to   = ["type=local,dest=.docker-cache,mode=max"]
}

# ── Test stage (runs lint + unit tests inside Docker) ────────────────────────
target "test" {
  inherits = ["_base"]
  target   = "test"
  tags     = ["${REGISTRY}/geokrety-stats-frontend:test"]
  platforms = ["linux/amd64"]
}

# ── Production runtime (multi-platform) ──────────────────────────────────────
target "runtime" {
  inherits = ["_base"]
  target   = "runtime"
  platforms = ["linux/amd64", "linux/arm64"]
  tags = [
    "${REGISTRY}/geokrety-stats-frontend:latest",
    "${REGISTRY}/geokrety-stats-frontend:${VERSION}",
  ]
}
