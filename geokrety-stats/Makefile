# GeoKrety Points System – Makefile
# Usage: make help

BINARY    := geokrety-stats
CMD_PATH  := ./cmd/geokrety-stats
BUILD_DIR := ./bin
GO        := $(shell which go || echo go)
GOFLAGS   ?=

# ── Default target ─────────────────────────────────────────────────────────────
.DEFAULT_GOAL := help

.PHONY: help build test lint run replay \
        replay_2010 replay_2011 replay_2012 replay_2013 replay_2014 \
        replay_2015 replay_2016 replay_2017 replay_2018 replay_2019 \
        replay_2020 replay_2021 replay_2022 replay_2023 replay_2024 \
	fixtures_extract fixtures_refresh_refs fixtures_refresh_one fixtures_test \
	tidy clean

## help: Show this help message.
help:
	@echo "GeoKrety Points System"
	@echo ""
	@echo "Usage:"
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' | \
		sed 's/## //'
	@echo ""
	@echo "Replay targets: run_YYYY where YYYY = 2010..2024"

## build: Compile the daemon binary to ./bin/.
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "Binary: $(BUILD_DIR)/$(BINARY)"

## test: Run all unit tests.
test:
	$(GO) test ./... -v -timeout 120s

## fixtures_extract: Extract all replay fixture input CSVs from source DB.
fixtures_extract: build
	./scripts/replay_fixtures.sh extract-all

## fixtures_refresh_refs: Regenerate all expected fixture reference JSON files.
fixtures_refresh_refs: build
	./scripts/replay_fixtures.sh refresh-all

## fixtures_refresh_one: Regenerate one fixture reference JSON (use FIXTURE=name).
fixtures_refresh_one: build
	@if [ -z "$(FIXTURE)" ]; then \
		echo "Usage: make fixtures_refresh_one FIXTURE=<fixture_name>"; \
		exit 1; \
	fi
	./scripts/replay_fixtures.sh refresh-one $(FIXTURE)

## fixtures_test: Run integration replay fixture tests against expected references.
fixtures_test:
	GK_RUN_REPLAY_FIXTURE_TEST=1 $(GO) test -tags=integration ./internal/computers -run TestReplayFixturesAgainstReferenceJSON -count=1 -v

## lint: Run go vet (add golangci-lint for stricter checks).
lint:
	$(GO) vet ./...

## tidy: Tidy and verify Go modules.
tidy:
	$(GO) mod tidy
	$(GO) mod verify

## clean: Remove build artefacts.
clean:
	rm -rf $(BUILD_DIR)

# ── Runtime helpers ────────────────────────────────────────────────────────────

## run: Start the scoring daemon (AMQP subscriber mode).
run: build
	$(BUILD_DIR)/$(BINARY)

## replay: Run a full historical replay (all moves). Wipes stats schema first.
replay: build
	$(BUILD_DIR)/$(BINARY) -replay -truncate

# ── Per-year replay targets ────────────────────────────────────────────────────
define YEAR_TARGET
.PHONY: replay_$(1)
## replay_$(1): Replay all moves from $(1). Wipes stats schema first.
replay_$(1): build
	$(BUILD_DIR)/$(BINARY) -replay -year $(1) -truncate
endef

$(foreach year,2010 2011 2012 2013 2014 2015 2016 2017 2018 2019 2020 2021 2022 2023 2024,\
	$(eval $(call YEAR_TARGET,$(year))))
