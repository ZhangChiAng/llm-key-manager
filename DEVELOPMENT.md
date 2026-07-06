# llm-key-manager 开发指南

基于 Wails + Vue 3 + TypeScript 的桌面应用。

当前项目在远程 Linux 服务器上开发。日常开发以静态检查和 Windows 构建检查为主；真实运行效果在 Windows 桌面环境中验证。

## 开发循环

进入项目根目录：

```bash
cd .
```

远程 Linux 日常开发以静态检查和构建检查为主：

```bash
cd frontend
npm run lint
npm run format
npm run build
```

需要验证真实应用行为时，构建 Windows 产物并在 Windows 桌面环境中运行。前端调用通过 Wails 注入的 Go 后端绑定完成。

## 后端变更

修改 `app.go`、`main.go` 或其他 Go 文件后，运行完整 Wails 构建检查：

```bash
cd .
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64
```

如果新增、删除、重命名了暴露给前端调用的 Go 方法，或者修改了这些方法的参数和返回值，需要重新生成 Wails 前端绑定：

```bash
wails generate module
```

生成的绑定文件位于：

```text
frontend/wailsjs
```

## 检查命令

前端类型检查和生产构建检查：

```bash
cd frontend
npm run lint
npm run format
npm run build
```

Linux 上的完整 Wails 构建检查：

```bash
cd .
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64
```

远程无桌面 Linux 使用前面的构建检查命令；桌面运行效果在 Windows 环境中验证。

## Windows 构建循环

本项目假设 Linux 交叉编译 Windows 正常工作，不再维护 Windows 本机构建流程。

安装交叉编译工具：

```bash
sudo apt install -y mingw-w64 nsis
```

构建 Windows 可执行文件：

```bash
cd .
GOTOOLCHAIN=local wails build -clean -platform windows/amd64
```

构建 Windows 安装包：

```bash
GOTOOLCHAIN=local wails build -clean -platform windows/amd64 -nsis
```

构建产物位于：

```text
build/bin/
```

常见产物包括：

```text
build/bin/llm-key-manager.exe
```

如果使用 `-nsis`，还会生成 Windows 安装包。

发布前仍需把生成的 `.exe` 或安装包拿到 Windows 上实际运行一次，确认程序可以启动、WebView2 正常、文件路径和系统能力符合预期。
