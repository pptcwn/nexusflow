# NexusFlow Architecture

- Edge Layer: Cloudflare Worker (Progressive Injection)
- Decision Layer: Go Backend
- Detection Layer: Client-side (fingerprint.js)
- Protection: Redis Rate Limit + WAF