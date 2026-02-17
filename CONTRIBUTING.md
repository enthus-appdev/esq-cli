# Contributing to esq-cli

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/enthus-appdev/esq-cli.git
cd esq-cli
make build    # Build the binary
make test     # Run tests
make lint     # Run goimports + golangci-lint
```

Requires Go 1.24+ and [golangci-lint](https://golangci-lint.run/).

## Making Changes

1. Fork the repository and create a feature branch from `main`
2. Write your code and add tests where appropriate
3. Run `make lint && make test` to ensure all checks pass
4. Commit with a clear message describing the change
5. Open a pull request against `main`

## Code Style

- Run `goimports -w .` before committing
- Follow standard Go conventions

## Reporting Bugs

Open a [GitHub issue](https://github.com/enthus-appdev/esq-cli/issues) with:
- Steps to reproduce
- Expected vs actual behavior
- CLI version (`esq --version`) and OS

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
