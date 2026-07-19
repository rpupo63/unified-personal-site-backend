# unified-personal-site-backend

Go (chi) API for pupo.codes — Supabase Postgres, Linear sync, email/social integrations. Prod on Racknerd as `api.pupo.codes:8080`.

## Commands

```bash
go run main.go                          # API server
go build ./...
go run cmd/linear_sync/main.go          # Linear project sync (hourly timer in prod)
go run cmd/linear_sync/main.go --force  # force sync
./deploy-backend                        # deploy helper
```

## Architecture

- `api/` handlers · `database/` · `models/` · `services/` · `config/` · `cmd/` utilities
- Deploy: `~/Projects/DEPLOYMENT_SOP.md` (systemd + Caddy + Cloudflare)
- Secrets: `.env` (gitignored); see `.env.example`

## Where knowledge lives

- Sibling frontends live in separate git repos under `../frontend`, `../math-website`, `../project-and-blog-builder` (umbrella notes in `../AGENTS.md`).
- Durable facts → Memory MCP tags `user:beto`, `project:unified-personal-site-backend`
