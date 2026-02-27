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
        run_2010 run_2011 run_2012 run_2013 run_2014 \
        run_2015 run_2016 run_2017 run_2018 run_2019 \
        run_2020 run_2021 run_2022 run_2023 run_2024 \
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
.PHONY: run_$(1)
## run_$(1): Replay all moves from $(1). Wipes stats schema first.
run_$(1): build
	$(BUILD_DIR)/$(BINARY) -replay -year $(1) -truncate
endef

$(foreach year,2010 2011 2012 2013 2014 2015 2016 2017 2018 2019 2020 2021 2022 2023 2024,\
	$(eval $(call YEAR_TARGET,$(year))))
