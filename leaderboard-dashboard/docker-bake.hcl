# docker-bake.hcl
# Multistage build configuration for development and production

variable "REGISTRY" {
  default = "localhost"
}

variable "TAG" {
  default = "latest"
}

group "default" {
  targets = ["prod"]
}

group "dev" {
  targets = ["dev"]
}

# Production build: optimized, minimal size
target "prod" {
  dockerfile = "Dockerfile"
  args = {
    BUILD_TARGET = "prod"
  }
  platforms = ["linux/amd64", "linux/arm64"]
  tags = [
    "${REGISTRY}/leaderboard-dashboard:${TAG}",
    "${REGISTRY}/leaderboard-dashboard:latest"
  ]
}

# Development build: includes dev dependencies, watches for changes
target "dev" {
  dockerfile = "Dockerfile"
  args = {
    BUILD_TARGET = "dev"
  }
  platforms = ["linux/amd64"]
  tags = [
    "${REGISTRY}/leaderboard-dashboard:${TAG}-dev",
    "${REGISTRY}/leaderboard-dashboard:dev"
  ]
}
