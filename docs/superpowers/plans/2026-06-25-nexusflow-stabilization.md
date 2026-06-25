# NexusFlow Stabilization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert the current NexusFlow prototype into a buildable, testable, policy-compliant traffic classification and transparent experience-routing service.

**Architecture:** First make the Go decision service compile with clear package boundaries and deterministic tests. Then align the Cloudflare Worker environment contract, add privacy-safe behavior ingestion, and document deployment/operations.

**Tech Stack:** Go 1.23, Gin, go-redis, Cloudflare Workers, TypeScript, Wrangler.

---

## Scope Boundaries

- Keep allowed behavior limited to bot/fraud classification, rate limiting, operational safety, and transparent routing.
- Remove or rename misleading "safe/money/cloaking" language before production use.

## File Structure

- `core/main.go`: HTTP entrypoint only.
- `core/config/config.go`: environment-backed configuration.
- `core/fingerprint/fingerprint.go`: request and behavior signal schema.
- `core/decision/decision.go`: deterministic classification logic.
- `core/decision/decision_test.go`: scoring and mode tests.
- `core/redisclient/redis.go`: Redis client and rate-limit helpers.
- `worker/src/index.ts`: edge request handling and backend integration.
- `worker/wrangler.toml`: Worker vars that match TypeScript `Env`.
- `.env.example`: shared local/deploy environment reference.
- `docs/ARCHITECTURE.md`: current architecture and data flow.
- `docs/RUNBOOK.md`: local run, deploy, rollback, and verification steps.

---

### Task 1: Make the Go Core Buildable

**Files:**
- Modify: `core/main.go`
- Create: `core/config/config.go`
- Create: `core/fingerprint/fingerprint.go`
- Create: `core/decision/decision.go`
- Create: `core/decision/decision_test.go`
- Create: `core/redisclient/redis.go`
- Remove after migration: `core/config.go`, `core/fingerprint.go`, `core/decision.go`, `core/redis.go`

- [ ] **Step 1: Write decision tests first**

Create `core/decision/decision_test.go`:

```go
package decision

import (
	"testing"

	"nexusflow/core/config"
	"nexusflow/core/fingerprint"
)

func TestMakeDecisionReturnsReviewForAutomationSignals(t *testing.T) {
	cfg := config.Default()
	fp := fingerprint.Fingerprint{IsHeadless: true, Webdriver: true}

	got := MakeDecision(fp, cfg)

	if got.Mode != "review" {
		t.Fatalf("Mode = %q, want review", got.Mode)
	}
	if got.AllowProgressive {
		t.Fatalf("AllowProgressive = true, want false")
	}
}

func TestMakeDecisionAllowsTrustedHumanSignals(t *testing.T) {
	cfg := config.Default()
	fp := fingerprint.Fingerprint{
		IP:                            "203.0.113.10",
		Referer:                       "https://google.com/search?q=example",
		CanvasHash:                    "canvas-hash",
		WebGLVendor:                   "Intel Inc.",
		HardwareConcurrency:           8,
		PluginCount:                   3,
		BehaviorScore:                 0.8,
		WatchTime:                     9,
		HasInteraction:                true,
		TimeOnPage:                    15,
		MouseMoves:                    30,
		MousePatternScore:             85,
		JerkScore:                     82,
		AccelerationDistributionScore: 78,
		WebGLDetection: &fingerprint.WebGLDetection{
			MultipleCallsConsistent: true,
		},
	}

	got := MakeDecision(fp, cfg)

	if got.Mode != "standard" {
		t.Fatalf("Mode = %q, want standard", got.Mode)
	}
	if !got.AllowProgressive {
		t.Fatalf("AllowProgressive = false, want true")
	}
	if got.Confidence <= 0 {
		t.Fatalf("Confidence = %v, want > 0", got.Confidence)
	}
}
```

- [ ] **Step 2: Run the test and confirm current failure**

Run:

```bash
cd core
go test ./...
```

Expected before implementation: compile failure caused by mixed packages in `core/` and missing fields.

- [ ] **Step 3: Move config into its package**

Create `core/config/config.go`:

```go
package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	RedisURL          string
	StandardOrigin    string
	ReviewOrigin      string
	BlockCountries    []string
	RateLimitPerMin   int
	KillSwitchEnabled bool
	GlobalMode        string
}

func Default() *Config {
	return &Config{
		Port:              "8080",
		RedisURL:          "redis://localhost:6379",
		StandardOrigin:    "https://standard.yourdomain.com",
		ReviewOrigin:      "https://review.yourdomain.com",
		BlockCountries:    []string{"RU", "CN"},
		RateLimitPerMin:   40,
		KillSwitchEnabled: false,
		GlobalMode:        "review",
	}
}

func Load() *Config {
	cfg := Default()
	cfg.Port = getEnv("PORT", cfg.Port)
	cfg.RedisURL = getEnv("REDIS_URL", cfg.RedisURL)
	cfg.StandardOrigin = getEnv("STANDARD_ORIGIN", cfg.StandardOrigin)
	cfg.ReviewOrigin = getEnv("REVIEW_ORIGIN", cfg.ReviewOrigin)
	cfg.GlobalMode = getEnv("GLOBAL_MODE", cfg.GlobalMode)
	cfg.BlockCountries = splitEnv("BLOCK_COUNTRIES", cfg.BlockCountries)
	cfg.RateLimitPerMin = intEnv("RATE_LIMIT_PER_MIN", cfg.RateLimitPerMin)
	cfg.KillSwitchEnabled = boolEnv("KILL_SWITCH_ENABLED", cfg.KillSwitchEnabled)
	return cfg
}

func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes"
}
```

- [ ] **Step 4: Move fingerprint schema into its package**

Create `core/fingerprint/fingerprint.go` with all fields used by the decision package:

```go
package fingerprint

type Fingerprint struct {
	IP                            string                 `json:"ip"`
	Country                       string                 `json:"country"`
	UserAgent                     string                 `json:"userAgent"`
	Referer                       string                 `json:"referer"`
	Timestamp                     int64                  `json:"timestamp"`
	CanvasHash                    string                 `json:"canvasHash"`
	WebGLVendor                   string                 `json:"webglVendor"`
	HardwareConcurrency           int                    `json:"hardwareConcurrency"`
	BehaviorScore                 float64                `json:"behaviorScore"`
	WatchTime                     float64                `json:"watchTime"`
	HasInteraction                bool                   `json:"hasInteraction"`
	TimeOnPage                    float64                `json:"timeOnPage"`
	IsHeadless                    bool                   `json:"isHeadless"`
	Webdriver                     bool                   `json:"webdriver"`
	AutomationFlags               bool                   `json:"automationFlags"`
	PluginCount                   int                    `json:"pluginCount"`
	MouseMoves                    int                    `json:"mouseMoves"`
	MousePatternScore             float64                `json:"mousePatternScore"`
	JerkScore                     float64                `json:"jerkScore"`
	AccelerationDistributionScore float64                `json:"accelerationDistributionScore"`
	WebGLDetection                *WebGLDetection        `json:"webglDetection"`
	AudioContext                  map[string]interface{} `json:"audioContext"`
}

type WebGLDetection struct {
	Inconsistency           bool `json:"inconsistency"`
	SuspiciousVendor        bool `json:"suspiciousVendor"`
	MultipleCallsConsistent bool `json:"multipleCallsConsistent"`
}
```

- [ ] **Step 5: Move decision logic and use compliant modes**

Create `core/decision/decision.go` by migrating the existing scoring logic and replacing output modes:

```go
package decision

import (
	"strings"

	"nexusflow/core/config"
	"nexusflow/core/fingerprint"
)

type Decision struct {
	Mode             string  `json:"mode"`
	Variant          int     `json:"variant"`
	Score            float64 `json:"score"`
	Confidence       float64 `json:"confidence"`
	AllowProgressive bool    `json:"allowProgressive"`
}

func MakeDecision(fp fingerprint.Fingerprint, cfg *config.Config) Decision {
	if cfg.KillSwitchEnabled {
		return Decision{Mode: cfg.GlobalMode, Confidence: 0, AllowProgressive: false}
	}

	if fp.IsHeadless || fp.Webdriver || fp.AutomationFlags {
		return Decision{Mode: "review", Score: 0, Confidence: 0, AllowProgressive: false}
	}

	score := scoreFingerprint(fp)
	confidence := score / 115
	if confidence > 1 {
		confidence = 1
	}
	if confidence < 0 {
		confidence = 0
	}

	mode := "review"
	allowProgressive := false
	if score >= 58 && confidence >= 0.52 {
		mode = "standard"
		allowProgressive = true
	}

	return Decision{
		Mode:             mode,
		Variant:          generateVariant(fp.IP),
		Score:            score,
		Confidence:       confidence,
		AllowProgressive: allowProgressive,
	}
}

func scoreFingerprint(fp fingerprint.Fingerprint) float64 {
	score := 0.0

	if fp.WebGLDetection != nil {
		if fp.WebGLDetection.Inconsistency {
			score -= 20
		}
		if fp.WebGLDetection.SuspiciousVendor {
			score -= 15
		}
		if !fp.WebGLDetection.MultipleCallsConsistent {
			score -= 18
		}
	}

	if fp.MousePatternScore < 40 && fp.MouseMoves > 12 {
		score -= 22
	}
	if fp.JerkScore < 45 {
		score -= 20
	}
	if fp.AccelerationDistributionScore < 50 {
		score -= 22
	}
	if fp.HardwareConcurrency == 0 || fp.HardwareConcurrency > 32 {
		score -= 10
	}
	if fp.PluginCount == 0 {
		score -= 8
	}
	if fp.CanvasHash != "" {
		score += 18
	}
	if fp.WebGLVendor != "" {
		score += 12
	}
	if fp.HardwareConcurrency >= 2 {
		score += 8
	}
	if fp.BehaviorScore > 0.75 {
		score += 32
	} else if fp.BehaviorScore > 0.6 {
		score += 22
	} else if fp.BehaviorScore > 0.45 {
		score += 12
	}
	if fp.WatchTime > 5 {
		score += 12
	}
	if fp.HasInteraction {
		score += 10
	}
	if fp.TimeOnPage > 8 {
		score += 8
	}
	if isGoodReferer(fp.Referer) {
		score += 8
	}

	return score
}

func isGoodReferer(referer string) bool {
	good := []string{"facebook.com", "instagram.com", "google.com", "t.co", "youtube.com"}
	for _, value := range good {
		if strings.Contains(strings.ToLower(referer), value) {
			return true
		}
	}
	return false
}

func generateVariant(ip string) int {
	hash := 0
	for _, char := range ip {
		hash = (hash*31 + int(char)) % 10
	}
	return hash
}
```

- [ ] **Step 6: Move Redis client into its package**

Create `core/redisclient/redis.go`:

```go
package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"nexusflow/core/config"
)

var Client *redis.Client
var Ctx = context.Background()

func Init(cfg *config.Config) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		Client = nil
		return
	}
	Client = redis.NewClient(opt)
}

func IsRateLimited(ip string, cfg *config.Config) bool {
	if Client == nil || ip == "" {
		return false
	}

	key := "rate:" + ip
	count, err := Client.Incr(Ctx, key).Result()
	if err != nil {
		return false
	}
	if count == 1 {
		Client.Expire(Ctx, key, time.Minute)
	}
	return int(count) > cfg.RateLimitPerMin
}

func IsBlocked(ip string) bool {
	if Client == nil || ip == "" {
		return false
	}
	blocked, err := Client.SIsMember(Ctx, "blocked_ips", ip).Result()
	return err == nil && blocked
}
```

- [ ] **Step 7: Update `core/main.go` to use packages**

Keep the existing routes but validate JSON errors and use review mode for rate-limited requests.

- [ ] **Step 8: Delete old same-directory package files**

Run:

```bash
rm core/config.go core/fingerprint.go core/decision.go core/redis.go
```

- [ ] **Step 9: Verify Go tests pass**

Run:

```bash
cd core
go test ./...
```

Expected: all packages compile and `core/decision` tests pass.

- [ ] **Step 10: Commit**

```bash
git add core
git commit -m "fix: make decision core buildable"
```

---

### Task 2: Align Worker and Environment Contracts

**Files:**
- Modify: `worker/src/index.ts`
- Modify: `worker/wrangler.toml`
- Modify: `.env.example`

- [ ] **Step 1: Rename Worker env keys to the compliant contract**

Use:

```ts
export interface Env {
  BACKEND_URL: string;
  STANDARD_ORIGIN: string;
  REVIEW_ORIGIN: string;
}
```

- [ ] **Step 2: Route suspicious traffic to review origin**

In `worker/src/index.ts`, replace references to `SAFE_ORIGIN` with `REVIEW_ORIGIN` and `MONEY_ORIGIN` with `STANDARD_ORIGIN`.

- [ ] **Step 3: Update `worker/wrangler.toml` vars**

Add:

```toml
[vars]
ENVIRONMENT = "production"
BACKEND_URL = "https://your-backend.example.com"
STANDARD_ORIGIN = "https://standard.yourdomain.com"
REVIEW_ORIGIN = "https://review.yourdomain.com"
```

- [ ] **Step 4: Update `.env.example`**

Use:

```dotenv
PORT=8080
BACKEND_URL=http://localhost:8080
STANDARD_ORIGIN=https://standard.yourdomain.com
REVIEW_ORIGIN=https://review.yourdomain.com
REDIS_URL=redis://localhost:6379
BLOCK_COUNTRIES=RU,CN
RATE_LIMIT_PER_MIN=40
KILL_SWITCH_ENABLED=false
GLOBAL_MODE=review
ENVIRONMENT=development
```

- [ ] **Step 5: Verify Worker syntax**

Run:

```bash
cd worker
npx wrangler deploy --dry-run
```

Expected: Wrangler parses `src/index.ts` and `wrangler.toml` without missing env type errors.

- [ ] **Step 6: Commit**

```bash
git add worker .env.example
git commit -m "chore: align worker environment contract"
```

---

### Task 3: Add Privacy-Safe Behavior Ingestion

**Files:**
- Modify: `public/fingerprint.js`
- Modify: `core/main.go`
- Create: `docs/PRIVACY.md`

- [ ] **Step 1: Add a concise privacy notice document**

Create `docs/PRIVACY.md` documenting collected categories: request metadata, browser capability signals, coarse interaction metrics, and operational logs. State that the service must not collect credentials, message content, payment data, or precise geolocation.

- [ ] **Step 2: Reduce client payload**

Keep aggregate metrics only in `public/fingerprint.js`: counts, timing buckets, WebGL consistency flags, and browser capability counts. Do not send raw mouse coordinates or raw movement traces.

- [ ] **Step 3: Return strict behavior endpoint status**

Update `/behavior` in `core/main.go` to return HTTP 204 after parsing succeeds and HTTP 400 on malformed JSON.

- [ ] **Step 4: Verify manually**

Run local backend and load `public/s/index.html` through a local static server. Confirm `/behavior` receives aggregate data and no raw mouse coordinate arrays.

- [ ] **Step 5: Commit**

```bash
git add public/fingerprint.js core/main.go docs/PRIVACY.md
git commit -m "feat: add privacy-safe behavior ingestion"
```

---

### Task 4: Add Local Runbook and Verification Commands

**Files:**
- Modify: `README.md`
- Modify: `docs/ARCHITECTURE.md`
- Create: `docs/RUNBOOK.md`

- [ ] **Step 1: Update README purpose**

Rewrite the opening description as a transparent traffic classification and experience-routing service.

- [ ] **Step 2: Document local run commands**

Create `docs/RUNBOOK.md` with:

```bash
cd core
go mod download
go test ./...
go run .
```

and:

```bash
cd worker
npx wrangler dev
```

- [ ] **Step 3: Document verification endpoints**

Include:

```bash
curl http://localhost:8080/status
curl -X POST http://localhost:8080/decide \
  -H 'content-type: application/json' \
  -d '{"ip":"203.0.113.10","hardwareConcurrency":8,"pluginCount":3,"behaviorScore":0.8,"hasInteraction":true,"timeOnPage":15}'
```

- [ ] **Step 4: Update architecture data flow**

Document: request enters Worker, Worker asks backend for classification, backend returns `review` or `standard`, Worker fetches the corresponding origin.

- [ ] **Step 5: Commit**

```bash
git add README.md docs/ARCHITECTURE.md docs/RUNBOOK.md
git commit -m "docs: add runbook and compliant architecture"
```

---

### Task 5: Operational Safety

**Files:**
- Modify: `core/main.go`
- Modify: `core/config/config.go`
- Modify: `docs/RUNBOOK.md`

- [ ] **Step 1: Add request size limit**

In `core/main.go`, set `http.MaxBytesReader` or Gin middleware to cap JSON request bodies at 64KB for `/decide` and `/behavior`.

- [ ] **Step 2: Wire country blocking and rate limit**

Before `decision.MakeDecision`, return `review` mode when the country is blocked or the IP is rate-limited.

- [ ] **Step 3: Add kill switch instructions**

Document `KILL_SWITCH_ENABLED=true` and `GLOBAL_MODE=review` in `docs/RUNBOOK.md`.

- [ ] **Step 4: Verify**

Run:

```bash
cd core
go test ./...
```

Then send a blocked-country payload and confirm mode is `review`.

- [ ] **Step 5: Commit**

```bash
git add core docs/RUNBOOK.md
git commit -m "feat: add operational safety controls"
```

---

## Self-Review

- Spec coverage: This plan covers buildability, environment alignment, privacy-safe ingestion, documentation, and operational safety.
- Placeholder scan: No task relies on "TBD" or undefined future work.
- Type consistency: The plan standardizes mode names to `review` and `standard`, and environment names to `REVIEW_ORIGIN` and `STANDARD_ORIGIN`.
