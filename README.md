# TaxMan

## Overview
Welcome to the TaxMan! This project implements a backend API to manage taxes applied in different municipalities. The API provides JSON-based REST API endpoints to manage tax records and retrieve tax rates for specific municipalities on given dates. The API is implemented in Go, with a PostgreSQL database used for storing the data.

## Features

- Store and manage tax records for different municipalities.
- Add new tax records for municipalities individually.
- Query specific municipality taxes by municipality name and date.
- Expose functionality via APIs (no user interface required).
- Handle errors gracefully, ensuring internal errors are not exposed to the end user.
- Includes unit and integration tests for reliability.
- Dockerized for easy deployment and testing.

# API Endpoints
For detailed information on the API endpoints, please refer to the [API Documentation](docs/openapi.yaml).

## Setup and Installation
Running tests that require a database connection is posbile in two ways:
1. Using a docker container and the make command:
```bash
make test
```
This will start a PostgresSQL container and run the tests. The container will be removed after the tests are done. This includes the unit tests and the integration tests.

2. Using a local PostgresSQL database and setting the environment variables:
```bash
export TEST_DB_URL=postgres://user:password@localhost:5432/dbname
```
Then run the tests:
```bash
make test-local
```

## Running the server
To run the server, you can use the make command:
```bash
make run
```
This will start the server on port 8080. You can change the port by setting the PORT enviroment variable.

The service will be available at http://localhost:8080.


To the run the server in the background, you can use the make command:
```bash
make bg-run
```

### Store Interface Segregation
We have implemented Store Interface Segregation, which defines separate interfaces for different store functionalities. This ensures that the service is not dependent on the implementation of the store, promoting maintainability and flexibility. By breaking down the store interfaces based on the domain of the service, such as tax store interface and municipality store interface, we achieve:

1. Decoupling: The service is decoupled from the store implementation, making it easier to switch or modify store implementations without affecting the service logic.
2. Maintainability: Smaller, focused interfaces are easier to understand and maintain, reducing the complexity of the system.
3. Testability: Segregated interfaces make it easier to mock dependencies during testing, resulting in more reliable and isolated tests.

### Suggestions for future improvements
1. Validation of input data can be improved to ensure date ranges are matching the period types.
2. Implementing a caching mechanism to store frequently accessed tax rates can improve performance.
3. Adding more detailed logging and monitoring to track API usage and errors.
4. Implementing rate limiting to prevent abuse of the API.
5. Adding support for more advanced queries, such as filtering by tax rate or date range.
6. Implementing a API key or token-based authentication mechanism to secure the API and track usage.