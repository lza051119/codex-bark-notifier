## Codex Bark Notifier v0.1.2

修复 Windows 双击 exe 后立即退出、没有显示配置窗口的问题。GUI 已改为原生 Win32 控件，并移除 Walk tooltip 初始化依赖。

首个 Windows x64 公开版本。

- 原生 Go 配置窗口
- Bark 服务地址、Device Key 和开关配置
- Codex 任务结束通知接入
- Device Key 使用 Windows DPAPI 保护
- Bark 失败不会阻塞 Codex
- 附带 SHA-256 校验文件

使用前请阅读 [README.md](https://github.com/lza051119/codex-bark-notifier/blob/main/README.md)。
