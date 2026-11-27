# LLM Key Manager

LLM Key Manager 是一个运行在 Windows 和 macOS 上的桌面应用，用来在自己的电脑上集中保管各家大模型服务商的 API Key。它适合手上有多个 Key、需要在不同工具之间来回粘贴的个人用户。数据只保存在本机，不做云同步。

## 能做什么

保存一个 Key 时，填写提供商、名称和 Key 内容。提供商可以从预设里选，也可以直接输入别的名字。已保存的 Key 按提供商分组，最近修改的排在前面；提供商、名称和内容完全相同的 Key 不能重复保存。

列表里只显示脱敏后的内容。需要使用时点“复制”，完整的 Key 会直接进入系统剪贴板。

编辑时可以只改提供商或名称，Key 内容留空表示沿用原来的，填了新内容则替换。不需要的记录可以删除。

## 过期检测

点“过期检测”后，应用会用每个已保存的 Key 向对应服务商发送一次很小的生成请求，以此判断这个 Key 现在还能不能用。**这是真实的 API 请求，可能产生费用。**

只有下面三家预设提供商支持检测，每家使用固定的模型。其他提供商的 Key 会显示“不支持检测”，不会发出任何请求。

| 提供商          | 检测用的模型      |
| --------------- | ----------------- |
| DeepSeek        | `deepseek-flash`  |
| MiniMax（国内） | `MiniMax-M3`      |
| OfoxAI          | `qwen/qwen-flash` |

每个 Key 最多等待 30 秒，结果是“可用”“认证失败”“余额不足”“权限受限”“限流”“检测失败”之一。“可用”只说明检测那一刻这个模型正常返回了内容；“余额不足”“限流”“检测失败”也不等于 Key 已经失效，可能只是账户或网络的暂时状况。

检测进行中不能新增、编辑或删除记录。检测结果不会保存，修改记录或重启应用后需要重新检测。

## 数据存在哪里，有多安全

所有记录保存在当前用户目录下的一个文件里：

- Windows：`%LOCALAPPDATA%\LLM Key Manager\keys.json`
- macOS：`~/Library/Application Support/LLM Key Manager/keys.json`

文件中只有 Key 内容是加密的，提供商、名称和时间以明文保存。Windows 上由系统的 DPAPI 以当前登录用户的身份加密；macOS 上用 AES-256-GCM 加密，主密钥存放在登录钥匙串里（服务 `llm-key-manager`，账户 `master-key`）。

因此解密离不开原来的系统账户：单独备份 `keys.json` 不足以恢复数据，文件也不能拿到另一台电脑、另一个账户或另一个系统上使用。在 macOS 上如果删除了钥匙串里的主密钥，已保存的 Key 将无法再解密。

还有两点需要留意。复制出来的 Key 会留在系统剪贴板和剪贴板历史工具里，应用不会自动清除。这种保护也挡不住已经以你的身份在电脑上运行的恶意程序：它是个人工具，不是企业级的密钥管理系统。

## 开发指南

以下内容面向参与开发的人和 agent。各文件的作用、想改某个功能该看哪里，见[代码地图](docs/architecture.md)。

应用基于 Wails v2，后端是 Go，前端是 Vue 3、TypeScript 和 TDesign。只支持 Windows 和 macOS。

### 环境

开发和构建都在 macOS arm64 上进行，Windows 版本通过交叉编译得到。需要以下工具：

| 工具                     | 版本       |
| ------------------------ | ---------- |
| Go                       | 1.27.4     |
| Node.js                  | >=22       |
| Wails                    | v2.17.1    |
| golangci-lint            | 2.15.0     |
| Xcode Command Line Tools | 随系统版本 |

[Go](https://go.dev/dl/) 和 [Node.js](https://nodejs.org/en/download) 从官网安装，其余用命令安装：

```bash
xcode-select --install
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
```

这些工具都要能从 `PATH` 找到。

克隆仓库后安装前端依赖：

```bash
(cd frontend && npm ci)
```

### 检查与构建

[scripts/check.sh](scripts/check.sh) 是唯一的检查入口，具体命令以脚本为准：

```bash
scripts/check.sh          # 格式化，检查前端和文档，对 macOS 与 Windows 两个目标做 Go lint
scripts/check.sh --build  # 在此基础上打包
```

脚本不会安装依赖，也不会启动应用。`--build` 会先清空 `build/bin/`，然后生成：

- `build/bin/llm-key-manager.app`：macOS universal（arm64 + x86_64）
- `build/bin/llm-key-manager.exe`：Windows amd64 的裸可执行文件，不含安装包

Windows 上运行需要 WebView2 Runtime。`.app` 没有签名，拷贝到别的 Mac 上首次打开时可能被 Gatekeeper 拦截。

整个 `build/` 目录不纳入版本控制，其中的图标和打包资源由 Wails 按需重新生成。

### 更新绑定

`frontend/wailsjs/` 是 Wails 根据 `app.go` 生成的前端绑定，纳入版本控制。修改了 `App` 的导出方法或它们用到的结构体后，需要重新生成：

1. 确认 `frontend/dist` 存在。`main.go` 嵌入了它，缺失时 Go 代码无法编译；没有就先运行一次 `scripts/check.sh`。
2. 运行 `GOTOOLCHAIN=local wails generate module`。
3. 查看 `frontend/wailsjs/` 的改动是否符合预期，再运行 `scripts/check.sh --build`。

`--build` 会比较打包前后的绑定。如果打包过程又改动了 `frontend/wailsjs/`，说明绑定没有提前更新，脚本会报错。

### 本地运行

在仓库根目录运行 `wails dev`，或直接打开构建出的 `.app`。两种方式读写的都是上文所说的真实数据文件和钥匙串，没有单独的开发数据。`wails.json` 没有配置前端 watcher，前端改动后的热更新不保证可用。
