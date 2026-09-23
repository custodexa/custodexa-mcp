# custodexa-mcp

<p><a href="../../README.md">English</a> | <b>繁體中文</b> | <a href="../ja/README.md">日本語</a></p>

`custodexa-mcp` 讓只能啟動本機 stdio 服務的 MCP 宿主連上 Custodexa。MCP 服務端本身屬於 Custodexa，也就是你部署中的 `POST /api/v1/mcp` 端點，授權、連線、稽核、錄影與遮罩都在那裡進行。本程式對宿主而言是 stdio 服務端，對該端點而言是 HTTP 客戶端。它帶著 agent token 轉送每一個請求，自己不做任何判定。支援 streamable HTTP 服務端的宿主可以直接連到 Custodexa，不需要它。

## 你需要它嗎？

| 你的 MCP 宿主 | 該怎麼做 |
|---|---|
| 只能以 stdio 啟動本機服務 | 安裝 `custodexa-mcp`，並讓宿主指向它 |
| 支援遠端 streamable HTTP 服務端，且可自訂 `Authorization` 標頭 | 直接連到 `https://<your-custodexa>/api/v1/mcp` |

直連時，agent token 以 bearer token 送出。各宿主的設定項目見[設定宿主](#設定宿主)。

## 安裝

### 使用 Go

需要 Go 1.25 以上版本。

```bash
go install github.com/custodexa/custodexa-mcp@latest
```

二進位檔會放在 `$(go env GOPATH)/bin`；若你設了 `GOBIN`，則放在該目錄。

### 從 GitHub Release 下載

每個 [release](https://github.com/custodexa/custodexa-mcp/releases) 都為各平台附一個壓縮檔，命名為 `custodexa-mcp_<version>_<os>_<arch>`，其中 `<os>` 為 `linux`、`darwin`（macOS）或 `windows`，`<arch>` 為 `amd64` 或 `arm64`（格式為 `.tar.gz`，Windows 為 `.zip`），另附記載這些檔案 SHA-256 雜湊值的 `checksums.txt`。把你的壓縮檔與 `checksums.txt` 下載到同一個目錄，解壓縮前先驗證：

```bash
# Linux
sha256sum --ignore-missing -c checksums.txt

# macOS
shasum -a 256 --ignore-missing -c checksums.txt
```

你的壓縮檔必須顯示為 `OK`。在 Windows 上，請在 PowerShell 執行 `Get-FileHash -Algorithm SHA256 <archive>.zip`，將輸出與 `checksums.txt` 中對應的那一行比對。

### 從原始碼建置

```bash
git clone https://github.com/custodexa/custodexa-mcp.git
cd custodexa-mcp
go build -o custodexa-mcp .
```

## 設定宿主

把下列各項目加進檔案既有的內容中，並把 `custodexa.example.com` 換成你的 Custodexa 位址。

直連的宿主從自身環境的 `CUSTODEXA_AGENT_TOKEN` 變數讀取 token，因此檔案裡只放對該變數的參照。請在啟動宿主的環境中設定這個變數。

啟動轉接頭的宿主需要轉接頭的**絕對路徑**。請把下方的 `/usr/local/bin/custodexa-mcp` 換成你放置二進位檔的路徑；`go install` 預設放在 `$(go env GOPATH)/bin`。

### `Claude Code`

直連。專案層級的項目放在專案根目錄的 `.mcp.json`：

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

要套用到你所有的專案，請用 CLI 把同一個項目加進 `~/.claude.json`。請保留單引號，讓 shell 不展開 `${CUSTODEXA_AGENT_TOKEN}`，交給 `Claude Code` 展開：

```bash
claude mcp add-json --scope user custodexa '{"type":"http","url":"https://custodexa.example.com/api/v1/mcp","headers":{"Authorization":"Bearer ${CUSTODEXA_AGENT_TOKEN}"}}'
```

開新的工作階段即可載入。`Claude Code` 第一次遇到 `.mcp.json` 裡的服務端時會請你核准，`claude mcp get custodexa` 可查看是否已連上。

若你的版本無法透過 HTTP 連線，改用以下項目執行轉接頭：

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

直連。所有專案共用的設定放在 `~/.codex/config.toml`，單一專案則放在專案內的 `.codex/config.toml`，`Codex` 只在你信任的專案中讀取後者。

```toml
[mcp_servers.custodexa]
url = "https://custodexa.example.com/api/v1/mcp"
bearer_token_env_var = "CUSTODEXA_AGENT_TOKEN"
```

`codex mcp add custodexa --url https://custodexa.example.com/api/v1/mcp --bearer-token-env-var CUSTODEXA_AGENT_TOKEN` 會把同一個項目寫進 `~/.codex/config.toml`。開新的工作階段即可載入；`codex mcp list` 列出已設定的服務端，工作階段中的 `/mcp` 列出目前啟用的服務端。

若你的版本無法透過 HTTP 連線，改用以下項目執行轉接頭。`env_vars` 會轉交 `Codex` 環境中的 token：

```toml
[mcp_servers.custodexa]
command = "/usr/local/bin/custodexa-mcp"
env = { CUSTODEXA_MCP_URL = "https://custodexa.example.com/api/v1/mcp" }
env_vars = ["CUSTODEXA_AGENT_TOKEN"]
```

### `Gemini CLI`

直連。所有專案共用的設定放在 `~/.gemini/settings.json`，單一專案則放在專案根目錄的 `.gemini/settings.json`。streamable HTTP 位址填在 `httpUrl`；`url` 是給 SSE 服務端用的。

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

開新的工作階段即可載入；`/mcp list` 會列出此服務端與它的工具。

若你的版本無法透過 HTTP 連線，改用以下項目執行轉接頭。請把 token 列在 `env` 下：`Gemini CLI` 不會把名稱含 `TOKEN` 的繼承變數交給它啟動的服務端。

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

直連。所有專案共用的設定放在 `~/.config/opencode/opencode.json`，單一專案則放在專案根目錄的 `opencode.json`。Custodexa 只以 bearer token 認證，因此設定 `"oauth": false`，避免 `OpenCode` 在請求被拒時啟動 OAuth 登入。

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

`OpenCode` 2 也能讀取這種結構；它的原生結構把項目放在 `mcp.servers` 之下。重新啟動 `OpenCode` 即可載入變更；`opencode mcp list` 會顯示每個服務端的連線狀態。

若你的版本無法透過 HTTP 連線，改用以下項目執行轉接頭：

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

直連。所有專案共用的設定放在 `~/.cursor/mcp.json`，單一專案則放在專案根目錄的 `.cursor/mcp.json`。

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

重新啟動 `Cursor` 即可載入變更。

若你的版本無法透過 HTTP 連線，改用以下項目執行轉接頭：

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

執行轉接頭。依序開啟 Settings、Developer、Edit Config。設定檔在 macOS 上是 `~/Library/Application Support/Claude/claude_desktop_config.json`，在 Windows 上是 `%APPDATA%\Claude\claude_desktop_config.json`。

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

把 `<agent-token>` 換成 agent token。完全結束 `Claude Desktop` 再重新啟動，變更才會載入。轉接頭的錯誤訊息會寫進宿主的 MCP 服務端日誌（macOS 上為 `~/Library/Logs/Claude/mcp-server-custodexa.log`）。

在 `Claude Desktop` 設定中加入的自訂連接器是從廠商的雲端連線，而不是從你的電腦，因此只連得到公開在網際網路上的 Custodexa。

### 保護設定檔

`Claude Desktop` 的設定檔以明文保存 token，請設為只有你能讀取：

```bash
chmod 600 ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

若你在其他宿主的檔案裡直接寫入 token 本身，而不是變數參照，也請以同樣方式限制該檔案的權限，並且不要納入版本控制。

## 環境變數

| 變數 | 必填 | 預設值 | 說明 |
|---|---|---|---|
| `CUSTODEXA_MCP_URL` | 否 | `http://localhost:8080/api/v1/mcp` | Custodexa 的 MCP 端點。必須是帶有主機的 `http` 或 `https` URL，且不得包含憑證（`user:pass@`）、查詢字串或片段。不會跟隨重新導向，請填最終的 URL。 |
| `CUSTODEXA_AGENT_TOKEN` | 是 | 無 | agent token，每次向端點發出請求時以 `Authorization: Bearer` 送出。前後空白會被去除，因此只有空白的值視同未設定。 |

只要任一值被拒，或啟動時連不到端點，轉接頭就會把原因寫到 stderr，並以狀態碼 1 結束。

## 處理 token

- 轉接頭只從自身環境讀取 token。切勿把它放在 `args`，或放進 `sh -c "CUSTODEXA_AGENT_TOKEN=... custodexa-mcp"` 這類 shell 包裝指令：其他使用者能在行程清單中看到命令列。
- 只要 Custodexa 在另一台機器上，就一律使用 `https://`。純 `http://` 僅供開發期間連到你自己機器上的服務端。

## 工具

轉接頭提供端點在它啟動時列出的所有工具，schema 原樣不變，收到的結果也原樣回傳。Custodexa 目前列出十項：

| 工具 | 功能 |
|---|---|
| `list_assets` | 列出 agent 看得到的資產與其帳號，以及各資產目前能否連線。資產 id 只能從這裡取得。 |
| `request_access` | 申請存取指定的資產與帳號，附上理由與時長，並回傳 `request_id`。 |
| `check_request` | 回報申請的狀態、已取得與所需的核准數、到期時間，以及申請關閉的時間。 |
| `open_session` | 在已核准的申請下開啟有錄影的 SSH、Kubernetes 或資料庫會話，並回傳 `session_handle`。 |
| `close_session` | 關閉會話，並完成該會話的錄影與指令稽核紀錄。 |
| `run_command` | 將指令送到終端會話。`completed` 狀態表示提示字元再次出現，不代表指令執行成功。 |
| `send_keys` | 對會話送出 `ctrl-c`、`ctrl-d` 或 `enter`。 |
| `query` | 在資料庫會話中執行 SQL。 |
| `read_screen` | 回傳會話輸出的最後幾行，並去除控制序列。這是行緩衝區，不是畫面快照。 |
| `close_task` | 以一份報告結束 agent 的任務。24 小時內再次呼叫會新增一個報告版本。 |

每項工具的輸入 schema 可透過 `tools/list` 取得。完整契約，包括狀態、錯誤碼與運作限制，見 [Custodexa API 規格的 MCP agent 通道章節](https://github.com/custodexa/custodexa/blob/main/docs/API_SPEC.md)。

收到 `initialize` 時，轉接頭回報的 `serverInfo.name` 為 `custodexa-mcp`，`serverInfo.version` 為轉接頭自身的版本，例如 `1.0.0`。這兩個值都不代表它背後的 Custodexa 服務端。

### 典型流程

agent 處理一項任務時，通常依下列順序進行：

1. 呼叫 `list_assets`，找出要用的資產與帳號。
2. 呼叫 `request_access`，帶入資產 id，逐一為每項資產指定帳號，並附上理由與時長，取得 `open_session` 必需的 `request_id`。即使 `list_assets` 顯示某資產可以連線，也要這樣做。若 agent 手上已有的某項任務仍涵蓋這次每項資產所指定的全部帳號，回應為 `already_connectable` 並附上該任務的 `request_id`，不會另建申請。
3. 申請為 `pending` 時，以 `check_request` 判斷：尚未到期且核准有進展就繼續等，核准仍不足額就先做別的事、稍後再查，狀態為 `rejected`、`cancelled` 或 `expired` 時停止等待。申請被拒絕後，不要重送相同的申請。撤銷是針對個別資產，不會改變狀態，申請仍顯示 `approved`；等到申請內的資產都不能再使用，`closed_at` 才會有值。
4. 呼叫 `open_session`，帶入 `list_assets` 的 `asset_id` 與 `account_id`，以及 `request_id`，取得 `session_handle`。`request_id` 必填，Custodexa 不會代為選定。
5. 在終端會話以 `run_command` 送出指令，在資料庫會話以 `query` 執行 SQL。`timed_out` 表示結果未知，不要盲目重送同一條指令。`run_command` 回傳 `needs_input` 時，先以 `send_keys` 送出 `ctrl-c` 讓會話復位，再送下一條指令。要看最近的輸出，`read_screen` 會回傳最後幾行。
6. 會話內的工作完成後，呼叫 `close_session`。
7. 最後呼叫 `close_task`，帶入 `request_id` 與報告。

## Custodexa 強制執行的規則

這些規則位於 Custodexa 服務端，無論 agent 是直連還是透過本轉接頭連線，都同樣適用：

- agent 的自助建立出廠為關閉。
- agent token 建立時必須設定到期時間，其明文只在建立當下顯示一次。
- agent 只能連線到經 `request_access` 申請且已核准的資產。是否需要人員核准，取決於各資產的存取政策。
- 凡能歸屬到 agent 自身任務或會話的工具呼叫，都會連同參數與判定結果記入工具呼叫帳本，被拒絕的呼叫也在內。`list_assets`、`request_access` 與 `check_request` 不需要任務。其他無法歸屬的呼叫會在進入帳本前被拒絕，只記錄在 HTTP 請求稽核中。
- 遮罩套用於指令輸出、畫面讀取、查詢結果與任務報告，且只取代符合已啟用輸出規則的部分。出廠規則涵蓋卡號與私鑰標頭。會話錄影保留原始畫面。

## 授權

`custodexa-mcp` 以 GNU Affero General Public License v3.0 授權，見 [LICENSE](../../LICENSE)、[NOTICE](../../NOTICE) 與 [THIRD-PARTY-LICENSES.md](../../THIRD-PARTY-LICENSES.md)。

歡迎依 [DCO](../../DCO.md) 貢獻，詳見 [CONTRIBUTING.md](CONTRIBUTING.md)。回報漏洞的方式見 [SECURITY.md](SECURITY.md)。版本說明在 [CHANGELOG.md](../../CHANGELOG.md)。
