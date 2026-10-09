package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/PresleyGSilva/taskforge-go/internal/auth"
	"github.com/PresleyGSilva/taskforge-go/internal/store"
)

// CORS allows the web frontend (and local dev origins) to call the API.
// Allowed origins come from FRONTEND_URL env (comma-separated) plus common
// localhost dev ports; defaults to permissive localhost matching.
func CORS(allowed []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Org-ID, X-Workspace-ID, X-User")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a != "" && (a == "*" || a == origin) {
			return true
		}
	}
	// default: any localhost/127.0.0.1 origin (dev + docker port mapping)
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		return true
	}
	return false
}

type ctxKey string

const (
	CtxUserID ctxKey = "user_id"
	CtxRole   ctxKey = "role"
	CtxOrgID  ctxKey = "org_id"
)

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + msg + `"}}`))
}

func Auth(secret string, st store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimPrefix(h, "Bearer ")
			} else if q := r.URL.Query().Get("token"); q != "" {
				token = q
			}
			if token == "" {
				writeErr(w, 401, "UNAUTHORIZED", "missing token")
				return
			}
			uid, err := auth.Verify(secret, token)
			if err != nil {
				writeErr(w, 401, "UNAUTHORIZED", "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), CtxUserID, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Workspace loads X-Org-ID (alias X-Workspace-ID) membership role into context.
// If require is true, org header is mandatory.
func Workspace(st store.Store, require bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.Header.Get("X-Org-ID")
			if org == "" {
				org = r.Header.Get("X-Workspace-ID")
			}
			if org == "" {
				if require {
					writeErr(w, 400, "VALIDATION_ERROR", "missing X-Org-ID")
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			uid, _ := r.Context().Value(CtxUserID).(string)
			role, err := st.MemberRole(r.Context(), uid, org)
			if err != nil || role == "" {
				writeErr(w, 403, "FORBIDDEN", "not a workspace member")
				return
			}
			ctx := context.WithValue(r.Context(), CtxOrgID, org)
			ctx = context.WithValue(ctx, CtxRole, role)
			(next.ServeHTTP(w, r.WithContext(ctx)))
		})
	}
}

func UserID(r *http.Request) string {
	v, _ := r.Context().Value(CtxUserID).(string)
	return v
}

func Role(r *http.Request) string {
	v, _ := r.Context().Value(CtxRole).(string)
	return v
}

func OrgID(r *http.Request) string {
	v, _ := r.Context().Value(CtxOrgID).(string)
	return v
}

func RoleLevel(role string) int {
	switch role {
	case "owner":
		return 4
	case "admin":
		return 3
	case "member":
		return 2
	case "viewer":
		return 1
	}
	return 0
}

// CanWrite: owner/admin/member may write; viewer read-only.
func CanWrite(role string) bool { return RoleLevel(role) >= 2 }

// CanManage: owner/admin only (members, project delete).
func CanManage(role string) bool { return RoleLevel(role) >= 3 }

func RequireWrite(w http.ResponseWriter, r *http.Request) bool {
	if !CanWrite(Role(r)) {
		writeErr(w, 403, "FORBIDDEN", "viewer cannot write")
		return false
	}
	return true
}

func RequireManage(w http.ResponseWriter, r *http.Request) bool {
	if !CanManage(Role(r)) {
		writeErr(w, 403, "FORBIDDEN", "requires owner or admin")
		return false
	}
	return true
}
