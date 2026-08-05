package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"message-pusher/common"
	"message-pusher/model"
)

type customChannelConfig struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func applyCustomTemplate(template string, message *model.Message) string {
	template = common.Replace(template, "$url", message.URL, -1)
	template = common.Replace(template, "$to", message.To, -1)
	template = common.Replace(template, "$title", message.Title, -1)
	template = common.Replace(template, "$description", message.Description, -1)
	template = common.Replace(template, "$content", message.Content, -1)
	return template
}

func parseCustomChannelOther(raw string) (headers map[string]string, body string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ""
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &probe); err != nil {
		return nil, raw
	}
	if _, hasHeaders := probe["headers"]; !hasHeaders {
		return nil, raw
	}
	if _, hasBody := probe["body"]; !hasBody {
		return nil, raw
	}
	var cfg customChannelConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil || cfg.Body == "" {
		return nil, raw
	}
	return cfg.Headers, cfg.Body
}

func SendCustomMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	url := channel_.URL
	if strings.HasPrefix(url, "http:") && os.Getenv("CHANNEL_URL_ALLOW_NON_HTTPS") != "true" {
		return errors.New("自定义通道必须使用 HTTPS 协议")
	}
	if strings.HasPrefix(url, common.ServerAddress) {
		return errors.New("自定义通道不能使用本服务地址")
	}
	headers, bodyTemplate := parseCustomChannelOther(channel_.Other)
	body := applyCustomTemplate(bodyTemplate, message)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}
	contentTypeSet := false
	for key, value := range headers {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		req.Header.Set(key, applyCustomTemplate(value, message))
		if strings.EqualFold(key, "Content-Type") {
			contentTypeSet = true
		}
	}
	if !contentTypeSet {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}
