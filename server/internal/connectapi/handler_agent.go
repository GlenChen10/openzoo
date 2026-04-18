package connectapi
import "net/http"
func (h *Handler) listAgents(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	agents, err := h.agent.List(r.Context(), getStr(in, "workspace_id"))
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, map[string]interface{}{"agents": agents})
}
func (h *Handler) getAgent(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	a, err := h.agent.Get(r.Context(), getStr(in, "workspace_id"), getStr(in, "agent_id"))
	if err != nil { writeJSON(w, 404, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, a)
}
func (h *Handler) createAgent(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	ws := getStr(in, "workspace_id")
	a, err := h.agent.Create(r.Context(), ws, getStr(in, "name"), getStr(in, "description"), getStr(in, "instructions"), getStr(in, "runtime_id"))
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	h.publisher.PublishWorkspace(r.Context(), ws, "agent:created", a)
	writeJSON(w, 201, a)
}
func (h *Handler) updateAgent(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	fields := make(map[string]interface{})
	for k, v := range in { if k != "workspace_id" && k != "agent_id" { fields[k] = v } }
	a, err := h.agent.Update(r.Context(), getStr(in, "workspace_id"), getStr(in, "agent_id"), fields)
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, a)
}
func (h *Handler) archiveAgent(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	ws := getStr(in, "workspace_id")
	a, err := h.agent.Archive(r.Context(), ws, getStr(in, "agent_id"))
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	h.publisher.PublishWorkspace(r.Context(), ws, "agent:archived", a)
	writeJSON(w, 200, a)
}
func (h *Handler) restoreAgent(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	ws := getStr(in, "workspace_id")
	a, err := h.agent.Restore(r.Context(), ws, getStr(in, "agent_id"))
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	h.publisher.PublishWorkspace(r.Context(), ws, "agent:restored", a)
	writeJSON(w, 200, a)
}
func (h *Handler) listRuntimes(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	runtimes, err := h.runtime.List(r.Context(), getStr(in, "workspace_id"))
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, map[string]interface{}{"runtimes": runtimes})
}
func (h *Handler) registerRuntime(w http.ResponseWriter, r *http.Request) {
	in := readJSON(r)
	ws := getStr(in, "workspace_id")
	rt, err := h.runtime.Register(r.Context(), ws, getStr(in, "name"), getStr(in, "provider"), getStr(in, "runtime_mode"), getStr(in, "device_info"))
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	h.publisher.PublishWorkspace(r.Context(), ws, "runtime:registered", rt)
	writeJSON(w, 201, rt)
}
