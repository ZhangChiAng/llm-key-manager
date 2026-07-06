# LLM Key Manager

LLM Key Manager 是一个面向 Windows 的本地 API Key 管理桌面应用，用于集中保存、查看、复制和维护大模型服务商的访问密钥。

项目基于 Wails、Go、Vue 3 和 TypeScript 构建。密钥内容仅保存在本机，后端使用 Windows DPAPI 加密落盘；前端列表只展示脱敏后的 Key，复制时由后端解密并写入系统剪贴板，避免把明文密钥作为列表数据返回给界面。

## 功能特性

- 本地保存 API Key，支持记录提供商、名称和 Key 内容。
- 按提供商分组展示已保存记录，并按更新时间排序。
- 列表展示脱敏 Key，长 Key 保留前 10 位和后 4 位。
- 支持编辑提供商、名称，或在需要时替换 Key 内容。
- 支持一键复制明文 Key 到系统剪贴板。
- 支持删除不再使用的 Key 记录。
- 保存前会检查相同提供商、名称和 Key 内容的重复记录。

## 安全模型

当前版本采用本地单机安全模型：

- 密钥文件位于应用可执行文件同目录的 `keys.json`。
- `keys.json` 中的 Key 内容使用 Windows DPAPI 加密后再 Base64 编码保存。
- 加密和解密能力依赖当前 Windows 用户环境，项目不提供云同步、远程备份或跨平台解密。
- 前端 `ListKeys` 只接收脱敏后的 Key；复制明文 Key 时通过 Wails 后端直接写入剪贴板。
- 如果 `keys.json` 被手动破坏为非法格式，应用会返回本地存储格式异常提示。

这不是团队级密钥保险箱，也不替代企业 KMS、密码管理器或集中审计系统。它更适合个人在 Windows 桌面上管理多个 LLM 客户端、工具或服务商所需的 API Key。

## 技术栈

- 桌面框架：Wails v2
- 后端：Go
- 前端：Vue 3、TypeScript、Vite
- UI 组件：TDesign Vue Next
- 本地加密：Windows DPAPI
- 目标平台：Windows

## 界面与使用流程

应用主界面由两部分组成：

1. 顶部表单用于新增 Key，需要填写提供商、Key 名称和 Key 内容。
2. 下方列表按提供商分组展示已保存记录，可展开分组后执行编辑、复制或删除操作。

编辑记录时，Key 内容输入框默认留空；留空表示只更新提供商或名称，不修改已保存的密钥内容。

## 从源码构建

当前仓库按 Windows 桌面应用维护。可以在 Linux 环境中交叉编译 Windows 产物，但真实运行效果仍应在 Windows 桌面环境验证。

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

构建产物位于 `build/bin/`。

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
```

更完整的开发说明见 [DEVELOPMENT.md](DEVELOPMENT.md)。

## 项目结构

```text
.
├── app.go                 # Wails 后端绑定与本地密钥存储逻辑
├── main.go                # Wails 应用入口与窗口配置
├── dpapi_windows.go       # Windows DPAPI 加密/解密实现
├── frontend/src/          # Vue 3 前端源码
├── frontend/wailsjs/      # Wails 生成的前端绑定
├── build/windows/         # Windows 打包资源
└── docs/                  # 设计与实现说明
```

## 平台说明

本项目是 Windows-only 桌面应用。仓库中的 `build/darwin/` 是 Wails 模板遗留材料，并不表示项目支持 macOS。
