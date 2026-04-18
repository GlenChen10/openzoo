package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AgentService struct {
	db *pgxpool.Pool
}

func NewAgentService(db *pgxpool.Pool) *AgentService {
	return &AgentService{db: db}
}

type Agent struct {
	ID                string    `json:"id"`
	WorkspaceID       string    `json:"workspace_id"`
	RuntimeID         string    `json:"runtime_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Instructions      string    `json:"instructions"`
	AvatarURL         string    `json:"avatar_url"`
	Visibility        string    `json:"visibility"`
	Status            string    `json:"status"`
	MaxConcurrentTask int       `json:"max_concurrent_tasks"`
	OwnerID           string    `json:"owner_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	ArchivedAt        *time.Time `json:"archived_at"`
}

func (s *AgentService) List(ctx context.Context, workspaceID string) ([]Agent, error) {
	rows, err := s.db.Query(ctx, `SELECT id, workspace_id, runtime_id, name, COALESCE(description,''), COALESCE(instructions,''), COALESCE(avatar_url,''), COALESCE(visibility,'workspace'), COALESCE(status,'idle'), COALESCE(max_concurrent_tasks,1), COALESCE(owner_id,''), created_at, updated_at, archived_at FROM agents WHERE workspace_id = $1 AND archived_at IS NULL ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.RuntimeID, &a.Name, &a.Description, &a.Instructions, &a.AvatarURL, &a.Visibility, &a.Status, &a.MaxConcurrentTask, &a.OwnerID, &a.CreatedAt, &a.UpdatedAt, &a.ArchivedAt); err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, nil
}

func (s *AgentService) Get(ctx context.Context, workspaceID, agentID string) (*Agent, error) {
	var a Agent
	err := s.db.QueryRow(ctx, `SELECT id, workspace_id, runtime_id, name, COALESCE(description,''), COALESCE(instructions,''), COALESCE(avatar_url,''), COALESCE(visibility,'workspace'), COALESCE(status,'idle'), COALESCE(max_concurrent_tasks,1), COALESCE(owner_id,''), created_at, updated_at, archived_at FROM agents WHERE id = $1 AND workspace_id = $2`, agentID, workspaceID).
		Scan(&a.ID, &a.WorkspaceID, &a.RuntimeID, &a.Name, &a.Description, &a.Instructions, &a.AvatarURL, &a.Visibility, &a.Status, &a.MaxConcurrentTask, &a.OwnerID, &a.CreatedAt, &a.UpdatedAt, &a.ArchivedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AgentService) Create(ctx context.Context, workspaceID, name, description, instructions, runtimeID string) (*Agent, error) {
	id := uuid.New().String()
	now := time.Now()
	_, err := s.db.Exec(ctx, `INSERT INTO agents (id, workspace_id, runtime_id, name, description, instructions, visibility, status, max_concurrent_tasks, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,'workspace','idle',1,$7,$7)`,
		id, workspaceID, runtimeID, name, description, instructions, now)
	if err != nil {
		return nil, err
	}
	return &Agent{ID: id, WorkspaceID: workspaceID, RuntimeID: runtimeID, Name: name, Description: description, Instructions: instructions, Visibility: "workspace", Status: "idle", MaxConcurrentTask: 1, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *AgentService) Update(ctx context.Context, workspaceID, agentID string, fields map[string]interface{}) (*Agent, error) {
	now := time.Now()
	fields["updated_at"] = now
	setClauses, args, _ := buildSetClauses(fields, allowedAgentFields, 3)
	if setClauses == "" {
		return s.Get(ctx, workspaceID, agentID)
	}
	args = append([]interface{}{agentID, workspaceID}, args...)
	_, err := s.db.Exec(ctx, "UPDATE agents SET "+setClauses+" WHERE id = $1 AND workspace_id = $2", args...)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, workspaceID, agentID)
}

func (s *AgentService) Archive(ctx context.Context, workspaceID, agentID string) (*Agent, error) {
	now := time.Now()
	_, err := s.db.Exec(ctx, `UPDATE agents SET archived_at = $3, updated_at = $3 WHERE id = $1 AND workspace_id = $2`, agentID, workspaceID, now)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, workspaceID, agentID)
}

func (s *AgentService) Restore(ctx context.Context, workspaceID, agentID string) (*Agent, error) {
	_, err := s.db.Exec(ctx, `UPDATE agents SET archived_at = NULL, updated_at = NOW() WHERE id = $1 AND workspace_id = $2`, agentID, workspaceID)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, workspaceID, agentID)
}

type Skill struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspace_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Content     string                 `json:"content"`
	Config      map[string]interface{} `json:"config"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func (s *AgentService) ListSkills(ctx context.Context, workspaceID string) ([]Skill, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(content,''), COALESCE(config,'{}'::jsonb), COALESCE(created_by,''), created_at, updated_at
		 FROM skills WHERE workspace_id = $1 ORDER BY created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var skills []Skill
	for rows.Next() {
		var sk Skill
		if err := rows.Scan(&sk.ID, &sk.WorkspaceID, &sk.Name, &sk.Description, &sk.Content, &sk.Config, &sk.CreatedBy, &sk.CreatedAt, &sk.UpdatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, sk)
	}
	return skills, nil
}

func (s *AgentService) CreateSkill(ctx context.Context, workspaceID, name, description, content string) (*Skill, error) {
	id := uuid.New().String()
	now := time.Now()
	sk := &Skill{ID: id, WorkspaceID: workspaceID, Name: name, Description: description, Content: content, CreatedAt: now, UpdatedAt: now}
	_, err := s.db.Exec(ctx,
		`INSERT INTO skills (id, workspace_id, name, description, content, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$6)`,
		id, workspaceID, name, description, content, now)
	if err != nil {
		return nil, err
	}
	return sk, nil
}

func (s *AgentService) GetSkill(ctx context.Context, workspaceID, skillID string) (*Skill, error) {
	var sk Skill
	err := s.db.QueryRow(ctx,
		`SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(content,''), COALESCE(config,'{}'::jsonb), COALESCE(created_by,''), created_at, updated_at
		 FROM skills WHERE id = $1 AND workspace_id = $2`,
		skillID, workspaceID).
		Scan(&sk.ID, &sk.WorkspaceID, &sk.Name, &sk.Description, &sk.Content, &sk.Config, &sk.CreatedBy, &sk.CreatedAt, &sk.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sk, nil
}

func (s *AgentService) UpdateSkill(ctx context.Context, workspaceID, skillID string, fields map[string]interface{}) (*Skill, error) {
	now := time.Now()
	fields["updated_at"] = now
	i := 3
	args := []interface{}{skillID, workspaceID}
	setClauses := ""
	for k, v := range fields {
		if setClauses != "" {
			setClauses += ", "
		}
		setClauses += k + " = $" + strconv.Itoa(i)
		args = append(args, v)
		i++
	}
	_, err := s.db.Exec(ctx, "UPDATE skills SET "+setClauses+" WHERE id = $1 AND workspace_id = $2", args...)
	if err != nil {
		return nil, err
	}
	return s.GetSkill(ctx, workspaceID, skillID)
}

func (s *AgentService) DeleteSkill(ctx context.Context, workspaceID, skillID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM skills WHERE id = $1 AND workspace_id = $2`, skillID, workspaceID)
	return err
}
