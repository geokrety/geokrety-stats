# ── Dependencies stage ──────────────────────────────────────────────────────
FROM node:22-alpine AS dependencies

WORKDIR /app
COPY package*.json ./
RUN npm ci

# ── Build stage ─────────────────────────────────────────────────────────────
FROM dependencies AS builder

COPY . .
RUN npm run build

# ── Development stage (Vite with hot reload) ────────────────────────────────
FROM dependencies AS dev

WORKDIR /app
COPY . .

# Expose Vite dev server port
EXPOSE 5173
EXPOSE 3000

CMD ["npm", "run", "dev"]

# ── Production stage (nginx) ────────────────────────────────────────────────
FROM nginx:1.27-alpine AS prod

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# SPA fallback and reverse proxy config
RUN cat > /etc/nginx/conf.d/default.conf <<'EOF'
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to leaderboard-api
    location /api/ {
        proxy_pass         http://leaderboard-api:8080;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
    }

    # Proxy WebSocket
    location /ws {
        proxy_pass         http://leaderboard-api:8080;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host $host;
        proxy_read_timeout 3600s;
    }
}
EOF

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
