# Architecture

Dependencies point inward through these layers:

```text
cmd/remote  ->  internal/cli  ->  internal/application
                                      |       |
                                      v       v
                               internal/domain
                               internal/project
                               internal/bundle
```

- `cmd/remote` is the composition root and process exit boundary.
- `internal/cli` parses commands and renders user-facing output.
- `internal/application` coordinates scaffold, validate, and build use cases.
- `internal/domain` owns manifest types and pure validation policy.
- `internal/project` adapts application projects on the filesystem.
- `internal/bundle` owns source collection, infrastructure payloads,
  reproducible archives, and atomic artifact writes.

Domain code must not import CLI, filesystem, or archive packages. The CLI must
not implement domain or packaging policy. Infrastructure packages expose small
operations used by application workflows; they do not format CLI output.

Put a new manifest rule in `domain`, an archive exclusion in `bundle`, command
syntax in `cli`, and the order of a multi-step use case in `application`.
