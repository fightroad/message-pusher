package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"message-pusher/model"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var TelegramMaxMessageLength = 4096

const DefaultTelegramAPIBase = "https://api.telegram.org"

type telegramMessageRequest struct {
	ChatId    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type telegramMessageResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

type telegramChat struct {
	Id int64 `json:"id"`
}

type telegramUpdateMessage struct {
	Chat telegramChat `json:"chat"`
}

type telegramUpdate struct {
	Message       *telegramUpdateMessage `json:"message"`
	EditedMessage *telegramUpdateMessage `json:"edited_message"`
	ChannelPost   *telegramUpdateMessage `json:"channel_post"`
}

type telegramGetUpdatesResponse struct {
	Ok          bool             `json:"ok"`
	Description string           `json:"description"`
	Result      []telegramUpdate `json:"result"`
}

func getTelegramHTTPClient(proxyAddr string) (*http.Client, error) {
	return NewHTTPClient(proxyAddr, 30*time.Second)
}

func resolveTelegramAPIBase(apiBase string) string {
	if apiBase == "" {
		apiBase = DefaultTelegramAPIBase
	}
	return strings.TrimRight(apiBase, "/")
}

func GetTelegramChatId(apiBase, botToken, proxyAddr string) (string, error) {
	if botToken == "" {
		return "", errors.New("bot token is required")
	}
	client, err := getTelegramHTTPClient(proxyAddr)
	if err != nil {
		return "", err
	}
	apiBase = resolveTelegramAPIBase(apiBase)
	resp, err := client.Get(fmt.Sprintf("%s/bot%s/getUpdates", apiBase, botToken))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var res telegramGetUpdatesResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	if !res.Ok {
		if res.Description != "" {
			return "", errors.New(res.Description)
		}
		return "", errors.New("telegram getUpdates failed")
	}
	if len(res.Result) == 0 {
		return "", errors.New("请先向你的机器人发送一条任意消息")
	}
	for i := len(res.Result) - 1; i >= 0; i-- {
		update := res.Result[i]
		if update.Message != nil {
			return strconv.FormatInt(update.Message.Chat.Id, 10), nil
		}
		if update.EditedMessage != nil {
			return strconv.FormatInt(update.EditedMessage.Chat.Id, 10), nil
		}
		if update.ChannelPost != nil {
			return strconv.FormatInt(update.ChannelPost.Chat.Id, 10), nil
		}
	}
	return "", errors.New("未能解析会话 ID，请手动填写")
}

func SendTelegramMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	// https://core.telegram.org/bots/api#sendmessage
	messageRequest := telegramMessageRequest{
		ChatId: channel_.AccountId,
	}
	if message.To != "" {
		messageRequest.ChatId = message.To
	}
	if message.Content == "" {
		messageRequest.Text = message.Description
	} else {
		messageRequest.Text = message.Content
		messageRequest.ParseMode = "markdown"
	}
	text := messageRequest.Text
	apiBase := resolveTelegramAPIBase(channel_.URL)
	client, err := getTelegramHTTPClient(channel_.Other)
	if err != nil {
		return err
	}
	idx := 0
	for idx < len(text) {
		nextIdx := idx + TelegramMaxMessageLength
		if nextIdx > len(text) {
			// we have reach the end, must be valid
			nextIdx = len(text)
		} else {
			nextIdx = getNearestValidSplit(text, nextIdx, messageRequest.ParseMode)
		}
		messageRequest.Text = text[idx:nextIdx]
		idx = nextIdx
		jsonData, err := json.Marshal(messageRequest)
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("%s/bot%s/sendMessage", apiBase, channel_.Secret), "application/json",
			bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		var res telegramMessageResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if !res.Ok {
			return errors.New(res.Description)
		}
	}
	return nil
}

func getNearestValidSplit(s string, idx int, mode string) int {
	if mode == "markdown" {
		return getMarkdownNearestValidSplit(s, idx)
	} else {
		return getPlainTextNearestValidSplit(s, idx)
	}
}

func getPlainTextNearestValidSplit(s string, idx int) int {
	if idx >= len(s) {
		return idx
	}
	if idx == 0 {
		return 0
	}
	isStartByte := utf8.RuneStart(s[idx])
	if isStartByte {
		return idx
	} else {
		return getPlainTextNearestValidSplit(s, idx-1)
	}
}

func getMarkdownNearestValidSplit(s string, idx int) int {
	if idx >= len(s) {
		return idx
	}
	if idx == 0 {
		return 0
	}
	for i := idx; i >= 0; i-- {
		if s[i] == '\n' {
			return i + 1
		}
	}
	// unable to find a '\n'
	return idx
}
