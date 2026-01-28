package serverchan

import (
	"fmt"

	svrchan "github.com/easychen/serverchan-sdk-golang"
)

type Client struct {
	sendKey string
}

func NewClient(sendKey string) *Client {
	return &Client{sendKey: sendKey}
}

func (c *Client) Send(title, content string) error {
	if c.sendKey == "" {
		return fmt.Errorf("send key is empty")
	}

	_, err := svrchan.ScSend(c.sendKey, title, content, nil)
	return err
}

func (c *Client) SendAppOperation(appTitle, operation string, success bool) error {
	var title, content string
	status := "成功"
	if !success {
		status = "失败"
	}

	switch operation {
	case "resume":
		title = fmt.Sprintf("应用恢复%s: %s", status, appTitle)
		content = fmt.Sprintf("应用 **%s** 定时恢复任务执行%s", appTitle, status)
	case "pause":
		title = fmt.Sprintf("应用休眠%s: %s", status, appTitle)
		content = fmt.Sprintf("应用 **%s** 定时休眠任务执行%s", appTitle, status)
	default:
		title = fmt.Sprintf("应用操作%s: %s", status, appTitle)
		content = fmt.Sprintf("应用 **%s** 操作执行%s", appTitle, status)
	}

	return c.Send(title, content)
}
