# TaskForge

Multi-tenant Kanban platform — Go (Chi + pgx) API, Postgres 16, Next.js 15 frontend. Violet `#7C3AED`, dark-first.

## Features

- Auth with JWT access (15 min) + httpOnly refresh cookie; login returns `orgs[]`, UI persists `orgs[0].id` (`taskforge_org_id`)
- Workspaces, members (`owner|admin|member|viewer`), email invites with token accept
- Projects, columns (ordered), tasks with priority, tags, assignee, due dates
- Optimistic concurrency on tasks (`version`, `409 CONFLICT` on mismatch, `ETag` header)
- Drag-and-drop move (`POST /tasks/:id/move`), comments, checklists, attachment metadata
- Activity feed, notifications, dashboard counts, global search
- Realtime per-workspace WebSocket (`task.created|updated|moved|deleted`, `comment.added`, `column.updated`)
- OpenAPI at `GET /openapi.yaml`, health probes, structured logs, graceful shutdown

## Quickstart (docker compose)

```bash
cp .env.example .env
docker compose up --build
```

| Service | URL / port |
|---------|------------|
| API     | http://localhost:8083 (health: `/healthz`, `/readyz`, spec: `/openapi.yaml`) |
| Web     | http://localhost:3003 |
| Postgres| localhost:5436 (`taskforge`/`taskforge`/`taskforge`) |

Default ports can be overridden via `API_PORT`, `WEB_PORT`, `DB_PORT` env vars.

## Live realtime demo (2 browsers)

1. Open http://localhost:3003 in two browsers (or normal + incognito), log in with two users of the same workspace.
2. Open the same board in both.
3. Drag a card or post a comment in browser A — the card moves / comment appears in browser B instantly via WS events.

## Screenshots

- `docs/screenshots/board.png` — live Kanban board with cards and WS live indicator
- `docs/screenshots/dashboard.png` — productivity dashboard with real data

## API sketch

```
POST /api/v1/auth/register | /auth/login | /auth/refresh | /auth/logout
GET|POST /api/v1/workspaces   GET|PATCH|DELETE /api/v1/workspaces/:id
GET|POST /api/v1/workspaces/:id/members   POST /api/v1/workspaces/:id/invites
GET /api/v1/invites/:token/accept
GET|POST /api/v1/projects?workspace_id=   GET|PATCH|DELETE /api/v1/projects/:id
GET|POST /api/v1/projects/:projectId/columns   PATCH|DELETE /api/v1/columns/:id
GET|POST /api/v1/tasks?...   GET|PATCH|DELETE /api/v1/tasks/:id
POST /api/v1/tasks/:id/move   GET|POST /api/v1/tasks/:id/comments
GET|POST /api/v1/tasks/:id/checklists   POST /api/v1/tasks/:id/attachments (JSON metadata)
GET /api/v1/activities?workspace_id=   GET /api/v1/notifications (+ PATCH /:id/read)
GET /api/v1/dashboard?workspace_id=   GET /api/v1/search?q=&workspace_id=
GET /ws?workspace=<id>
```

Tenant headers: `X-Org-ID` (`X-Workspace-ID` alias) + `X-User` where applicable.

## Dev

```bash
make build | make up | make down | make test | make seed
```

See `docs/ARCHITECTURE.md`. License: MIT.
