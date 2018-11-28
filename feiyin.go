package feiyin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/redis.v5"
)

type Client struct {
	MemberCode  string
	Appid       string
	Secret      string
	ApiBase     string
	AccessToken string
	Cache       *redis.Client
}

type ResponseError struct {
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
}

type AccessToken struct {
	ResponseError
	AccessToken string `json:"access_token"`
	Appid       string `json:"appid"`
	ExpiresIn   int    `json:"expires_in"`
}

type Device struct {
	DeviceNo        string `json:"device_no"`         //设备编码
	Since           string `json:"since"`             //打印机激活时间
	Status          string `json:"status"`            //打印机的连接状态，包括：online 在线 offline 离线 overheat 打印头过热 error 打印机系统故障
	PaperStatus     string `json:"paper_status"`      //打印纸张的状态，包括：nomal 正常 lack 缺纸
	LastConnectedAt string `json:"last_connected_at"` //最近连接时间
}

type PrintResponse struct {
	ResponseError
	MsgNo string `json:"msg_no"`
}

type MsgStatus struct {
	ResponseError
	MsgNo     string `json:"msg_no"`
	Status    string `json:"status"`
	MsgTime   string `json:"msg_time"`
	PrintTime string `json:"print_time"`
}

type ClearResponse struct {
	ResponseError
	ClearCnt int `json:"clear_cnt"`
}

type Template struct {
	Name      string `json:"name"`       //模板名称
	Content   string `json:"content"`    //模板内容
	Catalog   string `json:"catalog"`    //模板归类
	Desc      string `json:"desc"`       //模板说明
	UpdatedAt string `json:"updated_at"` //模板最后更新时间
}

type TemplateAddResponse struct {
	ResponseError
	TemplateId string `json:"template_id"`
}

type MemberDevice struct {
	DeviceNo string `json:"device_no"`
	Model    string `json:"model"`
	Memo     string `json:"memo"`
}

type Member struct {
	ResponseError
	Uid       string         `json:"uid"`
	Name      string         `json:"name"`
	CreatedAt string         `json:"created_at,omitempty"`
	Devices   []MemberDevice `json:"devices,omitempty"`
}

func NewClient(membercode string, appid string, secret string, redisAddr string) (*Client, error) {

	client := new(Client)
	client.MemberCode = membercode
	client.Appid = appid
	client.Secret = secret

	client.ApiBase = "https://api.open.feyin.net"

	if redisAddr != "" {

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisAddr,
			DB:   3,
		})

		_, err := redisClient.Ping().Result()

		if err != nil {
			return nil, err
		}

		client.Cache = redisClient
	}

	return client, nil
}

func (this *Client) Members() ([]Member, error) {

	body, err := this.httpGet(fmt.Sprintf("/app/%s/members", this.Appid), nil)

	if err != nil {
		return nil, err
	}

	var list []Member

	err = json.Unmarshal(body, &list)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (this *Client) Member(uid string) (*Member, error) {

	body, err := this.httpGet(fmt.Sprintf("/member/%s", uid), nil)

	if err != nil {
		return nil, err
	}

	var resp Member

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return nil, err
	}

	if resp.ErrCode != 0 {
		return nil, errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return &resp, nil
}

//本接口为重要接口，出于安全性考虑，这个接口只能对开发者自己名下的设备进行操作。
func (this *Client) DeviceBind(deviceNo string) error {

	body, err := this.httpPost(fmt.Sprintf("/device/%s/bind", deviceNo), nil)

	if err != nil {
		return err
	}

	var resp ResponseError

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return err
	}

	if resp.ErrCode != 0 {
		return errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return nil
}

//本接口为重要接口，出于安全性考虑，这个接口只能对开发者自己名下的设备进行操作。
func (this *Client) DeviceUnbind(deviceNo string) error {

	body, err := this.httpPost(fmt.Sprintf("/device/%s/unbind", deviceNo), nil)

	if err != nil {
		return err
	}

	var resp ResponseError

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return err
	}

	if resp.ErrCode != 0 {
		return errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return nil
}

func (this *Client) TemplateEdit(templateId string, name string, content string, catalog string, desc string) error {

	param := map[string]interface{}{
		"name":    name,
		"content": content,
		"catalog": catalog,
		"desc":    desc,
	}

	body, err := this.httpPost(fmt.Sprintf("/template/%s", templateId), param)

	if err != nil {
		return err
	}

	var resp ResponseError

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return err
	}

	if resp.ErrCode != 0 {
		return errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return nil
}

func (this *Client) TemplateAdd(name string, content string, catalog string, desc string) (string, error) {

	param := map[string]interface{}{
		"name":    name,
		"content": content,
		"catalog": catalog,
		"desc":    desc,
	}

	body, err := this.httpPost("/template", param)

	if err != nil {
		return "", err
	}

	var resp TemplateAddResponse

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return "", err
	}

	if resp.ErrCode != 0 {
		return "", errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return resp.TemplateId, nil
}

func (this *Client) Template(templateId string) (*Template, error) {

	body, err := this.httpGet(fmt.Sprintf("/template/detail/%s", templateId), nil)

	if err != nil {
		return nil, err
	}

	var data Template

	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (this *Client) Templates() ([]Template, error) {

	body, err := this.httpGet("/templates", nil)

	if err != nil {
		return nil, err
	}

	var list []Template

	err = json.Unmarshal(body, &list)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (this *Client) DeviceClear(deviceNo string) (int, error) {

	apiPath := fmt.Sprintf("/device/%s/msg/clear", deviceNo)

	body, err := this.httpPost(apiPath, nil)

	if err != nil {
		return 0, err
	}

	var resp ClearResponse

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return 0, err
	}

	if resp.ErrCode != 0 {
		return 0, errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return resp.ClearCnt, nil
}

func (this *Client) MsgCancel(msgNo string) error {

	apiPath := fmt.Sprintf("/msg/%s/cancel", msgNo)

	body, err := this.httpPost(apiPath, nil)

	if err != nil {
		return err
	}

	var resp ResponseError

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return err
	}

	if resp.ErrCode != 0 {
		return errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return nil
}

func (this *Client) MsgStatus(msgNo string) (*MsgStatus, error) {

	apiPath := fmt.Sprintf("/msg/%s/status", msgNo)

	body, err := this.httpGet(apiPath, nil)

	if err != nil {
		return nil, err
	}

	var resp MsgStatus

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return nil, err
	}

	if resp.ErrCode != 0 {
		return nil, errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return &resp, nil
}

func (this *Client) Print(deviceNo string, content string, templateId string, templateData map[string]interface{}) (string, error) {

	param := map[string]interface{}{}

	if content != "" { //普通

		param = map[string]interface{}{
			"device_no":   deviceNo,
			"msg_content": content,
			"appid":       this.Appid,
		}

	} else { //模版

		param = map[string]interface{}{
			"device_no":     deviceNo,
			"template_id":   templateId,
			"template_data": templateData,
			"appid":         this.Appid,
		}
	}

	body, err := this.httpPost("/msg", param)

	if err != nil {
		return "", err
	}

	var resp PrintResponse

	err = json.Unmarshal(body, &resp)

	if err != nil {
		return "", err
	}

	if resp.ErrCode != 0 {
		return "", errors.New(fmt.Sprintf("%d:%s", resp.ErrCode, resp.ErrMsg))
	}

	return resp.MsgNo, nil
}

func (this *Client) Device(deviceNo string) (*Device, error) {

	body, err := this.httpGet(fmt.Sprintf("/device/%s/status", deviceNo), nil)

	if err != nil {
		return nil, err
	}

	device := new(Device)

	err = json.Unmarshal(body, device)

	if err != nil {
		return nil, err
	}

	return device, nil
}

func (this *Client) Devices() ([]Device, error) {

	body, err := this.httpGet("/devices", nil)

	if err != nil {
		return nil, err
	}

	var list []Device

	err = json.Unmarshal(body, &list)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (this *Client) refreshAccessToken() error {

	cacheKey := fmt.Sprintf("FeiyinSDKAccessToken_%s%s", this.MemberCode, this.Appid)

	var cache string
	var err error

	if this.Cache != nil {
		cache, err = this.Cache.Get(cacheKey).Result()
	}

	if err != nil || cache == "" {

		param := map[string]interface{}{
			"code":   this.MemberCode,
			"secret": this.Secret,
			"appid":  this.Appid,
		}

		body, err := this.httpGet("/token", param)

		if err != nil {
			return err
		}

		token := new(AccessToken)

		err = json.Unmarshal(body, token)

		if err != nil {
			return err
		}

		this.AccessToken = token.AccessToken

		if this.Cache != nil {
			this.Cache.Set(cacheKey, this.AccessToken, time.Duration(token.ExpiresIn)*time.Second)
		}
	} else {
		this.AccessToken = cache
	}

	return nil
}

func (this *Client) httpGet(apiPath string, param map[string]interface{}) ([]byte, error) {

	if param == nil {
		param = map[string]interface{}{}
	}

	if apiPath != "/token" {
		this.refreshAccessToken()
		param["access_token"] = this.AccessToken
	}

	return httpGet(fmt.Sprintf("%s%s", this.ApiBase, apiPath), param)
}

func (this *Client) httpPost(apiPath string, param map[string]interface{}) ([]byte, error) {

	connect := "?"

	if strings.Contains(apiPath, "?") {
		connect = "&"
	}

	this.refreshAccessToken()
	apiPath = fmt.Sprintf("%s%saccess_token=%s", apiPath, connect, this.AccessToken)

	return httpPost(fmt.Sprintf("%s%s", this.ApiBase, apiPath), param)
}
