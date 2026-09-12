# llm-key-manager 开发指南

本项目是基于 Wails、Go、Vue 3 和 TypeScript 的 Windows 个人应用，前端是唯一的用户交互入口。表单必填校验由前端负责；后端负责输入标准化、重复 Key 检查、加密存储和系统调用。

项目不再维护单元测试，也不引入测试框架、测试脚本或测试专用依赖注入。验证统一由编译、lint 和用户在 Windows 上进行的端到端测试完成。

## 开发环境

当前在远程 Linux 服务器上开发并交叉编译 Windows 产物。准备 Go、Node.js 24 或以上版本、Wails CLI 和 golangci-lint；安装 Windows 交叉编译和打包工具：

```bash
sudo apt install -y mingw-w64 nsis
```

从项目根目录安装锁文件指定的前端依赖：

```bash
cd frontend
npm ci
```

远程无桌面环境不要运行 `wails dev`。桌面行为通过构建出的 Windows 程序验证。

## 修改后的检查

每次完成代码修改，先格式化，再检查前后端。即使只修改一侧，也需要完成另一侧的检查。

在项目根目录格式化修改过的 Go 文件，例如：

```bash
gofmt -w app.go key_store.go
```

如果新增、删除、重命名导出的 `App` 方法，或更改参数、返回值，在前端检查前重新生成并审查 Wails 绑定：

```bash
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails generate module
```

绑定位于 `frontend/wailsjs/`。保持现有接口的内部重构不需要单独重新生成绑定。

前端格式化、lint、文档检查和生产构建：

```bash
cd frontend
npm run lint:fix
npm run format:fix
npm run lint
npm run format
npm run docs:lint
npm run build
```

回到项目根目录，执行后端 lint 和 Windows 集成构建：

```bash
cd ..
GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64
```

构建产物为 `build/bin/llm-key-manager.exe`。需要安装包时执行：

```bash
GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64 -nsis
```

## 用户端到端验收

将生成的 `.exe` 或安装包交由用户在 Windows 桌面环境中运行，使用临时测试记录验收：

1. 应用正常启动，WebView2 加载完成，已有记录可以读取。
2. 新增 Key，确认必填提示、列表脱敏、提供商分组和更新时间排序。
3. 保存相同提供商、名称和内容的 Key，确认重复提示。
4. 编辑提供商或名称，Key 内容留空；复制确认原密钥保留。再填写新内容，确认密钥被替换。
5. 复制 Key，确认系统剪贴板内容正确。
6. 删除测试记录，确认列表和数量更新。
7. 关闭并重新启动，确认保存和删除结果保留。

写入失败时表单内容应保留；如果写入已经成功而后续列表读取失败，界面应明确提示操作成功、需要重新刷新。

完成报告分别列出已通过的编译和 lint，以及用户端到端验收结果。用户尚未验收时标记为待验收；命令无法执行时记录完整命令、失败原因和已完成的部分检查。
