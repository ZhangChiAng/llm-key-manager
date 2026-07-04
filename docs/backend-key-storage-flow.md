                              后端密钥存储流程图
╔═════════════════════════════════════════════════════════════════════════════════════════╗

一、数据结构定义
──────────────────────────────────────────────────────────────────────────────────────

   结构体                                  说明
   ─────────────────────────────────────── ──────────────────────────────────────────────
   App                                    后端状态容器，持有 Wails 运行时 ctx
   ├── ctx  context.Context               用于调用 runtime.ClipboardSetText 等前端 API

   KeyRecord          (导出，返给前端)     对应 storedKeyRecord 的「脱敏版本」
   ├── ID            string               唯一标识
   ├── Provider      string               API 提供商
   ├── Name          string               自定义名称
   ├── MaskedValue   string               脱敏后的 key（如 sk-xxx***xxx）
   └── CreatedAt     string               RFC3339 时间

   storedKeyRecord    (内部，落盘)         磁盘存储的「加密版本」
   ├── ID            string               唯一标识
   ├── Provider      string               API 提供商
   ├── Name          string               自定义名称
   ├── EncryptedValue string              base64( DPAPI加密字节流 )
   └── CreatedAt     string               RFC3339 时间


   App ────── 读写 ──────► storedKeyRecord[] ──── JSON序列化 ────► keys.json


二、导出方法 → 内部函数 调用链
──────────────────────────────────────────────────────────────────────────────────────

  前端 (Vue/TS)                      内部函数 & 平台层                 OS / 磁盘
  ══════════════                     ══════════════════              ═════════════

  SaveKey(provider,name,value)
  │
  ├─ TrimSpace(三个参数)
  ├─ readStoredKeys()
  │      ├─ keyStorePath() ───────────► os.Executable() ──► filepath.Join
  │      ├─ os.ReadFile(path) ─────────────────────────────────► keys.json
  │      ├─ json.Unmarshal(data, &records)
  │      └─ validateStoredKeys(records)
  │
  ├─ protectKeyValue(value) ──────────► dpapi_windows.go (DPAPI加密)
  │                                     dpapi_unsupported.go (空实现)
  │
  ├─ base64.StdEncoding.EncodeToString(加密字节流)
  ├─ strconv.FormatInt(time.Now().UnixNano(), 10) ──► ID 生成
  ├─ time.Now().Format(time.RFC3339) ──► CreatedAt 生成
  │
  ├─ new storedKeyRecord{...}
  │
  └─ writeStoredKeys(records)
         ├─ keyStorePath()
         ├─ json.MarshalIndent(records)
         └─ os.WriteFile(path, data, 0666) ─────────────────────► keys.json


  ListKeys()
  │
  ├─ readStoredKeys()
  │
  └─ for each storedKeyRecord:
         ├─ decryptStoredValue(record)
         │      ├─ base64.StdEncoding.DecodeString(EncryptedValue)
         │      └─ unprotectKeyValue(字节流) ─► dpapi_xxx.go (DPAPI解密)
         │
         ├─ maskKeyValue(明文)
         │      └─ "sk-abcde...1234" ──► "sk-abcde***1234"
         │
         └─ new KeyRecord{ ID, Provider, Name, MaskedValue, CreatedAt }


  CopyKey(id)
  │
  ├─ readStoredKeys()
  │
  ├─ 遍历找到 record.ID == id:
  │      ├─ decryptStoredValue(record) ──► 得到明文
  │      └─ runtime.ClipboardSetText(ctx, 明文) ────────────────► 系统剪贴板
  │
  └─ 未找到 ──► error "key record not found"


  DeleteKey(id)
  │
  ├─ readStoredKeys()
  ├─ 过滤: 移除 record.ID == id 的记录
  │
  └─ writeStoredKeys(过滤后的 records)



三、数据流转：加密写入路径
──────────────────────────────────────────────────────────────────────────────────────

  用户输入                   处理中                      落盘字段
  ══════════               ════════════               ═════════════════════
  provider ──► TrimSpace ──────────────────────────► storedKeyRecord.Provider
  name     ──► TrimSpace ──────────────────────────► storedKeyRecord.Name
  value    ──► TrimSpace ──► DPAPI加密 ──► base64 ─► storedKeyRecord.EncryptedValue

  time.Now() ──► UnixNano() ──► FormatInt ─────────► storedKeyRecord.ID
  time.Now() ──► Format(RFC3339) ──────────────────► storedKeyRecord.CreatedAt

                                       │
                                       ▼
                                    keys.json  (JSON 数组)



四、数据流转：解密读取路径
──────────────────────────────────────────────────────────────────────────────────────

  keys.json (JSON数组)
       │
       ▼
  storedKeyRecord[] ── 逐条读取 ──┬──► ID        ────────────────┬──────────────► KeyRecord.ID
                                  ├──► Provider  ────────────────┤  直接传递     KeyRecord.Provider
                                  ├──► Name      ────────────────┤              KeyRecord.Name
                                  ├──► EncryptedValue ──────────┐│
                                  │        │                    ││
                                  │        ▼                    ││
                                  │   base64.DecodeString       ││
                                  │        │                    ││
                                  │        ▼                    ││
                                  │   unprotectKeyValue         ││
                                  │   (DPAPI解密 → 明文)        ││
                                  │        │                    ││
                                  │        ▼                    ││
                                  │   maskKeyValue() ───────────┤▼───► KeyRecord.MaskedValue
                                  │                             │      ("sk-xxx***xxx")
                                  └──► CreatedAt ───────────────┘      KeyRecord.CreatedAt

                                                                       │
                                                                       ▼
                                                                  前端列表展示


  CopyKey 分支:
  ────────────
  同一解密路径 ──► 明文 ──► runtime.ClipboardSetText() ──► 系统剪贴板



五、方法 ↔ 数据 ↔ 磁盘 IO 速查表
──────────────────────────────────────────────────────────────────────────────────────

  方法        输入              操作的数据结构           输出类型        磁盘 IO
  ──────     ────────────────  ──────────────────────  ─────────────  ────────
  SaveKey    3 个 string       写入 storedKeyRecord     error          写
  ListKeys   无                读取 → 脱敏 → KeyRecord  []KeyRecord    读
  CopyKey    id (string)       读取 → 解密 → 明文       error          读
  DeleteKey  id (string)       读取 → 过滤 → 重写       error          读 + 写


六、内部函数一览
──────────────────────────────────────────────────────────────────────────────────────

  函数                       职责
  ──────────────────────     ─────────────────────────────────────────────────────
  readStoredKeys()           打开 keys.json → JSON反序列化 → 校验 → 返回 []storedKeyRecord
  writeStoredKeys(records)   []storedKeyRecord → JSON序列化 → 写入 keys.json
  validateStoredKeys(...)    检查每条记录的 ID/Provider/Name/EncryptedValue/CreatedAt 非空
  keyStorePath()             返回 <exe所在目录>/keys.json
  decryptStoredValue(...)    base64解码 EncryptedValue → DPAPI解密 → 明文
  maskKeyValue(...)          明文长度>14时前缀10字符+***+后缀4字符，否则全***
  protectKeyValue(...)       DPAPI加密明文 → 返回字节流 (平台相关)
  unprotectKeyValue(...)     DPAPI解密字节流 → 返回明文   (平台相关)

╚═════════════════════════════════════════════════════════════════════════════════════════╝
