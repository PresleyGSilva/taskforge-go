# TaskForge — Architecture

TaskForge is a multi-tenant Kanban platform: a modular-monolith Go API
(Chi + pgx, slog) backed by Postgres 16, plus a Next.js 15 + TypeScript +
Tailwind frontend. Brand identity: violet `#7C3AED`, dark-first UI.

## Components

- `backend/` — Go modular monolith. Chi router, pgx pool, slog structured
  logs, JWT access (15 min, HS256) + refresh in httpOnly cookie, bcrypt
  passwords. Versioned SQL migrations, seed with fictitious data,
  `GET /openapi.yaml`, health probes (`/healthz`, `/readyz`), graceful
  shutdown.
- `frontend/` — Next.js 15 + TypeScript + Tailwind. `npm run start` serves
  the production build; `NEXT_PUBLIC_API_URL` is baked via Docker `ARG`.
- `docker-compose.yml` — services `db` (postgres:16), `api` (build
  `./backend`), `web` (build `./frontend`). Ports come from env with
  defaults API `8083`, web `3003`, DB host `5436`.

## API contract

- Base path: `/api/v1`.
- Envelopes: success `{ "data": ... }`, errors
  `{ "error": { "code": "NOT_FOUND", "message": "..." } }` with codes
  `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`.
- Auth: `POST /auth/register {name,email,password}` and `POST /auth/login`
  return `{ data: { user, access_token, orgs: [{ id, name }] } }` and set a
  `refresh_token` cookie; `POST /auth/refresh` (cookie) returns a new
  access token; `POST /auth/logout` clears it. `Authorization: Bearer`.
- Frontend persists the first org: `localStorage` key `taskforge_org_id`.
- Tenancy: every tenant-scoped request sends `X-Org-ID` (workspace id;
  `X-Workspace-ID` accepted as alias) and `X-User` where applicable.

## Data model

- `users(id, name, email unique, password_hash, created_at)`
- `workspaces(id, name, owner_id, created_at)` + `workspace_members(workspace_id, user_id, role)`
- `invites(token, workspace_id, email, role, accepted_at, expires_at)`
- `projects(id, workspace_id → workspaces, name, description, created_at)`
- `columns(id, project_id → projects, name, position)`
- `tasks(id, project_id, column_id → columns, title, description, priority, due_date, assignee_id → users, position, version, tags[], created_at, updated_at)`
- `checklists(id, task_id, title)` + `checklist_items(id, checklist_id, title, done, position)`
- `comments(id, task_id, author_id → users, body, created_at)`
- `attachments(id, task_id, filename, size, mime, created_at)` (metadata only)
- `activities(id, workspace_id, task_id nullable, actor_id, kind, payload jsonb, created_at)`
- `notifications(id, user_id, task_id nullable, kind, read_at, created_at)`

Money is not applicable (Kanban domain; no monetary floats anywhere).

## Optimistic concurrency (tasks)

Every task carries an integer `version`. `PATCH /tasks/:id` and
`POST /tasks/:id/move { column_id, position, version }` require the caller
to send the last-seen `version`. The update runs as a guarded
`UPDATE ... WHERE id = $1 AND version = $2` (incrementing `version`);
on mismatch the API returns `409 { error: { code: "CONFLICT" } }` and the
client must refetch. Responses include the new `version` plus an `ETag`
header so caches and the realtime layer can order events.

## Realtime (WebSocket)

- `GET /ws?workspace=<id>` with `Authorization: Bearer` (or `?token=`).
- Server keeps one hub per workspace; each connection joins the hub of the
  workspace id in the query after verifying membership.
- Events are JSON `{ "type": "...", "payload": {...} }` with types
  `task.created`, `task.updated`, `task.moved`, `task.deleted`,
  `comment.added`, `column.updated`.
- Mutations broadcast to every client of the same workspace except,
  optionally, the sender.

## Permissions (per workspace)

| Role          | Read | Write tasks/comments | Delete project | Manage members |
|---------------|------|----------------------|----------------|----------------|
| owner         | yes  | yes                  | yes            | yes            |
| admin         | yes  | yes                  | yes            | yes            |
| member        | yes  | yes                  | no             | no             |
| viewer        | yes  | no                   | no             | no             |

Enforced by auth + workspace-membership middleware resolving `X-Org-ID`
(`X-Workspace-ID` alias). Writes by `viewer` (and project deletes /
member management by `member`) return `403 { error: { code: "FORBIDDEN" } }`.

## Observability & operations

Structured slog logs, `/healthz` (liveness) and `/readyz` (DB check),
graceful shutdown on SIGTERM/SIGINT, versioned migrations run at boot.
