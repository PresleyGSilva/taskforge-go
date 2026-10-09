package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/PresleyGSilva/taskforge-go/internal/middleware"
	"github.com/PresleyGSilva/taskforge-go/internal/store"
	"github.com/PresleyGSilva/taskforge-go/internal/ws"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Store     store.Store
	JWTSecret string
	Hub       *ws.Hub
}

func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CORS(nil))
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) { ok(w, map[string]string{"status": "ok"}) })
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) { ok(w, map[string]string{"status": "ready"}) })
	r.Get("/openapi.yaml", serveOpenAPI)

	auth := middleware.Auth(s.JWTSecret, s.Store)
	wsOpt := middleware.Workspace(s.Store, false)
	wsReq := middleware.Workspace(s.Store, true)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", s.register)
		r.Post("/auth/login", s.login)
		r.Post("/auth/refresh", s.refresh)
		r.Post("/auth/logout", s.logout)

		r.Group(func(r chi.Router) {
			r.Use(auth)
			// workspaces (org header optional for list/create)
			r.With(wsOpt).Get("/workspaces", s.listWorkspaces)
			r.With(wsOpt).Post("/workspaces", s.createWorkspace)
			r.With(wsReq).Get("/workspaces/{id}", s.getWorkspace)
			r.With(wsReq).Patch("/workspaces/{id}", s.updateWorkspace)
			r.With(wsReq).Delete("/workspaces/{id}", s.deleteWorkspace)
			r.With(wsReq).Get("/workspaces/{id}/members", s.listMembers)
			r.With(wsReq).Post("/workspaces/{id}/members", s.addMember)
			r.With(wsReq).Post("/workspaces/{id}/invites", s.createInvite)
			r.Post("/invites/{token}/accept", s.acceptInvite)

			r.With(wsReq).Get("/projects", s.listProjects)
			r.With(wsReq).Post("/projects", s.createProject)
			r.With(wsReq).Get("/projects/{id}", s.getProject)
			r.With(wsReq).Patch("/projects/{id}", s.updateProject)
			r.With(wsReq).Delete("/projects/{id}", s.deleteProject)

			r.With(wsReq).Get("/projects/{projectId}/columns", s.listColumns)
			r.With(wsReq).Post("/projects/{projectId}/columns", s.createColumn)
			r.With(wsReq).Patch("/columns/{id}", s.updateColumn)
			r.With(wsReq).Delete("/columns/{id}", s.deleteColumn)

			r.With(wsReq).Get("/tasks", s.listTasks)
			r.With(wsReq).Post("/tasks", s.createTask)
			r.With(wsReq).Get("/tasks/{id}", s.getTask)
			r.With(wsReq).Patch("/tasks/{id}", s.patchTask)
			r.With(wsReq).Delete("/tasks/{id}", s.deleteTask)
			r.With(wsReq).Post("/tasks/{id}/move", s.moveTask)
			r.With(wsReq).Get("/tasks/{id}/comments", s.listComments)
			r.With(wsReq).Post("/tasks/{id}/comments", s.addComment)
			r.With(wsReq).Get("/tasks/{id}/checklists", s.listChecklists)
			r.With(wsReq).Post("/tasks/{id}/checklists", s.addChecklist)
			r.With(wsReq).Patch("/checklists/{id}", s.toggleChecklist)
			r.With(wsReq).Post("/tasks/{id}/attachments", s.addAttachment)
			r.With(wsReq).Get("/tasks/{id}/attachments", s.listAttachments)

			r.With(wsReq).Get("/activities", s.listActivities)
			r.With(wsOpt).Get("/notifications", s.listNotifications)
			r.With(wsOpt).Patch("/notifications/{id}/read", s.readNotification)
			r.With(wsReq).Get("/dashboard", s.dashboard)
			r.With(wsReq).Get("/search", s.search)
		})

		// WS: auth via Bearer or ?token=, workspace via ?workspace=
		r.Get("/ws", s.serveWS)
	})
	// top-level alias (frontend useWS connects here)
	r.Get("/ws", s.serveWS)
	return r
}

func ok(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func err(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": msg}})
}

func decode(r *http.Request, v any) bool {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v) == nil
}

func uidOf(r *http.Request) string  { return middleware.UserID(r) }
func roleOf(r *http.Request) string { return middleware.Role(r) }
func orgOf(r *http.Request) string  { return middleware.OrgID(r) }
