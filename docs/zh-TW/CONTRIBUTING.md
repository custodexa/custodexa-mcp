# 貢獻指南

<p><a href="../../CONTRIBUTING.md">English</a> | <b>繁體中文</b> | <a href="../ja/CONTRIBUTING.md">日本語</a></p>

感謝你為 `custodexa-mcp` 貢獻。本 repo 只包含 stdio 轉接頭。授權、稽核、遮罩以及工具本身都屬於 Custodexa 服務端，這些部分的修改請送到 [Custodexa 主 repo](https://github.com/custodexa/custodexa)。

## 開始之前

- 回報 bug 或提想法：開 issue，附上重現步驟或你的動機。
- 文件、錯字或明顯的小 bug：直接開 pull request。
- 行為變更：先開 issue，對做法取得共識後再寫程式碼。
- 安全問題：不要開 issue，請依 [SECURITY.md](SECURITY.md) 回報。

## 建置與測試

需要 Go 1.25 以上版本。

```bash
go build -o custodexa-mcp .
go test ./...
```

測試必須通過，行為變更需附上涵蓋它的測試。

## DCO 簽署（必要）

本專案以 AGPL-3.0 授權，並依 Developer Certificate of Origin 接受貢獻。不需要簽任何協議，只要每個 commit 加一行：

```bash
git commit -s -m "fix: describe your change"
```

`Signed-off-by` 這一行聲明這項修改是你寫的，或你有權以本專案的授權提交它。它不會轉讓你的著作權；全文見 [DCO.md](../../DCO.md)。Git 從你的 `user.name` 與 `user.email` 設定取得姓名與電子郵件，這些資料會永久留在公開歷史中，請使用你願意公開的資料。

- 忘了簽署：最後一個 commit 執行 `git commit --amend -s`，多個 commit 執行 `git rebase --signoff <base>`，然後 force push。
- 含未簽署 commit 的 pull request 不會被合併。審查照常進行，維護者會告訴你如何補上簽署。

## Pull request 檢查清單

- 每個 commit 都帶有 `Signed-off-by`。
- Commit 訊息遵循 [Conventional Commits](https://www.conventionalcommits.org/)（`feat:`、`fix:`、`refactor:`、`docs:`、`test:`、`chore:`）。
- `go test ./...` 通過。
- 程式碼、測試與範例中不含機密資訊；請使用 `<agent-token>` 這類佔位字串。
