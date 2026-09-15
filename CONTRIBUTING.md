# Contributing

## Before opening a pull request

Run the same core checks as CI:

```sh
make test
make test-race
make coverage
go vet ./...
```

CI runs tests and vet on Linux, macOS, and Windows, validates both installers,
runs the race detector, cross-compiles every supported OS/architecture pair,
and rejects total statement coverage below 80%.

## Testing changes

- Put pure domain-rule tests in `internal/domain`.
- Put filesystem adapter tests in `internal/project`.
- Put archive and artifact tests in `internal/bundle`.
- Put use-case and end-to-end tests in `internal/application`.
- Put command parsing, output, aliases, and error mapping tests in `internal/cli`.
- Test success, invalid input, boundary failures, and exact observable output.
- Reproduce every bug with a failing regression test before fixing it.

Keep commits focused. Structural refactors must preserve observable behavior and
should be separate from feature or bug-fix commits.
