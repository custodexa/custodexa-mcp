# custodexa-mcp

`custodexa-mcp` lets MCP hosts that can only launch local stdio servers reach Custodexa. The MCP server itself is part of Custodexa: the `POST /api/v1/mcp` endpoint of your deployment, where authorization, connections, auditing, recording and masking take place. This program is a stdio server to the host and an HTTP client to that endpoint. It forwards every request with an agent token and makes no decisions of its own. Hosts that support streamable HTTP servers can connect to Custodexa directly and do not need it.

## Do you need it?

| Your MCP host | What to do |
|---|---|
| Launches local servers over stdio only | Install `custodexa-mcp` and point the host at it |
| Supports remote streamable HTTP servers with a custom `Authorization` header | Connect directly to `https://<your-custodexa>/api/v1/mcp` |

A direct connection sends the agent token as a bearer token. In `Claude Code`, for example, a project `.mcp.json` can read the token from your environment, so it stays out of the file:

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

Hosts start the relay themselves, so give them the binary's **absolute path** and pass both variables through the server's `env` block. Replace `/usr/local/bin/custodexa-mcp` below with the absolute path where you placed the binary; `go install` puts it in `$(go env GOPATH)/bin` by default. Replace the URL and token with your own as well.

### `Claude Desktop`

Open Settings, then Developer, then Edit Config. The file is `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS and `%APPDATA%\Claude\claude_desktop_config.json` on Windows.

```json
{
  "mcpServers": {
    "custodexa": {
      "command": "/usr/local/bin/custodexa-mcp",
      "env": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "cxa_..."
      }
    }
  }
}
```

Quit and restart `Claude Desktop` to load the change. The relay's error messages go to the host's MCP server log (`~/Library/Logs/Claude/mcp-server-custodexa.log` on macOS).

### `Cursor`

Use `~/.cursor/mcp.json` for all projects, or `.cursor/mcp.json` in a project root.

```json
{
  "mcpServers": {
    "custodexa": {
      "type": "stdio",
      "command": "/usr/local/bin/custodexa-mcp",
      "env": {
        "CUSTODEXA_MCP_URL": "https://custodexa.example.com/api/v1/mcp",
        "CUSTODEXA_AGENT_TOKEN": "cxa_..."
      }
    }
  }
}
```

`Cursor` can fill the token from its own environment with `"CUSTODEXA_AGENT_TOKEN": "${env:CUSTODEXA_AGENT_TOKEN}"`. `Cursor` also accepts remote servers through `url` and `headers`, so it can connect to Custodexa directly.

### Protect the config file

A config file holding the token in plain text should be readable by you alone:

```bash
chmod 600 ~/Library/Application\ Support/Claude/claude_desktop_config.json
chmod 600 ~/.cursor/mcp.json
```

Keep a project-level file that holds a token out of version control.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `CUSTODEXA_MCP_URL` | No | `http://localhost:8080/api/v1/mcp` | The Custodexa MCP endpoint. Must be an `http` or `https` URL with a host, and must not contain credentials (`user:pass@`), a query string or a fragment. Redirects are not followed, so give the final URL. |
| `CUSTODEXA_AGENT_TOKEN` | Yes | none | The agent token (`cxa_...`), sent as `Authorization: Bearer` on every request to the endpoint. Surrounding whitespace is removed, so a blank value counts as missing. |

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
| `check_request` | Reports a request's status, approvals received and required, and expiry. |
| `open_session` | Opens a recorded SSH, Kubernetes or database session under an approved request and returns a `session_handle`. |
| `close_session` | Closes a session and finalizes its recording and command audit. |
| `run_command` | Sends a command to a terminal session. A `completed` status means the prompt appeared again, not that the command succeeded. |
| `send_keys` | Sends `ctrl-c`, `ctrl-d` or `enter` to a session. |
| `query` | Runs SQL in a database session. |
| `read_screen` | Returns the last lines of a session's output with control sequences removed. It is a line buffer, not a screen snapshot. |
| `close_task` | Closes the agent's task with a report. Calling it again within 24 hours adds a new report version. |

The input schema of each tool is available through `tools/list`. The full contract, including statuses, error codes and operational limits, is in the [MCP agent channel section of the Custodexa API specification](https://github.com/custodexa/custodexa/blob/main/docs/API_SPEC.md).

On `initialize`, the relay reports `serverInfo.name` as `custodexa-mcp` and `serverInfo.version` as its own version, such as `1.0.0`. Neither value describes the Custodexa server behind it.

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
