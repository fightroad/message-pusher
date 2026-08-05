package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"message-pusher/model"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type gotifyOptions struct {
	Priority string `json:"priority"`
}

type gotifyMessageRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

type gotifyMessageResponse struct {
	Id        int    `json:"id"`
	Error     string `json:"error"`
	ErrorDesc string `json:"errorDescription"`
}

func parseGotifyOptions(raw string) gotifyOptions {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return gotifyOptions{}
	}
	var opts gotifyOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		return gotifyOptions{}
	}
	return opts
}

func SendGotifyMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	server := strings.TrimRight(strings.TrimSpace(channel_.URL), "/")
	token := strings.TrimSpace(channel_.Secret)
	if server == "" || token == "" {
		return errors.New("未配置 Gotify 服务器地址或应用令牌")
	}
	body := message.Content
	if body == "" {
		body = message.Description
	}
	opts := parseGotifyOptions(channel_.Other)
	priority := 0
	if p := strings.TrimSpace(opts.Priority); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			priority = v
		}
	}
	jsonData, err := json.Marshal(gotifyMessageRequest{
		Title:    message.Title,
		Message:  body,
		Priority: priority,
	})
	if err != nil {
		return err
	}
	u, err := url.Parse(server + "/message")
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	resp, err := http.Post(u.String(), "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var res gotifyMessageResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errors.New(strings.TrimSpace(string(respBody)))
		}
		return err
	}
	if res.Id == 0 || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := res.ErrorDesc
		if msg == "" {
			msg = res.Error
		}
		if msg == "" {
			msg = strings.TrimSpace(string(respBody))
		}
		if msg == "" {
			msg = resp.Status
		}
		return errors.New(msg)
	}
	return nil
}
