package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InboxService struct {
	db *pgxpool.Pool
}

func NewInboxService(db *pgxpool.Pool) *InboxService {
	return &InboxService{db: db}
}

type InboxItem struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	RecipientType string    `json:"recipient_type"`
	RecipientID   string    `json:"recipient_id"`
	ActorType     *string   `json:"actor_type"`
	ActorID       *string   `json:"actor_id"`
	Type          string    `json:"type"`
	Severity      string    `json:"severity"`
	IssueID       *string   `json:"issue_id"`
	Title         string    `json:"title"`
	Body          *string   `json:"body"`
	IssueStatus   *string   `json:"issue_status"`
	Read          bool      `json:"read"`
	Archived      bool      `json:"archived"`
	CreatedAt     time.Time `json:"created_at"`
	Details       *string   `json:"details_json"`
}

func (s *InboxService) List(ctx context.Context, workspaceID, recipientID string, unreadOnly bool, limit, offset int) ([]InboxItem, int, int, error) {
	query := `SELECT id, workspace_id, recipient_type, recipient_id, actor_type, actor_id, type, severity, issue_id, title, body, issue_status, read, archived, created_at, details FROM inbox_items WHERE workspace_id = $1 AND recipient_id = $2 AND archived = false`
	args := []interface{}{workspaceID, recipientID}
	argN := 3
	if unreadOnly {
		query += fmt.Sprintf(` AND read = false`)
	}
	countQuery := `SELECT COUNT(*) FROM (` + query + `) sub`
	var total int
	if err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, 0, err
	}
	var unreadCount int
	unreadQuery := `SELECT COUNT(*) FROM (` + query + ` AND read = false) sub`
	_ = s.db.QueryRow(ctx, unreadQuery, args...).Scan(&unreadCount)
	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	args = append(args, limit, offset)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	var items []InboxItem
	for rows.Next() {
		var it InboxItem
		if err := rows.Scan(&it.ID, &it.WorkspaceID, &it.RecipientType, &it.RecipientID, &it.ActorType, &it.ActorID, &it.Type, &it.Severity, &it.IssueID, &it.Title, &it.Body, &it.IssueStatus, &it.Read, &it.Archived, &it.CreatedAt, &it.Details); err != nil {
			return nil, 0, 0, err
		}
		items = append(items, it)
	}
	return items, total, unreadCount, nil
}

func (s *InboxService) MarkRead(ctx context.Context, workspaceID string, itemIDs []string) error {
	for _, id := range itemIDs {
		_, err := s.db.Exec(ctx, `UPDATE inbox_items SET read = true WHERE id = $1 AND workspace_id = $2`, id, workspaceID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InboxService) MarkArchived(ctx context.Context, workspaceID string, itemIDs []string) error {
	for _, id := range itemIDs {
		_, err := s.db.Exec(ctx, `UPDATE inbox_items SET archived = true WHERE id = $1 AND workspace_id = $2`, id, workspaceID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InboxService) MarkAllRead(ctx context.Context, workspaceID, recipientID string) error {
	_, err := s.db.Exec(ctx, `UPDATE inbox_items SET read = true WHERE workspace_id = $1 AND recipient_id = $2 AND read = false`, workspaceID, recipientID)
	return err
}

func (s *InboxService) Create(ctx context.Context, workspaceID, recipientType, recipientID, actorType, actorID, itemType, severity, issueID, title, body string) (*InboxItem, error) {
	id := uuid.New().String()
	now := time.Now()
	_, err := s.db.Exec(ctx, `INSERT INTO inbox_items (id, workspace_id, recipient_type, recipient_id, actor_type, actor_id, type, severity, issue_id, title, body, read, archived, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,false,false,$12)`,
		id, workspaceID, recipientType, recipientID, actorType, actorID, itemType, severity, issueID, title, body, now)
	if err != nil {
		return nil, err
	}
	return &InboxItem{ID: id, WorkspaceID: workspaceID, RecipientType: recipientType, RecipientID: recipientID, Type: itemType, Severity: severity, Title: title, CreatedAt: now}, nil
}

func CreateInboxItem(ctx context.Context, db *pgxpool.Pool, workspaceID, userID, actorType, entityID, itemType, title, body string) error {
	id := uuid.New().String()
	_, err := db.Exec(ctx,
		`INSERT INTO inbox_items (id, workspace_id, recipient_type, recipient_id, actor_type, actor_id, type, severity, issue_id, title, body, read, archived, created_at)
		 VALUES ($1, $2, 'user', $3, $4, $5, $6, 'info', $7, $8, $9, false, false, NOW())`,
		id, workspaceID, userID, actorType, entityID, itemType, entityID, title, body)
	return err
}

func CreateInboxForSubscribers(ctx context.Context, db *pgxpool.Pool, workspaceID, issueID, excludeUserID, itemType, title, body string) error {
	rows, err := db.Query(ctx,
		`SELECT user_id FROM issue_subscribers WHERE issue_id = $1 AND user_id != $2`,
		issueID, excludeUserID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		_ = CreateInboxItem(ctx, db, workspaceID, userID, "member", issueID, itemType, title, body)
	}
	return nil
}
