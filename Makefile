APP_NAME ?= ai-learn-english
GO ?= go
BIN_DIR ?= bin
BIN := $(BIN_DIR)/$(APP_NAME)

.PHONY: run build tidy clean

run:
	$(GO) run ./cmd

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)


