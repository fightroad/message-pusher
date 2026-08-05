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

type wechatTestAccountResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type WeChatTestAccountTokenStoreItem struct {
	AppID       string
	AppSecret   string
	AccessToken string
}

func (i *WeChatTestAccountTokenStoreItem) Key() string {
	return i.AppID + i.AppSecret
}

func (i *WeChatTestAccountTokenStoreItem) IsShared() bool {
	var count int64 = 0
	model.DB.Model(&model.Channel{}).Where("secret = ? and app_id = ? and type = ?",
		i.AppSecret, i.AppID, model.TypeWeChatTestAccount).Count(&count)
	return count > 1
}

func (i *WeChatTestAccountTokenStoreItem) IsFilled() bool {
	return i.AppID != "" && i.AppSecret != ""
}

func (i *WeChatTestAccountTokenStoreItem) Token() string {
	return i.AccessToken
}

func (i *WeChatTestAccountTokenStoreItem) Refresh() {
	// https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Get_access_token.html
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		i.AppID, i.AppSecret), nil)
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
	var res wechatTestAccountResponse
	err = json.NewDecoder(responseData.Body).Decode(&res)
	if err != nil {
		common.SysError("failed to decode wechatTestAccountResponse: " + err.Error())
		return
	}
	if res.ErrorCode != 0 {
		common.SysError(res.ErrorMessage)
		return
	}
	i.AccessToken = res.AccessToken
	common.SysLog("access token refreshed")
}

type wechatTestAccountRequestValue struct {
	Value string `json:"value"`
}

type wechatTestMessageRequest struct {
	ToUser     string `json:"touser"`
	TemplateId string `json:"template_id"`
	URL        string `json:"url"`
	Data       struct {
		Text        wechatTestAccountRequestValue `json:"text"` // alias for description, for compatibility
		Title       wechatTestAccountRequestValue `json:"title"`
		Description wechatTestAccountRequestValue `json:"description"`
		Content     wechatTestAccountRequestValue `json:"content"`
	} `json:"data"`
}

type wechatTestMessageResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
}

type wechatTestUserListResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
	Total        int    `json:"total"`
	Count        int    `json:"count"`
	Data         struct {
		OpenId []string `json:"openid"`
	} `json:"data"`
	NextOpenId string `json:"next_openid"`
}

func SendWeChatTestMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	// https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
	key := fmt.Sprintf("%s%s", channel_.AppId, channel_.Secret)
	accessToken := TokenStoreGetToken(key)
	target := channel_.AccountId
	if message.To != "" {
		target = message.To
	}
	openIds, err := resolveWeChatTestRecipients(accessToken, target)
	if err != nil {
		return err
	}
	if len(openIds) == 0 {
		return errors.New("没有可推送的用户 Open ID")
	}

	var errs []string
	for _, openId := range openIds {
		err := sendWeChatTestTemplateMessage(accessToken, openId, channel_.Other, message)
		if err != nil {
			errs = append(errs, openId+": "+err.Error())
		}
	}
	if len(errs) == len(openIds) {
		return errors.New("全部推送失败：" + strings.Join(errs, "; "))
	}
	if len(errs) > 0 {
		return errors.New("部分推送失败：" + strings.Join(errs, "; "))
	}
	return nil
}

func resolveWeChatTestRecipients(accessToken, target string) ([]string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, errors.New("未配置用户 Open ID")
	}
	if strings.EqualFold(target, "@all") {
		return fetchWeChatTestFollowerOpenIds(accessToken)
	}
	parts := strings.Split(target, "|")
	openIds := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			openIds = append(openIds, part)
		}
	}
	return openIds, nil
}

func fetchWeChatTestFollowerOpenIds(accessToken string) ([]string, error) {
	// https://developers.weixin.qq.com/doc/offiaccount/User_Management/Getting_a_User_List.html
	client := &http.Client{Timeout: 15 * time.Second}
	var openIds []string
	nextOpenId := ""
	for {
		url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/get?access_token=%s", accessToken)
		if nextOpenId != "" {
			url += "&next_openid=" + nextOpenId
		}
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}
		var res wechatTestUserListResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if res.ErrorCode != 0 {
			return nil, errors.New(res.ErrorMessage)
		}
		openIds = append(openIds, res.Data.OpenId...)
		if res.Count == 0 || res.NextOpenId == "" || res.NextOpenId == nextOpenId {
			break
		}
		nextOpenId = res.NextOpenId
	}
	return openIds, nil
}

func sendWeChatTestTemplateMessage(accessToken, openId, templateId string, message *model.Message) error {
	values := wechatTestMessageRequest{
		ToUser:     openId,
		TemplateId: templateId,
		URL:        message.URL,
	}
	values.Data.Text.Value = common.StripHTMLTags(message.Description)
	values.Data.Title.Value = common.StripHTMLTags(message.Title)
	values.Data.Description.Value = common.StripHTMLTags(message.Description)
	values.Data.Content.Value = common.StripHTMLTags(message.Content)
	jsonData, err := json.Marshal(values)
	if err != nil {
		return err
	}
	resp, err := http.Post(
		fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res wechatTestMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return err
	}
	if res.ErrorCode != 0 {
		return errors.New(res.ErrorMessage)
	}
	return nil
}
