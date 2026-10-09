package store

import (
	"context"
	"time"

	"github.com/PresleyGSilva/taskforge-go/internal/models"
)

type TaskFilter struct {
	ProjectID string
	ColumnID  string
	Q         string
	Priority  string
	Assignee  string
	Tag       string
}

type TaskPatch struct {
	Title       *string
	Description *string
	Priority    *string
	AssigneeID  **string
	ColumnID    *string
	Position    *int
	DueDate     **time.Time
	Tags        []string
	HasTags     bool
	Version     int
	HasVersion  bool
}

type Store interface {
	// users
	CreateUser(ctx context.Context, name, email, hash string) (models.User, error)
	UserByEmail(ctx context.Context, email string) (models.User, error)
	UserByID(ctx context.Context, id string) (models.User, error)
	// workspaces
	CreateWorkspace(ctx context.Context, name, ownerID string) (models.Workspace, error)
	ListWorkspaces(ctx context.Context, userID string) ([]models.Workspace, error)
	GetWorkspace(ctx context.Context, id string) (models.Workspace, error)
	UpdateWorkspace(ctx context.Context, id, name string) (models.Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	MemberRole(ctx context.Context, userID, workspaceID string) (string, error)
	ListMembers(ctx context.Context, workspaceID string) ([]models.Member, error)
	AddMember(ctx context.Context, workspaceID, userID, role string) error
	// invites
	CreateInvite(ctx context.Context, inv models.Invite) error
	GetInvite(ctx context.Context, token string) (models.Invite, error)
	AcceptInvite(ctx context.Context, token, userID string) (string, error)
	// projects
	CreateProject(ctx context.Context, workspaceID, name, desc string) (models.Project, error)
	ListProjects(ctx context.Context, workspaceID string) ([]models.Project, error)
	GetProject(ctx context.Context, id string) (models.Project, error)
	UpdateProject(ctx context.Context, id, name, desc string) (models.Project, error)
	DeleteProject(ctx context.Context, id string) error
	// columns
	CreateColumn(ctx context.Context, projectID, name string, position int) (models.Column, error)
	ListColumns(ctx context.Context, projectID string) ([]models.Column, error)
	UpdateColumn(ctx context.Context, id, name string, position *int) (models.Column, error)
	DeleteColumn(ctx context.Context, id string) error
	// tasks
	CreateTask(ctx context.Context, t models.Task) (models.Task, error)
	ListTasks(ctx context.Context, f TaskFilter) ([]models.Task, error)
	GetTask(ctx context.Context, id string) (models.Task, error)
	PatchTask(ctx context.Context, id string, p TaskPatch) (models.Task, bool, error)
	MoveTask(ctx context.Context, id, columnID string, position, version int) (models.Task, bool, error)
	DeleteTask(ctx context.Context, id string) error
	// comments
	AddComment(ctx context.Context, taskID, userID, body string) (models.Comment, error)
	ListComments(ctx context.Context, taskID string) ([]models.Comment, error)
	// checklists
	AddChecklist(ctx context.Context, taskID, title string, position int) (models.Checklist, error)
	ListChecklists(ctx context.Context, taskID string) ([]models.Checklist, error)
	ToggleChecklist(ctx context.Context, id string, done *bool, title *string) (models.Checklist, error)
	// attachments
	AddAttachment(ctx context.Context, a models.Attachment) (models.Attachment, error)
	ListAttachments(ctx context.Context, taskID string) ([]models.Attachment, error)
	// activities
	LogActivity(ctx context.Context, workspaceID string, taskID, userID *string, action, detail string) error
	ListActivities(ctx context.Context, workspaceID, taskID string) ([]models.Activity, error)
	// notifications
	ListNotifications(ctx context.Context, userID string) ([]models.Notification, error)
	MarkNotificationRead(ctx context.Context, id, userID string) error
	CreateNotification(ctx context.Context, n models.Notification) error
	// dashboard & search
	Dashboard(ctx context.Context, workspaceID string) (map[string]any, error)
	Search(ctx context.Context, workspaceID, q string) (map[string]any, error)
}
