package notification

import (
	"fmt"
	"log"
	"os"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// NotificationPayload 通知事件的上下文数据
type NotificationPayload struct {
	ServiceName string
	ClusterName string
	Env         string
	Namespace   string
	ImageTag    string
	TriggerUser string
	FailReason  string
	DeployID    uint
	BuildID     uint
}

// Dispatcher 通知调度器，在关键事件发生时异步推送到绑定的渠道
type Dispatcher struct {
	ruleRepo    repository.NotificationRuleRepository
	svcRuleRepo repository.ServiceNotificationRuleRepository
	logRepo     repository.NotificationLogRepository
	sender      *WebhookSender
}

func NewDispatcher(
	ruleRepo repository.NotificationRuleRepository,
	svcRuleRepo repository.ServiceNotificationRuleRepository,
	logRepo repository.NotificationLogRepository,
	sender *WebhookSender,
) *Dispatcher {
	return &Dispatcher{
		ruleRepo:    ruleRepo,
		svcRuleRepo: svcRuleRepo,
		logRepo:     logRepo,
		sender:      sender,
	}
}

// Dispatch 异步发送通知，不阻塞业务流程
func (d *Dispatcher) Dispatch(serviceID uint, eventType string, payload NotificationPayload) {
	go d.dispatch(serviceID, eventType, payload)
}

func (d *Dispatcher) dispatch(serviceID uint, eventType string, payload NotificationPayload) {
	// 查找服务级规则
	channels := d.resolveChannels(serviceID, eventType)
	if len(channels) == 0 {
		return
	}

	title := buildTitle(eventType, payload.ServiceName)
	content := buildContent(eventType, payload)

	for _, ch := range channels {
		if ch.WebhookURL == "" {
			continue
		}
		err := d.sender.Send(ch.WebhookURL, ch.Type, title, content)
		logEntry := &model.NotificationLog{
			ServiceID: serviceID,
			ChannelID: ch.ID,
			EventType: eventType,
			Title:     title,
			Content:   content,
			Status:    "sent",
			CreatedAt: time.Now(),
		}
		if err != nil {
			logEntry.Status = "failed"
			logEntry.ErrorMsg = err.Error()
			log.Printf("[Dispatcher] 发送通知失败 svc=%d event=%s channel=%s: %v", serviceID, eventType, ch.Name, err)
		}
		_ = d.logRepo.Create(logEntry)
	}
}

// resolveChannels 三层优先级: 服务级绑定 > 全局事件自定义 > 全局 All 默认
func (d *Dispatcher) resolveChannels(serviceID uint, eventType string) []model.NotificationChannel {
	// 第一层: 服务级绑定（一个渠道覆盖所有事件）
	svcRule, err := d.svcRuleRepo.FindByService(serviceID)
	if err == nil && svcRule != nil && svcRule.Channel != nil {
		return []model.NotificationChannel{*svcRule.Channel}
	}

	// 第二层: 全局事件自定义（具体事件类型）
	specificRules, err := d.ruleRepo.FindByEventType(eventType)
	if err == nil && len(specificRules) > 0 {
		var channels []model.NotificationChannel
		for _, r := range specificRules {
			if r.Channel != nil {
				channels = append(channels, *r.Channel)
			}
		}
		if len(channels) > 0 {
			return channels
		}
	}

	// 第三层: 全局 All 默认渠道
	allRules, err := d.ruleRepo.FindByEventType(model.EventAll)
	if err == nil && len(allRules) > 0 {
		var channels []model.NotificationChannel
		for _, r := range allRules {
			if r.Channel != nil {
				channels = append(channels, *r.Channel)
			}
		}
		return channels
	}

	return nil
}

// 事件标题映射
var eventLabels = map[string]string{
	model.EventBuildSuccess:      "构建成功",
	model.EventBuildFailed:       "构建失败",
	model.EventBuildCancelled:    "构建取消",
	model.EventDeploySuccess:     "部署成功",
	model.EventDeployFailed:      "部署失败",
	model.EventApprovalPending:   "待审批",
	model.EventPodUnhealthy:      "Pod 异常",
	model.EventRollbackTriggered: "回滚触发",
	model.EventDeployCancelled:   "部署取消",
}

func buildTitle(eventType, serviceName string) string {
	label := eventLabels[eventType]
	if label == "" {
		label = eventType
	}
	return fmt.Sprintf("[DeployHub] %s — %s", label, serviceName)
}

func buildContent(eventType string, p NotificationPayload) string {
	lines := []string{}
	lines = append(lines, fmt.Sprintf("服务: %s", p.ServiceName))
	if p.ClusterName != "" {
		env := p.Env
		if env == "" {
			env = "-"
		}
		lines = append(lines, fmt.Sprintf("集群: %s (%s)", p.ClusterName, env))
	}
	if p.Namespace != "" {
		lines = append(lines, fmt.Sprintf("命名空间: %s", p.Namespace))
	}
	if p.ImageTag != "" {
		lines = append(lines, fmt.Sprintf("镜像: %s", p.ImageTag))
	}
	if p.TriggerUser != "" {
		lines = append(lines, fmt.Sprintf("触发人: %s", p.TriggerUser))
	}
	if p.FailReason != "" {
		lines = append(lines, fmt.Sprintf("原因: %s", p.FailReason))
	}
	lines = append(lines, fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))

	// 跳转链接
	baseURL := os.Getenv("DEPLOYHUB_BASE_URL")
	if baseURL != "" {
		if p.DeployID > 0 {
			lines = append(lines, fmt.Sprintf("查看: %s/deployments/%d", baseURL, p.DeployID))
		} else if p.BuildID > 0 {
			lines = append(lines, fmt.Sprintf("查看: %s/builds?id=%d", baseURL, p.BuildID))
		}
	}

	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	return result
}
