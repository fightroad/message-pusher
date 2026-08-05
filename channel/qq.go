package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"message-pusher/common"
	"message-pusher/model"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultQQBotAPIBase = "https://api.bot.qq.com"
	qqBotAccessTokenURL = "https://bots.qq.com/app/getAppAccessToken"
)

type qqBotTokenRequest struct {
	AppID        string `json:"appId"`
	ClientSecret string `json:"clientSecret"`
}

type qqBotTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   any    `json:"expires_in"`
	Code        int    `json:"code"`
	Message     string `json:"message"`
	ErrCode     int    `json:"err_code"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type QQBotTokenStoreItem struct {
	AppID       string
	AppSecret   string
	AccessToken string
}

func (i *QQBotTokenStoreItem) Key() string {
	return i.AppID + i.AppSecret
}

func (i *QQBotTokenStoreItem) IsShared() bool {
	var count int64 = 0
	model.DB.Model(&model.Channel{}).Where("secret = ? and app_id = ? and type = ?",
		i.AppSecret, i.AppID, model.TypeQQBot).Count(&count)
	return count > 1
}

func (i *QQBotTokenStoreItem) IsFilled() bool {
	return i.AppID != "" && i.AppSecret != ""
}

func (i *QQBotTokenStoreItem) Token() string {
	return i.AccessToken
}

func (i *QQBotTokenStoreItem) Refresh() {
	tokenRequest := qqBotTokenRequest{
		AppID:        i.AppID,
		ClientSecret: i.AppSecret,
	}
	tokenRequestData, err := json.Marshal(tokenRequest)
	if err != nil {
		common.SysError("failed to marshal qq bot token request: " + err.Error())
		return
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(qqBotAccessTokenURL, "application/json", bytes.NewBuffer(tokenRequestData))
	if err != nil {
		common.SysError("failed to refresh qq bot access token: " + err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		common.SysError("failed to read qq bot token response: " + err.Error())
		return
	}
	var res qqBotTokenResponse
	if err := json.Unmarshal(body, &res); err != nil {
		common.SysError("failed to decode qq bot token response: " + err.Error())
		return
	}
	if res.AccessToken == "" {
		msg := res.Message
		if msg == "" {
			msg = res.ErrorDesc
		}
		if msg == "" {
			msg = res.Error
		}
		if msg == "" {
			msg = strings.TrimSpace(string(body))
		}
		if msg == "" {
			msg = "empty access_token"
		}
		common.SysError("failed to refresh qq bot access token: " + msg)
		return
	}
	i.AccessToken = res.AccessToken
	common.SysLog("qq bot access token refreshed")
}

type qqBotMessageRequest struct {
	MsgType int    `json:"msg_type"`
	Content string `json:"content"`
}

type qqBotMessageResponse struct {
	ID      string `json:"id"`
	Code    int    `json:"code"`
	ErrCode int    `json:"err_code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
}

func parseQQBotTarget(target string) (string, string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", "", errors.New("请填写推送目标 OpenID，例如 user:xxx 或 group:xxx")
	}
	parts := strings.SplitN(target, ":", 2)
	if len(parts) == 1 {
		return "user", parts[0], nil
	}
	kind := strings.ToLower(parts[0])
	id := strings.TrimSpace(parts[1])
	if id == "" {
		return "", "", errors.New("无效的 QQ 机器人推送目标")
	}
	switch kind {
	case "user", "c2c":
		return "user", id, nil
	case "group":
		return "group", id, nil
	default:
		return "", "", errors.New("推送目标类型仅支持 user / group，例如 user:OPENID 或 group:OPENID")
	}
}

func buildQQBotText(message *model.Message) string {
	text := message.Content
	if text == "" {
		text = message.Description
	}
	if message.Title != "" {
		if text == "" {
			return message.Title
		}
		return message.Title + "\n" + text
	}
	return text
}

func SendQQBotMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	if channel_.AppId == "" || channel_.Secret == "" {
		return errors.New("请填写 QQ 机器人 AppID 与 AppSecret")
	}
	rawTarget := message.To
	if rawTarget == "" {
		rawTarget = channel_.AccountId
	}
	targetType, targetID, err := parseQQBotTarget(rawTarget)
	if err != nil {
		return err
	}
	text := buildQQBotText(message)
	if text == "" {
		return errors.New("消息内容为空")
	}

	reqBody := qqBotMessageRequest{MsgType: 0, Content: text}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	var path string
	if targetType == "group" {
		path = fmt.Sprintf("/v2/groups/%s/messages", targetID)
	} else {
		path = fmt.Sprintf("/v2/users/%s/messages", targetID)
	}

	key := channel_.AppId + channel_.Secret
	accessToken := TokenStoreGetToken(key)
	if accessToken == "" {
		item := &QQBotTokenStoreItem{
			AppID:     channel_.AppId,
			AppSecret: channel_.Secret,
		}
		TokenStoreAddItem(item)
		accessToken = TokenStoreGetToken(key)
	}
	if accessToken == "" {
		return errors.New("获取 QQ 机器人 access_token 失败，请检查 AppID / AppSecret")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, DefaultQQBotAPIBase+path, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "QQBot "+accessToken)
	req.Header.Set("X-Union-Appid", channel_.AppId)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var res qqBotMessageResponse
	_ = json.Unmarshal(body, &res)
	code := res.Code
	if code == 0 {
		code = res.ErrCode
	}
	if code != 0 || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := res.Message
		if msg == "" {
			msg = res.Msg
		}
		if msg == "" {
			msg = strings.TrimSpace(string(body))
		}
		if msg == "" {
			msg = resp.Status
		}
		return errors.New(msg)
	}
	return nil
}
