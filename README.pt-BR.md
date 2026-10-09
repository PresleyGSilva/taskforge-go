# TaskForge

Plataforma Kanban multi-tenant — API Go (Chi + pgx), Postgres 16, frontend Next.js 15. Violeta `#7C3AED`, dark-first.

## Funcionalidades

- Autenticação JWT access (15 min) + cookie httpOnly de refresh; login retorna `orgs[]`, UI persiste `orgs[0].id` (`taskforge_org_id`)
- Workspaces, membros (`owner|admin|member|viewer`), convites por e-mail com aceite via token
- Projetos, colunas ordenadas, tarefas com prioridade, tags, responsável e vencimento
- Concorrência otimista nas tarefas (`version`, `409 CONFLICT` em divergência, header `ETag`)
- Arrastar-e-soltar (`POST /tasks/:id/move`), comentários, checklists, metadados de anexos
- Feed de atividades, notificações, dashboard com contagens, busca global
- WebSocket por workspace (`task.created|updated|moved|deleted`, `comment.added`, `column.updated`)
- OpenAPI em `GET /openapi.yaml`, healthchecks, logs estruturados, shutdown gracioso

## Início rápido (docker compose)

```bash
cp .env.example .env
docker compose up --build
```

| Serviço  | URL / porta |
|----------|-------------|
| API      | http://localhost:8083 (saúde: `/healthz`, `/readyz`, spec: `/openapi.yaml`) |
| Web      | http://localhost:3003 |
| Postgres | localhost:5436 (`taskforge`/`taskforge`/`taskforge`) |

Portas padrão sobrescrevíveis via `API_PORT`, `WEB_PORT`, `DB_PORT`.

## Demo ao vivo do WS (2 navegadores)

1. Abra http://localhost:3003 em dois navegadores (ou normal + anônimo), com dois usuários do mesmo workspace.
2. Abra o mesmo quadro nos dois.
3. Arraste um cartão ou comente no navegador A — o cartão move / o comentário aparece no B na hora via eventos WS.

## Capturas de tela

- `docs/screenshots/board.png` — quadro Kanban ao vivo com cards e indicador WS
- `docs/screenshots/dashboard.png` — dashboard de produtividade com dados reais

## Resumo da API

```
POST /api/v1/auth/register | /auth/login | /auth/refresh | /auth/logout
GET|POST /api/v1/workspaces   GET|PATCH|DELETE /api/v1/workspaces/:id
GET|POST /api/v1/workspaces/:id/members   POST /api/v1/workspaces/:id/invites
GET /api/v1/invites/:token/accept
GET|POST /api/v1/projects?workspace_id=   GET|PATCH|DELETE /api/v1/projects/:id
GET|POST /api/v1/projects/:projectId/columns   PATCH|DELETE /api/v1/columns/:id
GET|POST /api/v1/tasks?...   GET|PATCH|DELETE /api/v1/tasks/:id
POST /api/v1/tasks/:id/move   GET|POST /api/v1/tasks/:id/comments
GET|POST /api/v1/tasks/:id/checklists   POST /api/v1/tasks/:id/attachments (metadados JSON)
GET /api/v1/activities?workspace_id=   GET /api/v1/notifications (+ PATCH /:id/read)
GET /api/v1/dashboard?workspace_id=   GET /api/v1/search?q=&workspace_id=
GET /ws?workspace=<id>
```

Headers de tenant: `X-Org-ID` (alias `X-Workspace-ID`) + `X-User` onde aplicável.

## Desenvolvimento

```bash
make build | make up | make down | make test | make seed
```

Veja `docs/ARCHITECTURE.md`. Licença: MIT.
