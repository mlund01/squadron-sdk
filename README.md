# Squadron SDK

Go SDK for building [Squadron](https://github.com/mlund01/squadron) plugins.

Plugins extend Squadron agents with custom tools — browser automation, database queries, API integrations, or anything else you can write in Go. Each plugin is a standalone binary that Squadron manages as a subprocess, communicating over gRPC.

## Quick Start

### 1. Create a new Go module

```bash
mkdir plugin_example && cd plugin_example
go mod init github.com/yourname/plugin_example
go get github.com/mlund01/squadron-sdk
```

### 2. Implement the plugin

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"

    squadron "github.com/mlund01/squadron-sdk"
)

type GreetInput struct {
    Name string `json:"name" jsonschema:"required,description=The name to greet"`
}

func main() {
    app := squadron.New()

    var greeting string
    app.Configure(func(settings map[string]string) error {
        greeting = settings["greeting"]
        if greeting == "" {
            greeting = "Hello"
        }
        return nil
    })

    squadron.Tool(app, "greet", "Greet someone by name",
        func(ctx context.Context, in GreetInput) (string, error) {
            return fmt.Sprintf("%s, %s!", greeting, in.Name), nil
        })

    app.Serve()
}
```

The schema is reflected from `GreetInput`'s struct tags via
[github.com/invopop/jsonschema](https://github.com/invopop/jsonschema). Use
the standard `jsonschema:` tag keys (`required`, `description=...`,
`enum=...,enum=...`, `minimum=N`, `maximum=N`, `default=...`, `pattern=...`,
etc.). For descriptions that need commas, use the dedicated
`jsonschema_description:` tag instead.

Pass `struct{}` as the input type for tools that take no parameters.

### 3. Build and install locally

```bash
mkdir -p ~/.squadron/plugins/example/local
go build -o ~/.squadron/plugins/example/local/plugin .
```

Or use the Squadron CLI:

```bash
squadron plugin build example .
```

### 4. Use it in your config

```hcl
plugin "example" {
  version = "local"
  settings = {
    greeting = "Hey there"
  }
}

agent "assistant" {
  model = models.anthropic.claude_sonnet_4
  tools = [plugins.example.greet]
}
```

### 5. Test it

```bash
# List available tools
squadron plugin tools example

# Call a tool directly
squadron plugin call example greet '{"name": "World"}'

# Use all tools from the plugin
# tools = [plugins.example.all]
```

## API layers

There are two ways to write a plugin:

### App + Tool[I, O] (recommended)

A typed, generics-driven API. The schema is derived from the input struct,
the payload is unmarshaled and the result marshaled automatically:

```go
app := squadron.New()
app.Configure(func(settings map[string]string) error { ... })
squadron.Tool(app, name, description, func(ctx, in InputType) (Output, error) { ... })
app.Serve()
```

Behind the scenes this builds a `ToolProvider` (see below) that ships full
JSON Schema bytes — `enum`, `default`, `minimum`/`maximum`, `$defs`, etc. all
flow through to the LLM verbatim.

### ToolProvider (low-level)

The underlying contract, useful when tools are dynamic or the typed API
doesn't fit:

```go
type ToolProvider interface {
    Configure(settings map[string]string) error
    Call(ctx context.Context, toolName string, payload string) (string, error)
    GetToolInfo(toolName string) (*ToolInfo, error)
    ListTools() ([]*ToolInfo, error)
}
```

`ToolInfo.RawSchema` (a `json.RawMessage`) ships verbatim and overrides the
typed `Schema` if both are set — use it to hand-write rich JSON Schema.

The entry point is `squadron.Serve(provider)` (or `app.Serve()`, which is
the same thing).

## Settings

Plugins receive settings from the HCL config via `Configure()`:

```hcl
plugin "example" {
  version = "local"
  settings = {
    api_url  = "https://api.example.com"
    timeout  = "30"
    headless = "true"
  }
}
```

All values are strings — parse them in `Configure()` as needed.

## Local Development

During development, use `version = "local"` to point Squadron at your local build:

```bash
# Build to the local plugin path
mkdir -p ~/.squadron/plugins/myplugin/local
go build -o ~/.squadron/plugins/myplugin/local/plugin .

# Or use the CLI shorthand
squadron plugin build myplugin .

# Reference in config
# plugin "myplugin" { version = "local" }
```

The plugin binary must be named `plugin` (or `plugin.exe` on Windows) and placed at:

```
~/.squadron/plugins/<name>/<version>/plugin
```

Plugin processes are managed by Squadron — they start when first referenced and persist for the duration of the mission (so state like database connections or browser sessions is maintained across tasks).

## Releasing

To distribute your plugin, set up [GoReleaser](https://goreleaser.com/) with GitHub Actions. Squadron auto-downloads published plugins on first use.

### `.goreleaser.yml`

```yaml
version: 2

project_name: plugin_example

builds:
  - id: plugin
    binary: plugin
    main: .
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

release:
  github:
    owner: yourname
    name: plugin_example
```

### `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Publish a release

```bash
git tag v0.0.1
git push origin v0.0.1
```

GoReleaser builds cross-platform binaries, generates checksums, and creates a GitHub release. Users reference your plugin by source and version:

```hcl
plugin "example" {
  source  = "github.com/yourname/plugin_example"
  version = "v0.0.1"
}
```

Squadron downloads the correct binary for the user's platform, verifies the checksum, and installs it automatically.

## Existing Plugins

| Plugin | Description |
|--------|-------------|
| [plugin_playwright](https://github.com/mlund01/plugin_playwright) | Browser automation (navigate, click, screenshot, aria snapshots) |
| [plugin_pinger](https://github.com/mlund01/plugin_pinger) | Minimal example plugin (ping/pong/echo) |
| [plugin_databricks_sql](https://github.com/mlund01/plugin_databricks_sql) | Databricks SQL queries and schema exploration |

## License

MIT
