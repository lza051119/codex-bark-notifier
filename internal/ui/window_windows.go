//go:build windows

package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"github.com/lza051119/codex-bark-notifier/internal/bark"
	"github.com/lza051119/codex-bark-notifier/internal/codex"
	"github.com/lza051119/codex-bark-notifier/internal/config"
	"golang.org/x/sys/windows"
)

const (
	windowClassName = "CodexBarkNotifierNativeWindow"
	controlURL      = 1001
	controlKey      = 1002
	controlEnabled  = 1003
	buttonSave      = 1101
	buttonTest      = 1102
	buttonInstall   = 1103
	buttonUninstall = 1104
	buttonREADME    = 1105
	controlStatus   = 1201
)

var (
	activeNativeWindow *nativeWindow
	nativeWindowProc   = windows.NewCallback(nativeWndProc)
)

type nativeWindow struct {
	hwnd    win.HWND
	store   *config.Store
	manager *codex.Manager
	editURL win.HWND
	editKey win.HWND
	check   win.HWND
	status  win.HWND
}

// GUIBackend is covered by a regression test so the application cannot
// silently reintroduce the Walk tooltip subsystem that fails during startup.
func GUIBackend() string {
	return "win32"
}

func Show(store *config.Store, manager *codex.Manager) error {
	settings, err := store.Load()
	if err != nil {
		return err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ui := &nativeWindow{store: store, manager: manager}
	activeNativeWindow = ui
	defer func() { activeNativeWindow = nil }()

	if err := ui.create(settings); err != nil {
		return err
	}
	defer func() {
		if ui.hwnd != 0 {
			win.DestroyWindow(ui.hwnd)
			ui.hwnd = 0
		}
	}()

	var msg win.MSG
	for {
		result := win.GetMessage(&msg, 0, 0, 0)
		if result == win.BOOL(-1) {
			return errors.New("Windows message loop failed")
		}
		if result == 0 {
			return nil
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}

func (ui *nativeWindow) create(settings config.Settings) error {
	instance := win.GetModuleHandle(nil)
	className := syscall.StringToUTF16Ptr(windowClassName)
	class := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		LpfnWndProc:   nativeWindowProc,
		HInstance:     instance,
		HbrBackground: win.GetSysColorBrush(win.COLOR_WINDOW),
		LpszClassName: className,
	}
	atom := win.RegisterClassEx(&class)
	if atom == 0 && !errors.Is(windows.GetLastError(), windows.ERROR_CLASS_ALREADY_EXISTS) {
		return fmt.Errorf("Windows window class registration failed: %w", windows.GetLastError())
	}

	title := syscall.StringToUTF16Ptr("Codex Bark Notifier")
	ui.hwnd = win.CreateWindowEx(
		0,
		className,
		title,
		win.WS_OVERLAPPEDWINDOW|win.WS_VISIBLE,
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		650,
		430,
		0,
		0,
		instance,
		nil,
	)
	if ui.hwnd == 0 {
		if atom != 0 {
			win.UnregisterClass(className)
		}
		return fmt.Errorf("Windows window creation failed: %w", windows.GetLastError())
	}

	if err := ui.createControls(settings); err != nil {
		win.DestroyWindow(ui.hwnd)
		ui.hwnd = 0
		if atom != 0 {
			win.UnregisterClass(className)
		}
		return err
	}
	win.ShowWindow(ui.hwnd, win.SW_SHOW)
	win.UpdateWindow(ui.hwnd)
	return nil
}

func (ui *nativeWindow) createControls(settings config.Settings) error {
	font := uintptr(win.GetStockObject(win.DEFAULT_GUI_FONT))
	var controlErr error
	require := func(name string, hwnd win.HWND) win.HWND {
		if hwnd == 0 && controlErr == nil {
			controlErr = fmt.Errorf("Windows %s creation failed: %w", name, windows.GetLastError())
		}
		return hwnd
	}
	label := func(text string, x, y, width, height int32) win.HWND {
		return require("label", ui.createControl("STATIC", text, win.WS_CHILD|win.WS_VISIBLE, x, y, width, height, 0, font))
	}
	edit := func(text string, style uint32, x, y, width, height int32, id int) win.HWND {
		return require("input", ui.createControl("EDIT", text, win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|style, x, y, width, height, id, font))
	}
	button := func(text string, x, y, width, height int32, id int) win.HWND {
		return require("button", ui.createControl("BUTTON", text, win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|win.BS_PUSHBUTTON, x, y, width, height, id, font))
	}

	label("Bark 服务地址", 24, 24, 200, 22)
	ui.editURL = edit(settings.ServerURL, win.ES_LEFT|win.ES_AUTOHSCROLL, 24, 48, 590, 28, controlURL)
	label("Device Key（仅保存在本机）", 24, 88, 300, 22)
	ui.editKey = edit(settings.DeviceKey, win.ES_LEFT|win.ES_AUTOHSCROLL|win.ES_PASSWORD, 24, 112, 590, 28, controlKey)
	ui.check = ui.createControl("BUTTON", "开启 Codex 手机提醒", win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|win.BS_AUTOCHECKBOX, 24, 154, 250, 28, controlEnabled, font)
	ui.check = require("enabled checkbox", ui.check)
	if settings.Enabled {
		win.SendMessage(ui.check, win.BM_SETCHECK, win.BST_CHECKED, 0)
	}
	label("任务结束后，Codex 会调用此程序发送 Bark 通知。", 24, 194, 590, 22)

	button("保存配置", 24, 238, 110, 32, buttonSave)
	button("发送测试", 144, 238, 110, 32, buttonTest)
	button("安装 Codex 接入", 264, 238, 140, 32, buttonInstall)
	button("卸载 Codex 接入", 414, 238, 140, 32, buttonUninstall)
	button("打开 README", 24, 280, 120, 32, buttonREADME)
	ui.status = label(StatusText(settings.Enabled, settings.DeviceKey), 24, 332, 590, 28)
	if controlErr != nil {
		return controlErr
	}
	return nil
}

func (ui *nativeWindow) createControl(className, text string, style uint32, x, y, width, height int32, id int, font uintptr) win.HWND {
	class := syscall.StringToUTF16Ptr(className)
	title := syscall.StringToUTF16Ptr(text)
	hwnd := win.CreateWindowEx(
		0,
		class,
		title,
		style,
		x,
		y,
		width,
		height,
		ui.hwnd,
		win.HMENU(uintptr(id)),
		win.GetModuleHandle(nil),
		nil,
	)
	if hwnd != 0 && font != 0 {
		win.SendMessage(hwnd, win.WM_SETFONT, font, 1)
	}
	return hwnd
}

func nativeWndProc(hwnd win.HWND, message uint32, wParam, lParam uintptr) uintptr {
	ui := activeNativeWindow
	switch message {
	case win.WM_COMMAND:
		if ui != nil && uint16(wParam>>16) == win.BN_CLICKED {
			ui.command(int(uint16(wParam)))
		}
		return 0
	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, message, wParam, lParam)
}

func (ui *nativeWindow) command(id int) {
	switch id {
	case buttonSave:
		ui.save(true)
	case buttonTest:
		if ui.save(false) {
			settings := ui.settings()
			client := bark.NewClient(settings.ServerURL, settings.DeviceKey)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := client.Send(ctx, "Codex Bark 测试", "手机提醒配置正常。\\nCodex Bark Notifier")
			cancel()
			if err != nil {
				ui.message("测试通知发送失败。", win.MB_ICONERROR)
				return
			}
			ui.setStatus("测试通知已发送。")
		}
	case buttonInstall:
		path, err := installExecutable()
		if err == nil {
			_, err = ui.manager.Install(path)
		}
		if err != nil {
			ui.message("Codex 接入安装失败。", win.MB_ICONERROR)
			return
		}
		ui.message("Codex 接入已安装。", win.MB_ICONINFORMATION)
	case buttonUninstall:
		if err := ui.manager.Uninstall(); err != nil {
			ui.message("卸载失败：当前 Codex 配置可能已被其他程序修改。", win.MB_ICONWARNING)
			return
		}
		ui.message("Codex 接入已卸载并恢复备份。", win.MB_ICONINFORMATION)
	case buttonREADME:
		if err := openREADME(); err != nil {
			ui.message("README 打开失败。", win.MB_ICONERROR)
		}
	}
}

func (ui *nativeWindow) settings() config.Settings {
	return config.Settings{
		ServerURL: readControlText(ui.editURL),
		DeviceKey: readControlText(ui.editKey),
		Enabled:   win.SendMessage(ui.check, win.BM_GETCHECK, 0, 0) == uintptr(win.BST_CHECKED),
	}
}

func (ui *nativeWindow) save(showMessage bool) bool {
	settings := ui.settings()
	if err := ValidateSettings(settings.ServerURL, settings.DeviceKey, settings.Enabled); err != nil {
		ui.message(err.Error(), win.MB_ICONWARNING)
		return false
	}
	if err := ui.store.Save(settings); err != nil {
		ui.message("配置保存失败。", win.MB_ICONERROR)
		return false
	}
	ui.setStatus(StatusText(settings.Enabled, settings.DeviceKey))
	if showMessage {
		ui.message("配置已保存。", win.MB_ICONINFORMATION)
	}
	return true
}

func (ui *nativeWindow) setStatus(text string) {
	setControlText(ui.status, text)
}

func (ui *nativeWindow) message(text string, icon uint32) {
	message := syscall.StringToUTF16Ptr(text)
	title := syscall.StringToUTF16Ptr("Codex Bark Notifier")
	win.MessageBox(ui.hwnd, message, title, win.MB_OK|icon)
}

func readControlText(hwnd win.HWND) string {
	length := int(win.SendMessage(hwnd, win.WM_GETTEXTLENGTH, 0, 0))
	buffer := make([]uint16, length+1)
	if len(buffer) > 0 {
		win.SendMessage(hwnd, win.WM_GETTEXT, uintptr(len(buffer)), uintptr(unsafe.Pointer(&buffer[0])))
	}
	return strings.TrimSpace(syscall.UTF16ToString(buffer))
}

func setControlText(hwnd win.HWND, text string) {
	if hwnd != 0 {
		win.SendMessage(hwnd, win.WM_SETTEXT, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
	}
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
