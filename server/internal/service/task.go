package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openzoo-ai/openzoo/server/internal/queue"
)

type TaskService struct {
	db    *pgxpool.Pool
	queue queue.TaskQueue
}

func NewTaskService(db *pgxpool.Pool) *TaskService {
	return &TaskService{db: db, queue: queue.NewTaskQueueFromEnv()}
}

type Task struct {
	ID          string     `json:"id"`
	AgentID     string     `json:"agent_id"`
	RuntimeID   string     `json:"runtime_id"`
	IssueID     string     `json:"issue_id"`
	Status      string     `json:"status"`
	Priority    int        `json:"priority"`
	DispatchedAt *time.Time `json:"dispatched_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Result      *string    `json:"result_json"`
	Error       *string    `json:"error"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TaskMessage struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	IssueID   string    `json:"issue_id"`
	Seq       int       `json:"seq"`
	Type      string    `json:"type"`
	Tool      *string   `json:"tool"`
	Content   *string   `json:"content"`
	InputJSON *string   `json:"input_json"`
	Output    *string   `json:"output"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *TaskService) Create(ctx context.Context, issueID, runtimeID, agentID, prompt string) (*Task, error) {
	id := uuid.New().String()
	now := time.Now()
	_, err := s.db.Exec(ctx, `INSERT INTO tasks (id, agent_id, runtime_id, issue_id, status, priority, created_at) VALUES ($1,$2,$3,$4,'queued',0,$5)`,
		id, agentID, runtimeID, issueID, now)
	if err != nil {
		return nil, err
	}
	task := &Task{ID: id, AgentID: agentID, RuntimeID: runtimeID, IssueID: issueID, Status: "queued", CreatedAt: now}
	_ = s.queue.Enqueue(ctx, queue.TaskEvent{
		TaskID:     task.ID,
		IssueID:    task.IssueID,
		RuntimeID:  task.RuntimeID,
		AgentID:    task.AgentID,
		Status:     task.Status,
		OccurredAt: time.Now().Unix(),
	})
	return task, nil
}

func (s *TaskService) Get(ctx context.Context, taskID string) (*Task, error) {
	var t Task
	err := s.db.QueryRow(ctx, `SELECT id, agent_id, runtime_id, issue_id, status, COALESCE(priority,0), dispatched_at, started_at, completed_at, result, error, created_at FROM tasks WHERE id = $1`, taskID).
		Scan(&t.ID, &t.AgentID, &t.RuntimeID, &t.IssueID, &t.Status, &t.Priority, &t.DispatchedAt, &t.StartedAt, &t.CompletedAt, &t.Result, &t.Error, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TaskService) List(ctx context.Context, workspaceID, issueID, agentID, status string, limit, offset int) ([]Task, int, error) {
	query := `SELECT t.id, t.agent_id, t.runtime_id, t.issue_id, t.status, COALESCE(t.priority,0), t.dispatched_at, t.started_at, t.completed_at, t.result, t.error, t.created_at FROM tasks t JOIN issues i ON t.issue_id = i.id WHERE i.workspace_id = $1`
	args := []interface{}{workspaceID}
	argN := 2
	if issueID != "" {
		query += fmt.Sprintf(` AND t.issue_id = $%d`, argN)
		args = append(args, issueID)
		argN++
	}
	if agentID != "" {
		query += fmt.Sprintf(` AND t.agent_id = $%d`, argN)
		args = append(args, agentID)
		argN++
	}
	if status != "" {
		query += fmt.Sprintf(` AND t.status = $%d`, argN)
		args = append(args, status)
		argN++
	}
	var total int
	countQ := `SELECT COUNT(*) FROM (` + query + `) sub`
	if err := s.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(` ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, limit, offset)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.AgentID, &t.RuntimeID, &t.IssueID, &t.Status, &t.Priority, &t.DispatchedAt, &t.StartedAt, &t.CompletedAt, &t.Result, &t.Error, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, nil
}

func (s *TaskService) UpdateStatus(ctx context.Context, taskID, status string, taskErr, result *string) (*Task, error) {
	now := time.Now()
	switch status {
	case "dispatched":
		s.db.Exec(ctx, `UPDATE tasks SET status = $2, dispatched_at = $3 WHERE id = $1`, taskID, status, now)
	case "running":
		s.db.Exec(ctx, `UPDATE tasks SET status = $2, started_at = $3 WHERE id = $1`, taskID, status, now)
	case "completed", "failed", "cancelled":
		s.db.Exec(ctx, `UPDATE tasks SET status = $2, completed_at = $3 WHERE id = $1`, taskID, status, now)
	default:
		s.db.Exec(ctx, `UPDATE tasks SET status = $2 WHERE id = $1`, taskID, status)
	}
	if taskErr != nil {
		s.db.Exec(ctx, `UPDATE tasks SET error = $2 WHERE id = $1`, taskID, *taskErr)
	}
	if result != nil {
		s.db.Exec(ctx, `UPDATE tasks SET result = $2 WHERE id = $1`, taskID, *result)
	}
	updated, err := s.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}
	event := queue.TaskEvent{
		TaskID:     updated.ID,
		IssueID:    updated.IssueID,
		RuntimeID:  updated.RuntimeID,
		AgentID:    updated.AgentID,
		Status:     updated.Status,
		OccurredAt: time.Now().Unix(),
	}
	if status == "failed" {
		_ = s.queue.DeadLetter(ctx, event, "task_status_failed")
	} else {
		_ = s.queue.Enqueue(ctx, event)
	}
	return updated, nil
}

func (s *TaskService) ListMessages(ctx context.Context, taskID string, limit, offset int) ([]TaskMessage, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.db.Query(ctx, `SELECT id, task_id, issue_id, seq, type, tool, content, input_json, output, created_at FROM task_messages WHERE task_id = $1 ORDER BY seq ASC LIMIT $2 OFFSET $3`, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []TaskMessage
	for rows.Next() {
		var m TaskMessage
		if err := rows.Scan(&m.ID, &m.TaskID, &m.IssueID, &m.Seq, &m.Type, &m.Tool, &m.Content, &m.InputJSON, &m.Output, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

type ClaimedTask struct {
	ID          string `json:"id"`
	AgentID     string `json:"agent_id"`
	RuntimeID   string `json:"runtime_id"`
	IssueID     string `json:"issue_id"`
	WorkspaceID string `json:"workspace_id"`
}

func (s *TaskService) ClaimTask(ctx context.Context, runtimeID string) (*ClaimedTask, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var taskID, agentID, runtimeIDVal, issueID, workspaceID string
	err = tx.QueryRow(ctx, `
		UPDATE tasks SET status = 'dispatched', dispatched_at = NOW()
		WHERE id = (
			SELECT t.id FROM tasks t
			JOIN issues i ON t.issue_id = i.id
			WHERE t.status = 'queued'
			ORDER BY t.priority DESC, t.created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING t.id, t.agent_id, t.runtime_id, t.issue_id, i.workspace_id
	`).Scan(&taskID, &agentID, &runtimeIDVal, &issueID, &workspaceID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	_ = s.queue.Enqueue(ctx, queue.TaskEvent{
		TaskID: taskID, IssueID: issueID, RuntimeID: runtimeIDVal,
		AgentID: agentID, Status: "dispatched", OccurredAt: time.Now().Unix(),
	})

	return &ClaimedTask{
		ID: taskID, AgentID: agentID, RuntimeID: runtimeIDVal,
		IssueID: issueID, WorkspaceID: workspaceID,
	}, nil
}

func (s *TaskService) StartTask(ctx context.Context, taskID string) error {
	_, err := s.db.Exec(ctx, `UPDATE tasks SET status = 'running', started_at = NOW() WHERE id = $1 AND status = 'dispatched'`, taskID)
	if err != nil {
		return err
	}
	task, err := s.Get(ctx, taskID)
	if err != nil {
		return nil
	}
	_ = s.queue.Enqueue(ctx, queue.TaskEvent{
		TaskID: taskID, IssueID: task.IssueID, RuntimeID: task.RuntimeID,
		AgentID: task.AgentID, Status: "running", OccurredAt: time.Now().Unix(),
	})
	return nil
}

func (s *TaskService) CompleteTask(ctx context.Context, taskID string, resultJSON, comment string) error {
	_, err := s.db.Exec(ctx, `UPDATE tasks SET status = 'completed', completed_at = NOW(), result = $2 WHERE id = $1`, taskID, resultJSON)
	if err != nil {
		return err
	}
	task, err := s.Get(ctx, taskID)
	if err != nil {
		return nil
	}
	if comment != "" && task.IssueID != "" {
		commentID := uuid.New().String()
		s.db.Exec(ctx,
			`INSERT INTO comments (id, issue_id, workspace_id, author_type, author_id, content, type, created_at, updated_at)
			 VALUES ($1, $2, (SELECT workspace_id FROM issues WHERE id = $2), 'agent', $3, $4, 'progress_update', NOW(), NOW())`,
			commentID, task.IssueID, task.AgentID, comment)
	}
	_ = s.queue.Enqueue(ctx, queue.TaskEvent{
		TaskID: taskID, IssueID: task.IssueID, RuntimeID: task.RuntimeID,
		AgentID: task.AgentID, Status: "completed", OccurredAt: time.Now().Unix(),
	})
	return nil
}

func (s *TaskService) FailTask(ctx context.Context, taskID string, errMsg string) error {
	_, err := s.db.Exec(ctx, `UPDATE tasks SET status = 'failed', completed_at = NOW(), error = $2 WHERE id = $1`, taskID, errMsg)
	if err != nil {
		return err
	}
	task, err := s.Get(ctx, taskID)
	if err != nil {
		return nil
	}
	_ = s.queue.DeadLetter(ctx, queue.TaskEvent{
		TaskID: taskID, IssueID: task.IssueID, RuntimeID: task.RuntimeID,
		AgentID: task.AgentID, Status: "failed", OccurredAt: time.Now().Unix(),
	}, errMsg)
	return nil
}

type InboundMessage struct {
	Seq     int32  `json:"seq"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Tool    string `json:"tool,omitempty"`
	CallID  string `json:"call_id,omitempty"`
	Input   string `json:"input,omitempty"`
	Output  string `json:"output,omitempty"`
}

func (s *TaskService) ReportMessages(ctx context.Context, taskID string, messages []InboundMessage) error {
	task, err := s.Get(ctx, taskID)
	if err != nil {
		return err
	}
	for _, msg := range messages {
		msgID := uuid.New().String()
		var toolJSON, inputJSON *string
		if msg.Tool != "" {
			b, _ := json.Marshal(map[string]string{"name": msg.Tool})
			s := string(b)
			toolJSON = &s
		}
		if msg.Input != "" {
			inputJSON = &msg.Input
		}
		s.db.Exec(ctx,
			`INSERT INTO task_messages (id, task_id, issue_id, seq, type, tool, content, input_json, output, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`,
			msgID, taskID, task.IssueID, msg.Seq, msg.Type, toolJSON, msg.Content, inputJSON, msg.Output)
	}
	return nil
}
