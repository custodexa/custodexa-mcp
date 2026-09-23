# Reporting Security Issues

**English** | [繁體中文](docs/zh-TW/SECURITY.md) | [日本語](docs/ja/SECURITY.md)

> If the language versions ever diverge, this English text governs.

## Please do not open a public issue

Use GitHub's private vulnerability reporting. Pick the repository by where the problem lives:

| The problem is in | Report it here |
|---|---|
| This relay: how it handles the agent token, the endpoint URL, or the messages it forwards | [custodexa-mcp private report](https://github.com/custodexa/custodexa-mcp/security/advisories/new) |
| The Custodexa server: authorization, request approval, the tool-call ledger, masking, recording, or anything else behind `POST /api/v1/mcp` | [custodexa private report](https://github.com/custodexa/custodexa/security/advisories/new) |

If you are not sure which one applies, use this repository.

Only the maintainer can see reports sent through this channel, and nothing becomes public before an advisory is published.

## What to include if you can

None of these are required, and a report is never rejected for missing items:

1. The relay version (`serverInfo.version` from `initialize`, or the release tag) and the Custodexa version it talked to
2. Your MCP host and operating system
3. Reproduction steps
4. Impact: what an attacker gains, and what access they need first
5. How you would like to be credited, or that you prefer to stay anonymous

Never include real tokens, credentials, personal data or customer data. If reproduction depends on specific data, describe its shape instead of pasting it.

## Ground rules for testing

- Test only deployments you own or are authorized in writing to test. There is no public test instance.
- No data destruction, no access to data that is not yours, and no denial-of-service load testing.

## What you can expect

This project makes no response-time (SLA) promises. Reports are handled in this order instead:

| Severity | Handling |
|---|---|
| Critical | Interrupts current work and takes priority over feature development |
| High | Scheduled ahead of the next batch of work |
| Medium / Low | Enters the backlog with other security items |

Anything already exploited in the wild, or with a public proof of concept, moves up one level. Severity is never lowered because a fix is hard. Deciding not to fix is a legitimate outcome, and it always comes with the reasoning.

## Disclosure

Please hold the details until a fix ships. If no fix has shipped within 90 days of your report, you may publish without asking. That is your right, not a promise that a fix will exist by then.

Fixes are announced as GitHub Security Advisories, with the impact, affected versions, mitigations and credit to the reporter unless you prefer to stay anonymous.

## Supported versions

Security fixes ship only in the latest release.

## What this project does not have

- A bug bounty.
- A formal legal safe harbor. The maintainer will not act against good-faith reporters who follow the rules above, but that is a statement of intent, not a legal commitment.
