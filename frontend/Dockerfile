# ============================================================
# Stage 1: dependencies
# ============================================================
FROM node:22-alpine AS deps

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy lockfile and manifests only for layer-cache efficiency
COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

# ============================================================
# Stage 2: builder
# ============================================================
FROM deps AS builder

ARG BUILD_VERSION=dev
ARG BUILD_DATE=unknown

# Copy full source
COPY . .

# Embed version info
ENV VITE_APP_VERSION=${BUILD_VERSION}
ENV VITE_APP_BUILD_DATE=${BUILD_DATE}

RUN pnpm build

# ============================================================
# Stage 3: test / lint (optional in CI)
# ============================================================
FROM builder AS test
RUN pnpm lint && pnpm test:unit --run

# ============================================================
# Stage 4: production runtime (nginx)
# ============================================================
FROM nginx:1.27-alpine AS runtime

# Security: run as non-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install tzdata for proper timezone handling
RUN apk add --no-cache tzdata

# Replace default nginx config with SPA-friendly one
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Correct permissions
RUN chown -R appuser:appgroup /usr/share/nginx/html \
    && chown -R appuser:appgroup /var/cache/nginx \
    && touch /var/run/nginx.pid \
    && chown appuser:appgroup /var/run/nginx.pid

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]
