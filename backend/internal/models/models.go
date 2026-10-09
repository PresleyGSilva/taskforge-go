package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type OrgRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	UserID      string `json:"user_id"`
	WorkspaceID string `json:"workspace_id"`
	Role        string `json:"role"`
}

type Project struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Column struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Position  int    `json:"position"`
}

type Task struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	ColumnID    string     `json:"column_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Position    int        `json:"position"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Checklist struct {
	ID       string `json:"id"`
	TaskID   string `json:"task_id"`
	Title    string `json:"title"`
	Done     bool   `json:"done"`
	Position int    `json:"position"`
}

type Comment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Attachment struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	Mime       string    `json:"mime"`
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type Activity struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	TaskID      *string   `json:"task_id,omitempty"`
	UserID      *string   `json:"user_id,omitempty"`
	Action      string    `json:"action"`
	Detail      string    `json:"detail,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Notification struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	WorkspaceID string    `json:"workspace_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body,omitempty"`
	Read        bool      `json:"read"`
	CreatedAt   time.Time `json:"created_at"`
}

type Invite struct {
	Token       string    `json:"token"`
	WorkspaceID string    `json:"workspace_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Accepted    bool      `json:"accepted"`
	CreatedAt   time.Time `json:"created_at"`
}
