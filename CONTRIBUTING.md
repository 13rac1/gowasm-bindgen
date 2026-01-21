# Contributing to gowasm-bindgen

## Development Setup

1. Clone the repository
2. Install Go 1.25+
3. Install golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
4. Run `make check` to verify everything works

## Making Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run `make check` (formats, lints, tests)
5. Commit with a clear message
6. Open a pull request

## Code Style

- Follow standard Go conventions
- Run `make format` before committing
- All lints must pass (`make lint`)
- Add tests for new functionality

## Pull Request Guidelines

- Keep PRs focused on a single change
- Update documentation if needed
- Add tests for new features
- Ensure CI passes
