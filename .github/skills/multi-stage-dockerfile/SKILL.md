---
name: multi-stage-dockerfile
description: 'Create optimized multi-stage Dockerfiles for any language or framework'
---

Your goal is to help me create efficient multi-stage Dockerfiles that follow best practices, resulting in smaller, more secure container images.

## Multi-Stage Structure

- Use a builder stage for compilation, dependency installation, and other build-time operations
- Use a separate runtime stage that only includes what's needed to run the application
- Copy only the necessary artifacts from the builder stage to the runtime stage
- Use meaningful stage names with the `AS` keyword (e.g., `FROM node:18 AS builder`)
- Place stages in logical order: dependencies → build → test → runtime

## Base Images

- Start with official, minimal base images when possible
- Specify exact version tags to ensure reproducible builds (e.g., `python:3.11-slim` not just `python`)
- Consider distroless images for runtime stages where appropriate
- Use Alpine-based images for smaller footprints when compatible with your application
- Ensure the runtime image has the minimal necessary dependencies

## Layer Optimization

- Organize commands to maximize layer caching
- Place commands that change frequently (like code changes) after commands that change less frequently (like dependency installation)
- Use `.dockerignore` to prevent unnecessary files from being included in the build context
- Combine related RUN commands with `&&` to reduce layer count
- Consider using COPY --chown to set permissions in one step

## Security Practices

- Avoid running containers as root - use `USER` instruction to specify a non-root user
- Remove build tools and unnecessary packages from the final image
- Scan the final image for vulnerabilities
- Set restrictive file permissions
- Use multi-stage builds to avoid including build secrets in the final image

## Performance Considerations

- Use build arguments for configuration that might change between environments
- Leverage build cache efficiently by ordering layers from least to most frequently changing
- Consider parallelization in build steps when possible
- Set appropriate environment variables like NODE_ENV=production to optimize runtime behavior
- Use appropriate healthchecks for the application type with the HEALTHCHECK instruction

## Docker Bake (docker-bake.hcl)

Docker Bake is a high-level build definition language that simplifies building complex multi-platform images with multiple targets. Use `docker-bake.hcl` for:

### Configuration Benefits

- **Matrix builds**: Build multiple platforms (linux/amd64, linux/arm64) and variants in one command
- **Build arguments**: Define and override build arguments for flexibility
- **Output control**: Specify where built images go (registries, local, OCI format)
- **Caching strategy**: Configure build cache behavior and remote caching
- **Target definitions**: Define multiple Docker targets with shared configuration

### Basic Structure

```hcl
# docker-bake.hcl

group "default" {
  targets = ["app"]
}

target "app" {
  dockerfile = "Dockerfile"
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    NODE_ENV = "production"
  }
  tags = [
    "geokrety/myapp:latest",
    "geokrety/myapp:${VERSION}"
  ]
  output = ["type=registry"]
}
```

### Common Patterns

**Development vs Production Builds:**
```hcl
target "dev" {
  dockerfile = "Dockerfile"
  tags = ["geokrety/myapp:dev"]
  cache-from = ["type=local,src=.docker-cache"]
  cache-to = ["type=local,dest=.docker-cache"]
}

target "prod" {
  inherits = ["dev"]
  dockerfile = "Dockerfile.prod"
  tags = ["geokrety/myapp:latest"]
  output = ["type=registry"]
}
```

**Multi-Platform Builds:**
```hcl
target "app" {
  platforms = [
    "linux/amd64",
    "linux/arm64",
    "linux/arm/v7"
  ]
  output = ["type=oci,dest=./output"]
}
```

**Using Build Arguments:**
```hcl
variable "VERSION" {
  default = "latest"
}

target "app" {
  args = {
    BUILD_VERSION = VERSION
    BUILD_DATE = timestamp()
  }
}
```

### Building with Docker Bake

```bash
# Build default target (requires group "default")
docker buildx bake

# Build specific target
docker buildx bake app

# Build multiple targets
docker buildx bake app database

# Build with variable override
docker buildx bake --set app.args.version=1.2.3

# Build all platforms and push to registry
docker buildx bake app --push

# Build and load locally (single platform only)
docker buildx bake app --load
```

### Advanced Features

- **Caching**: Use `cache-from` and `cache-to` for faster rebuilds locally and in CI/CD
- **Source control**: Use `source` attribute to reference external Dockerfile sources
- **Inheritance**: Targets can inherit from other targets to reduce duplication
- **Conditional logic**: Use variables and HCL expressions for dynamic configuration
- **CI/CD Integration**: Perfect for GitHub Actions with `docker/bake-action`
