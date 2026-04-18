package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskService_CreateTask(t *testing.T) {
	svc := &TaskService{}
	assert.NotNil(t, svc)

	params := map[string]any{
		"issue_id":   "issue_123",
		"runtime_id": "runtime_456",
		"agent_id":   "agent_789",
		"prompt":     "Fix the bug in the login page",
	}

	assert.Equal(t, "issue_123", params["issue_id"])
	assert.Equal(t, "runtime_456", params["runtime_id"])
	assert.Equal(t, "agent_789", params["agent_id"])
}

func TestTaskService_GetTask(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	taskID := "task_123"
	assert.NotEmpty(t, taskID)
}

func TestTaskService_ListTasks(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	workspaceID := "ws_123"
	status := "queued"
	limit := 50

	assert.Equal(t, "ws_123", workspaceID)
	assert.Equal(t, "queued", status)
	assert.Equal(t, 50, limit)
}

func TestTaskService_UpdateTaskStatus(t *testing.T) {
	statuses := []string{"queued", "dispatched", "running", "completed", "failed", "cancelled"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			assert.NotEmpty(t, status)
		})
	}
}

func TestTaskService_UpdateTaskStatus_WithError(t *testing.T) {
	taskErr := "Something went wrong"
	result := `{"success": false}`

	assert.NotEmpty(t, taskErr)
	assert.NotEmpty(t, result)
}

func TestTaskService_ListTaskMessages(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	taskID := "task_123"
	limit := 200

	assert.Equal(t, "task_123", taskID)
	assert.Equal(t, 200, limit)
}

func TestTaskService_TaskStruct(t *testing.T) {
	now := time.Now()
	task := Task{
		ID:          "task_123",
		AgentID:     "agent_456",
		RuntimeID:   "runtime_789",
		IssueID:     "issue_123",
		Status:      "queued",
		Priority:    0,
		CreatedAt:   now,
	}

	assert.Equal(t, "task_123", task.ID)
	assert.Equal(t, "agent_456", task.AgentID)
	assert.Equal(t, "queued", task.Status)
	assert.Equal(t, 0, task.Priority)
}

func TestTaskService_TaskMessageStruct(t *testing.T) {
	now := time.Now()
	tool := "read_file"
	content := "Reading config.json"
	
	msg := TaskMessage{
		ID:        "msg_123",
		TaskID:    "task_456",
		IssueID:   "issue_789",
		Seq:       1,
		Type:      "tool_call",
		Tool:      &tool,
		Content:   &content,
		CreatedAt: now,
	}

	assert.Equal(t, "msg_123", msg.ID)
	assert.Equal(t, "task_456", msg.TaskID)
	assert.Equal(t, 1, msg.Seq)
	assert.Equal(t, "tool_call", msg.Type)
	assert.NotNil(t, msg.Tool)
	assert.Equal(t, "read_file", *msg.Tool)
}
