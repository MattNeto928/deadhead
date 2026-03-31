# Contributing

Thanks for your interest in contributing to deadhead.

## Prerequisites

- Go 1.26+
- Google Chrome (required to run the tool; not required to run tests)

## Getting started

```bash
git clone https://github.com/mattneto928/deadhead
cd deadhead
go build -o deadhead ./cmd/...
```

## Running tests

```bash
go test ./...
```

Tests mock all HTTP calls and do not require a running browser or network access.

## Making changes

1. Fork the repository and create a branch from `main`.
2. Make your changes.
3. Add or update tests for any logic you change. All packages have test coverage — keep it that way.
4. Run `go test ./...` and confirm everything passes.
5. Run `go vet ./...` and fix any warnings.
6. Open a pull request against `main` with a clear description of what changed and why.

## Code style

- Standard `gofmt` formatting. Run `gofmt -w .` before committing.
- Exported types and functions must have doc comments.
- Avoid adding dependencies. This project intentionally has a small dependency footprint.

## What to work on

Check the [issue tracker](https://github.com/mattneto928/deadhead/issues) for open issues. Issues labeled `good first issue` are a good starting point.

## Reporting bugs

Open an issue with:
- The exact command you ran
- What you expected to happen
- What actually happened (include any error output)
- Your OS and Go version (`go version`)

## Security issues

Please do not open a public issue for security vulnerabilities. Email the maintainer directly instead.
