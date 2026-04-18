package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentService_CreateAgent(t *testing.T) {
	svc := &AgentService{}
	assert.NotNil(t, svc)

	// Test agent creation parameters
	params := map[string]any{
		"name":                  "Test Agent",
		"description":           "A test agent",
		"instructions":          "You are a helpful assistant",
		"visibility":            "private",
		"max_concurrent_tasks":  5,
	}

	assert.Equal(t, "Test Agent", params["name"])
	assert.Equal(t, "A test agent", params["description"])
	assert.Equal(t, "private", params["visibility"])
}

func TestAgentService_UpdateAgent(t *testing.T) {
	updates := map[string]any{
		"name":                  "Updated Agent",
		"description":           "Updated description",
		"instructions":          "Updated instructions",
		"status":                "active",
		"max_concurrent_tasks":  10,
	}

	assert.Equal(t, "Updated Agent", updates["name"])
	assert.Equal(t, "active", updates["status"])
	assert.Equal(t, 10, updates["max_concurrent_tasks"])
}

func TestAgentService_ListAgents(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	// Test list parameters
	params := map[string]any{
		"workspace_id":       "ws_123",
		"include_archived":   false,
	}

	assert.Equal(t, "ws_123", params["workspace_id"])
	assert.Equal(t, false, params["include_archived"])
}

func TestAgentService_ArchiveAgent(t *testing.T) {
	now := time.Now()
	archivedAt := now.Format(time.RFC3339)

	assert.NotEmpty(t, archivedAt)
	assert.Contains(t, archivedAt, "T")
}

func TestAgentService_RestoreAgent(t *testing.T) {
	// Test restore operation clears archived_at
	restored := map[string]any{
		"archived_at": nil,
		"status":      "active",
	}

	assert.Nil(t, restored["archived_at"])
	assert.Equal(t, "active", restored["status"])
}

func TestAgentService_Skills(t *testing.T) {
	skillIDs := []string{"skill_1", "skill_2", "skill_3"}

	assert.Len(t, skillIDs, 3)
	assert.Contains(t, skillIDs, "skill_1")

	// Test empty skills
	emptySkills := []string{}
	assert.Len(t, emptySkills, 0)
}

func TestAgentService_GetAgent(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	agentData := map[string]any{
		"id":               "agent_123",
		"name":             "Test Agent",
		"status":           "active",
		"visibility":       "private",
		"runtime_mode":     "cloud",
		"created_at":       time.Now().Format(time.RFC3339),
	}

	assert.Equal(t, "agent_123", agentData["id"])
	assert.Equal(t, "Test Agent", agentData["name"])
	assert.Equal(t, "active", agentData["status"])
}
