# Contributing

<p><b>English</b> | <a href="docs/zh-TW/CONTRIBUTING.md">繁體中文</a> | <a href="docs/ja/CONTRIBUTING.md">日本語</a></p>

Thanks for contributing to `custodexa-mcp`. This repository holds only the stdio relay. Authorization, auditing, masking and the tools themselves belong to the Custodexa server, so changes to them go to the [main Custodexa repository](https://github.com/custodexa/custodexa).

## Before you start

- Bug reports and ideas: open an issue with reproduction steps or your motivation.
- Docs, typos, or an obvious small bug: open a pull request directly.
- Behavior changes: open an issue first and agree on the approach before writing code.
- Security problems: do not open an issue; follow [SECURITY.md](SECURITY.md).

## Build and test

Requires Go 1.25 or later.

```bash
go build -o custodexa-mcp .
go test ./...
```

Tests must pass, and a behavior change needs a test that covers it.

## DCO sign-off (required)

The project is licensed under AGPL-3.0 and accepts contributions under the Developer Certificate of Origin. There is no agreement to sign, just one line per commit:

```bash
git commit -s -m "fix: describe your change"
```

The `Signed-off-by` line states that you wrote the change, or have the right to submit it under the project's license. It does not transfer your copyright; the full text is in [DCO.md](DCO.md). Git takes the name and email from your `user.name` and `user.email` settings, and they stay in public history permanently, so use details you are comfortable publishing.

- Forgot to sign: run `git commit --amend -s` for the last commit, or `git rebase --signoff <base>` for several, then force-push.
- Pull requests with unsigned commits are not merged. Review still goes ahead, and a maintainer will point out how to fix the sign-off.

## Pull request checklist

- Every commit carries `Signed-off-by`.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`).
- `go test ./...` passes.
- No secrets in code, tests or examples; use placeholders such as `<agent-token>`.
