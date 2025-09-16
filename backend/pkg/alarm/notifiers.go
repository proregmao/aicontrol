package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	FromAddress  string
	ToAddresses  []string
	logger       *log.Logger
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(smtpHost string, smtpPort int, username, password, fromAddress string, toAddresses []string) *EmailNotifier {
	return &EmailNotifier{
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
		Username:    username,
		Password:    password,
		FromAddress: fromAddress,
		ToAddresses: toAddresses,
		logger:      log.New(log.Writer(), "[EMAIL_NOTIFIER] ", log.LstdFlags),
	}
}

// Send 发送邮件通知
func (n *EmailNotifier) Send(alarm *AlarmLog) error {
	subject := fmt.Sprintf("[%s] %s", alarm.Level, alarm.Title)
	body := fmt.Sprintf(`
告警详情:
- 规则名称: %s
- 告警级别: %s
- 告警描述: %s
- 数据源: %s
- 首次触发: %s
- 最后触发: %s
- 触发次数: %d

原始数据:
%s
`, alarm.RuleName, alarm.Level, alarm.Description, alarm.Source,
		alarm.FirstTime.Format("2006-01-02 15:04:05"),
		alarm.LastTime.Format("2006-01-02 15:04:05"),
		alarm.Count, alarm.Data)

	// 这里应该实现真正的SMTP发送逻辑
	// 为了演示，我们只是记录日志
	n.logger.Printf("发送邮件通知: %s -> %v", subject, n.ToAddresses)
	n.logger.Printf("邮件内容: %s", body)

	return nil
}

// GetType 获取通知器类型
func (n *EmailNotifier) GetType() string {
	return "email"
}

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
	WebhookURL string
	Secret     string
	logger     *log.Logger
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier(webhookURL, secret string) *DingTalkNotifier {
	return &DingTalkNotifier{
		WebhookURL: webhookURL,
		Secret:     secret,
		logger:     log.New(log.Writer(), "[DINGTALK_NOTIFIER] ", log.LstdFlags),
	}
}

// Send 发送钉钉通知
func (n *DingTalkNotifier) Send(alarm *AlarmLog) error {
	// 构建钉钉消息
	message := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": fmt.Sprintf("[%s] %s", alarm.Level, alarm.Title),
			"text": fmt.Sprintf(`## 🚨 系统告警通知

**告警规则:** %s  
**告警级别:** %s  
**告警描述:** %s  
**数据源:** %s  
**首次触发:** %s  
**最后触发:** %s  
**触发次数:** %d  

> 请及时处理相关问题！`,
				alarm.RuleName,
				alarm.Level,
				alarm.Description,
				alarm.Source,
				alarm.FirstTime.Format("2006-01-02 15:04:05"),
				alarm.LastTime.Format("2006-01-02 15:04:05"),
				alarm.Count),
		},
	}

	// 序列化消息
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化钉钉消息失败: %v", err)
	}

	// 发送HTTP请求
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		return fmt.Errorf("发送钉钉消息失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("钉钉API返回错误状态码: %d", resp.StatusCode)
	}

	n.logger.Printf("钉钉通知已发送: %s", alarm.Title)
	return nil
}

// GetType 获取通知器类型
func (n *DingTalkNotifier) GetType() string {
	return "dingtalk"
}

// WebhookNotifier Webhook通知器
type WebhookNotifier struct {
	URL     string
	Headers map[string]string
	logger  *log.Logger
}

// NewWebhookNotifier 创建Webhook通知器
func NewWebhookNotifier(url string, headers map[string]string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:     url,
		Headers: headers,
		logger:  log.New(log.Writer(), "[WEBHOOK_NOTIFIER] ", log.LstdFlags),
	}
}

// Send 发送Webhook通知
func (n *WebhookNotifier) Send(alarm *AlarmLog) error {
	// 构建Webhook消息
	payload := map[string]interface{}{
		"alarm_id":    alarm.ID,
		"rule_id":     alarm.RuleID,
		"rule_name":   alarm.RuleName,
		"level":       alarm.Level,
		"title":       alarm.Title,
		"description": alarm.Description,
		"source":      alarm.Source,
		"status":      alarm.Status,
		"count":       alarm.Count,
		"first_time":  alarm.FirstTime.Unix(),
		"last_time":   alarm.LastTime.Unix(),
		"data":        alarm.Data,
		"timestamp":   time.Now().Unix(),
	}

	// 序列化消息
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化Webhook消息失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", n.URL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range n.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Webhook请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook返回错误状态码: %d", resp.StatusCode)
	}

	n.logger.Printf("Webhook通知已发送: %s -> %s", alarm.Title, n.URL)
	return nil
}

// GetType 获取通知器类型
func (n *WebhookNotifier) GetType() string {
	return "webhook"
}

// SMSNotifier 短信通知器
type SMSNotifier struct {
	APIKey      string
	APISecret   string
	ServiceURL  string
	PhoneNumbers []string
	logger      *log.Logger
}

// NewSMSNotifier 创建短信通知器
func NewSMSNotifier(apiKey, apiSecret, serviceURL string, phoneNumbers []string) *SMSNotifier {
	return &SMSNotifier{
		APIKey:      apiKey,
		APISecret:   apiSecret,
		ServiceURL:  serviceURL,
		PhoneNumbers: phoneNumbers,
		logger:      log.New(log.Writer(), "[SMS_NOTIFIER] ", log.LstdFlags),
	}
}

// Send 发送短信通知
func (n *SMSNotifier) Send(alarm *AlarmLog) error {
	// 构建短信内容
	content := fmt.Sprintf("[%s]%s: %s，触发时间: %s",
		alarm.Level,
		alarm.Title,
		alarm.Description,
		alarm.LastTime.Format("15:04:05"))

	// 限制短信长度
	if len(content) > 70 {
		content = content[:67] + "..."
	}

	// 构建短信API请求
	smsData := map[string]interface{}{
		"api_key":    n.APIKey,
		"content":    content,
		"phones":     n.PhoneNumbers,
		"timestamp":  time.Now().Unix(),
	}

	// 序列化请求数据
	smsJSON, err := json.Marshal(smsData)
	if err != nil {
		return fmt.Errorf("序列化短信数据失败: %v", err)
	}

	// 发送HTTP请求
	resp, err := http.Post(n.ServiceURL, "application/json", bytes.NewBuffer(smsJSON))
	if err != nil {
		return fmt.Errorf("发送短信请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("短信API返回错误状态码: %d", resp.StatusCode)
	}

	n.logger.Printf("短信通知已发送: %s -> %v", alarm.Title, n.PhoneNumbers)
	return nil
}

// GetType 获取通知器类型
func (n *SMSNotifier) GetType() string {
	return "sms"
}

// ConsoleNotifier 控制台通知器（用于测试）
type ConsoleNotifier struct {
	logger *log.Logger
}

// NewConsoleNotifier 创建控制台通知器
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{
		logger: log.New(log.Writer(), "[CONSOLE_NOTIFIER] ", log.LstdFlags),
	}
}

// Send 发送控制台通知
func (n *ConsoleNotifier) Send(alarm *AlarmLog) error {
	n.logger.Printf(`
========== 告警通知 ==========
告警ID: %d
规则名称: %s
告警级别: %s
告警标题: %s
告警描述: %s
数据源: %s
状态: %s
触发次数: %d
首次触发: %s
最后触发: %s
原始数据: %s
=============================`,
		alarm.ID,
		alarm.RuleName,
		alarm.Level,
		alarm.Title,
		alarm.Description,
		alarm.Source,
		alarm.Status,
		alarm.Count,
		alarm.FirstTime.Format("2006-01-02 15:04:05"),
		alarm.LastTime.Format("2006-01-02 15:04:05"),
		alarm.Data)

	return nil
}

// GetType 获取通知器类型
func (n *ConsoleNotifier) GetType() string {
	return "console"
}
