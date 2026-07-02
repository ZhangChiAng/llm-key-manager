# llm-key-manager

基于 Wails + Vue 3 + TypeScript 的桌面应用。

当前项目在远程 Linux 服务器上开发。日常查看界面效果使用 Vite + SSH 端口转发；Windows 发布包通过 Linux 交叉编译生成。

## 开发循环

进入项目根目录：

```bash
cd .
```

在远程 Linux 服务器上启动前端开发服务：

```bash
cd frontend
npm run dev -- --host 127.0.0.1
```

在本地电脑上打开 SSH 端口转发：

```bash
ssh -L 5173:127.0.0.1:5173 <用户名>@<服务器地址>
```

然后在本地浏览器访问：

```text
http://localhost:5173
```

这足够用于查看 Vue 和 TDesign 的界面效果。修改 `frontend/src` 下的前端代码后，Vite 会自动热更新。

## 后端变更

修改 `app.go`、`main.go` 或其他 Go 文件后，运行完整 Wails 构建检查：

```bash
cd .
GOTOOLCHAIN=local wails build -clean
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
npm run build
```

Linux 上的完整 Wails 构建检查：

```bash
cd .
GOTOOLCHAIN=local wails build -clean
```

当前服务器是无桌面远程环境，不建议把下面命令作为日常开发流程：

```bash
wails dev
```

因为它会尝试启动 Linux 桌面窗口。

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
