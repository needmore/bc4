# Integration Tests

This directory contains integration tests for bc4 that test against real Basecamp API endpoints.

## Running Integration Tests

Integration tests are disabled by default and use a build tag. To run them:

```bash
# Run integration tests
go test -tags=integration ./test/integration

# Run with verbose output
go test -tags=integration -v ./test/integration

# Run a specific test
go test -tags=integration -v -run TestAPIConnection ./test/integration
```

## Configuration

Integration tests require the following environment variables:

- `BC4_TEST_ACCOUNT_ID`: Your Basecamp account ID
- `BC4_TEST_ACCESS_TOKEN`: A valid access token for the account
- `BC4_TEST_PROJECT_ID`: (Optional) A project ID to use for testing
- `BC4_TEST_SKIP_CLEANUP`: (Optional) Set to "true" to skip cleanup of created resources

Example:
```bash
export BC4_TEST_ACCOUNT_ID="1234567"
export BC4_TEST_ACCESS_TOKEN="your-access-token"
export BC4_TEST_PROJECT_ID="98765432"
export BC4_TEST_SKIP_CLEANUP="false"

go test -tags=integration -v ./test/integration
```

## Test Coverage

The integration tests cover:

1. **API Connection**: Basic connectivity and authentication
2. **Project Operations**: Listing and retrieving projects
3. **Todo Operations**: Creating and completing todos
4. **Campfire Operations**: Listing campfires and retrieving messages
5. **Card Table Operations**: Accessing card tables and columns
6. **Configuration**: Loading and saving configuration files

## Safety Measures

- Tests will skip if required environment variables are not set
- Created resources are cleaned up by default (unless `BC4_TEST_SKIP_CLEANUP` is set)
- Tests use separate test data and avoid modifying existing resources
- All operations are read-only unless explicitly testing write operations

## Writing New Integration Tests

When adding new integration tests:

1. Use the `integration` build tag
2. Check for required configuration using `getTestConfig()`
3. Skip tests gracefully if prerequisites are not met
4. Clean up any resources created during tests
5. Log important actions for debugging