package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PresleyGSilva/taskforge-go/internal/auth"
	"github.com/PresleyGSilva/taskforge-go/internal/httpapi"
	"github.com/PresleyGSilva/taskforge-go/internal/models"
	"github.com/PresleyGSilva/taskforge-go/internal/store"
	"github.com/PresleyGSilva/taskforge-go/internal/ws"
)

const testSecret = "test-secret"

func testServer() (*httpapi.Server, *store.Fake) {
	f := store.NewFake()
	return &httpapi.Server{Store: f, JWTSecret: testSecret, Hub: ws.NewHub()}, f
}

func tokenFor(uid string) string {
	t, _ := auth.SignAccess(testSecret, uid)
	return t
}

func doReq(t *testing.T, h http.Handler, method, path, token, org, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if org != "" {
		r.Header.Set("X-Org-ID", org)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	return rr
}

func TestWorkspacePermissions(t *testing.T) {
	s, f := testServer()
	h := httpapi.NewRouter(s)
	wsID := "ws-1"
	f.SeedMember("u-owner", wsID, "owner")
	f.SeedMember("u-member", wsID, "member")
	f.SeedMember("u-viewer", wsID, "viewer")
	f.SeedProject(models.Project{ID: "p-1", WorkspaceID: wsID, Name: "P"})

	// viewer 403 on POST task
	rr := doReq(t, h, "POST", "/api/v1/tasks", tokenFor("u-viewer"), wsID, `{"project_id":"p-1","column_id":"c-1","title":"x"}`)
	if rr.Code != 403 {
		t.Fatalf("viewer POST task: got %d want 403", rr.Code)
	}

	// member 403 on DELETE project
	rr = doReq(t, h, "DELETE", "/api/v1/projects/p-1", tokenFor("u-member"), wsID, "")
	if rr.Code != 403 {
		t.Fatalf("member DELETE project: got %d want 403", rr.Code)
	}

	// owner 200 on DELETE project
	rr = doReq(t, h, "DELETE", "/api/v1/projects/p-1", tokenFor("u-owner"), wsID, "")
	if rr.Code != 200 {
		t.Fatalf("owner DELETE project: got %d want 200", rr.Code)
	}
}

func TestOptimisticConcurrency(t *testing.T) {
	s, f := testServer()
	h := httpapi.NewRouter(s)
	wsID := "ws-1"
	f.SeedMember("u-owner", wsID, "owner")
	f.SeedProject(models.Project{ID: "p-1", WorkspaceID: wsID, Name: "P"})
	f.SeedTask(models.Task{ID: "t-1", ProjectID: "p-1", ColumnID: "c-1", Title: "T", Version: 1})

	// stale version -> 409
	rr := doReq(t, h, "PATCH", "/api/v1/tasks/t-1", tokenFor("u-owner"), wsID, `{"title":"new","version":99}`)
	if rr.Code != 409 {
		t.Fatalf("stale PATCH: got %d want 409", rr.Code)
	}

	// correct version -> 200 and increments
	rr = doReq(t, h, "PATCH", "/api/v1/tasks/t-1", tokenFor("u-owner"), wsID, `{"title":"new","version":1}`)
	if rr.Code != 200 {
		t.Fatalf("ok PATCH: got %d want 200 body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Data models.Task `json:"data"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Data.Version != 2 {
		t.Fatalf("version not incremented: got %d want 2", resp.Data.Version)
	}
	if resp.Data.Title != "new" {
		t.Fatalf("title not updated")
	}
}

func TestWSHub(t *testing.T) {
	hub := ws.NewHub()
	a := &ws.Client{Workspace: "ws-1", Send: make(chan []byte, 4)}
	b := &ws.Client{Workspace: "ws-1", Send: make(chan []byte, 4)}
	other := &ws.Client{Workspace: "ws-2", Send: make(chan []byte, 4)}
	hub.Add(a)
	hub.Add(b)
	hub.Add(other)

	hub.Emit("ws-1", "task.created", map[string]string{"id": "t-1"}, nil)

	for _, c := range []*ws.Client{a, b} {
		select {
		case msg := <-c.Send:
			var m map[string]any
			if err := json.Unmarshal(msg, &m); err != nil {
				t.Fatalf("bad json: %v", err)
			}
			if m["type"] != "task.created" {
				t.Fatalf("wrong type: %v", m["type"])
			}
		default:
			t.Fatalf("client in ws-1 did not receive broadcast")
		}
	}
	select {
	case msg := <-other.Send:
		t.Fatalf("client in other workspace received broadcast: %s", msg)
	default:
	}
}
