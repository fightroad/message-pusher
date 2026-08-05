package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"message-pusher/common"
	"message-pusher/model"
	"net/http"
	"strings"
)

type barkOptions struct {
	Group     string `json:"group"`
	Sound     string `json:"sound"`
	Icon      string `json:"icon"`
	Level     string `json:"level"`
	IsArchive string `json:"is_archive"`
	URL       string `json:"url"`
}

type barkMessageRequest struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	URL       string `json:"url,omitempty"`
	Group     string `json:"group,omitempty"`
	Sound     string `json:"sound,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Level     string `json:"level,omitempty"`
	IsArchive string `json:"isArchive,omitempty"`
}

type barkMessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func parseBarkOptions(raw string) barkOptions {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return barkOptions{}
	}
	var opts barkOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		return barkOptions{}
	}
	return opts
}

func SendBarkMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	opts := parseBarkOptions(channel_.Other)
	url := fmt.Sprintf("%s/%s", channel_.URL, channel_.Secret)
	req := barkMessageRequest{
		Title:     message.Title,
		Body:      message.Content,
		URL:       message.URL,
		Group:     strings.TrimSpace(opts.Group),
		Sound:     strings.TrimSpace(opts.Sound),
		Icon:      strings.TrimSpace(opts.Icon),
		Level:     strings.TrimSpace(opts.Level),
		IsArchive: strings.TrimSpace(opts.IsArchive),
	}
	if message.Content == "" {
		req.Body = message.Description
	}
	jumpURL := strings.TrimSpace(opts.URL)
	if jumpURL != "" && (message.URL == "" || strings.HasPrefix(message.URL, common.ServerAddress+"/message/")) {
		req.URL = jumpURL
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res barkMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return err
	}
	if res.Code != 200 {
		return errors.New(res.Message)
	}
	return nil
}
