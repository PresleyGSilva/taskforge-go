package store

import (
	"context"
	"time"

	"github.com/PresleyGSilva/taskforge-go/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pgx struct{ Pool *pgxpool.Pool }

func NewPgx(pool *pgxpool.Pool) *Pgx { return &Pgx{Pool: pool} }

func (p *Pgx) CreateUser(ctx context.Context, name, email, hash string) (models.User, error) {
	var u models.User
	err := p.Pool.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES($1,$2,$3) RETURNING id,name,email,created_at`, name, email, hash).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		return u, err
	}
	u.PasswordHash = hash
	return u, nil
}

func (p *Pgx) UserByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := p.Pool.QueryRow(ctx, `SELECT id,name,email,password_hash,created_at FROM users WHERE lower(email)=lower($1)`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return u, ErrNotFound
	}
	return u, nil
}

func (p *Pgx) UserByID(ctx context.Context, id string) (models.User, error) {
	var u models.User
	err := p.Pool.QueryRow(ctx, `SELECT id,name,email,password_hash,created_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return u, ErrNotFound
	}
	return u, nil
}

func (p *Pgx) CreateWorkspace(ctx context.Context, name, ownerID string) (models.Workspace, error) {
	var w models.Workspace
	err := p.Pool.QueryRow(ctx, `INSERT INTO workspaces(name) VALUES($1) RETURNING id,name,created_at`, name).Scan(&w.ID, &w.Name, &w.CreatedAt)
	if err != nil {
		return w, err
	}
	_, err = p.Pool.Exec(ctx, `INSERT INTO workspace_members(user_id,workspace_id,role) VALUES($1,$2,'owner') ON CONFLICT (user_id,workspace_id) DO UPDATE SET role='owner'`, ownerID, w.ID)
	return w, err
}

func (p *Pgx) ListWorkspaces(ctx context.Context, userID string) ([]models.Workspace, error) {
	rows, err := p.Pool.Query(ctx, `SELECT w.id,w.name,w.created_at FROM workspaces w JOIN workspace_members m ON m.workspace_id=w.id WHERE m.user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Workspace
	for rows.Next() {
		var w models.Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (p *Pgx) GetWorkspace(ctx context.Context, id string) (models.Workspace, error) {
	var w models.Workspace
	err := p.Pool.QueryRow(ctx, `SELECT id,name,created_at FROM workspaces WHERE id=$1`, id).Scan(&w.ID, &w.Name, &w.CreatedAt)
	if err != nil {
		return w, ErrNotFound
	}
	return w, nil
}

func (p *Pgx) UpdateWorkspace(ctx context.Context, id, name string) (models.Workspace, error) {
	var w models.Workspace
	err := p.Pool.QueryRow(ctx, `UPDATE workspaces SET name=$2 WHERE id=$1 RETURNING id,name,created_at`, id, name).Scan(&w.ID, &w.Name, &w.CreatedAt)
	if err != nil {
		return w, ErrNotFound
	}
	return w, nil
}

func (p *Pgx) DeleteWorkspace(ctx context.Context, id string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM workspaces WHERE id=$1`, id)
	return err
}

func (p *Pgx) MemberRole(ctx context.Context, userID, workspaceID string) (string, error) {
	var role string
	err := p.Pool.QueryRow(ctx, `SELECT role FROM workspace_members WHERE user_id=$1 AND workspace_id=$2`, userID, workspaceID).Scan(&role)
	if err != nil {
		return "", nil
	}
	return role, nil
}

func (p *Pgx) ListMembers(ctx context.Context, workspaceID string) ([]models.Member, error) {
	rows, err := p.Pool.Query(ctx, `SELECT user_id,workspace_id,role FROM workspace_members WHERE workspace_id=$1`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Member
	for rows.Next() {
		var m models.Member
		_ = rows.Scan(&m.UserID, &m.WorkspaceID, &m.Role)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (p *Pgx) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	_, err := p.Pool.Exec(ctx, `INSERT INTO workspace_members(user_id,workspace_id,role) VALUES($1,$2,$3) ON CONFLICT (user_id,workspace_id) DO UPDATE SET role=EXCLUDED.role`, userID, workspaceID, role)
	return err
}

func (p *Pgx) CreateInvite(ctx context.Context, inv models.Invite) error {
	_, err := p.Pool.Exec(ctx, `INSERT INTO invites(token,workspace_id,email,role) VALUES($1,$2,$3,$4)`, inv.Token, inv.WorkspaceID, inv.Email, inv.Role)
	return err
}

func (p *Pgx) GetInvite(ctx context.Context, token string) (models.Invite, error) {
	var inv models.Invite
	err := p.Pool.QueryRow(ctx, `SELECT token,workspace_id,email,role,accepted,created_at FROM invites WHERE token=$1`, token).Scan(&inv.Token, &inv.WorkspaceID, &inv.Email, &inv.Role, &inv.Accepted, &inv.CreatedAt)
	if err != nil {
		return inv, ErrNotFound
	}
	return inv, nil
}

func (p *Pgx) AcceptInvite(ctx context.Context, token, userID string) (string, error) {
	inv, err := p.GetInvite(ctx, token)
	if err != nil {
		return "", err
	}
	_, err = p.Pool.Exec(ctx, `INSERT INTO workspace_members(user_id,workspace_id,role) VALUES($1,$2,$3) ON CONFLICT (user_id,workspace_id) DO NOTHING`, userID, inv.WorkspaceID, inv.Role)
	if err != nil {
		return "", err
	}
	_, _ = p.Pool.Exec(ctx, `UPDATE invites SET accepted=true WHERE token=$1`, token)
	return inv.WorkspaceID, nil
}

func (p *Pgx) CreateProject(ctx context.Context, workspaceID, name, desc string) (models.Project, error) {
	var pr models.Project
	err := p.Pool.QueryRow(ctx, `INSERT INTO projects(workspace_id,name,description) VALUES($1,$2,$3) RETURNING id,workspace_id,name,coalesce(description,''),created_at`, workspaceID, name, desc).Scan(&pr.ID, &pr.WorkspaceID, &pr.Name, &pr.Description, &pr.CreatedAt)
	return pr, err
}

func (p *Pgx) ListProjects(ctx context.Context, workspaceID string) ([]models.Project, error) {
	var rows interface {
	}
	_ = rows
	var q string
	var args []any
	if workspaceID == "" {
		q = `SELECT id,workspace_id,name,coalesce(description,''),created_at FROM projects ORDER BY created_at`
	} else {
		q = `SELECT id,workspace_id,name,coalesce(description,''),created_at FROM projects WHERE workspace_id=$1 ORDER BY created_at`
		args = append(args, workspaceID)
	}
	r, err := p.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Project
	for r.Next() {
		var pr models.Project
		_ = r.Scan(&pr.ID, &pr.WorkspaceID, &pr.Name, &pr.Description, &pr.CreatedAt)
		out = append(out, pr)
	}
	return out, r.Err()
}

func (p *Pgx) GetProject(ctx context.Context, id string) (models.Project, error) {
	var pr models.Project
	err := p.Pool.QueryRow(ctx, `SELECT id,workspace_id,name,coalesce(description,''),created_at FROM projects WHERE id=$1`, id).Scan(&pr.ID, &pr.WorkspaceID, &pr.Name, &pr.Description, &pr.CreatedAt)
	if err != nil {
		return pr, ErrNotFound
	}
	return pr, nil
}

func (p *Pgx) UpdateProject(ctx context.Context, id, name, desc string) (models.Project, error) {
	var pr models.Project
	err := p.Pool.QueryRow(ctx, `UPDATE projects SET name=coalesce(nullif($2,''),name), description=$3 WHERE id=$1 RETURNING id,workspace_id,name,coalesce(description,''),created_at`, id, name, desc).Scan(&pr.ID, &pr.WorkspaceID, &pr.Name, &pr.Description, &pr.CreatedAt)
	if err != nil {
		return pr, ErrNotFound
	}
	return pr, nil
}

func (p *Pgx) DeleteProject(ctx context.Context, id string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	return err
}

func (p *Pgx) CreateColumn(ctx context.Context, projectID, name string, position int) (models.Column, error) {
	var c models.Column
	err := p.Pool.QueryRow(ctx, `INSERT INTO columns(project_id,name,position) VALUES($1,$2,$3) RETURNING id,project_id,name,position`, projectID, name, position).Scan(&c.ID, &c.ProjectID, &c.Name, &c.Position)
	return c, err
}

func (p *Pgx) ListColumns(ctx context.Context, projectID string) ([]models.Column, error) {
	r, err := p.Pool.Query(ctx, `SELECT id,project_id,name,position FROM columns WHERE project_id=$1 ORDER BY position`, projectID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Column
	for r.Next() {
		var c models.Column
		_ = r.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Position)
		out = append(out, c)
	}
	return out, r.Err()
}

func (p *Pgx) UpdateColumn(ctx context.Context, id, name string, position *int) (models.Column, error) {
	var c models.Column
	pos := -1
	if position != nil {
		pos = *position
	}
	err := p.Pool.QueryRow(ctx, `UPDATE columns SET name=CASE WHEN $2='' THEN name ELSE $2 END, position=CASE WHEN $3<0 THEN position ELSE $3 END WHERE id=$1 RETURNING id,project_id,name,position`, id, name, pos).Scan(&c.ID, &c.ProjectID, &c.Name, &c.Position)
	if err != nil {
		return c, ErrNotFound
	}
	return c, nil
}

func (p *Pgx) DeleteColumn(ctx context.Context, id string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM columns WHERE id=$1`, id)
	return err
}

func scanTask(cols []any) models.Task { return models.Task{} }

func (p *Pgx) CreateTask(ctx context.Context, t models.Task) (models.Task, error) {
	var out models.Task
	err := p.Pool.QueryRow(ctx, `INSERT INTO tasks(project_id,column_id,title,description,priority,assignee_id,tags,due_date,position) VALUES($1,$2,$3,$4,coalesce(nullif($5,''),''),$6,coalesce($7,'{}'::text[]),$8,$9) RETURNING id,project_id,column_id,title,coalesce(description,''),coalesce(priority,''),assignee_id,coalesce(tags,'{}'),due_date,position,version,created_at`,
		t.ProjectID, t.ColumnID, t.Title, t.Description, t.Priority, t.AssigneeID, t.Tags, t.DueDate, t.Position).Scan(&out.ID, &out.ProjectID, &out.ColumnID, &out.Title, &out.Description, &out.Priority, &out.AssigneeID, &out.Tags, &out.DueDate, &out.Position, &out.Version, &out.CreatedAt)
	return out, err
}

func (p *Pgx) ListTasks(ctx context.Context, fl TaskFilter) ([]models.Task, error) {
	q := `SELECT id,project_id,column_id,title,coalesce(description,''),coalesce(priority,''),assignee_id,coalesce(tags,'{}'),due_date,position,version,created_at FROM tasks WHERE ($1='' OR project_id::text=$1) AND ($2='' OR column_id::text=$2) AND ($3='' OR priority=$3) AND ($4='' OR assignee_id::text=$4) AND ($5='' OR $5=ANY(tags)) AND ($6='' OR title ILIKE '%'||$6||'%' OR coalesce(description,'') ILIKE '%'||$6||'%') ORDER BY position`
	r, err := p.Pool.Query(ctx, q, fl.ProjectID, fl.ColumnID, fl.Priority, fl.Assignee, fl.Tag, fl.Q)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Task
	for r.Next() {
		var t models.Task
		_ = r.Scan(&t.ID, &t.ProjectID, &t.ColumnID, &t.Title, &t.Description, &t.Priority, &t.AssigneeID, &t.Tags, &t.DueDate, &t.Position, &t.Version, &t.CreatedAt)
		out = append(out, t)
	}
	return out, r.Err()
}

func (p *Pgx) GetTask(ctx context.Context, id string) (models.Task, error) {
	var t models.Task
	err := p.Pool.QueryRow(ctx, `SELECT id,project_id,column_id,title,coalesce(description,''),coalesce(priority,''),assignee_id,coalesce(tags,'{}'),due_date,position,version,created_at FROM tasks WHERE id=$1`, id).Scan(&t.ID, &t.ProjectID, &t.ColumnID, &t.Title, &t.Description, &t.Priority, &t.AssigneeID, &t.Tags, &t.DueDate, &t.Position, &t.Version, &t.CreatedAt)
	if err != nil {
		return t, ErrNotFound
	}
	return t, nil
}

func (p *Pgx) PatchTask(ctx context.Context, id string, pt TaskPatch) (models.Task, bool, error) {
	cur, err := p.GetTask(ctx, id)
	if err != nil {
		return cur, false, err
	}
	if pt.HasVersion && pt.Version != cur.Version {
		return cur, true, ErrConflict
	}
	applyPatch(&cur, pt)
	var out models.Task
	err = p.Pool.QueryRow(ctx, `UPDATE tasks SET title=$2,description=$3,priority=coalesce(nullif($4,''),''),assignee_id=$5,column_id=$6,position=$7,due_date=$8,tags=coalesce($9,'{}'::text[]),version=version+1 WHERE id=$1 AND version=$10 RETURNING id,project_id,column_id,title,coalesce(description,''),coalesce(priority,''),assignee_id,coalesce(tags,'{}'),due_date,position,version,created_at`,
		id, cur.Title, cur.Description, cur.Priority, cur.AssigneeID, cur.ColumnID, cur.Position, cur.DueDate, cur.Tags, cur.Version).Scan(&out.ID, &out.ProjectID, &out.ColumnID, &out.Title, &out.Description, &out.Priority, &out.AssigneeID, &out.Tags, &out.DueDate, &out.Position, &out.Version, &out.CreatedAt)
	if err != nil {
		return cur, true, ErrConflict
	}
	return out, false, nil
}

func (p *Pgx) MoveTask(ctx context.Context, id, columnID string, position, version int) (models.Task, bool, error) {
	cur, err := p.GetTask(ctx, id)
	if err != nil {
		return cur, false, err
	}
	if version != 0 && version != cur.Version {
		return cur, true, ErrConflict
	}
	var out models.Task
	err = p.Pool.QueryRow(ctx, `UPDATE tasks SET column_id=$2,position=$3,version=version+1 WHERE id=$1 AND version=$4 RETURNING id,project_id,column_id,title,coalesce(description,''),coalesce(priority,''),assignee_id,coalesce(tags,'{}'),due_date,position,version,created_at`,
		id, columnID, position, cur.Version).Scan(&out.ID, &out.ProjectID, &out.ColumnID, &out.Title, &out.Description, &out.Priority, &out.AssigneeID, &out.Tags, &out.DueDate, &out.Position, &out.Version, &out.CreatedAt)
	if err != nil {
		return cur, true, ErrConflict
	}
	return out, false, nil
}

func (p *Pgx) DeleteTask(ctx context.Context, id string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func (p *Pgx) AddComment(ctx context.Context, taskID, userID, body string) (models.Comment, error) {
	var c models.Comment
	err := p.Pool.QueryRow(ctx, `INSERT INTO comments(task_id,user_id,body) VALUES($1,$2,$3) RETURNING id,task_id,user_id,body,created_at`, taskID, userID, body).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Body, &c.CreatedAt)
	return c, err
}

func (p *Pgx) ListComments(ctx context.Context, taskID string) ([]models.Comment, error) {
	r, err := p.Pool.Query(ctx, `SELECT id,task_id,user_id,body,created_at FROM comments WHERE task_id=$1 ORDER BY created_at`, taskID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Comment
	for r.Next() {
		var c models.Comment
		_ = r.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Body, &c.CreatedAt)
		out = append(out, c)
	}
	return out, r.Err()
}

func (p *Pgx) AddChecklist(ctx context.Context, taskID, title string, position int) (models.Checklist, error) {
	var c models.Checklist
	err := p.Pool.QueryRow(ctx, `INSERT INTO checklists(task_id,title,position) VALUES($1,$2,$3) RETURNING id,task_id,title,done,position`, taskID, title, position).Scan(&c.ID, &c.TaskID, &c.Title, &c.Done, &c.Position)
	return c, err
}

func (p *Pgx) ListChecklists(ctx context.Context, taskID string) ([]models.Checklist, error) {
	r, err := p.Pool.Query(ctx, `SELECT id,task_id,title,done,position FROM checklists WHERE task_id=$1 ORDER BY position`, taskID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Checklist
	for r.Next() {
		var c models.Checklist
		_ = r.Scan(&c.ID, &c.TaskID, &c.Title, &c.Done, &c.Position)
		out = append(out, c)
	}
	return out, r.Err()
}

func (p *Pgx) ToggleChecklist(ctx context.Context, id string, done *bool, title *string) (models.Checklist, error) {
	var c models.Checklist
	err := p.Pool.QueryRow(ctx, `SELECT id,task_id,title,done,position FROM checklists WHERE id=$1`, id).Scan(&c.ID, &c.TaskID, &c.Title, &c.Done, &c.Position)
	if err != nil {
		return c, ErrNotFound
	}
	if done != nil {
		c.Done = *done
	} else {
		c.Done = !c.Done
	}
	if title != nil {
		c.Title = *title
	}
	err = p.Pool.QueryRow(ctx, `UPDATE checklists SET done=$2,title=$3 WHERE id=$1 RETURNING id,task_id,title,done,position`, id, c.Done, c.Title).Scan(&c.ID, &c.TaskID, &c.Title, &c.Done, &c.Position)
	return c, err
}

func (p *Pgx) AddAttachment(ctx context.Context, a models.Attachment) (models.Attachment, error) {
	var out models.Attachment
	err := p.Pool.QueryRow(ctx, `INSERT INTO attachments(task_id,filename,size,mime,uploaded_by) VALUES($1,$2,$3,$4,$5) RETURNING id,task_id,filename,size,mime,uploaded_by,created_at`, a.TaskID, a.Filename, a.Size, a.Mime, a.UploadedBy).Scan(&out.ID, &out.TaskID, &out.Filename, &out.Size, &out.Mime, &out.UploadedBy, &out.CreatedAt)
	return out, err
}

func (p *Pgx) ListAttachments(ctx context.Context, taskID string) ([]models.Attachment, error) {
	r, err := p.Pool.Query(ctx, `SELECT id,task_id,filename,size,mime,uploaded_by,created_at FROM attachments WHERE task_id=$1 ORDER BY created_at`, taskID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Attachment
	for r.Next() {
		var a models.Attachment
		_ = r.Scan(&a.ID, &a.TaskID, &a.Filename, &a.Size, &a.Mime, &a.UploadedBy, &a.CreatedAt)
		out = append(out, a)
	}
	return out, r.Err()
}

func (p *Pgx) LogActivity(ctx context.Context, workspaceID string, taskID, userID *string, action, detail string) error {
	_, err := p.Pool.Exec(ctx, `INSERT INTO activities(workspace_id,task_id,user_id,action,detail) VALUES($1,$2,$3,$4,$5)`, workspaceID, taskID, userID, action, detail)
	return err
}

func (p *Pgx) ListActivities(ctx context.Context, workspaceID, taskID string) ([]models.Activity, error) {
	r, err := p.Pool.Query(ctx, `SELECT id,workspace_id,task_id,user_id,action,coalesce(detail,''),created_at FROM activities WHERE ($1='' OR workspace_id::text=$1) AND ($2='' OR task_id::text=$2) ORDER BY created_at DESC LIMIT 200`, workspaceID, taskID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Activity
	for r.Next() {
		var a models.Activity
		_ = r.Scan(&a.ID, &a.WorkspaceID, &a.TaskID, &a.UserID, &a.Action, &a.Detail, &a.CreatedAt)
		out = append(out, a)
	}
	return out, r.Err()
}

func (p *Pgx) ListNotifications(ctx context.Context, userID string) ([]models.Notification, error) {
	r, err := p.Pool.Query(ctx, `SELECT id,user_id,workspace_id,title,coalesce(body,''),read,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []models.Notification
	for r.Next() {
		var n models.Notification
		_ = r.Scan(&n.ID, &n.UserID, &n.WorkspaceID, &n.Title, &n.Body, &n.Read, &n.CreatedAt)
		out = append(out, n)
	}
	return out, r.Err()
}

func (p *Pgx) MarkNotificationRead(ctx context.Context, id, userID string) error {
	_, err := p.Pool.Exec(ctx, `UPDATE notifications SET read=true WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (p *Pgx) CreateNotification(ctx context.Context, n models.Notification) error {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	_, err := p.Pool.Exec(ctx, `INSERT INTO notifications(id,user_id,workspace_id,title,body) VALUES($1,$2,$3,$4,$5)`, n.ID, n.UserID, n.WorkspaceID, n.Title, n.Body)
	return err
}

func (p *Pgx) Dashboard(ctx context.Context, workspaceID string) (map[string]any, error) {
	byCol := map[string]int{}
	r, err := p.Pool.Query(ctx, `SELECT c.name,count(t.id) FROM columns c JOIN projects pr ON pr.id=c.project_id LEFT JOIN tasks t ON t.column_id=c.id WHERE pr.workspace_id=$1 GROUP BY c.name`, workspaceID)
	if err != nil {
		return nil, err
	}
	for r.Next() {
		var name string
		var n int
		_ = r.Scan(&name, &n)
		byCol[name] = n
	}
	r.Close()
	byPri := map[string]int{}
	r2, err := p.Pool.Query(ctx, `SELECT coalesce(t.priority,'none'),count(*) FROM tasks t JOIN projects pr ON pr.id=t.project_id WHERE pr.workspace_id=$1 GROUP BY 1`, workspaceID)
	if err != nil {
		return nil, err
	}
	for r2.Next() {
		var k string
		var n int
		_ = r2.Scan(&k, &n)
		byPri[k] = n
	}
	r2.Close()
	var overdue int
	_ = p.Pool.QueryRow(ctx, `SELECT count(*) FROM tasks t JOIN projects pr ON pr.id=t.project_id WHERE pr.workspace_id=$1 AND t.due_date IS NOT NULL AND t.due_date < now()`, workspaceID).Scan(&overdue)
	byMember := map[string]int{}
	r3, err := p.Pool.Query(ctx, `SELECT coalesce(t.assignee_id::text,'unassigned'),count(*) FROM tasks t JOIN projects pr ON pr.id=t.project_id WHERE pr.workspace_id=$1 GROUP BY 1`, workspaceID)
	if err != nil {
		return nil, err
	}
	for r3.Next() {
		var k string
		var n int
		_ = r3.Scan(&k, &n)
		byMember[k] = n
	}
	r3.Close()
	_ = time.Now()
	return map[string]any{"by_column": byCol, "by_priority": byPri, "overdue": overdue, "by_member": byMember}, nil
}

func (p *Pgx) Search(ctx context.Context, workspaceID, q string) (map[string]any, error) {
	r, err := p.Pool.Query(ctx, `SELECT t.id,t.project_id,t.column_id,t.title,coalesce(t.description,''),coalesce(t.priority,''),t.assignee_id,coalesce(t.tags,'{}'),t.due_date,t.position,t.version,t.created_at FROM tasks t JOIN projects pr ON pr.id=t.project_id WHERE pr.workspace_id=$1 AND (t.title ILIKE '%'||$2||'%' OR coalesce(t.description,'') ILIKE '%'||$2||'%') LIMIT 50`, workspaceID, q)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var tasks []models.Task
	for r.Next() {
		var t models.Task
		_ = r.Scan(&t.ID, &t.ProjectID, &t.ColumnID, &t.Title, &t.Description, &t.Priority, &t.AssigneeID, &t.Tags, &t.DueDate, &t.Position, &t.Version, &t.CreatedAt)
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	return map[string]any{"tasks": tasks}, nil
}
