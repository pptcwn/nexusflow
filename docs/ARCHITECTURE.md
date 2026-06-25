# NexusFlow Architecture

NexusFlow classifies traffic for operational safety and transparent experience routing. It does not hide destinations from users or alter content after routing; it chooses a configured origin based on explicit backend classification.

## Components

- Client: loads the public page and sends privacy-safe aggregate behavior signals to `/behavior`.
- Cloudflare Worker: receives requests at the edge, checks Cloudflare bot/threat headers, asks the backend for a decision, and fetches the selected origin.
- Go decision backend: parses request fingerprints, applies deterministic scoring, rate-limit checks, and safety controls, then returns `review` or `standard`.
- Redis: optional backing service for per-IP rate limiting and blocked-IP lookups.

## Data Flow

```text
Client request -> Cloudflare Worker -> Go decision backend -> review or standard mode -> Worker fetches matching origin
```

## Environment Contract

- `BACKEND_URL`: URL for the Go decision backend.
- `STANDARD_ORIGIN`: origin for normal trusted traffic.
- `REVIEW_ORIGIN`: origin for conservative review traffic.
- `REDIS_URL`: Redis connection URL for backend rate-limit helpers.
- `BLOCK_COUNTRIES`: comma-separated country codes routed to `review`.
- `RATE_LIMIT_PER_MIN`: per-IP request threshold.
- `KILL_SWITCH_ENABLED`: when true, backend returns `GLOBAL_MODE`.
- `GLOBAL_MODE`: emergency mode, normally `review`.

## Runtime Boundaries

- Worker-to-backend requests carry request metadata only: IP, country, user agent, referer, and timestamp.
- Browser behavior ingestion uses aggregate metrics only. Raw mouse coordinates, movement traces, credentials, payment data, message content, and precise geolocation are prohibited.
- Backend decisions are deterministic and should be covered by Go tests before scoring changes.
