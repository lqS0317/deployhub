package notification

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSender_Feishu(t *testing.T) {
	var received map[string]interface{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{httpClient: server.Client()}
	err := sender.Send(server.URL, "feishu", "测试标题", "测试内容")
	require.NoError(t, err)

	assert.Equal(t, "text", received["msg_type"])
	content := received["content"].(map[string]interface{})
	assert.Contains(t, content["text"], "测试标题")
	assert.Contains(t, content["text"], "测试内容")
}

func TestWebhookSender_DingTalk(t *testing.T) {
	var received map[string]interface{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{httpClient: server.Client()}
	err := sender.Send(server.URL, "dingtalk", "测试标题", "测试内容")
	require.NoError(t, err)

	assert.Equal(t, "text", received["msgtype"])
	text := received["text"].(map[string]interface{})
	assert.Contains(t, text["content"], "测试标题")
	assert.Contains(t, text["content"], "测试内容")
}

func TestWebhookSender_Slack(t *testing.T) {
	var received map[string]interface{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{httpClient: server.Client()}
	err := sender.Send(server.URL, "slack", "测试标题", "测试内容")
	require.NoError(t, err)

	assert.Contains(t, received["text"], "测试标题")
	assert.Contains(t, received["text"], "测试内容")
}

func TestWebhookSender_Generic(t *testing.T) {
	var received map[string]interface{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{httpClient: server.Client()}
	err := sender.Send(server.URL, "generic", "测试标题", "测试内容")
	require.NoError(t, err)

	assert.Equal(t, "测试标题", received["title"])
	assert.Equal(t, "测试内容", received["content"])
}

func TestWebhookSender_NonHTTPS(t *testing.T) {
	sender := NewWebhookSender()
	err := sender.Send("http://example.com/hook", "generic", "标题", "内容")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTPS")
}

func TestWebhookSender_InvalidChannelType(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{httpClient: server.Client()}
	err := sender.Send(server.URL, "unknown", "标题", "内容")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的渠道类型")
}

func TestWebhookSender_ServerError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := &WebhookSender{httpClient: server.Client()}
	err := sender.Send(server.URL, "generic", "标题", "内容")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
