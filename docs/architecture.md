# 代码地图

本文说明每个文件做什么，以及想改某个功能时该打开哪个文件。行为细节以代码和其中的注释为准，这里不重复。增删、移动文件，或重命名这里提到的符号时，请在同一次改动里更新本文。

环境和命令见 [README 的开发指南](../README.md#开发指南)。

## 一次操作怎么走

界面上的每个操作都沿同一条路径到达后端：

```text
frontend/src/components/*.vue    界面
  → composables/useKeyManager.ts 前端唯一调用后端的地方，持有全部共享状态
  → frontend/wailsjs/            Wails 生成的绑定
  → app.go                       App 的导出方法
  → internal/keystore/           读写 keys.json，加解密
    internal/probe/              向服务商发检测请求
```

`keystore` 和 `probe` 互不依赖，由 `app.go` 串起来。例如检测时，`app.go` 先从 keystore 取出记录并解密，再把明文交给 probe。

## 文件

### 后端

| 文件                                   | 作用                                                                                                                                                                                |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `main.go`                              | 程序入口：窗口标题、大小和背景色，嵌入 `frontend/dist`，把 `App` 绑定给前端                                                                                                         |
| `app.go`                               | 前端能调用的全部后端方法：`SaveKey`、`ListKeys`、`UpdateKey`、`CopyKey`、`DeleteKey`、`ListProviderPresets`、`ProbeKey`；返回给前端的结构体 `KeyRecord`、`KeyProbeResult`；写剪贴板 |
| `internal/keystore/store.go`           | `Store` 的增删改查、重复判定、脱敏                                                                                                                                                  |
| `internal/keystore/file.go`            | `keys.json` 的路径、落盘结构、读取时的校验、原子写入                                                                                                                                |
| `internal/keystore/protect_windows.go` | Windows：DPAPI 加解密，数据目录                                                                                                                                                     |
| `internal/keystore/protect_darwin.go`  | macOS：AES-GCM 加解密，钥匙串主密钥的读取与创建，数据目录                                                                                                                           |
| `internal/probe/probe.go`              | 预设提供商表、检测请求、把响应归类为状态                                                                                                                                            |

### 前端

以下路径都在 `frontend/` 下。

| 文件                               | 作用                                                                                                   |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `src/main.ts`                      | 创建 Vue 应用，注册全局使用的 TDesign 组件                                                             |
| `src/App.vue`                      | 页面外壳：上下两个面板，启动时加载列表                                                                 |
| `src/components/AddKeyPanel.vue`   | 上方面板：标题、“刷新”按钮、新增表单、页面级的成功与错误提示                                           |
| `src/components/KeyListPanel.vue`  | “已保存 Key”面板：“过期检测”按钮、按提供商分组与折叠、排序、表格、每行的编辑、复制、删除按钮、检测结果 |
| `src/components/KeyEditDialog.vue` | 编辑弹窗及弹窗内的错误提示                                                                             |
| `src/components/KeyForm.vue`       | 新增和编辑共用的三个输入框与必填校验                                                                   |
| `src/components/ProviderInput.vue` | 提供商输入框：预设下拉加自由输入                                                                       |
| `src/composables/useKeyManager.ts` | 记录列表、预设、提示消息、各操作的进行中状态和检测结果；封装对后端的全部调用，把后端错误转成中文提示   |
| `src/probeStatus.ts`               | 检测状态的类型与中文标签                                                                               |
| `src/style.css`                    | 全局样式：字体、页面底色、面板、提示文字颜色。组件自己的样式写在各自的 `<style scoped>` 里             |
| `wailsjs/`                         | 生成的绑定，不手改，更新方式见 README 的“更新绑定”                                                     |

### 工具与配置

| 文件                                                                                          | 作用                                         |
| --------------------------------------------------------------------------------------------- | -------------------------------------------- |
| `scripts/check.sh`                                                                            | 格式化、检查、打包的全部命令                 |
| `frontend/package.json`                                                                       | 前端依赖，以及脚本调用的 `fix`、`check` 命令 |
| `wails.json`                                                                                  | 应用名、输出文件名、前端构建命令             |
| `.golangci.yml`                                                                               | Go lint 规则                                 |
| `frontend/eslint.config.js`、`frontend/.prettierrc.json`、`frontend/.markdownlint-cli2.jsonc` | 前端代码与 Markdown 文档的 lint 和格式规则   |
| `frontend/vite.config.ts`、`frontend/tsconfig.json`                                           | 前端构建与类型检查配置                       |

## 按需求找代码

左列是想改的东西，右列是位置。同一行列出多处时，它们需要一起改。

### 过期检测

| 想改的东西                             | 位置                                                                                              |
| -------------------------------------- | ------------------------------------------------------------------------------------------------- |
| 增减预设提供商，换检测模型或接口地址   | `probe.go` 的 `providerPresets`；README 的预设表                                                  |
| 检测请求的内容、超时、响应大小上限     | `probe.go` 的 `probeCompletion` 前半部分                                                          |
| 响应如何判定为各种状态                 | `probe.go` 的 `probeCompletion` 后半部分；MiniMax 的业务码在 `providerPresets` 的 `businessCodes` |
| 增减或重命名检测状态                   | `probe.go` 的 `Status` 常量；`probeStatus.ts` 的 `ProbeStatus` 和 `probeStatusLabels`             |
| 状态的中文文案                         | `probeStatus.ts` 的 `probeStatusLabels`                                                           |
| 哪些记录会被检测，自定义提供商如何处理 | `app.go` 的 `ProbeKey`；`useKeyManager.ts` 的 `isProbeSupported`、`probeState`                    |
| 并行方式，检测期间禁止哪些操作         | `useKeyManager.ts` 的 `probeKeys` 及各操作开头的 `isProbing` 判断；按钮禁用条件在各组件模板里     |
| 检测结果何时保留、何时清除             | `useKeyManager.ts` 的 `refreshKeys`、`updateKey`、`deleteKey`                                     |
| 检测结果在表格里的样子                 | `KeyListPanel.vue` 的 `.probe-result`                                                             |

### 保存、编辑、删除、复制

| 想改的东西                                 | 位置                                                                                                        |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------- |
| 必填校验及其提示                           | `KeyForm.vue` 的 `submit`                                                                                   |
| 新增时的处理：去首尾空白、生成 ID 和时间   | `store.go` 的 `Store.Add`                                                                                   |
| 编辑时的处理：留空沿用原 Key、无变化不写入 | `store.go` 的 `Store.Update`；“保存”按钮置灰的条件在 `KeyEditDialog.vue` 的 `isUnchanged`                   |
| 重复判定                                   | `store.go` 的 `ensureUniqueKeyRecord`                                                                       |
| 脱敏规则                                   | `store.go` 的 `maskKeyValue`                                                                                |
| 删除                                       | `store.go` 的 `Store.Delete`；确认气泡在 `KeyListPanel.vue`                                                 |
| 复制到剪贴板                               | `app.go` 的 `CopyKey`                                                                                       |
| 成功提示、失败提示的文案                   | `useKeyManager.ts` 的各操作函数、`refreshAfterWrite`、`getErrorMessage`                                     |
| 后端错误码                                 | `store.go` 的 `errKeyDuplicate`、`file.go` 的 `errKeyStoreInvalid`；`useKeyManager.ts` 的 `getErrorMessage` |
| 增加或修改后端接口                         | `app.go`；重新生成绑定（见 README 的“更新绑定”）；`useKeyManager.ts`                                        |

### 存储与加密

| 想改的东西                     | 位置                                                                                       |
| ------------------------------ | ------------------------------------------------------------------------------------------ |
| 数据文件的位置和文件名         | `file.go` 的 `filePath` 与常量；两个 `protect_*.go` 的 `keyStoreBaseDir`                   |
| 文件里保存哪些字段             | `file.go` 的 `storedKeyRecord`；返回给前端的字段在 `app.go` 的 `KeyRecord`                 |
| 读取时的校验，写入方式         | `file.go` 的 `load`、`validate`、`writeFileAtomically`                                     |
| 加密和解密                     | 两个 `protect_*.go` 的 `protectKeyValue`、`unprotectKeyValue`                              |
| 钥匙串主密钥的读取、创建和缓存 | `protect_darwin.go` 的 `loadMasterKey`、`readKeychainMasterKey`、`createKeychainMasterKey` |

### 界面

| 想改的东西                           | 位置                                                                |
| ------------------------------------ | ------------------------------------------------------------------- |
| 分组、排序、展开折叠、时间的显示格式 | `KeyListPanel.vue` 的 `groupedKeys`、`toggleProvider`、`formatDate` |
| 表格的列、列宽、行内按钮             | `KeyListPanel.vue` 的模板                                           |
| 输入框、占位文字、表单布局           | `KeyForm.vue`；提供商输入框在 `ProviderInput.vue`                   |
| 编辑弹窗                             | `KeyEditDialog.vue`                                                 |
| 提示消息显示的位置                   | 页面级在 `AddKeyPanel.vue`，弹窗内在 `KeyEditDialog.vue`            |
| 窄窗口下的布局                       | `style.css` 和各组件里的 `@media (max-width: 760px)`                |
| 窗口标题、大小、背景色               | `main.go`                                                           |

## 已有数据依赖的格式

用户电脑上已有的 `keys.json` 和钥匙串项靠下面这些约定保持可读。改动其中任何一项，旧记录都会无法加载或解密。

- **文件位置**：平台数据目录下的 `LLM Key Manager/keys.json`。Windows 的数据目录是 `%LOCALAPPDATA%`，macOS 是 `~/Library/Application Support`。
- **文件内容**：JSON 数组，每条记录有 `id`、`provider`、`name`、`encryptedValue`、`createdAt`、`updatedAt` 六个字段，都不能为空。
- **`encryptedValue`**：平台密文的标准 Base64 编码。Windows 的密文是不带额外熵的 DPAPI 输出；macOS 的密文是 AES-256-GCM 的 `nonce || ciphertext+tag`。
- **macOS 主密钥**：32 字节，以十六进制文本存为登录钥匙串的通用密码项，服务 `llm-key-manager`，账户 `master-key`。
- **预设提供商的名称**：已保存的记录按提供商名称精确匹配预设，改名会让旧记录变成“不支持检测”。
