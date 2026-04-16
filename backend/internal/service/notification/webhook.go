package notification

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WebhookSender Webhook 消息发送器
type WebhookSender struct {
	httpClient *http.Client
}

// NewWebhookSender 创建 Webhook 发送器
func NewWebhookSender() *WebhookSender {
	return &WebhookSender{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			},
		},
	}
}

// Send 向指定渠道发送 Webhook 通知，webhookURL 必须使用 HTTPS
func (s *WebhookSender) Send(webhookURL, channelType, title, content string) error {
	parsed, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("无效的 Webhook URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("Webhook URL 必须使用 HTTPS 协议")
	}

	payload, err := buildPayload(channelType, title, content)
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化 Webhook 消息失败: %w", err)
	}

	resp, err := s.httpClient.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("发送 Webhook 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook 返回非成功状态码: %d", resp.StatusCode)
	}
	return nil
}

// buildPayload 根据渠道类型构建请求体
func buildPayload(channelType, title, content string) (interface{}, error) {
	combined := fmt.Sprintf("%s\n%s", title, content)

	switch channelType {
	case "feishu":
		return map[string]interface{}{
			"msg_type": "text",
			"content":  map[string]string{"text": combined},
		}, nil
	case "dingtalk":
		return map[string]interface{}{
			"msgtype": "text",
			"text":    map[string]string{"content": combined},
		}, nil
	case "slack":
		return map[string]string{"text": combined}, nil
	case "generic":
		return map[string]string{"title": title, "content": content}, nil
	default:
		return nil, fmt.Errorf("不支持的渠道类型: %s", channelType)
	}
}
