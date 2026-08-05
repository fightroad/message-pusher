package channel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"message-pusher/model"
	"net/http"
	"net/url"
	"strings"
)

type ntfyOptions struct {
	Priority string `json:"priority"`
	Icon     string `json:"icon"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func parseNtfyOptions(raw string) ntfyOptions {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ntfyOptions{}
	}
	var opts ntfyOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		return ntfyOptions{}
	}
	return opts
}

func encodeNtfyHeader(text string) string {
	if text == "" {
		return ""
	}
	for _, r := range text {
		if r > 127 {
			return fmt.Sprintf("=?utf-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(text)))
		}
	}
	return text
}

func SendNtfyMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	topic := strings.TrimSpace(channel_.AccountId)
	if topic == "" {
		return errors.New("未配置 ntfy topic")
	}
	server := strings.TrimRight(strings.TrimSpace(channel_.URL), "/")
	if server == "" {
		server = "https://ntfy.sh"
	}
	opts := parseNtfyOptions(channel_.Other)
	body := message.Content
	if body == "" {
		body = message.Description
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", server, url.PathEscape(topic)), bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	if message.Title != "" {
		req.Header.Set("Title", encodeNtfyHeader(message.Title))
	}
	priority := strings.TrimSpace(opts.Priority)
	if priority == "" {
		priority = "3"
	}
	req.Header.Set("Priority", priority)
	if icon := strings.TrimSpace(opts.Icon); icon != "" {
		req.Header.Set("Icon", icon)
	}
	if message.URL != "" {
		req.Header.Set("Click", message.URL)
	}
	token := strings.TrimSpace(channel_.Secret)
	username := strings.TrimSpace(opts.Username)
	password := strings.TrimSpace(opts.Password)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if username != "" && password != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = resp.Status
		}
		return errors.New(msg)
	}
	return nil
}
