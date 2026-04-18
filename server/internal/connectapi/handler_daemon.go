package connectapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/openzoo-ai/openzoo/server/internal/service"
)

func (h *Handler) registerDaemon(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	wsID := getStr(in, "workspace_id")
	if wsID == "" {
		writeJSON(w, 400, map[string]string{"error": "workspace_id is required"})
		return
	}
	name := getStr(in, "name")
	if name == "" {
		name = "daemon-" + uuid.New().String()[:8]
	}
	daemon, err := h.daemon.Register(r.Context(), name, getStr(in, "runtime_id"), wsID, getInt(in, "pid"), getInt(in, "port"))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, daemon)
}

func (h *Handler) unregisterDaemon(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	daemonID := getStr(in, "daemon_id")
	if daemonID == "" {
		writeJSON(w, 400, map[string]string{"error": "daemon_id is required"})
		return
	}
	if err := h.daemon.Unregister(r.Context(), daemonID); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) listDaemons(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	wsID := getStr(in, "workspace_id")
	daemons, err := h.daemon.List(r.Context(), wsID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if daemons == nil {
		daemons = []service.Daemon{}
	}
	writeJSON(w, 200, map[string]interface{}{"daemons": daemons})
}

func (h *Handler) getDaemonDetail(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	daemonID := getStr(in, "daemon_id")
	if daemonID == "" {
		writeJSON(w, 400, map[string]string{"error": "daemon_id is required"})
		return
	}
	daemon, err := h.daemon.Get(r.Context(), daemonID)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "daemon not found"})
		return
	}
	writeJSON(w, 200, daemon)
}

func (h *Handler) daemonHeartbeat(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	daemonID := getStr(in, "daemon_id")
	if daemonID == "" {
		writeJSON(w, 400, map[string]string{"error": "daemon_id is required"})
		return
	}
	if err := h.daemon.Heartbeat(r.Context(), daemonID); err != nil {
		writeJSON(w, 404, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) daemonStats(w http.ResponseWriter, r *http.Request) {
	stats := h.daemon.GetStats(r.Context())
	writeJSON(w, 200, stats)
}

func (h *Handler) claimTask(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	runtimeID := getStr(in, "runtime_id")
	task, err := h.task.ClaimTask(r.Context(), runtimeID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if task == nil {
		writeJSON(w, 200, map[string]interface{}{"task": nil})
		return
	}
	writeJSON(w, 200, map[string]interface{}{"task": task})
}

func (h *Handler) startTask(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	taskID := getStr(in, "task_id")
	if taskID == "" {
		writeJSON(w, 400, map[string]string{"error": "task_id is required"})
		return
	}
	if err := h.task.StartTask(r.Context(), taskID); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) completeTask(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	taskID := getStr(in, "task_id")
	if taskID == "" {
		writeJSON(w, 400, map[string]string{"error": "task_id is required"})
		return
	}
	resultData := in["result"]
	var resultJSON, comment string
	if resultData != nil {
		b, _ := json.Marshal(resultData)
		resultJSON = string(b)
		if m, ok := resultData.(map[string]interface{}); ok {
			if c, ok := m["comment"].(string); ok {
				comment = c
			}
		}
	}
	if err := h.task.CompleteTask(r.Context(), taskID, resultJSON, comment); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) failTask(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	taskID := getStr(in, "task_id")
	if taskID == "" {
		writeJSON(w, 400, map[string]string{"error": "task_id is required"})
		return
	}
	errMsg := getStr(in, "error")
	if err := h.task.FailTask(r.Context(), taskID, errMsg); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) reportMessages(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	taskID := getStr(in, "task_id")
	if taskID == "" {
		writeJSON(w, 400, map[string]string{"error": "task_id is required"})
		return
	}
	var messages []service.InboundMessage
	if raw, ok := in["messages"]; ok {
		b, _ := json.Marshal(raw)
		json.Unmarshal(b, &messages)
	}
	if err := h.task.ReportMessages(r.Context(), taskID, messages); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
