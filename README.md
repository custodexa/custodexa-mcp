# custodexa-mcp

<p><b>English</b> | <a href="docs/zh-TW/README.md">繁體中文</a> | <a href="docs/ja/README.md">日本語</a></p>

`custodexa-mcp` lets MCP hosts that can only launch local stdio servers reach Custodexa. The MCP server itself is part of Custodexa: the `POST /api/v1/mcp` endpoint of your deployment, where authorization, connections, auditing, recording and masking take place. This program is a stdio server to the host and an HTTP client to that endpoint. It forwards every request with an agent token and makes no decisions of its own. Hosts that support streamable HTTP servers can connect to Custodexa directly and do not need it.

## Do you need it?

| Your MCP host | What to do |
|---|---|
| Launches local servers over stdio only | Install `custodexa-mcp` and point the host at it |
| Supports remote streamable HTTP servers with a custom `Authorization` header | Connect directly to `https://<your-custodexa>/api/v1/mcp` |

A direct connection sends the agent token as a bearer token. [Configure your host](#configure-your-host) has an entry for each host.

## Install

### With Go

Requires Go 1.25 or later.

```bash
go install github.com/custodexa/custodexa-mcp@latest
```

The binary lands in `$(go env GOPATH)/bin`, or in `GOBIN` if you set it.

### From a GitHub Release

Each [release](https://github.com/custodexa/custodexa-mcp/releases) carries an archive per platform, named `custodexa-mcp_<version>_<os>_<arch>` with `<os>` one of `linux`, `darwin` (macOS) or `windows` and `<arch>` one of `amd64` or `arm64` (`.tar.gz`, or `.zip` on Windows), and a `checksums.txt` with their SHA-256 sums. Download your archive and `checksums.txt` into the same directory, then verify before unpacking:

```bash
# Linux
sha256sum --ignore-missing -c checksums.txt

# macOS
shasum -a 256 --ignore-missing -c checksums.txt
```

Your archive must be reported as `OK`. On Windows, compare the output of `Get-FileHash -Algorithm SHA256 <archive>.zip` in PowerShell with the matching line in `checksums.txt`.

### From source

```bash
git clone https://github.com/custodexa/custodexa-mcp.git
cd custodexa-mcp
go build -o custodexa-mcp .
```

## Configure your host

Add each entry below to what the file already contains, and replace `custodexa.example.com` with the address of your Custodexa.

Hosts that connect directly read the token from the `CUSTODEXA_AGENT_TOKEN` variable in their own environment, so the file holds only a reference to it. Set that variable in the environment you start the host from.

Hosts that start the relay need its **absolute path**. Replace `/usr/local/bin/custodexa-mcp` below with the path where you placed the binary; `go install` puts it in `$(go env GOPATH)/bin` by default.

### `Claude Code`

Connects directly. A project entry goes in `.mcp.json` at the project root:

```json
{
  "mcpServers": {
    "custodexa": {
      "type": "http",
      "url": "https://custodexa.example.com/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

For all your projects, add the same entry to `~/.claude.json` with the CLI. Keep the single quotes, so your shell leaves `${CUSTODEXA_AGENT_TOKEN}` for `Claude Code` to expand:

```bash
claude mcp add-json --scope user custodexa '{"type":"http","url":"https://custodexa.example.com/api/v1/mcp","headers":{"Authorization":"Bearer ${CUSTODEXA_AGENT_TOKEN}"}}'
```

Start a new session to load it. `Claude Code` asks you to approve a server from `.mcp.json` the first time, and `claude mcp get custodexa` shows whether it connected.

If your version cannot connect over HTTP, run the relay with this entry instead:

```json
{
  "mcpServers": {
    "custodexa": {
      "type": "stdio",
      "command": "/usr/local/bin/custodexa-mcp",
      "env": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "${CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

### `Codex`

Connects directly. Use `~/.codex/config.toml` for all projects, or `.codex/config.toml` in a project, which `Codex` reads only in projects you trust.

```toml
[mcp_servers.custodexa]
url = "https://custodexa.example.com/api/v1/mcp"
bearer_token_env_var = "CUSTODEXA_AGENT_TOKEN"
```

`codex mcp add custodexa --url https://custodexa.example.com/api/v1/mcp --bearer-token-env-var CUSTODEXA_AGENT_TOKEN` writes the same entry to `~/.codex/config.toml`. Start a new session to load it; `codex mcp list` shows the configured servers and `/mcp` in a session shows the active ones.

If your version cannot connect over HTTP, run the relay with this entry instead. `env_vars` forwards the token from the environment of `Codex`:

```toml
[mcp_servers.custodexa]
command = "/usr/local/bin/custodexa-mcp"
env = { CUSTODEXA_MCP_URL = "https://custodexa.example.com/api/v1/mcp" }
env_vars = ["CUSTODEXA_AGENT_TOKEN"]
```

### `Gemini CLI`

Connects directly. Use `~/.gemini/settings.json` for all projects, or `.gemini/settings.json` in a project root. The streamable HTTP address goes in `httpUrl`; `url` is for SSE servers.

```json
{
  "mcpServers": {
    "custodexa": {
      "httpUrl": "https://custodexa.example.com/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

Start a new session to load it; `/mcp list` shows the server and its tools.

If your version cannot connect over HTTP, run the relay with this entry instead. List the token under `env`: `Gemini CLI` withholds inherited variables whose names contain `TOKEN` from the servers it starts.

```json
{
  "mcpServers": {
    "custodexa": {
      "command": "/usr/local/bin/custodexa-mcp",
      "env": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "${CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

### `OpenCode`

Connects directly. Use `~/.config/opencode/opencode.json` for all projects, or `opencode.json` in a project root. Custodexa authenticates with the bearer token alone, so `"oauth": false` keeps `OpenCode` from starting an OAuth sign-in when a request is refused.

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "custodexa": {
      "type": "remote",
      "url": "https://custodexa.example.com/api/v1/mcp",
      "oauth": false,
      "headers": {
        "Authorization": "Bearer {env:CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

`OpenCode` 2 also reads this layout; its native layout places the entry under `mcp.servers`. Restart `OpenCode` to load the change; `opencode mcp list` shows each server's connection status.

If your version cannot connect over HTTP, run the relay with this entry instead:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "custodexa": {
      "type": "local",
      "command": ["/usr/local/bin/custodexa-mcp"],
      "environment": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "{env:CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

### `Cursor`

Connects directly. Use `~/.cursor/mcp.json` for all projects, or `.cursor/mcp.json` in a project root.

```json
{
  "mcpServers": {
    "custodexa": {
      "url": "https://custodexa.example.com/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${env:CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

Restart `Cursor` to load the change.

If your version cannot connect over HTTP, run the relay with this entry instead:

```json
{
  "mcpServers": {
    "custodexa": {
      "type": "stdio",
      "command": "/usr/local/bin/custodexa-mcp",
      "env": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "${env:CUSTODEXA_AGENT_TOKEN}"
      }
    }
  }
}
```

### `Claude Desktop`

Runs the relay. Open Settings, then Developer, then Edit Config. The file is `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS and `%APPDATA%\Claude\claude_desktop_config.json` on Windows.

```json
{
  "mcpServers": {
    "custodexa": {
      "command": "/usr/local/bin/custodexa-mcp",
      "env": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "<agent-token>"
      }
    }
  }
}
```

Replace `<agent-token>` with the agent token. Quit and restart `Claude Desktop` to load the change. The relay's error messages go to the host's MCP server log (`~/Library/Logs/Claude/mcp-server-custodexa.log` on macOS).

A custom connector added in the `Claude Desktop` settings connects from the vendor's cloud rather than from your machine, so it reaches only a Custodexa exposed to the public internet.

### Protect the config file

The `Claude Desktop` config holds the token in plain text, so make it readable by you alone:

```bash
chmod 600 ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

If you put the token itself into any other host's file instead of a variable reference, restrict that file the same way and keep it out of version control.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `CUSTODEXA_MCP_URL` | No | `http://localhost:8080/api/v1/mcp` | The Custodexa MCP endpoint. Must be an `http` or `https` URL with a host, and must not contain credentials (`user:pass@`), a query string or a fragment. Redirects are not followed, so give the final URL. |
| `CUSTODEXA_AGENT_TOKEN` | Yes | none | The agent token, sent as `Authorization: Bearer` on every request to the endpoint. Surrounding whitespace is removed, so a blank value counts as missing. |

If either value is rejected, or the endpoint cannot be reached at startup, the relay writes the reason to stderr and exits with status 1.

## Handling the token

- The relay reads the token only from its environment. Never put it in `args` or in a shell wrapper such as `sh -c "CUSTODEXA_AGENT_TOKEN=... custodexa-mcp"`: command lines are visible to other users in the process list.
- Use `https://` whenever Custodexa runs on another machine. Plain `http://` is for a server on your own machine during development.

## Tools

The relay offers whatever tools the endpoint lists when it starts, with their schemas unchanged, and returns results as it receives them. Custodexa currently lists ten:

| Tool | What it does |
|---|---|
| `list_assets` | Lists the assets the agent can see, their accounts, and whether each can be connected now. Asset ids come only from here. |
| `request_access` | Asks for access to named assets and accounts, with a reason and a duration, and returns a `request_id`. |
| `check_request` | Reports a request's status, approvals received and required, expiry, and when the request closed. |
| `open_session` | Opens a recorded SSH, Kubernetes or database session under an approved request and returns a `session_handle`. |
| `close_session` | Closes a session and finalizes its recording and command audit. |
| `run_command` | Sends a command to a terminal session. A `completed` status means the prompt appeared again, not that the command succeeded. |
| `send_keys` | Sends `ctrl-c`, `ctrl-d` or `enter` to a session. |
| `query` | Runs SQL in a database session. |
| `read_screen` | Returns the last lines of a session's output with control sequences removed. It is a line buffer, not a screen snapshot. |
| `close_task` | Closes the agent's task with a report. Calling it again within 24 hours adds a new report version. |

The input schema of each tool is available through `tools/list`. The full contract, including statuses, error codes and operational limits, is in the [MCP agent channel section of the Custodexa API specification](https://github.com/custodexa/custodexa/blob/main/docs/API_SPEC.md).

On `initialize`, the relay reports `serverInfo.name` as `custodexa-mcp` and `serverInfo.version` as its own version, such as `1.0.0`. Neither value describes the Custodexa server behind it.

### Typical flow

An agent usually works through a task in this order:

1. Call `list_assets` to find the asset and the account to use.
2. Call `request_access` with the asset ids, naming the accounts for each asset, plus a reason and a duration, to get the `request_id` that `open_session` requires. This applies even to an asset that `list_assets` shows as connectable. If a task the agent already holds still covers every account named for every requested asset, the response is `already_connectable` with that task's `request_id`, and no new request is created.
3. While the request is `pending`, use `check_request` to decide: keep waiting while it has not expired and approvals are coming in, do other work and check again later while approvals are still short, and stop waiting once it is `rejected`, `cancelled` or `expired`. After a rejection, do not send the same request again. Revocation applies to individual assets and does not change the status, which stays `approved`; `closed_at` is set once none of the request's assets can still be used.
4. Call `open_session` with the `asset_id` and `account_id` from `list_assets` and the `request_id` to get a `session_handle`. The `request_id` is required; Custodexa never picks one for you.
5. Send commands with `run_command` in a terminal session, or SQL with `query` in a database session. After `timed_out` the result is unknown, so do not resend the same command blindly. If `run_command` returns `needs_input`, send `ctrl-c` with `send_keys` to reset the session before the next command. To see recent output, `read_screen` returns the last lines.
6. Call `close_session` when the work in that session is done.
7. Finish with `close_task`, passing the `request_id` and a report.

## What Custodexa enforces

These rules live on the Custodexa server and apply the same way whether an agent connects directly or through this relay:

- Self-service creation of agents ships turned off.
- An agent token must have an expiry set when it is created, and its plain text is shown only once, at creation.
- An agent can connect only to assets granted by a `request_access` request that has been approved. Whether a person has to approve it depends on the access policy of each asset.
- Every tool call that can be tied to the agent's own task or session is recorded in the tool-call ledger with its arguments and decision, refusals included. `list_assets`, `request_access` and `check_request` need no task. Any other call that cannot be tied to one is refused before it reaches the ledger and is recorded only in the HTTP request audit.
- Masking applies to command output, screen reads, query results and task reports, and replaces only the parts that match an enabled output rule. The shipped rules cover card numbers and private key headers. Session recordings keep the original screen.

## License

`custodexa-mcp` is licensed under the GNU Affero General Public License v3.0; see [LICENSE](LICENSE), [NOTICE](NOTICE) and [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md).

Contributions are welcome under the [DCO](DCO.md); see [CONTRIBUTING.md](CONTRIBUTING.md). Report vulnerabilities as described in [SECURITY.md](SECURITY.md). Release notes are in [CHANGELOG.md](CHANGELOG.md).
