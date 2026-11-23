# Contributing to devkit

Thanks for taking the time to contribute.

## Getting started

1. Fork the repository and create a branch from `main`.
2. Install Go 1.22 or newer.
3. Make your change and add tests for it.
4. Run the checks below.
5. Open a pull request that explains what changed and why.

## Checks

All of these must pass before a pull request is merged (CI runs them on Linux, macOS and Windows):

```sh
gofmt -l .        # must print nothing
go vet ./...
go test -cover ./...
go build ./...
```

## Guidelines

- Keep commands pipe-friendly: read from a file or standard input, write results to standard output, errors to standard error.
- Avoid new dependencies unless they are clearly worth it.
- Prefer table-driven tests.
- Use clear commit messages in the imperative mood, for example `Add sha256 command`.
- Update `README.md` and `CHANGELOG.md` when behavior changes.

## Reporting bugs and requesting features

Use the issue templates. For security problems, see [SECURITY.md](SECURITY.md) instead of opening a public issue.
