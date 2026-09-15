package application

import "fmt"

func scaffoldFiles(id, name string) map[string]string {
	return map[string]string{
		"application.json":           manifestTemplate(id, name),
		"README.md":                  "# " + name + "\n\nA Remote application.\n\n## Development\n\n```sh\nremote validate\nremote build\n```\n",
		"infra/LICENSE":              mitLicense,
		"infra/install.sh":           installTemplate(id),
		"infra/README.md":            "# Infrastructure\n\n`install.sh` runs as root in the target container and must be idempotent. Every other file in this directory is bundled into `APP_PACKAGE_DIR` by `remote build`.\n",
		"backend/main.go":            backendTemplate(id),
		"ui/scripts/main.js":         uiTemplate,
		"ui/style/style.css":         ".remote-app-message { color: var(--text-secondary, inherit); }\n",
		"ui/views/panel.html":        "<p class=\"remote-app-message\">" + name + " is running.</p>\n",
		"skills/" + id + "/SKILL.md": "# " + name + "\n\nUse this application when the installed capability is needed.\n",
		".gitignore":                 "dist/\n*.zip\ninfra/payload.tar.gz\n",
	}
}

func manifestTemplate(id, name string) string {
	return fmt.Sprintf(`{
  "id": %q,
  "name": %q,
  "description": "Describe what this application provides.",
  "category": "development",
  "version": "0.1.0",
  "icon": "layers",
  "scopes": ["project"],
  "env": [],
  "service": %q,
  "install": "infra/install.sh",
  "backend": { "access": "registered", "timeoutMs": 10000 },
  "ui": {
    "entry": "scripts/main.js",
    "styles": ["style/style.css"],
    "views": { "panel": "views/panel.html" }
  }
}
`, id, name, id)
}

func installTemplate(id string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

# This script runs as root and may run repeatedly. Keep it idempotent.
# Auxiliary infra files are available under $APP_PACKAGE_DIR/infra.
install -d -m 0755 /etc/%s

echo "install: %s is ready"
`, id, id)
}

func backendTemplate(id string) string {
	return fmt.Sprintf(`package main

import (
    "github.com/futrx-com/remote.futrx.com/pkg/appplugin"
    "github.com/futrx-com/remote.futrx.com/pkg/appplugin/pluginrpc"
)

type backend struct { mux *appplugin.Mux }

func main() { pluginrpc.Serve(newBackend()) }
func newBackend() *backend { return &backend{mux: appplugin.NewMux()} }
func (b *backend) Describe() (appplugin.Descriptor, error) {
    return appplugin.Descriptor{Name: %q, Version: "1", APIVersion: appplugin.APIVersion, Routes: b.mux.Routes()}, nil
}
func (b *backend) Init(appplugin.Instance) error { return nil }
func (b *backend) Handle(r appplugin.Request) (appplugin.Response, error) { return b.mux.Serve(r), nil }
`, id)
}

const uiTemplate = `export default function activate(remote) {
  remote.ui.register(remote.slots.applicationsPanel, async (host) => {
    host.innerHTML = await remote.views.load("panel");
  });
}
`

const mitLicense = `MIT License

Copyright (c) 2026 Application author

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
