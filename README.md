# Codex Bark Notifier

一个 Windows 原生小工具：当 Codex 任务结束时，自动向手机发送 Bark 提醒。

## 功能

- Go 原生 Windows x64 exe，不需要安装 Node.js。
- App 内配置 Bark 服务地址、Device Key 和手机提醒开关。
- App 内发送测试通知。
- App 内安装/卸载 Codex 任务结束接入。
- 保留 Codex 原有的 Windows 本地通知。
- Bark 网络失败不会阻塞 Codex 任务结束。
- Device Key 使用当前 Windows 用户的 DPAPI 加密保存。
- 不把 Device Key、Codex 配置或本机日志上传到 GitHub。

## 手机端需要安装什么

### iPhone / iPad

请安装官方 Bark App：

- [Bark on App Store](https://apps.apple.com/us/app/bark-custom-notifications/id1403753865)
- [Bark 官方开源仓库](https://github.com/Finb/Bark)
- [Bark 官方使用文档](https://bark.day.app/#/en-us/)

首次打开 Bark 后，App 会显示一个推送测试地址。这个地址中的服务地址和设备密钥就是本程序需要的配置：

```text
https://api.day.app/你的设备密钥/...
```

在 iPhone 系统设置中允许 Bark 通知。如果希望使用重要提醒、响铃或静音模式下提醒，还需要按 iOS 提示允许相应权限。

### Android

上游 Bark 官方客户端是 iOS App。Android 用户需要使用能够兼容 Bark HTTP API 的第三方客户端或自建兼容服务，并在本程序中填写对应的服务地址和设备密钥。请先用 App 的测试功能确认推送链路可用。

## Windows 安装

1. 从 GitHub Release 下载 `codex-bark-notifier-windows-x64.exe`。
2. 双击运行程序。
3. 填写 Bark 服务地址，默认是：

   ```text
   https://api.day.app
   ```

4. 填写 Bark Device Key。
5. 点击“保存配置”。
6. 点击“发送测试”，确认手机能收到通知。
7. 点击“安装 Codex 接入”。
8. 打开或重新启动 Codex。之后每次 Codex 任务结束都会调用这个 exe。

程序会把稳定副本安装到：

```text
%LOCALAPPDATA%\CodexBarkNotifier\codex-bark-notifier.exe
```

## 开启和关闭提醒

在 App 中勾选或取消“开启 Codex 手机提醒”，然后点击“保存配置”。

- 关闭后，Codex 仍然正常运行，原有 Windows 本地通知也不受影响。
- 关闭只跳过 Bark 手机推送。
- 重新打开 App 即可再次修改配置。

## 自建 Bark Server

如果你使用自建服务器，在“Bark 服务地址”中填写服务器根地址，例如：

```text
https://bark.example.com
```

不要填写完整的测试推送 URL，也不要重复填写设备密钥。程序会自动拼接：

```text
服务器地址 / 设备密钥 / 标题 / 内容
```

## 卸载 Codex 接入

打开 App，点击“卸载 Codex 接入”。程序只会在检测到 Codex 配置仍是本程序安装的版本时恢复备份；如果配置已经被其他程序修改，程序会停止恢复，避免覆盖用户的新配置。

## 隐私说明

通知内容包含当前项目文件夹名称，以及最多 180 个字符的 Codex 最后回复摘要。请不要在 Codex 最终回复中放置不适合发送到手机通知的敏感信息。

Device Key 保存在 Windows DPAPI 保护的本地配置中。日志只记录时间和成功/失败状态，不记录 Device Key 或完整事件内容。

## 从源码构建

需要 Go 1.26 或更高版本：

```powershell
go test ./...
.\scripts\build-release.ps1 -Version 0.1.2
```

生成文件：

```text
dist\codex-bark-notifier-windows-x64.exe
dist\SHA256SUMS.txt
```

## 安全提醒

请不要把以下内容上传到公开仓库：

- Bark Device Key
- `%USERPROFILE%\.codex\config.toml`
- 本机通知日志
- 包含真实项目内容或个人信息的测试事件

## License

MIT License，详见 [LICENSE](LICENSE)。
