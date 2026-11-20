# LLM Key Manager

LLM Key Manager 是一个面向 Windows 和 macOS 的本地 API Key 管理桌面应用，用于集中保存、查看、复制和维护大模型服务商的访问密钥。

项目基于 Wails、Go、Vue 3 和 TypeScript 构建。密钥内容仅保存在本机，后端在 Windows 上使用 DPAPI、在 macOS 上使用 AES-256-GCM 加密落盘；前端列表只展示脱敏后的 Key，复制时由后端解密并写入系统剪贴板，避免把明文密钥作为列表数据返回给界面。

## 功能特性

- 本地保存 API Key，支持记录提供商、名称和 Key 内容。
- 按提供商分组展示已保存记录，并按更新时间排序。
- 列表展示脱敏 Key，长 Key 保留前 10 位和后 4 位。
- 支持编辑提供商、名称，或在需要时替换 Key 内容。
- 支持一键复制明文 Key 到系统剪贴板。
- 支持删除不再使用的 Key 记录。
- 支持并行检测预设提供商的已保存 Key，逐条显示检测结论。
- 保存前会检查相同提供商、名称和 Key 内容的重复记录。

## 安全模型

当前版本采用本地单机安全模型：

- 密钥文件位于当前用户的本机应用数据目录：Windows 为 `%LOCALAPPDATA%\LLM Key Manager\keys.json`，macOS 为 `~/Library/Application Support/LLM Key Manager/keys.json`。
- Windows 上，Key 内容使用 Windows DPAPI 加密后再 Base64 编码保存。
- macOS 上，Key 内容使用 AES-256-GCM 加密，随机 nonce 与密文拼接后 Base64 编码保存。32 字节主密钥在首次保存时随机生成，存放在登录钥匙串的通用密码项中（服务 `llm-key-manager`，账户 `master-key`），应用通过系统自带的 `/usr/bin/security` 读取；重新构建未签名应用后不会反复弹出钥匙串授权提示。删除该钥匙串项后，已保存的 Key 无法再解密。
- 两个平台的解密能力都绑定当前系统用户，项目不提供云同步、远程备份或跨平台解密，`keys.json` 不能在 Windows 和 macOS 之间互相拷贝使用。
- 前端 `ListKeys` 只接收脱敏后的 Key；复制明文 Key 时通过 Wails 后端直接写入剪贴板。
- 如果 `keys.json` 被手动破坏为非法格式，应用会返回本地存储格式异常提示。

这不是团队级密钥保险箱，也不替代企业 KMS、密码管理器或集中审计系统。它更适合个人在 Windows 或 macOS 桌面上管理多个 LLM 客户端、工具或服务商所需的 API Key。

## 技术栈

- 桌面框架：Wails v2
- 后端：Go
- 前端：Vue 3、TypeScript、Vite
- UI 组件：TDesign Vue Next
- 本地加密：Windows DPAPI；macOS AES-256-GCM（主密钥保存在钥匙串）
- 目标平台：Windows、macOS

## 界面与使用流程

应用主界面由两部分组成：

1. 顶部表单用于新增 Key，需要填写提供商、Key 名称和 Key 内容。
2. 下方列表按提供商分组展示已保存记录，可展开分组后执行编辑、复制或删除操作。

编辑记录时，Key 内容输入框默认留空；留空表示只更新提供商或名称，不修改已保存的密钥内容。

### 过期检测

点击“过期检测”立即并行检测所有预设提供商的已保存 Key，每条发送一次真实请求，会产生少量 API 用量。展开提供商分组后，Key 内容旁会逐条显示转圈标记，完成后显示“可用”“认证失败”“余额不足”“权限受限”“限流”或“检测失败”。自定义提供商显示“不支持检测”，不会发送请求。

| 提供商          | 请求地址                                     | 固定模型          |
| --------------- | -------------------------------------------- | ----------------- |
| DeepSeek        | `https://api.deepseek.com/chat/completions`  | `deepseek-flash`  |
| MiniMax（国内） | `https://api.minimax.cn/v1/chat/completions` | `MiniMax-M3`      |
| OfoxAI          | `https://api.ofox.io/v1/chat/completions`    | `qwen/qwen-flash` |

请求仅发送用户消息 `Hi`，使用 `stream: false`、`max_tokens: 1`，DeepSeek 和 MiniMax 显式关闭思考。每条请求最多等待 30 秒，随应用生命周期取消，不自动重试、切换模型或跟随重定向。密钥由后端解密并通过 Bearer 请求头发送给对应提供商，前端只传记录 ID、接收结论；不记录或展示请求头、明文密钥及原始响应体。

“可用”仅表示当前固定模型返回了有效生成响应，输出达到长度限制也算成功。余额不足、限流、网络异常不能据此判为 Key 过期；模型不可用、异常响应及未知错误显示“检测失败”。固定模型名称可能随服务商调整，需在实际账户中确认可调用性。

检测期间禁用重复检测、保存、编辑和删除；没有可检测记录时禁用按钮。结果仅保存在本次运行的内存中，普通刷新保留未变更记录的结果，修改或删除后清除对应结果，重启后清空。检测不会更改存储格式或记录更新时间。

接口参考：[DeepSeek](https://api-docs.deepseek.com/api/create-chat-completion/)、[MiniMax](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[MiniMax 错误码](https://platform.minimax.cn/docs/api-reference/errorcode)、[OfoxAI](https://ofox.ai/docs/api)。

## 从源码构建

当前仓库按 Windows 和 macOS 桌面应用维护。可以在 Linux 环境中交叉编译 Windows 产物；macOS 产物需在安装了 Xcode Command Line Tools 的 Mac 上构建。真实运行效果应分别在 Windows 和 macOS 桌面环境验证。

准备前端依赖：

```bash
cd frontend
npm install
```

构建 Windows 可执行文件：

```bash
cd ..
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64
```

构建 Windows 安装包：

```bash
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64 -nsis
```

在 Mac 上构建 macOS 应用（同时包含 Apple Silicon 和 Intel 架构）：

```bash
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform darwin/universal
```

构建产物位于 `build/bin/`（Windows 为 `llm-key-manager.exe`，macOS 为 `llm-key-manager.app`）。

## 开发检查

前端检查：

```bash
cd frontend
npm run lint
npm run format
npm run build
```

后端检查：

```bash
cd ..
GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
GOOS=windows GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
GOOS=darwin CGO_ENABLED=0 GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
```

更完整的开发说明见 [DEVELOPMENT.md](DEVELOPMENT.md)。

## 项目结构

```text
.
├── app.go                  # Wails 后端绑定与密钥操作编排
├── key_store.go            # 加密存储、重复检查和脱敏辅助函数
├── main.go                 # Wails 应用入口与窗口配置
├── platform_windows.go     # Windows DPAPI 加密/解密与存储目录
├── platform_darwin.go      # macOS AES-GCM 加密/解密、钥匙串主密钥与存储目录
├── platform_unsupported.go # 仅供 Linux 开发机类型检查的报错实现
├── frontend/src/           # Vue 3 前端源码
├── frontend/wailsjs/       # Wails 生成的前端绑定
├── build/windows/          # Windows 打包资源
├── build/darwin/           # macOS 打包资源
└── docs/                   # 设计与实现说明
```

## 平台说明

本项目支持 Windows 和 macOS 桌面，不支持 Linux 桌面。`platform_unsupported.go` 只用于让 Linux 开发机通过类型检查和 lint，在该环境调用存储会返回不支持错误。
