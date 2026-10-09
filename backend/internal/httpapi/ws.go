package httpapi

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/PresleyGSilva/taskforge-go/internal/auth"
	"github.com/PresleyGSilva/taskforge-go/internal/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	token := ""
	if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
		token = h[7:]
	} else {
		token = r.URL.Query().Get("token")
	}
	uid, uerr := auth.Verify(s.JWTSecret, token)
	if uerr != nil {
		err(w, 401, "UNAUTHORIZED", "invalid token")
		return
	}
	workspace := r.URL.Query().Get("workspace")
	if workspace == "" {
		err(w, 400, "VALIDATION_ERROR", "workspace query param required")
		return
	}
	role, _ := s.Store.MemberRole(r.Context(), uid, workspace)
	if role == "" {
		err(w, 403, "FORBIDDEN", "not a workspace member")
		return
	}
	conn, uerr2 := upgrader.Upgrade(w, r, nil)
	if uerr2 != nil {
		return
	}
	c := &ws.Client{Workspace: workspace, Send: make(chan []byte, 64)}
	s.Hub.Add(c)
	defer func() {
		s.Hub.Remove(c)
		conn.Close()
	}()
	go func() {
		for msg := range c.Send {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()
	for {
		if _, _, rerr := conn.ReadMessage(); rerr != nil {
			break
		}
	}
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	for _, p := range []string{"openapi.yaml", "backend/openapi.yaml", "../openapi.yaml"} {
		if _, serr := os.Stat(p); serr == nil {
			http.ServeFile(w, r, p)
			return
		}
	}
	// fallback: search upward
	if abs, aerr := filepath.Abs("."); aerr == nil {
		_ = abs
	}
	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(404)
	_, _ = w.Write([]byte("openapi.yaml not found"))
}
