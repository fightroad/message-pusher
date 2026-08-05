package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"message-pusher/model"
	"net/http"
	"strings"
	"time"
)

type discordMessageRequest struct {
	Content string `json:"content"`
}

type discordMessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func SendDiscordMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	if message.Content == "" {
		message.Content = message.Description
	}
	messageRequest := discordMessageRequest{
		Content: message.Content,
	}
	// https://discord.com/developers/docs/reference#message-formatting
	if message.To != "" {
		messageRequest.Content = ""
		ids := strings.Split(message.To, "|")
		for _, id := range ids {
			messageRequest.Content = "<@" + id + "> " + messageRequest.Content
		}
		messageRequest.Content = messageRequest.Content + message.Content
	}

	jsonData, err := json.Marshal(messageRequest)
	if err != nil {
		return err
	}
	client, err := NewHTTPClient(strings.TrimSpace(channel_.Other), 30*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.Post(channel_.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var res discordMessageResponse
	if err := json.Unmarshal(body, &res); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := strings.TrimSpace(string(body))
			if msg == "" {
				msg = resp.Status
			}
			return errors.New(msg)
		}
		return err
	}
	if res.Code != 0 {
		return errors.New(res.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if res.Message != "" {
			return errors.New(res.Message)
		}
		return errors.New(resp.Status)
	}
	return nil
}
