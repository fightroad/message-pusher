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
	"time"
)

const defaultWeChatCorpAPIHost = "https://qyapi.weixin.qq.com"

type wechatCorpAccountResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type corpAppOptions struct {
	ClientType string `json:"client_type"`
	Proxy      string `json:"proxy"`
}

type WeChatCorpAccountTokenStoreItem struct {
	CorpId      string
	AgentSecret string
	AgentId     string
	ApiHost     string
	Proxy       string
	AccessToken string
}

func (i *WeChatCorpAccountTokenStoreItem) Key() string {
	return i.CorpId + i.AgentId + i.AgentSecret + "|" + i.ApiHost + "|" + i.Proxy
}

func (i *WeChatCorpAccountTokenStoreItem) IsShared() bool {
	appId := fmt.Sprintf("%s|%s", i.CorpId, i.AgentId)
	var count int64 = 0
	model.DB.Model(&model.Channel{}).Where("secret = ? and app_id = ? and type = ?",
		i.AgentSecret, appId, model.TypeWeChatCorpAccount).Count(&count)
	return count > 1
}

func (i *WeChatCorpAccountTokenStoreItem) IsFilled() bool {
	return i.CorpId != "" && i.AgentSecret != "" && i.AgentId != ""
}

func (i *WeChatCorpAccountTokenStoreItem) Token() string {
	return i.AccessToken
}

func normalizeWeChatCorpAPIHost(raw string) string {
	host := strings.TrimRight(strings.TrimSpace(raw), "/")
	if host == "" {
		return defaultWeChatCorpAPIHost
	}
	return host
}

func parseCorpAppOptions(raw string) corpAppOptions {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return corpAppOptions{ClientType: "app"}
	}
	var opts corpAppOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		return corpAppOptions{ClientType: "app"}
	}
	if opts.ClientType == "" {
		opts.ClientType = "app"
	}
	return opts
}

func (i *WeChatCorpAccountTokenStoreItem) Refresh() {
	// https://work.weixin.qq.com/api/doc/90000/90135/91039
	client, err := NewHTTPClient(i.Proxy, 15*time.Second)
	if err != nil {
		common.SysError("failed to create wechat corp http client: " + err.Error())
		return
	}
	apiHost := normalizeWeChatCorpAPIHost(i.ApiHost)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		apiHost, i.CorpId, i.AgentSecret), nil)
	if err != nil {
		common.SysError(err.Error())
		return
	}
	responseData, err := client.Do(req)
	if err != nil {
		common.SysError("failed to refresh access token: " + err.Error())
		return
	}
	defer responseData.Body.Close()
	var res wechatCorpAccountResponse
	err = json.NewDecoder(responseData.Body).Decode(&res)
	if err != nil {
		common.SysError("failed to decode wechatCorpAccountResponse: " + err.Error())
		return
	}
	if res.ErrorCode != 0 {
		common.SysError(res.ErrorMessage)
		return
	}
	i.AccessToken = res.AccessToken
	common.SysLog("access token refreshed")
}

type wechatCorpMessageRequest struct {
	MessageType string `json:"msgtype"`
	ToUser      string `json:"touser"`
	AgentId     string `json:"agentid"`
	TextCard    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
	} `json:"textcard"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

type wechatCorpMessageResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
}

func parseWechatCorpAccountAppId(appId string) (string, string, error) {
	parts := strings.Split(appId, "|")
	if len(parts) != 2 {
		return "", "", errors.New("无效的微信企业号配置")
	}
	return parts[0], parts[1], nil
}

func SendWeChatCorpMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	// https://developer.work.weixin.qq.com/document/path/90236
	corpId, agentId, err := parseWechatCorpAccountAppId(channel_.AppId)
	if err != nil {
		return err
	}
	opts := parseCorpAppOptions(channel_.Other)
	apiHost := normalizeWeChatCorpAPIHost(channel_.URL)
	userId := channel_.AccountId
	clientType := opts.ClientType
	agentSecret := channel_.Secret
	messageRequest := wechatCorpMessageRequest{
		ToUser:  userId,
		AgentId: agentId,
	}
	if message.To != "" {
		messageRequest.ToUser = message.To
	}
	if message.Content == "" {
		if message.Title == "" {
			messageRequest.MessageType = "text"
			messageRequest.Text.Content = message.Description
		} else {
			messageRequest.MessageType = "textcard"
			messageRequest.TextCard.Title = message.Title
			messageRequest.TextCard.Description = message.Description
			messageRequest.TextCard.URL = message.URL
			if messageRequest.TextCard.URL == "" {
				messageRequest.TextCard.URL = common.ServerAddress
			}
		}
	} else {
		if clientType == "plugin" {
			messageRequest.MessageType = "textcard"
			messageRequest.TextCard.Title = message.Title
			messageRequest.TextCard.Description = message.Description
			messageRequest.TextCard.URL = message.URL
			if messageRequest.TextCard.URL == "" {
				messageRequest.TextCard.URL = common.ServerAddress
			}
		} else {
			messageRequest.MessageType = "markdown"
			messageRequest.Markdown.Content = message.Content
		}
	}
	jsonData, err := json.Marshal(messageRequest)
	if err != nil {
		return err
	}
	item := &WeChatCorpAccountTokenStoreItem{
		CorpId:      corpId,
		AgentId:     agentId,
		AgentSecret: agentSecret,
		ApiHost:     apiHost,
		Proxy:       strings.TrimSpace(opts.Proxy),
	}
	accessToken := TokenStoreGetToken(item.Key())
	client, err := NewHTTPClient(item.Proxy, 15*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.Post(fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", apiHost, accessToken), "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res wechatCorpMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return err
	}
	if res.ErrorCode != 0 {
		return errors.New(res.ErrorMessage)
	}
	return nil
}
