.PHONY: dev install build prod

# Run the dev server on the host (uses pnpm if available, falls back to npm)
dev:
	@echo "Starting frontend dev server on host..."
	@if command -v pnpm >/dev/null 2>&1; then \
		pnpm dev; \
	elif command -v npm >/dev/null 2>&1; then \
		npm run dev; \
	elif command -v npx >/dev/null 2>&1; then \
		npx vite --host 0.0.0.0 --port 5173; \
	else \
		echo "Please install pnpm or npm (or use npx) to run the dev server on host."; exit 1; \
	fi

# Install dependencies on host
install:
	@if command -v pnpm >/dev/null 2>&1; then pnpm install; elif command -v npm >/dev/null 2>&1; then npm ci; else echo "Please install pnpm or npm"; exit 1; fi

# Build production assets on host
build:
	@if command -v pnpm >/dev/null 2>&1; then pnpm build; else npm run build; fi

# Start production container for testing (keeps Dockerfile and compose unchanged)
prod:
	docker compose --profile prod up --force-recreate --build
