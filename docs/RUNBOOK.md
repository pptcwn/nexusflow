# Runbook

## Local Backend

In WSL on this machine, use the explicit Windows Go binary if `go` is not on PATH:

```bash
cd core
'/mnt/c/Program Files/Go/bin/go.exe' mod download
'/mnt/c/Program Files/Go/bin/go.exe' test ./...
'/mnt/c/Program Files/Go/bin/go.exe' run .
```

On a normal Go installation:

```bash
cd core
go mod download
go test ./...
go run .
```

## Local Worker

In WSL on this machine, use Windows `npx.cmd` if the WSL wrapper fails with `node: Permission denied`:

```bash
cd worker
cmd.exe /c "E:\Nodejs\npx.cmd wrangler dev"
```

On a normal Node.js installation:

```bash
cd worker
npx wrangler dev
```

## Verification

Backend health:

```bash
curl http://localhost:8080/status
```

Decision endpoint:

```bash
curl -X POST http://localhost:8080/decide \
  -H 'content-type: application/json' \
  -d '{"ip":"203.0.113.10","hardwareConcurrency":8,"pluginCount":3,"behaviorScore":0.8,"hasInteraction":true,"timeOnPage":15}'
```

Behavior endpoint:

```bash
curl -i -X POST http://localhost:8080/behavior \
  -H 'content-type: application/json' \
  -d '{"mouseMoves":3,"timeOnPageBucket":"5-15","hasInteraction":true}'
```

Worker config validation:

```bash
cd worker
npx wrangler deploy --dry-run
```

## Deploy

Check Cloudflare auth:

```bash
cd worker
npx wrangler whoami
```

Deploy Worker:

```bash
cd worker
npx wrangler deploy
```

## Rollback

List versions and roll back:

```bash
cd worker
npx wrangler versions list
npx wrangler rollback <version-id>
```

## Kill Switch

Set these backend environment variables to route all decisions to review mode:

```dotenv
KILL_SWITCH_ENABLED=true
GLOBAL_MODE=review
```
