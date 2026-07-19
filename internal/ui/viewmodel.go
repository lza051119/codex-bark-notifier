package ui

import (
	"errors"
	"net/url"
	"strings"
)

func ValidateSettings(serverURL, deviceKey string, enabled bool) error {
	parsed, err := url.Parse(strings.TrimSpace(serverURL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("请输入有效的 Bark 服务地址")
	}
	if enabled && strings.TrimSpace(deviceKey) == "" {
		return errors.New("开启手机提醒时必须填写 Device Key")
	}
	return nil
}

func StatusText(enabled bool, _ string) string {
	if enabled {
		return "手机提醒：已开启"
	}
	return "手机提醒：已关闭"
}
