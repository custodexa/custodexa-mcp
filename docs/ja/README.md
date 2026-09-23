# custodexa-mcp

<p><a href="../../README.md">English</a> | <a href="../zh-TW/README.md">繁體中文</a> | <b>日本語</b></p>

`custodexa-mcp` は、ローカルの stdio サーバーしか起動できない MCP ホストを Custodexa につなぐためのプログラムです。MCP サーバー本体は Custodexa の一部で、お使いのデプロイの `POST /api/v1/mcp` エンドポイントがそれにあたります。認可、接続、監査、録画、マスクはすべてそこで行われます。本プログラムは、ホストに対しては stdio サーバーとして、そのエンドポイントに対しては HTTP クライアントとして動作します。すべてのリクエストをエージェントトークン付きで転送し、自身では何も判断しません。streamable HTTP サーバーに対応したホストは Custodexa へ直接接続できるため、本プログラムは必要ありません。

## 必要かどうか

| お使いの MCP ホスト | 対応 |
|---|---|
| ローカルサーバーを stdio でしか起動できない | `custodexa-mcp` をインストールし、ホストの接続先に指定する |
| カスタムの `Authorization` ヘッダーを付けたリモート streamable HTTP サーバーに対応している | `https://<your-custodexa>/api/v1/mcp` へ直接接続する |

直接接続では、エージェントトークンを bearer トークンとして送ります。ホストごとの設定は [ホストの設定](#ホストの設定) にあります。

## インストール

### Go でインストール

Go 1.25 以降が必要です。

```bash
go install github.com/custodexa/custodexa-mcp@latest
```

バイナリは `$(go env GOPATH)/bin` に置かれます。`GOBIN` を設定している場合はそちらに置かれます。

### GitHub Release から

各 [リリース](https://github.com/custodexa/custodexa-mcp/releases) には、プラットフォームごとのアーカイブと、それらの SHA-256 値を収めた `checksums.txt` が含まれます。アーカイブ名は `custodexa-mcp_<version>_<os>_<arch>` で、`<os>` は `linux`、`darwin`（macOS）、`windows` のいずれか、`<arch>` は `amd64` または `arm64` です（形式は `.tar.gz`、Windows では `.zip`）。お使いのアーカイブと `checksums.txt` を同じディレクトリにダウンロードし、展開する前に検証してください。

```bash
# Linux
sha256sum --ignore-missing -c checksums.txt

# macOS
shasum -a 256 --ignore-missing -c checksums.txt
```

お使いのアーカイブについて `OK` と表示されることを確認してください。Windows では、PowerShell で `Get-FileHash -Algorithm SHA256 <archive>.zip` を実行し、その出力を `checksums.txt` の該当する行と照合してください。

### ソースから

```bash
git clone https://github.com/custodexa/custodexa-mcp.git
cd custodexa-mcp
go build -o custodexa-mcp .
```

## ホストの設定

以下の各設定は、ファイルの既存の内容に追記してください。また、`custodexa.example.com` はお使いの Custodexa のアドレスに置き換えてください。

直接接続するホストは、自身の環境にある `CUSTODEXA_AGENT_TOKEN` 変数からトークンを読み取ります。そのため、ファイルに書くのはこの変数への参照だけです。この変数は、ホストを起動する環境で設定してください。

リレーを起動するホストには、リレーの**絶対パス**が必要です。以下の `/usr/local/bin/custodexa-mcp` は、バイナリを置いたパスに置き換えてください。`go install` を使った場合の既定の配置先は `$(go env GOPATH)/bin` です。

### `Claude Code`

直接接続します。プロジェクト単位の設定は、プロジェクトルートの `.mcp.json` に書きます。

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

すべてのプロジェクトで使う場合は、CLI で同じ設定を `~/.claude.json` に追加します。シェルが `${CUSTODEXA_AGENT_TOKEN}` を展開せず、`Claude Code` に展開させるため、シングルクォートはそのまま残してください。

```bash
claude mcp add-json --scope user custodexa '{"type":"http","url":"https://custodexa.example.com/api/v1/mcp","headers":{"Authorization":"Bearer ${CUSTODEXA_AGENT_TOKEN}"}}'
```

新しいセッションを開始すると読み込まれます。`.mcp.json` にあるサーバーについては、初回に `Claude Code` が承認を求めます。接続できたかどうかは `claude mcp get custodexa` で確認できます。

お使いのバージョンが HTTP で接続できない場合は、代わりに次の設定でリレーを実行してください。

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

直接接続します。すべてのプロジェクトで使う場合は `~/.codex/config.toml` を、プロジェクト単位ではプロジェクト内の `.codex/config.toml` を使います。後者は、信頼済みのプロジェクトでのみ `Codex` が読み込みます。

```toml
[mcp_servers.custodexa]
url = "https://custodexa.example.com/api/v1/mcp"
bearer_token_env_var = "CUSTODEXA_AGENT_TOKEN"
```

`codex mcp add custodexa --url https://custodexa.example.com/api/v1/mcp --bearer-token-env-var CUSTODEXA_AGENT_TOKEN` を実行すると、同じ設定が `~/.codex/config.toml` に書き込まれます。新しいセッションを開始すると読み込まれます。設定済みのサーバーは `codex mcp list` で、稼働中のサーバーはセッション内の `/mcp` で確認できます。

お使いのバージョンが HTTP で接続できない場合は、代わりに次の設定でリレーを実行してください。`env_vars` は、`Codex` の環境にあるトークンをリレーへ引き渡します。

```toml
[mcp_servers.custodexa]
command = "/usr/local/bin/custodexa-mcp"
env = { CUSTODEXA_MCP_URL = "https://custodexa.example.com/api/v1/mcp" }
env_vars = ["CUSTODEXA_AGENT_TOKEN"]
```

### `Gemini CLI`

直接接続します。すべてのプロジェクトで使う場合は `~/.gemini/settings.json` を、プロジェクト単位ではプロジェクトルートの `.gemini/settings.json` を使います。streamable HTTP のアドレスは `httpUrl` に書きます。`url` は SSE サーバー用です。

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

新しいセッションを開始すると読み込まれます。`/mcp list` でサーバーとそのツールを確認できます。

お使いのバージョンが HTTP で接続できない場合は、代わりに次の設定でリレーを実行してください。トークンは `env` に記載してください。`Gemini CLI` は、名前に `TOKEN` を含む変数を、自身が起動するサーバーに引き継がないためです。

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

直接接続します。すべてのプロジェクトで使う場合は `~/.config/opencode/opencode.json` を、プロジェクト単位ではプロジェクトルートの `opencode.json` を使います。Custodexa は bearer トークンだけで認証するため、`"oauth": false` を指定して、リクエストが拒否されたときに `OpenCode` が OAuth のサインインを始めないようにします。

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

`OpenCode` 2 もこの構成を読み込みます。ただし、同バージョン本来の構成では、設定を `mcp.servers` の下に置きます。変更を読み込むには `OpenCode` を再起動してください。各サーバーの接続状態は `opencode mcp list` で確認できます。

お使いのバージョンが HTTP で接続できない場合は、代わりに次の設定でリレーを実行してください。

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

直接接続します。すべてのプロジェクトで使う場合は `~/.cursor/mcp.json` を、プロジェクト単位ではプロジェクトルートの `.cursor/mcp.json` を使います。

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

変更を読み込むには `Cursor` を再起動してください。

お使いのバージョンが HTTP で接続できない場合は、代わりに次の設定でリレーを実行してください。

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

リレーを実行します。Settings、Developer、Edit Config の順に開いてください。ファイルの場所は、macOS では `~/Library/Application Support/Claude/claude_desktop_config.json`、Windows では `%APPDATA%\Claude\claude_desktop_config.json` です。

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

`<agent-token>` はエージェントトークンに置き換えてください。変更を読み込むには、`Claude Desktop` をいったん終了してから起動し直してください。リレーのエラーメッセージは、ホストの MCP サーバーログ（macOS では `~/Library/Logs/Claude/mcp-server-custodexa.log`）に出力されます。

`Claude Desktop` の設定で追加したカスタムコネクタは、お使いのマシンではなくベンダーのクラウドから接続します。そのため、到達できるのはインターネットに公開された Custodexa だけです。

### 設定ファイルの保護

`Claude Desktop` の設定ファイルにはトークンが平文で保存されるため、自分だけが読めるようにしてください。

```bash
chmod 600 ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

ほかのホストのファイルに変数参照ではなくトークンそのものを書いた場合も、同じようにアクセスを制限し、バージョン管理の対象から外してください。

## 環境変数

| 変数 | 必須 | 既定値 | 備考 |
|---|---|---|---|
| `CUSTODEXA_MCP_URL` | いいえ | `http://localhost:8080/api/v1/mcp` | Custodexa の MCP エンドポイントです。ホストを含む `http` または `https` の URL でなければならず、資格情報（`user:pass@`）、クエリ文字列、フラグメントは含められません。リダイレクトは追跡しないため、最終的な URL を指定してください。 |
| `CUSTODEXA_AGENT_TOKEN` | はい | なし | エージェントトークンです。エンドポイントへのすべてのリクエストで `Authorization: Bearer` として送信されます。前後の空白は取り除かれるため、空白だけの値は未設定として扱われます。 |

いずれかの値が受け付けられない場合、または起動時にエンドポイントへ到達できない場合、リレーは理由を stderr に書き出し、終了ステータス 1 で終了します。

## トークンの扱い

- リレーはトークンを自身の環境変数からのみ読み取ります。`args` や、`sh -c "CUSTODEXA_AGENT_TOKEN=... custodexa-mcp"` のようなシェルラッパーには決して書かないでください。コマンドラインはプロセス一覧を通じてほかのユーザーからも見えます。
- Custodexa が別のマシンで動作している場合は、必ず `https://` を使ってください。平文の `http://` は、開発中に自分のマシンで動かすサーバー向けです。

## ツール

リレーは、自身の起動時にエンドポイントが一覧として返したツールを、スキーマを変えずにそのまま提供し、結果も受け取ったとおりに返します。現在、Custodexa が一覧に載せているツールは次の 10 個です。

| ツール | 機能 |
|---|---|
| `list_assets` | エージェントが参照できる資産、そのアカウント、それぞれに現在接続できるかどうかを一覧にします。資産 ID はここからしか取得できません。 |
| `request_access` | 指定した資産とアカウントへのアクセスを、理由と期間を添えて申請し、`request_id` を返します。 |
| `check_request` | 申請の状態、得られた承認の数と必要な承認の数、有効期限、申請が閉じた日時を返します。 |
| `open_session` | 承認済みの申請のもとで、録画される SSH、Kubernetes、またはデータベースのセッションを開き、`session_handle` を返します。 |
| `close_session` | セッションを閉じ、その録画とコマンド監査を確定させます。 |
| `run_command` | 端末セッションにコマンドを送ります。ステータス `completed` はプロンプトが再び表示されたことを示すだけで、コマンドが成功したことを示すものではありません。 |
| `send_keys` | `ctrl-c`、`ctrl-d`、`enter` のいずれかをセッションに送ります。 |
| `query` | データベースセッションで SQL を実行します。 |
| `read_screen` | セッション出力の末尾の行を、制御シーケンスを取り除いて返します。これは行バッファーであり、画面のスナップショットではありません。 |
| `close_task` | 報告を添えてエージェントのタスクを閉じます。24 時間以内に再度呼び出すと、報告の新しい版が追加されます。 |

各ツールの入力スキーマは `tools/list` で取得できます。ステータス、エラーコード、運用上の制限を含む完全な仕様は、[Custodexa API 仕様書の MCP エージェントチャネルの節](https://github.com/custodexa/custodexa/blob/main/docs/API_SPEC.md) にあります。

`initialize` に対して、リレーは `serverInfo.name` に `custodexa-mcp` を、`serverInfo.version` に `1.0.0` のような自身のバージョンを返します。どちらの値も、背後にある Custodexa サーバーを表すものではありません。

### 典型的な流れ

エージェントがタスクを進めるときは、通常次の順序になります。

1. `list_assets` を呼び出し、使う資産とアカウントを見つけます。
2. 資産 ID を指定し、資産ごとにアカウントを明示したうえで、理由と期間を添えて `request_access` を呼び出し、`open_session` に必要な `request_id` を取得します。`list_assets` で接続できると示された資産でも同じです。エージェントがすでに持っているタスクが、申請する各資産に指定したすべてのアカウントをまだカバーしている場合、応答は `already_connectable` となってそのタスクの `request_id` が添えられ、新しい申請は作られません。
3. 申請が `pending` の間は、`check_request` で判断します。期限前で承認が進んでいれば待ち続け、承認がまだ足りなければ別の作業をして後で確認し直し、`rejected`、`cancelled`、`expired` のいずれかになったら待つのをやめます。拒否された後は、同じ申請を再送しないでください。取り消しは資産ごとに行われ、状態は `approved` のまま変わりません。申請内のどの資産も使えなくなった時点で `closed_at` が設定されます。
4. `list_assets` の `asset_id` と `account_id`、および `request_id` を指定して `open_session` を呼び出し、`session_handle` を取得します。`request_id` は必須で、Custodexa が代わりに選ぶことはありません。
5. 端末セッションでは `run_command` でコマンドを、データベースセッションでは `query` で SQL を送ります。`timed_out` の場合は結果が不明なので、同じコマンドをむやみに再送しないでください。`run_command` が `needs_input` を返したら、`send_keys` で `ctrl-c` を送ってセッションをリセットしてから、次のコマンドを送ります。直近の出力を見るには、`read_screen` で末尾の行を取得します。
6. そのセッションでの作業が終わったら、`close_session` を呼び出します。
7. 最後に `request_id` と報告を指定して `close_task` を呼び出します。

## Custodexa が強制すること

以下のルールは Custodexa サーバー側にあり、エージェントが直接接続する場合も、このリレーを経由する場合も同じように適用されます。

- エージェントをセルフサービスで作成する機能は、出荷時には無効になっています。
- エージェントトークンは作成時に有効期限を設定する必要があり、その平文は作成時に一度だけ表示されます。
- エージェントが接続できるのは、承認済みの `request_access` 申請で許可された資産だけです。人による承認が必要かどうかは、各資産のアクセスポリシーによって決まります。
- エージェント自身のタスクまたはセッションに結び付けられるツール呼び出しは、拒否されたものも含め、引数と判定結果とともにツール呼び出し台帳に記録されます。`list_assets`、`request_access`、`check_request` にはタスクが必要ありません。それ以外の呼び出しでタスクやセッションに結び付けられないものは、台帳に届く前に拒否され、HTTP リクエストの監査にのみ記録されます。
- マスクは、コマンド出力、画面の読み取り、クエリ結果、タスク報告に適用され、有効な出力ルールに一致した部分だけを置き換えます。出荷時のルールが対象とするのは、カード番号と秘密鍵ヘッダーです。セッションの録画には元の画面がそのまま残ります。

## ライセンス

`custodexa-mcp` は GNU Affero General Public License v3.0 のもとでライセンスされています。[LICENSE](../../LICENSE)、[NOTICE](../../NOTICE)、[THIRD-PARTY-LICENSES.md](../../THIRD-PARTY-LICENSES.md) を参照してください。

コントリビューションは [DCO](../../DCO.md) のもとで歓迎します。詳しくは [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。脆弱性は [SECURITY.md](SECURITY.md) に記載の方法で報告してください。リリースノートは [CHANGELOG.md](../../CHANGELOG.md) にあります。
