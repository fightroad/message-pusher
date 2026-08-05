package channel

import (
	"encoding/json"
	"errors"
	"io"
	"message-pusher/model"
	"net/http"
	"net/url"
	"strings"
)

type pushMeOptions struct {
	Date string `json:"date"`
	Type string `json:"type"`
}

func parsePushMeOptions(raw string) pushMeOptions {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pushMeOptions{}
	}
	var opts pushMeOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		return pushMeOptions{}
	}
	return opts
}

func SendPushMeMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	pushKey := strings.TrimSpace(channel_.Secret)
	if pushKey == "" {
		return errors.New("未配置 PushMe push_key")
	}
	apiURL := strings.TrimSpace(channel_.URL)
	if apiURL == "" {
		apiURL = "https://push.i-i.me/"
	}
	body := message.Content
	if body == "" {
		body = message.Description
	}
	opts := parsePushMeOptions(channel_.Other)
	form := url.Values{}
	form.Set("push_key", pushKey)
	form.Set("title", message.Title)
	form.Set("content", body)
	if date := strings.TrimSpace(opts.Date); date != "" {
		form.Set("date", date)
	}
	if typ := strings.TrimSpace(opts.Type); typ != "" {
		form.Set("type", typ)
	}
	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(string(respBody))
	if resp.StatusCode == http.StatusOK && text == "success" {
		return nil
	}
	if text == "" {
		text = resp.Status
	}
	return errors.New(text)
}
