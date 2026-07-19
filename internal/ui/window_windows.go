//go:build windows

package ui

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/lza051119/codex-bark-notifier/internal/bark"
	"github.com/lza051119/codex-bark-notifier/internal/codex"
	"github.com/lza051119/codex-bark-notifier/internal/config"
)

func Show(store *config.Store, manager *codex.Manager) error {
	settings, err := store.Load()
	if err != nil {
		return err
	}

	var urlEdit, keyEdit *walk.LineEdit
	var enabledBox *walk.CheckBox
	var statusLabel *walk.Label
	var mainWindow *walk.MainWindow

	showStatus := func(message string, icon walk.MsgBoxStyle) {
		if statusLabel != nil {
			statusLabel.SetText(message)
		}
		if icon != 0 && mainWindow != nil {
			walk.MsgBox(mainWindow, "Codex Bark Notifier", message, walk.MsgBoxOK|icon)
		}
	}

	readSettings := func() config.Settings {
		return config.Settings{
			ServerURL: strings.TrimSpace(urlEdit.Text()),
			DeviceKey: strings.TrimSpace(keyEdit.Text()),
			Enabled:   enabledBox.Checked(),
		}
	}

	saveSettings := func(showMessage bool) error {
		current := readSettings()
		if err := ValidateSettings(current.ServerURL, current.DeviceKey, current.Enabled); err != nil {
			showStatus(err.Error(), walk.MsgBoxIconWarning)
			return err
		}
		if err := store.Save(current); err != nil {
			showStatus("配置保存失败。", walk.MsgBoxIconError)
			return err
		}
		showStatus(StatusText(current.Enabled, current.DeviceKey), 0)
		if showMessage {
			walk.MsgBox(mainWindow, "Codex Bark Notifier", "配置已保存。", walk.MsgBoxOK|walk.MsgBoxIconInformation)
		}
		return nil
	}

	if err := (MainWindow{
		AssignTo: &mainWindow,
		Title:    "Codex Bark Notifier",
		MinSize:  Size{Width: 560, Height: 360},
		Size:     Size{Width: 620, Height: 430},
		Layout:   VBox{Margins: Margins{Left: 16, Top: 16, Right: 16, Bottom: 16}, Spacing: 10},
		Children: []Widget{
			Label{Text: "Bark 服务地址"},
			LineEdit{AssignTo: &urlEdit, Text: settings.ServerURL},
			Label{Text: "Device Key（仅保存在本机）"},
			LineEdit{AssignTo: &keyEdit, Text: settings.DeviceKey, PasswordMode: true},
			CheckBox{AssignTo: &enabledBox, Text: "开启 Codex 手机提醒", Checked: settings.Enabled},
			Label{Text: "任务结束后，Codex 会调用此程序发送 Bark 通知。"},
			Composite{
				Layout: HBox{Spacing: 8},
				Children: []Widget{
					PushButton{Text: "保存配置", OnClicked: func() { _ = saveSettings(true) }},
					PushButton{Text: "发送测试", OnClicked: func() {
						if err := saveSettings(false); err != nil {
							return
						}
						current := readSettings()
						client := bark.NewClient(current.ServerURL, current.DeviceKey)
						if err := client.Send(context.Background(), "Codex Bark 测试", "手机提醒配置正常。\\nCodex Bark Notifier"); err != nil {
							showStatus("测试通知发送失败。", walk.MsgBoxIconError)
							return
						}
						showStatus("测试通知已发送。", walk.MsgBoxIconInformation)
					}},
					PushButton{Text: "安装 Codex 接入", OnClicked: func() {
						path, err := installExecutable()
						if err == nil {
							_, err = manager.Install(path)
						}
						if err != nil {
							showStatus("Codex 接入安装失败。", walk.MsgBoxIconError)
							return
						}
						showStatus("Codex 接入已安装。", walk.MsgBoxIconInformation)
					}},
					PushButton{Text: "卸载 Codex 接入", OnClicked: func() {
						if err := manager.Uninstall(); err != nil {
							showStatus("卸载失败：当前 Codex 配置可能已被其他程序修改。", walk.MsgBoxIconWarning)
							return
						}
						showStatus("Codex 接入已卸载并恢复备份。", walk.MsgBoxIconInformation)
					}},
					PushButton{Text: "打开 README", OnClicked: func() {
						if err := openREADME(); err != nil {
							showStatus("README 打开失败。", walk.MsgBoxIconError)
						}
					}},
				},
			},
			Label{AssignTo: &statusLabel, Text: StatusText(settings.Enabled, settings.DeviceKey)},
		},
	}).Create(); err != nil {
		return err
	}
	mainWindow.Run()
	return nil
}

func installExecutable() (string, error) {
	current, err := os.Executable()
	if err != nil {
		return "", err
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if strings.TrimSpace(localAppData) == "" {
		return "", errors.New("LOCALAPPDATA is not set")
	}
	dir := filepath.Join(localAppData, "CodexBarkNotifier")
	destination := filepath.Join(dir, "codex-bark-notifier.exe")
	if filepath.Clean(current) == filepath.Clean(destination) {
		return destination, nil
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	in, err := os.Open(current)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(destination)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	return destination, nil
}

func openREADME() error {
	return exec.Command("explorer.exe", "https://github.com/lza051119/codex-bark-notifier#readme").Start()
}
