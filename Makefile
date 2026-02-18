.PHONY: build run test clean install run-csv

BINARY_NAME=peretran
BIN_DIR=./bin
BUILD_DIR=$(BIN_DIR)/$(BINARY_NAME)

install:
	go mod download

build: clean
	@mkdir -p $(BIN_DIR)
	go build -o $(BUILD_DIR)

run: build
	$(BUILD_DIR) $(ARGS)

run-csv: build
	$(BUILD_DIR) translate csv $(ARGS)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)

help:
	@echo "Available targets:"
	@echo "  make install     - Download Go dependencies"
	@echo "  make build       - Build binary to ./bin/peretran"
	@echo "  make run         - Build and run with ARGS=..."
	@echo "  make run-csv     - Build and run 'translate csv' subcommand with ARGS=..."
	@echo "  make test        - Run tests"
	@echo "  make vet         - Run go vet"
	@echo "  make clean       - Remove binary"
	@echo ""
	@echo "Examples:"
	@echo "  make run ARGS='-i input.txt -o output.txt -t es'"
	@echo "  make run-csv ARGS='-i data.csv -o out.csv -t uk -l 1 -l 3'"
