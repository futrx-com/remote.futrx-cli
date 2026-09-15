# Remote CLI

The developer CLI for creating and packaging Remote applications.

## Install

Install the latest release on Linux or macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/futrx-com/remote.futrx-cli/main/install.sh | sh
```

The installer supports AMD64 and ARM64, verifies the release checksum, and
places `remote` in `~/.local/bin` by default. Install a specific release or
choose another directory with environment variables on the `sh` command:

```sh
curl -fsSL https://raw.githubusercontent.com/futrx-com/remote.futrx-cli/main/install.sh \
  | REMOTE_VERSION=v0.1.0 REMOTE_INSTALL_DIR="$HOME/bin" sh
```

## Install from source

```sh
go install github.com/futrx-com/remote.futrx-cli/cmd/remote@latest
```

During local development:

```sh
go build -o ./bin/remote ./cmd/remote
```

## Create an application

```sh
remote create app my-app
cd my-app
remote validate
remote build
```

`create` produces a complete composable application with `ui/`, `backend/`,
`infra/`, `skills/`, `application.json`, documentation, and an MIT license.
Delete capabilities the application does not need and update the manifest.

`build` validates the package and writes `<id>-<version>.zip` beside the app.
The ZIP is reproducible. Infrastructure support files are turned into
`infra/payload.tar.gz` in memory, so an app does not need to ship or run a
`package.sh` script. Legacy `infra/package.sh` and generated payloads are
excluded from the package automatically.

## Commands

```text
remote create app <name> [--dir PATH]
remote app create <name> [--dir PATH]
remote validate [PATH]
remote build [PATH] [-o FILE]
remote package [PATH] [-o FILE]  # alias for build
remote version
```
