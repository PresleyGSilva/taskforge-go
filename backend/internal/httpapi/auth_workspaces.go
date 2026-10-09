package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/PresleyGSilva/taskforge-go/internal/auth"
	"github.com/PresleyGSilva/taskforge-go/internal/models"
	"github.com/PresleyGSilva/taskforge-go/internal/store"
	"github.com/google/uuid"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(r, &in) || in.Email == "" || in.Password == "" || in.Name == "" {
		err(w, 400, "VALIDATION_ERROR", "name, email and password are required")
		return
	}
	hash, _ := auth.HashPassword(in.Password)
	u, aerr := s.Store.CreateUser(r.Context(), in.Name, in.Email, hash)
	if aerr != nil {
		err(w, 409, "CONFLICT", "email already registered")
		return
	}
	s.finishAuth(w, r, u)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(r, &in) || in.Email == "" || in.Password == "" {
		err(w, 400, "VALIDATION_ERROR", "email and password are required")
		return
	}
	u, aerr := s.Store.UserByEmail(r.Context(), in.Email)
	if aerr != nil || auth.CheckPassword(u.PasswordHash, in.Password) != nil {
		err(w, 401, "UNAUTHORIZED", "invalid credentials")
		return
	}
	s.finishAuth(w, r, u)
}

func (s *Server) finishAuth(w http.ResponseWriter, r *http.Request, u models.User) {
	access, _ := auth.SignAccess(s.JWTSecret, u.ID)
	refresh, _ := auth.SignRefresh(s.JWTSecret, u.ID)
	workspaces, _ := s.Store.ListWorkspaces(r.Context(), u.ID)
	orgs := []models.OrgRef{}
	for _, ws := range workspaces {
		orgs = append(orgs, models.OrgRef{ID: ws.ID, Name: ws.Name})
	}
	if orgs == nil {
		orgs = []models.OrgRef{}
	}
	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token", Value: refresh, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(7 * 24 * time.Hour),
	})
	ok(w, map[string]any{
		"user":         map[string]string{"id": u.ID, "name": u.Name, "email": u.Email},
		"access_token": access,
		"orgs":         orgs,
	})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	c, cerr := r.Cookie("refresh_token")
	if cerr != nil {
		err(w, 401, "UNAUTHORIZED", "missing refresh token")
		return
	}
	uid, verr := auth.Verify(s.JWTSecret, c.Value)
	if verr != nil {
		err(w, 401, "UNAUTHORIZED", "invalid refresh token")
		return
	}
	u, uerr := s.Store.UserByID(r.Context(), uid)
	if uerr != nil {
		if se, ok := s.Store.(*store.Fake); ok {
			_ = se
		}
		err(w, 401, "UNAUTHORIZED", "user not found")
		return
	}
	access, _ := auth.SignAccess(s.JWTSecret, u.ID)
	ok(w, map[string]any{"access_token": access})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	ok(w, map[string]bool{"ok": true})
}

// workspaces

func (s *Server) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListWorkspaces(r.Context(), uidOf(r))
	if list == nil {
		list = []models.Workspace{}
	}
	ok(w, list)
}

func (s *Server) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Name) == "" {
		err(w, 400, "VALIDATION_ERROR", "name is required")
		return
	}
	ws, aerr := s.Store.CreateWorkspace(r.Context(), in.Name, uidOf(r))
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot create workspace")
		return
	}
	w.WriteHeader(201)
	ok(w, ws)
}

func (s *Server) getWorkspace(w http.ResponseWriter, r *http.Request) {
	ws, aerr := s.Store.GetWorkspace(r.Context(), chiID(r, "id"))
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "workspace not found")
		return
	}
	ok(w, ws)
}

func (s *Server) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireManage(w, r) {
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	_ = decode(r, &in)
	ws, aerr := s.Store.UpdateWorkspace(r.Context(), chiID(r, "id"), in.Name)
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "workspace not found")
		return
	}
	ok(w, ws)
}

func (s *Server) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	if !requireManage(w, r) {
		return
	}
	_ = s.Store.DeleteWorkspace(r.Context(), chiID(r, "id"))
	ok(w, map[string]bool{"ok": true})
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	m, _ := s.Store.ListMembers(r.Context(), orgOf(r))
	if m == nil {
		m = []models.Member{}
	}
	ok(w, m)
}

func (s *Server) addMember(w http.ResponseWriter, r *http.Request) {
	if !requireManage(w, r) {
		return
	}
	var in struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}
	_ = decode(r, &in)
	if in.Role == "" {
		in.Role = "member"
	}
	uid := in.UserID
	if uid == "" && in.Email != "" {
		if u, uerr := s.Store.UserByEmail(r.Context(), in.Email); uerr == nil {
			uid = u.ID
		} else {
			err(w, 404, "NOT_FOUND", "user not found")
			return
		}
	}
	if uid == "" {
		err(w, 400, "VALIDATION_ERROR", "user_id or email required")
		return
	}
	_ = s.Store.AddMember(r.Context(), orgOf(r), uid, in.Role)
	ok(w, map[string]bool{"ok": true})
}

func (s *Server) createInvite(w http.ResponseWriter, r *http.Request) {
	if !requireManage(w, r) {
		return
	}
	var in struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if !decode(r, &in) || in.Email == "" {
		err(w, 400, "VALIDATION_ERROR", "email is required")
		return
	}
	if in.Role == "" {
		in.Role = "member"
	}
	token := uuid.NewString()
	_ = s.Store.CreateInvite(r.Context(), models.Invite{Token: token, WorkspaceID: orgOf(r), Email: in.Email, Role: in.Role})
	w.WriteHeader(201)
	ok(w, map[string]string{"token": token})
}

func (s *Server) acceptInvite(w http.ResponseWriter, r *http.Request) {
	wsID, aerr := s.Store.AcceptInvite(r.Context(), chiID(r, "token"), uidOf(r))
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "invite not found")
		return
	}
	ok(w, map[string]string{"workspace_id": wsID})
}
