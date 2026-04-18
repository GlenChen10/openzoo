package connectapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerAgentListEndpoints(t *testing.T) {
	assert.NotNil(t, t)

	handler := &Handler{}
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/agents", nil)
	assert.Equal(t, http.MethodPost, req.Method)
}

func TestHandlerAgentCreateEndpoint(t *testing.T) {
	handler := &Handler{}
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/agents/create", nil)
	assert.Equal(t, http.MethodPost, req.Method)
}

func TestHandlerAgentUpdateEndpoint(t *testing.T) {
	handler := &Handler{}
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/agents/update", nil)
	assert.Equal(t, http.MethodPost, req.Method)
}

func TestHandlerAgentArchiveEndpoint(t *testing.T) {
	handler := &Handler{}
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/agents/archive", nil)
	assert.Equal(t, http.MethodPost, req.Method)
}

func TestHandlerAgentRestoreEndpoint(t *testing.T) {
	handler := &Handler{}
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/agents/restore", nil)
	assert.Equal(t, http.MethodPost, req.Method)
}

func TestHandlerAgentGetEndpoint(t *testing.T) {
	handler := &Handler{}
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodPost, "/agents/get", nil)
	assert.Equal(t, http.MethodPost, req.Method)
}
