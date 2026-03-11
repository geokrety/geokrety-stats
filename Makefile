VENV := uv
PY := $(VENV)/bin/python
PIP := $(VENV)/bin/pip
ZENSICAL := $(VENV)/bin/zensical

.PHONY: init install build serve preview clean help

help:
	@echo "Zensical documentation management targets:"
	@echo "  make init       - Initialize uv Python virtual environment"
	@echo "  make install    - Install zensical static site generator"
	@echo "  make build      - Build the documentation site"
	@echo "  make serve      - Serve documentation with live reload (for editing)"
	@echo "  make preview    - Build and serve site preview with clickable link"
	@echo "  make clean      - Remove built site directory"

init:
	@echo "Initializing Python virtual environment in $(VENV)..."
	test -d $(VENV) || python3 -m venv $(VENV)
	$(PIP) install --upgrade pip setuptools wheel
	@echo "✓ Virtual environment ready at $(VENV)"

install: init
	@echo "Installing zensical static site generator..."
	$(PIP) install zensical
	@echo "✓ Dependencies installed"

build:
	@echo "Building documentation site..."
	$(ZENSICAL) build
	@echo "✓ Site built at site/"

serve:
	@echo "Starting Zensical live server..."
	@echo "Visit http://127.0.0.1:8160 in your browser"
	@echo "Press Ctrl+C to stop"
	$(ZENSICAL) serve --dev-addr=localhost:8160

preview: build
	@echo "Starting HTTP server on port 8160..."
	@URL="http://127.0.0.1:8160"; \
	( cd site && $(PY) -m http.server 8160 2>/dev/null & ); \
	sleep 1; \
	printf '\033]8;;%s\a%s\033]8;;\a\n' "$$URL" "📖 Documentation Preview:" ; \
	echo "Serving at $$URL"; \
	echo "Press Ctrl+C to stop..."; \
	wait

clean:
	@echo "Removing built site..."
	rm -rf site
	@echo "✓ Cleaned"
