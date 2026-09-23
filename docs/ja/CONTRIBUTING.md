# コントリビューションガイド

<p><a href="../../CONTRIBUTING.md">English</a> | <a href="../zh-TW/CONTRIBUTING.md">繁體中文</a> | <b>日本語</b></p>

`custodexa-mcp` へのコントリビューションをありがとうございます。このリポジトリに含まれるのは stdio リレーだけです。認可、監査、マスク、そしてツールそのものは Custodexa サーバーが担うため、それらに関する変更は [Custodexa 本体のリポジトリ](https://github.com/custodexa/custodexa) へ送ってください。

## 始める前に

- バグの報告やアイデア: 再現手順、または提案の動機を添えて issue を立ててください。
- ドキュメントや誤字の修正、明らかな小さなバグの修正: そのままプルリクエストを送ってください。
- 振る舞いの変更: まず issue を立て、コードを書く前に方針について合意してください。
- セキュリティ上の問題: issue は立てず、[SECURITY.md](SECURITY.md) の手順に従ってください。

## ビルドとテスト

Go 1.25 以降が必要です。

```bash
go build -o custodexa-mcp .
go test ./...
```

テストはすべて通る必要があります。振る舞いを変更する場合は、その変更をカバーするテストが必要です。

## DCO の署名（必須）

本プロジェクトのライセンスは AGPL-3.0 で、コントリビューションは Developer Certificate of Origin のもとで受け入れます。署名する契約書はなく、コミットごとに 1 行を付けるだけです。

```bash
git commit -s -m "fix: describe your change"
```

`Signed-off-by` の行は、あなたがその変更を書いたこと、またはプロジェクトのライセンスのもとで提出する権利を持っていることを表明するものです。著作権が移転することはありません。全文は [DCO.md](../../DCO.md) にあります。Git は名前とメールアドレスを `user.name` と `user.email` の設定から取得し、それらは公開履歴に永久に残ります。公開して差し支えない情報をお使いください。

- 署名を忘れた場合: 直前のコミットなら `git commit --amend -s`、複数なら `git rebase --signoff <base>` を実行し、force-push してください。
- 署名のないコミットを含むプルリクエストはマージされません。レビューは通常どおり進み、メンテナが署名の直し方をお知らせします。

## プルリクエストのチェックリスト

- すべてのコミットに `Signed-off-by` が付いている。
- コミットメッセージが [Conventional Commits](https://www.conventionalcommits.org/) に従っている（`feat:`、`fix:`、`refactor:`、`docs:`、`test:`、`chore:`）。
- `go test ./...` が通る。
- コード、テスト、例にシークレットが含まれていない。`<agent-token>` のようなプレースホルダーを使っている。
