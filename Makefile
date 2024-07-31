.PHONY: all build run run-bg test test-local clean

all: test

# Build the Docker image
build:
	@docker-compose build

# Run the Docker container and clean up resources afterward
run: build
	@docker-compose up server; \
	STATUS=$$?; \
	$(MAKE) clean; \
	exit $$STATUS

# Run the Docker container in the background
run-bg: build
	@docker-compose up -d server; \
	STATUS=$$?; \
	$(MAKE) clean; \
	exit $$STATUS

# Run the tests in Docker
test: build
	@docker-compose run --rm tester; \
	STATUS=$$?; \
	$(MAKE) clean; \
	exit $$STATUS

# Run the tests locally
test-local:
	@go test -v ./...

# Clean up the Docker environment
clean:
	@docker-compose down -v
