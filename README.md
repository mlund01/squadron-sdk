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
    "encoding/json"
    "fmt"

    squadron "github.com/mlund01/squadron-sdk"
)

var tools = map[string]*squadron.ToolInfo{
    "greet": {
        Name:        "greet",
        Description: "Greet someone by name",
        Schema: squadron.Schema{
            Type: squadron.TypeObject,
            Properties: squadron.PropertyMap{
                "name": {
                    Type:        squadron.TypeString,
                    Description: "The name to greet",
                },
            },
            Required: []string{"name"},
        },
    },
}

type Plugin struct {
    greeting string
}

func (p *Plugin) Configure(settings map[string]string) error {
    if v, ok := settings["greeting"]; ok {
        p.greeting = v
    }
    return nil
}

func (p *Plugin) Call(toolName string, payload string) (string, error) {
    switch toolName {
    case "greet":
        var params struct {
            Name string `json:"name"`
        }
        if err := json.Unmarshal([]byte(payload), &params); err != nil {
            return "", fmt.Errorf("invalid payload: %w", err)
        }
        greeting := p.greeting
        if greeting == "" {
            greeting = "Hello"
        }
        return fmt.Sprintf("%s, %s!", greeting, params.Name), nil
    default:
        return "", fmt.Errorf("unknown tool: %s", toolName)
    }
}

func (p *Plugin) GetToolInfo(toolName string) (*squadron.ToolInfo, error) {
    info, ok := tools[toolName]
    if !ok {
        return nil, fmt.Errorf("unknown tool: %s", toolName)
    }
    return info, nil
}

func (p *Plugin) ListTools() ([]*squadron.ToolInfo, error) {
    result := make([]*squadron.ToolInfo, 0, len(tools))
    for _, info := range tools {
        result = append(result, info)
    }
    return result, nil
}

func main() {
    squadron.Serve(&Plugin{})
}
```

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

## Interface

Every plugin implements the `ToolProvider` interface:

```go
type ToolProvider interface {
    Configure(settings map[string]string) error
    Call(toolName string, payload string) (string, error)
    GetToolInfo(toolName string) (*ToolInfo, error)
    ListTools() ([]*ToolInfo, error)
}
```

| Method | Purpose |
|--------|---------|
| `Configure` | Receive settings from HCL config. Use this to initialize connections, set options, etc. |
| `Call` | Handle a tool invocation. `payload` is a JSON string matching the tool's schema. |
| `GetToolInfo` | Return metadata and schema for a specific tool. |
| `ListTools` | Return metadata for all tools the plugin provides. |

The entry point is always `squadron.Serve(&YourPlugin{})` in `main()`.

## Schema Types

Tool parameters are defined using JSON Schema types:

```go
squadron.TypeString   // "string"
squadron.TypeNumber   // "number"
squadron.TypeInteger  // "integer"
squadron.TypeBoolean  // "boolean"
squadron.TypeArray    // "array"
squadron.TypeObject   // "object"
```

Nested objects and arrays are supported:

```go
Schema: squadron.Schema{
    Type: squadron.TypeObject,
    Properties: squadron.PropertyMap{
        "query": {
            Type:        squadron.TypeString,
            Description: "SQL query to execute",
        },
        "options": {
            Type:        squadron.TypeObject,
            Description: "Query options",
            Properties: squadron.PropertyMap{
                "timeout": {Type: squadron.TypeInteger, Description: "Timeout in seconds"},
                "limit":   {Type: squadron.TypeInteger, Description: "Max rows to return"},
            },
        },
        "tags": {
            Type:        squadron.TypeArray,
            Description: "Tags for the query",
            Items:       &squadron.Property{Type: squadron.TypeString},
        },
    },
    Required: []string{"query"},
},
```

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
