# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

---

## [0.0.1] - 2024-01-XX

### Added
- Initial release of Mockchaos library
- HTTP mocking with configurable handlers, status codes, latencies, and headers
- gRPC mocking with dynamic service registration from `.proto` files
- Latency simulation with random duration selection from multiple options
- JSON-based configuration for both HTTP and gRPC mocks
- Test utilities package (`httptest`) for HTTP integration testing
- Test utilities package (`grpctest`) for gRPC integration testing
- Standalone HTTP mock server implementation
- Standalone gRPC mock server implementation
- CLI tool (`chaosmock`) for running standalone mock servers
- Support for randomized status codes and latencies
- Automatic recursive loading of JSON mock files from directories
- gRPC reflection support for service discovery
- Example integration tests for both HTTP and gRPC
- Kubernetes deployment examples
- Comprehensive documentation and README

### Features
- **HTTP Mocking**: Create mock HTTP servers with configurable endpoints, responses, status codes, latencies, and headers
- **gRPC Mocking**: Create mock gRPC servers with automatic service registration from proto files
- **Latency Simulation**: Simulate network delays with random selection from multiple duration options
- **JSON Configuration**: Define mocks using simple JSON files for easy maintenance
- **Test Utilities**: Built-in helpers for integration testing (`httptest` and `grpctest` packages)
- **Standalone Servers**: Run mock servers as standalone applications or in Kubernetes
- **Chaos Engineering**: Test service resilience with variable latencies and error conditions

[Unreleased]: https://github.com/nawafswe/mockchaos/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/nawafswe/mockchaos/releases/tag/v0.0.1