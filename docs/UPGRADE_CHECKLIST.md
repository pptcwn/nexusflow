# Upgrade Checklist

- Review `worker/wrangler.toml` compatibility date.
- Review Go version in `core/go.mod`.
- Validate environment variables in `.env.example`.
- Re-test `public/fingerprint.js` against target browsers.
- Confirm routing behavior for both `public/s/` and `public/m/`.
