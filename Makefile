.PHONY: all build run test clean test-local

all: test

# Build the Docker image
build:
	@docker-compose build

# Run the Docker container
run: build
	@docker-compose up server

# Run the Docker container in the background
run-bg: build
	@docker-compose up -d server

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
