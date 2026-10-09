package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/PresleyGSilva/taskforge-go/internal/middleware"
	"github.com/PresleyGSilva/taskforge-go/internal/models"
	"github.com/PresleyGSilva/taskforge-go/internal/store"
	"github.com/go-chi/chi/v5"
)

func chiID(r *http.Request, key string) string { return chi.URLParam(r, key) }

func requireWrite(w http.ResponseWriter, r *http.Request) bool {
	return middleware.RequireWrite(w, r)
}
func requireManage(w http.ResponseWriter, r *http.Request) bool {
	return middleware.RequireManage(w, r)
}

func isNotFound(err error) bool { return err == store.ErrNotFound }

// projects

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	wsID := r.URL.Query().Get("workspace_id")
	if wsID == "" {
		wsID = orgOf(r)
	}
	list, _ := s.Store.ListProjects(r.Context(), wsID)
	if list == nil {
		list = []models.Project{}
	}
	ok(w, list)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		WorkspaceID string `json:"workspace_id"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Name) == "" {
		err(w, 400, "VALIDATION_ERROR", "name is required")
		return
	}
	wsID := in.WorkspaceID
	if wsID == "" {
		wsID = orgOf(r)
	}
	pr, aerr := s.Store.CreateProject(r.Context(), wsID, in.Name, in.Description)
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot create project")
		return
	}
	w.WriteHeader(201)
	ok(w, pr)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	pr, aerr := s.Store.GetProject(r.Context(), chiID(r, "id"))
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "project not found")
		return
	}
	ok(w, pr)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	_ = decode(r, &in)
	pr, aerr := s.Store.UpdateProject(r.Context(), chiID(r, "id"), in.Name, in.Description)
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "project not found")
		return
	}
	ok(w, pr)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	if !requireManage(w, r) {
		return
	}
	_ = s.Store.DeleteProject(r.Context(), chiID(r, "id"))
	ok(w, map[string]bool{"ok": true})
}

// columns

func (s *Server) listColumns(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListColumns(r.Context(), chiID(r, "projectId"))
	if list == nil {
		list = []models.Column{}
	}
	ok(w, list)
}

func (s *Server) createColumn(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Name     string `json:"name"`
		Position int    `json:"position"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Name) == "" {
		err(w, 400, "VALIDATION_ERROR", "name is required")
		return
	}
	c, aerr := s.Store.CreateColumn(r.Context(), chiID(r, "projectId"), in.Name, in.Position)
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot create column")
		return
	}
	w.WriteHeader(201)
	ok(w, c)
}

func (s *Server) updateColumn(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Name     *string `json:"name"`
		Position *int    `json:"position"`
	}
	_ = decode(r, &in)
	name := ""
	if in.Name != nil {
		name = *in.Name
	}
	c, aerr := s.Store.UpdateColumn(r.Context(), chiID(r, "id"), name, in.Position)
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "column not found")
		return
	}
	s.Hub.Emit(orgOf(r), "column.updated", c, nil)
	ok(w, c)
}

func (s *Server) deleteColumn(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	_ = s.Store.DeleteColumn(r.Context(), chiID(r, "id"))
	ok(w, map[string]bool{"ok": true})
}

// tasks

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, _ := s.Store.ListTasks(r.Context(), store.TaskFilter{
		ProjectID: q.Get("project_id"), ColumnID: q.Get("column_id"), Q: q.Get("q"),
		Priority: q.Get("priority"), Assignee: q.Get("assignee"), Tag: q.Get("tag"),
	})
	if list == nil {
		list = []models.Task{}
	}
	ok(w, list)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		ProjectID   string   `json:"project_id"`
		ColumnID    string   `json:"column_id"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Priority    string   `json:"priority"`
		AssigneeID  *string  `json:"assignee_id"`
		Tags        []string `json:"tags"`
		DueDate     *string  `json:"due_date"`
		Position    int      `json:"position"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Title) == "" || in.ProjectID == "" || in.ColumnID == "" {
		err(w, 400, "VALIDATION_ERROR", "project_id, column_id and title are required")
		return
	}
	var due *time.Time
	if in.DueDate != nil && *in.DueDate != "" {
		if t, terr := time.Parse(time.RFC3339, *in.DueDate); terr == nil {
			due = &t
		}
	}
	t, aerr := s.Store.CreateTask(r.Context(), models.Task{
		ProjectID: in.ProjectID, ColumnID: in.ColumnID, Title: in.Title,
		Description: in.Description, Priority: in.Priority, AssigneeID: in.AssigneeID,
		Tags: in.Tags, DueDate: due, Position: in.Position, Version: 1,
	})
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot create task")
		return
	}
	uid := uidOf(r)
	org := orgOf(r)
	_ = s.Store.LogActivity(r.Context(), org, &t.ID, &uid, "task.created", t.Title)
	s.Hub.Emit(org, "task.created", t, nil)
	w.WriteHeader(201)
	w.Header().Set("ETag", etag(t.Version))
	ok(w, t)
}

func etag(v int) string {
	return `W/"` + strings.TrimSpace(strings.Repeat(" ", 0)) + `"v` + itoa(v) + `"`
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	t, aerr := s.Store.GetTask(r.Context(), chiID(r, "id"))
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "task not found")
		return
	}
	w.Header().Set("ETag", etag(t.Version))
	ok(w, t)
}

func (s *Server) patchTask(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Priority    *string  `json:"priority"`
		AssigneeID  *string  `json:"assignee_id"`
		ColumnID    *string  `json:"column_id"`
		Position    *int     `json:"position"`
		DueDate     *string  `json:"due_date"`
		Tags        []string `json:"tags"`
		Version     *int     `json:"version"`
	}
	raw := map[string]any{}
	_ = decode(r, &raw)
	// re-decode into struct via second pass is overkill; decode again not possible. Parse from raw map manually:
	_ = in
	p := store.TaskPatch{}
	if v, ok := raw["title"].(string); ok {
		vv := v
		p.Title = &vv
	}
	if v, ok := raw["description"].(string); ok {
		vv := v
		p.Description = &vv
	}
	if v, ok := raw["priority"].(string); ok {
		vv := v
		p.Priority = &vv
	}
	if v, has := raw["assignee_id"]; has {
		if v == nil {
			p.AssigneeID = new(*string)
		} else if vs, ok := v.(string); ok {
			pp := &vs
			p.AssigneeID = &pp
		}
	}
	if v, ok := raw["column_id"].(string); ok {
		vv := v
		p.ColumnID = &vv
	}
	if v, ok := raw["position"].(float64); ok {
		vv := int(v)
		p.Position = &vv
	}
	if v, has := raw["due_date"]; has {
		if v == nil {
			p.DueDate = new(*time.Time)
		} else if vs, ok := v.(string); ok && vs != "" {
			if t, terr := time.Parse(time.RFC3339, vs); terr == nil {
				pp := &t
				p.DueDate = &pp
			}
		}
	}
	if v, has := raw["tags"]; has && v != nil {
		p.HasTags = true
		if arr, ok := v.([]any); ok {
			for _, e := range arr {
				if es, ok := e.(string); ok {
					p.Tags = append(p.Tags, es)
				}
			}
		}
	}
	if v, ok := raw["version"].(float64); ok {
		vv := int(v)
		p.Version = vv
		p.HasVersion = true
	}
	if !p.HasVersion {
		err(w, 400, "VALIDATION_ERROR", "version is required")
		return
	}
	t, conflict, aerr := s.Store.PatchTask(r.Context(), chiID(r, "id"), p)
	if aerr != nil {
		if conflict {
			w.Header().Set("ETag", etag(t.Version))
			err(w, 409, "CONFLICT", "version mismatch")
			return
		}
		if isNotFound(aerr) {
			err(w, 404, "NOT_FOUND", "task not found")
			return
		}
		err(w, 400, "VALIDATION_ERROR", "cannot update task")
		return
	}
	uid := uidOf(r)
	org := orgOf(r)
	_ = s.Store.LogActivity(r.Context(), org, &t.ID, &uid, "task.updated", t.Title)
	s.Hub.Emit(org, "task.updated", t, nil)
	w.Header().Set("ETag", etag(t.Version))
	ok(w, t)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	id := chiID(r, "id")
	_ = s.Store.DeleteTask(r.Context(), id)
	uid := uidOf(r)
	org := orgOf(r)
	_ = s.Store.LogActivity(r.Context(), org, &id, &uid, "task.deleted", id)
	s.Hub.Emit(org, "task.deleted", map[string]string{"id": id}, nil)
	ok(w, map[string]bool{"ok": true})
}

func (s *Server) moveTask(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		ColumnID string `json:"column_id"`
		Position int    `json:"position"`
		Version  int    `json:"version"`
	}
	if !decode(r, &in) || in.ColumnID == "" {
		err(w, 400, "VALIDATION_ERROR", "column_id is required")
		return
	}
	t, conflict, aerr := s.Store.MoveTask(r.Context(), chiID(r, "id"), in.ColumnID, in.Position, in.Version)
	if aerr != nil {
		if conflict {
			err(w, 409, "CONFLICT", "version mismatch")
			return
		}
		if isNotFound(aerr) {
			err(w, 404, "NOT_FOUND", "task not found")
			return
		}
		err(w, 400, "VALIDATION_ERROR", "cannot move task")
		return
	}
	uid := uidOf(r)
	org := orgOf(r)
	_ = s.Store.LogActivity(r.Context(), org, &t.ID, &uid, "task.moved", t.ColumnID)
	s.Hub.Emit(org, "task.moved", t, nil)
	w.Header().Set("ETag", etag(t.Version))
	ok(w, t)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListComments(r.Context(), chiID(r, "id"))
	if list == nil {
		list = []models.Comment{}
	}
	ok(w, list)
}

func (s *Server) addComment(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Body) == "" {
		err(w, 400, "VALIDATION_ERROR", "body is required")
		return
	}
	c, aerr := s.Store.AddComment(r.Context(), chiID(r, "id"), uidOf(r), in.Body)
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot add comment")
		return
	}
	tid := chiID(r, "id")
	uid := uidOf(r)
	org := orgOf(r)
	_ = s.Store.LogActivity(r.Context(), org, &tid, &uid, "comment.added", in.Body)
	s.Hub.Emit(org, "comment.added", c, nil)
	w.WriteHeader(201)
	ok(w, c)
}

func (s *Server) listChecklists(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListChecklists(r.Context(), chiID(r, "id"))
	if list == nil {
		list = []models.Checklist{}
	}
	ok(w, list)
}

func (s *Server) addChecklist(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Title    string `json:"title"`
		Position int    `json:"position"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Title) == "" {
		err(w, 400, "VALIDATION_ERROR", "title is required")
		return
	}
	c, aerr := s.Store.AddChecklist(r.Context(), chiID(r, "id"), in.Title, in.Position)
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot add checklist")
		return
	}
	w.WriteHeader(201)
	ok(w, c)
}

func (s *Server) toggleChecklist(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Done  *bool   `json:"done"`
		Title *string `json:"title"`
	}
	_ = decode(r, &in)
	c, aerr := s.Store.ToggleChecklist(r.Context(), chiID(r, "id"), in.Done, in.Title)
	if aerr != nil {
		err(w, 404, "NOT_FOUND", "checklist not found")
		return
	}
	ok(w, c)
}

func (s *Server) addAttachment(w http.ResponseWriter, r *http.Request) {
	if !requireWrite(w, r) {
		return
	}
	var in struct {
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		Mime     string `json:"mime"`
	}
	if !decode(r, &in) || strings.TrimSpace(in.Filename) == "" {
		err(w, 400, "VALIDATION_ERROR", "filename is required")
		return
	}
	a, aerr := s.Store.AddAttachment(r.Context(), models.Attachment{
		TaskID: chiID(r, "id"), Filename: in.Filename, Size: in.Size, Mime: in.Mime, UploadedBy: uidOf(r),
	})
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot add attachment")
		return
	}
	w.WriteHeader(201)
	ok(w, a)
}

func (s *Server) listAttachments(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListAttachments(r.Context(), chiID(r, "id"))
	if list == nil {
		list = []models.Attachment{}
	}
	ok(w, list)
}

func (s *Server) listActivities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	wsID := q.Get("workspace_id")
	if wsID == "" {
		wsID = orgOf(r)
	}
	list, _ := s.Store.ListActivities(r.Context(), wsID, q.Get("task_id"))
	if list == nil {
		list = []models.Activity{}
	}
	ok(w, list)
}

func (s *Server) listNotifications(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListNotifications(r.Context(), uidOf(r))
	if list == nil {
		list = []models.Notification{}
	}
	ok(w, list)
}

func (s *Server) readNotification(w http.ResponseWriter, r *http.Request) {
	_ = s.Store.MarkNotificationRead(r.Context(), chiID(r, "id"), uidOf(r))
	ok(w, map[string]bool{"ok": true})
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	wsID := r.URL.Query().Get("workspace_id")
	if wsID == "" {
		wsID = orgOf(r)
	}
	d, aerr := s.Store.Dashboard(r.Context(), wsID)
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "cannot load dashboard")
		return
	}
	ok(w, d)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	wsID := q.Get("workspace_id")
	if wsID == "" {
		wsID = orgOf(r)
	}
	res, aerr := s.Store.Search(r.Context(), wsID, q.Get("q"))
	if aerr != nil {
		err(w, 400, "VALIDATION_ERROR", "search failed")
		return
	}
	ok(w, res)
}
