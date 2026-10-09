package store

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PresleyGSilva/taskforge-go/internal/models"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("version conflict")
var ErrExists = errors.New("exists")

// Fake is an in-memory Store for tests.
type Fake struct {
	mu            sync.Mutex
	users         map[string]models.User
	byEmail       map[string]string
	workspaces    map[string]models.Workspace
	members       map[string]string // userID + "|" + workspaceID -> role
	projects      map[string]models.Project
	columns       map[string]models.Column
	tasks         map[string]models.Task
	comments      map[string][]models.Comment
	checklists    map[string][]models.Checklist
	attachments   map[string][]models.Attachment
	activities    []models.Activity
	notifications map[string]models.Notification
	invites       map[string]models.Invite
}

func NewFake() *Fake {
	return &Fake{
		users: map[string]models.User{}, byEmail: map[string]string{},
		workspaces: map[string]models.Workspace{}, members: map[string]string{},
		projects: map[string]models.Project{}, columns: map[string]models.Column{},
		tasks: map[string]models.Task{}, comments: map[string][]models.Comment{},
		checklists: map[string][]models.Checklist{}, attachments: map[string][]models.Attachment{},
		notifications: map[string]models.Notification{}, invites: map[string]models.Invite{},
	}
}

func nid() string { return uuid.NewString() }

func (f *Fake) CreateUser(ctx context.Context, name, email, hash string) (models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byEmail[strings.ToLower(email)]; ok {
		return models.User{}, ErrExists
	}
	u := models.User{ID: nid(), Name: name, Email: email, PasswordHash: hash, CreatedAt: time.Now()}
	f.users[u.ID] = u
	f.byEmail[strings.ToLower(email)] = u.ID
	return u, nil
}

func (f *Fake) UserByEmail(ctx context.Context, email string) (models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.byEmail[strings.ToLower(email)]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return f.users[id], nil
}

func (f *Fake) UserByID(ctx context.Context, id string) (models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return u, nil
}

func (f *Fake) CreateWorkspace(ctx context.Context, name, ownerID string) (models.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w := models.Workspace{ID: nid(), Name: name, CreatedAt: time.Now()}
	f.workspaces[w.ID] = w
	f.members[ownerID+"|"+w.ID] = "owner"
	return w, nil
}

func (f *Fake) ListWorkspaces(ctx context.Context, userID string) ([]models.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Workspace
	for _, w := range f.workspaces {
		if _, ok := f.members[userID+"|"+w.ID]; ok {
			out = append(out, w)
		}
	}
	return out, nil
}

func (f *Fake) GetWorkspace(ctx context.Context, id string) (models.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w, ok := f.workspaces[id]
	if !ok {
		return models.Workspace{}, ErrNotFound
	}
	return w, nil
}

func (f *Fake) UpdateWorkspace(ctx context.Context, id, name string) (models.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w, ok := f.workspaces[id]
	if !ok {
		return models.Workspace{}, ErrNotFound
	}
	w.Name = name
	f.workspaces[id] = w
	return w, nil
}

func (f *Fake) DeleteWorkspace(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.workspaces, id)
	return nil
}

func (f *Fake) MemberRole(ctx context.Context, userID, workspaceID string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.members[userID+"|"+workspaceID]
	if !ok {
		return "", nil
	}
	return r, nil
}

func (f *Fake) ListMembers(ctx context.Context, workspaceID string) ([]models.Member, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Member
	for k, r := range f.members {
		parts := strings.SplitN(k, "|", 2)
		if len(parts) == 2 && parts[1] == workspaceID {
			out = append(out, models.Member{UserID: parts[0], WorkspaceID: workspaceID, Role: r})
		}
	}
	return out, nil
}

func (f *Fake) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.members[userID+"|"+workspaceID] = role
	return nil
}

func (f *Fake) CreateInvite(ctx context.Context, inv models.Invite) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.invites[inv.Token] = inv
	return nil
}

func (f *Fake) GetInvite(ctx context.Context, token string) (models.Invite, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	inv, ok := f.invites[token]
	if !ok {
		return models.Invite{}, ErrNotFound
	}
	return inv, nil
}

func (f *Fake) AcceptInvite(ctx context.Context, token, userID string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	inv, ok := f.invites[token]
	if !ok {
		return "", ErrNotFound
	}
	inv.Accepted = true
	f.invites[token] = inv
	f.members[userID+"|"+inv.WorkspaceID] = inv.Role
	return inv.WorkspaceID, nil
}

func (f *Fake) CreateProject(ctx context.Context, workspaceID, name, desc string) (models.Project, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p := models.Project{ID: nid(), WorkspaceID: workspaceID, Name: name, Description: desc, CreatedAt: time.Now()}
	f.projects[p.ID] = p
	return p, nil
}

func (f *Fake) ListProjects(ctx context.Context, workspaceID string) ([]models.Project, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Project
	for _, p := range f.projects {
		if workspaceID == "" || p.WorkspaceID == workspaceID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *Fake) GetProject(ctx context.Context, id string) (models.Project, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.projects[id]
	if !ok {
		return models.Project{}, ErrNotFound
	}
	return p, nil
}

func (f *Fake) UpdateProject(ctx context.Context, id, name, desc string) (models.Project, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.projects[id]
	if !ok {
		return models.Project{}, ErrNotFound
	}
	if name != "" {
		p.Name = name
	}
	p.Description = desc
	f.projects[id] = p
	return p, nil
}

func (f *Fake) DeleteProject(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.projects, id)
	return nil
}

func (f *Fake) CreateColumn(ctx context.Context, projectID, name string, position int) (models.Column, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := models.Column{ID: nid(), ProjectID: projectID, Name: name, Position: position}
	f.columns[c.ID] = c
	return c, nil
}

func (f *Fake) ListColumns(ctx context.Context, projectID string) ([]models.Column, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Column
	for _, c := range f.columns {
		if c.ProjectID == projectID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (f *Fake) UpdateColumn(ctx context.Context, id, name string, position *int) (models.Column, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.columns[id]
	if !ok {
		return models.Column{}, ErrNotFound
	}
	if name != "" {
		c.Name = name
	}
	if position != nil {
		c.Position = *position
	}
	f.columns[id] = c
	return c, nil
}

func (f *Fake) DeleteColumn(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.columns, id)
	return nil
}

func (f *Fake) CreateTask(ctx context.Context, t models.Task) (models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t.ID = nid()
	if t.Version == 0 {
		t.Version = 1
	}
	t.CreatedAt = time.Now()
	f.tasks[t.ID] = t
	return t, nil
}

func (f *Fake) ListTasks(ctx context.Context, fl TaskFilter) ([]models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Task
	for _, t := range f.tasks {
		if fl.ProjectID != "" && t.ProjectID != fl.ProjectID {
			continue
		}
		if fl.ColumnID != "" && t.ColumnID != fl.ColumnID {
			continue
		}
		if fl.Priority != "" && t.Priority != fl.Priority {
			continue
		}
		if fl.Assignee != "" && (t.AssigneeID == nil || *t.AssigneeID != fl.Assignee) {
			continue
		}
		if fl.Tag != "" {
			found := false
			for _, tg := range t.Tags {
				if tg == fl.Tag {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		if fl.Q != "" && !strings.Contains(strings.ToLower(t.Title+" "+t.Description), strings.ToLower(fl.Q)) {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (f *Fake) GetTask(ctx context.Context, id string) (models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

func applyPatch(t *models.Task, p TaskPatch) {
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Description != nil {
		t.Description = *p.Description
	}
	if p.Priority != nil {
		t.Priority = *p.Priority
	}
	if p.AssigneeID != nil {
		t.AssigneeID = *p.AssigneeID
	}
	if p.ColumnID != nil {
		t.ColumnID = *p.ColumnID
	}
	if p.Position != nil {
		t.Position = *p.Position
	}
	if p.DueDate != nil {
		t.DueDate = *p.DueDate
	}
	if p.HasTags {
		t.Tags = p.Tags
	}
}

func (f *Fake) PatchTask(ctx context.Context, id string, p TaskPatch) (models.Task, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return models.Task{}, false, ErrNotFound
	}
	if p.HasVersion && p.Version != t.Version {
		return t, true, ErrConflict
	}
	applyPatch(&t, p)
	t.Version++
	f.tasks[id] = t
	return t, false, nil
}

func (f *Fake) MoveTask(ctx context.Context, id, columnID string, position, version int) (models.Task, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return models.Task{}, false, ErrNotFound
	}
	if version != 0 && version != t.Version {
		return t, true, ErrConflict
	}
	t.ColumnID = columnID
	t.Position = position
	t.Version++
	f.tasks[id] = t
	return t, false, nil
}

func (f *Fake) DeleteTask(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.tasks, id)
	return nil
}

func (f *Fake) AddComment(ctx context.Context, taskID, userID, body string) (models.Comment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := models.Comment{ID: nid(), TaskID: taskID, UserID: userID, Body: body, CreatedAt: time.Now()}
	f.comments[taskID] = append(f.comments[taskID], c)
	return c, nil
}

func (f *Fake) ListComments(ctx context.Context, taskID string) ([]models.Comment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.comments[taskID], nil
}

func (f *Fake) AddChecklist(ctx context.Context, taskID, title string, position int) (models.Checklist, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := models.Checklist{ID: nid(), TaskID: taskID, Title: title, Position: position}
	f.checklists[taskID] = append(f.checklists[taskID], c)
	return c, nil
}

func (f *Fake) ListChecklists(ctx context.Context, taskID string) ([]models.Checklist, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.checklists[taskID], nil
}

func (f *Fake) ToggleChecklist(ctx context.Context, id string, done *bool, title *string) (models.Checklist, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for taskID, list := range f.checklists {
		for i, c := range list {
			if c.ID == id {
				if done != nil {
					c.Done = *done
				} else {
					c.Done = !c.Done
				}
				if title != nil {
					c.Title = *title
				}
				f.checklists[taskID][i] = c
				return c, nil
			}
		}
	}
	return models.Checklist{}, ErrNotFound
}

func (f *Fake) AddAttachment(ctx context.Context, a models.Attachment) (models.Attachment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a.ID = nid()
	a.CreatedAt = time.Now()
	f.attachments[a.TaskID] = append(f.attachments[a.TaskID], a)
	return a, nil
}

func (f *Fake) ListAttachments(ctx context.Context, taskID string) ([]models.Attachment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.attachments[taskID], nil
}

func (f *Fake) LogActivity(ctx context.Context, workspaceID string, taskID, userID *string, action, detail string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activities = append(f.activities, models.Activity{ID: nid(), WorkspaceID: workspaceID, TaskID: taskID, UserID: userID, Action: action, Detail: detail, CreatedAt: time.Now()})
	return nil
}

func (f *Fake) ListActivities(ctx context.Context, workspaceID, taskID string) ([]models.Activity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Activity
	for _, a := range f.activities {
		if workspaceID != "" && a.WorkspaceID != workspaceID {
			continue
		}
		if taskID != "" && (a.TaskID == nil || *a.TaskID != taskID) {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

func (f *Fake) ListNotifications(ctx context.Context, userID string) ([]models.Notification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Notification
	for _, n := range f.notifications {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (f *Fake) MarkNotificationRead(ctx context.Context, id, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, ok := f.notifications[id]
	if !ok {
		return ErrNotFound
	}
	n.Read = true
	f.notifications[id] = n
	return nil
}

func (f *Fake) CreateNotification(ctx context.Context, n models.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if n.ID == "" {
		n.ID = nid()
	}
	f.notifications[n.ID] = n
	return nil
}

func (f *Fake) Dashboard(ctx context.Context, workspaceID string) (map[string]any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	byCol := map[string]int{}
	byPri := map[string]int{}
	overdue := 0
	byMember := map[string]int{}
	now := time.Now()
	for _, t := range f.tasks {
		p, ok := f.projects[t.ProjectID]
		if !ok || p.WorkspaceID != workspaceID {
			continue
		}
		byCol[t.ColumnID]++
		byPri[t.Priority]++
		if t.DueDate != nil && t.DueDate.Before(now) {
			overdue++
		}
		if t.AssigneeID != nil {
			byMember[*t.AssigneeID]++
		}
	}
	return map[string]any{"by_column": byCol, "by_priority": byPri, "overdue": overdue, "by_member": byMember}, nil
}

func (f *Fake) Search(ctx context.Context, workspaceID, q string) (map[string]any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var tasks []models.Task
	for _, t := range f.tasks {
		p, ok := f.projects[t.ProjectID]
		if !ok || p.WorkspaceID != workspaceID {
			continue
		}
		if q == "" || strings.Contains(strings.ToLower(t.Title+" "+t.Description), strings.ToLower(q)) {
			tasks = append(tasks, t)
		}
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	return map[string]any{"tasks": tasks}, nil
}

// Seed helpers for tests.
func (f *Fake) SeedMember(userID, workspaceID, role string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.workspaces[workspaceID]; !ok {
		f.workspaces[workspaceID] = models.Workspace{ID: workspaceID, Name: "ws"}
	}
	if _, ok := f.users[userID]; !ok {
		f.users[userID] = models.User{ID: userID, Name: userID, Email: userID + "@x.com"}
	}
	f.members[userID+"|"+workspaceID] = role
}

func (f *Fake) SeedProject(p models.Project) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.projects[p.ID] = p
}

func (f *Fake) SeedTask(t models.Task) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t.Version == 0 {
		t.Version = 1
	}
	f.tasks[t.ID] = t
}
