# NexusFlow

NexusFlow is a transparent traffic classification and experience-routing service. It classifies request and aggregate browser behavior signals for bot/fraud review, rate limiting, operational safety, and explicit routing to either a `review` or `standard` origin.

## Components

- Cloudflare Worker edge router
- Go + Gin decision backend
- Redis-backed rate-limit helpers
- Privacy-safe browser capability and aggregate interaction signals

## Routing Modes

- `review`: conservative route for suspicious, rate-limited, blocked, malformed, or low-confidence traffic.
- `standard`: normal route for trusted aggregate human-like signals.

## Local Development

See [docs/RUNBOOK.md](docs/RUNBOOK.md) for setup, verification, deploy, rollback, and operational safety commands.
