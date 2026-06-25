# NexusFlow Phase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan phase-by-phase. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the NexusFlow stabilization work as ordered phases with clear quality gates, commit boundaries, and deploy readiness checks.

**Architecture:** Stabilize from the inside out. First make the Go decision core buildable and testable, then align the Cloudflare Worker contract, then reduce behavior ingestion to privacy-safe aggregate signals, then document operations and add safety controls.

**Tech Stack:** Go 1.23, Gin, go-redis, Cloudflare Workers, TypeScript, Wrangler, GitHub.

---

## Phase Order

1. Phase 0: Baseline and repository hygiene
2. Phase 1: Go core build and deterministic decision tests
3. Phase 2: Worker and environment contract alignment
4. Phase 3: Privacy-safe behavior ingestion
5. Phase 4: Documentation and local operations
6. Phase 5: Operational safety controls
7. Phase 6: End-to-end verification and release branch

Each phase should produce one focused commit. Do not start the next phase until the previous phase's verification commands pass or the failure is documented in the phase notes.

---

### Phase 0: Baseline and Repository Hygiene

**Purpose:** Confirm the local repository is clean and synchronized before code changes.

**Files:**
- Read: `docs/superpowers/plans/2026-06-25-nexusflow-stabilization.md`
- Read: `README.md`
- Read: `docs/ARCHITECTURE.md`
- Read: `.gitignore`

- [ ] **Step 1: Confirm Git state**

Run:

```bash
git status --short --branch
git remote -v
git log --oneline -3
```

Expected:

```text
## main...origin/main
origin	https://github.com/pptcwn/nexusflow.git (fetch)
origin	https://github.com/pptcwn/nexusflow.git (push)
```

- [ ] **Step 2: Create implementation branch**

Run:

```bash
git switch -c codex/nexusflow-stabilization
```

Expected: current branch is `codex/nexusflow-stabilization`.

- [ ] **Step 3: Run baseline checks**

Run:

```bash
cd core
go test ./...
```

Expected at baseline: failure is acceptable if it matches the current known issue in the stabilization plan: mixed package boundaries or missing fields in the Go core.

- [ ] **Step 4: Record baseline result**

Add a short note to the implementation log in the final phase summary, not as a separate source file.

---

### Phase 1: Go Core Build and Decision Tests

**Purpose:** Make the Go backend compile with clean package boundaries and deterministic classification tests.

**Files:**
- Modify: `core/main.go`
- Create: `core/config/config.go`
- Create: `core/fingerprint/fingerprint.go`
- Create: `core/decision/decision.go`
- Create: `core/decision/decision_test.go`
- Create: `core/redisclient/redis.go`
- Delete after migration: `core/config.go`
- Delete after migration: `core/fingerprint.go`
- Delete after migration: `core/decision.go`
- Delete after migration: `core/redis.go`

- [ ] **Step 1: Implement the detailed Task 1 from the stabilization plan**

Use the exact code and commands from:

```text
docs/superpowers/plans/2026-06-25-nexusflow-stabilization.md
Task 1: Make the Go Core Buildable
```

- [ ] **Step 2: Verify Go packages**

Run:

```bash
cd core
go test ./...
```

Expected: all packages compile and `core/decision` tests pass.

- [ ] **Step 3: Smoke test backend startup**

Run:

```bash
cd core
go run .
```

In another shell:

```bash
curl http://localhost:8080/status
```

Expected: backend starts on the configured port and `/status` returns a successful JSON response.

- [ ] **Step 4: Commit Phase 1**

Run:

```bash
git add core
git commit -m "fix: make decision core buildable"
```

---

### Phase 2: Worker and Environment Contract Alignment

**Purpose:** Replace misleading environment names with `standard` and `review` routing terminology across edge and deploy configuration.

**Files:**
- Modify: `worker/src/index.ts`
- Modify: `worker/wrangler.toml`
- Modify: `.env.example`

- [ ] **Step 1: Implement the detailed Task 2 from the stabilization plan**

Use the exact contract from:

```text
docs/superpowers/plans/2026-06-25-nexusflow-stabilization.md
Task 2: Align Worker and Environment Contracts
```

Required Worker env shape:

```ts
export interface Env {
  BACKEND_URL: string;
  STANDARD_ORIGIN: string;
  REVIEW_ORIGIN: string;
}
```

- [ ] **Step 2: Verify Worker dry run**

Run:

```bash
cd worker
npx wrangler deploy --dry-run
```

Expected: Wrangler parses `worker/src/index.ts` and `worker/wrangler.toml` without env contract errors.

- [ ] **Step 3: Commit Phase 2**

Run:

```bash
git add worker .env.example
git commit -m "chore: align worker environment contract"
```

---

### Phase 3: Privacy-Safe Behavior Ingestion

**Purpose:** Ensure behavior collection uses aggregate operational signals only and the backend accepts or rejects behavior payloads deterministically.

**Files:**
- Modify: `public/fingerprint.js`
- Modify: `core/main.go`
- Create: `docs/PRIVACY.md`

- [ ] **Step 1: Implement privacy notice**

Create `docs/PRIVACY.md` with these required sections:

```markdown
# Privacy

NexusFlow collects only operational signals needed for bot and fraud classification, rate limiting, service reliability, and transparent experience routing.

## Collected Categories

- Request metadata: IP address, country, user agent, referer, and request timestamp.
- Browser capability signals: canvas hash, WebGL vendor, hardware concurrency, plugin count, and WebGL consistency flags.
- Coarse interaction metrics: interaction presence, timing buckets, total mouse movement count, and aggregate movement quality scores.
- Operational logs: classification mode, confidence score, rate-limit events, blocked-country events, and backend errors.

## Prohibited Data

NexusFlow must not collect credentials, message content, payment data, precise geolocation, raw mouse coordinate arrays, raw movement traces, or form field values.

## Retention

Operational logs should be retained only as long as needed for abuse prevention, debugging, and service reliability.
```

- [ ] **Step 2: Reduce client payload**

Update `public/fingerprint.js` so the behavior request sends aggregate metrics only. It must not send raw arrays named `mousePositions`, `mouseTrail`, `coordinates`, `rawMoves`, or equivalent raw pointer traces.

- [ ] **Step 3: Harden `/behavior` status handling**

Update `core/main.go` so malformed JSON returns HTTP 400 and successfully parsed behavior payloads return HTTP 204.

- [ ] **Step 4: Verify behavior payload**

Run backend:

```bash
cd core
go run .
```

Run static server:

```bash
python3 -m http.server 4173 --directory public
```

Open:

```text
http://localhost:4173/s/index.html
```

Expected: `/behavior` receives aggregate data and no raw mouse coordinate arrays.

- [ ] **Step 5: Commit Phase 3**

Run:

```bash
git add public/fingerprint.js core/main.go docs/PRIVACY.md
git commit -m "feat: add privacy-safe behavior ingestion"
```

---

### Phase 4: Documentation and Local Operations

**Purpose:** Make the project understandable for the next worker and give exact local run, verification, deploy, and rollback commands.

**Files:**
- Modify: `README.md`
- Modify: `docs/ARCHITECTURE.md`
- Create: `docs/RUNBOOK.md`

- [ ] **Step 1: Update README positioning**

Rewrite the opening of `README.md` to describe NexusFlow as a transparent traffic classification and experience-routing service.

- [ ] **Step 2: Add runbook**

Create `docs/RUNBOOK.md` with local backend commands:

```bash
cd core
go mod download
go test ./...
go run .
```

Worker commands:

```bash
cd worker
npx wrangler dev
```

Verification commands:

```bash
curl http://localhost:8080/status
curl -X POST http://localhost:8080/decide \
  -H 'content-type: application/json' \
  -d '{"ip":"203.0.113.10","hardwareConcurrency":8,"pluginCount":3,"behaviorScore":0.8,"hasInteraction":true,"timeOnPage":15}'
```

- [ ] **Step 3: Update architecture**

Update `docs/ARCHITECTURE.md` with this flow:

```text
Client request -> Cloudflare Worker -> Go decision backend -> review or standard mode -> Worker fetches matching origin
```

- [ ] **Step 4: Commit Phase 4**

Run:

```bash
git add README.md docs/ARCHITECTURE.md docs/RUNBOOK.md
git commit -m "docs: add runbook and compliant architecture"
```

---

### Phase 5: Operational Safety Controls

**Purpose:** Add protective behavior for oversized requests, blocked countries, rate-limited clients, and emergency kill switch routing.

**Files:**
- Modify: `core/main.go`
- Modify: `core/config/config.go`
- Modify: `docs/RUNBOOK.md`

- [ ] **Step 1: Add request body size limit**

Cap JSON request bodies for `/decide` and `/behavior` at 64KB using Gin middleware or `http.MaxBytesReader`.

- [ ] **Step 2: Apply blocked-country routing**

Before calling `decision.MakeDecision`, route requests from configured blocked countries to:

```json
{"mode":"review","allowProgressive":false}
```

- [ ] **Step 3: Apply Redis rate-limit routing**

Before calling `decision.MakeDecision`, route rate-limited IPs to:

```json
{"mode":"review","allowProgressive":false}
```

- [ ] **Step 4: Document kill switch**

Add to `docs/RUNBOOK.md`:

```dotenv
KILL_SWITCH_ENABLED=true
GLOBAL_MODE=review
```

Expected behavior: all decisions return `review` mode while the kill switch is enabled.

- [ ] **Step 5: Verify safety behavior**

Run:

```bash
cd core
go test ./...
```

Then run:

```bash
curl -X POST http://localhost:8080/decide \
  -H 'content-type: application/json' \
  -d '{"ip":"203.0.113.10","country":"RU","hardwareConcurrency":8,"pluginCount":3,"behaviorScore":0.8,"hasInteraction":true,"timeOnPage":15}'
```

Expected:

```json
{"mode":"review"}
```

Additional fields may be present, but `mode` must be `review`.

- [ ] **Step 6: Commit Phase 5**

Run:

```bash
git add core docs/RUNBOOK.md
git commit -m "feat: add operational safety controls"
```

---

### Phase 6: End-to-End Verification and Release Branch

**Purpose:** Confirm the full service path works and publish the stabilization branch.

**Files:**
- Read: all changed files
- No planned source edits unless verification finds a defect

- [ ] **Step 1: Run full backend test suite**

Run:

```bash
cd core
go test ./...
```

Expected: all tests pass.

- [ ] **Step 2: Run Worker validation**

Run:

```bash
cd worker
npx wrangler deploy --dry-run
```

Expected: dry run completes without syntax or config errors.

- [ ] **Step 3: Run backend decision smoke test**

Run:

```bash
cd core
go run .
```

In another shell:

```bash
curl -X POST http://localhost:8080/decide \
  -H 'content-type: application/json' \
  -d '{"ip":"203.0.113.10","country":"US","hardwareConcurrency":8,"pluginCount":3,"behaviorScore":0.8,"watchTime":9,"hasInteraction":true,"timeOnPage":15,"canvasHash":"canvas-hash","webglVendor":"Intel Inc."}'
```

Expected: response contains `"mode":"standard"` for trusted aggregate human-like signals.

- [ ] **Step 4: Inspect final diff**

Run:

```bash
git status --short --branch
git log --oneline --decorate -8
```

Expected: working tree is clean and the phase commits are visible on `codex/nexusflow-stabilization`.

- [ ] **Step 5: Push branch**

Run:

```bash
git push -u origin codex/nexusflow-stabilization
```

Expected: branch is available on `https://github.com/pptcwn/nexusflow.git`.

---

## Phase Gates

- Phase 1 gate: `cd core && go test ./...` passes.
- Phase 2 gate: `cd worker && npx wrangler deploy --dry-run` parses Worker and config.
- Phase 3 gate: `/behavior` returns 204 for valid aggregate payload and 400 for malformed JSON.
- Phase 4 gate: README, architecture, and runbook describe the same `review` and `standard` contract.
- Phase 5 gate: blocked country, rate limit, and kill switch all route to `review`.
- Phase 6 gate: backend tests, Worker dry run, and decision smoke test pass.

## Self-Review

- Spec coverage: This phase plan covers all five stabilization tasks and adds baseline plus final release verification.
- Placeholder scan: No phase depends on undefined future work or a TBD value.
- Type consistency: The plan consistently uses `review`, `standard`, `STANDARD_ORIGIN`, and `REVIEW_ORIGIN`.
