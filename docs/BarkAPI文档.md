# Bark API 官方文档（Markdown版）

> 该文档由 AI 生成

Bark 是一款专为 iOS 设计的开源、隐私优先的自定义推送工具，其 API 极简易用，支持通过 HTTP 请求向 iPhone/iPad 发送自定义推送通知，无需复杂配置，开箱即用。本文档详细介绍 Bark API 的所有用法、参数说明及常见场景示例。

## 一、API 基础信息

### 1.1 核心请求地址

官方公共服务端（默认，无需部署）：

```plain text
https://api.day.app/<your-key>/
```

自建服务端（私有化部署后）：

```plain text
https://<your-server-url>/<your-key>/
```

### 1.2 核心参数（必选）

| 参数位置     | 参数说明                                    | 示例                       |
|----------|-----------------------------------------|--------------------------|
| your-key | 设备唯一标识，打开 Bark App 首页即可获取（每个设备对应一个 Key） | abc123xyz（实际以自己的 Key 为准） |

### 1.3 支持请求方式

- GET：最简洁，适合简单推送（标题+内容）

- POST：支持更复杂的参数（如加密、自定义铃声），推荐用于高级场景

### 1.4 响应格式

所有请求均返回 JSON 格式，成功/失败均有明确提示：

```json
// 成功响应
{
  "code": 200,
  "message": "success",
  "timestamp": 1710000000
}

// 失败响应（如 Key 错误）
{
  "code": 400,
  "message": "invalid key",
  "timestamp": 1710000000
}

```

## 二、基础推送（最常用）

适用于简单场景，仅推送「标题+内容」，无需额外参数，支持 GET 请求快速调用。

### 2.1 语法格式（GET）

```plain text
https://api.day.app/<your-key>/<title>/<body>
```

### 2.2 参数说明

| 参数位置  | 参数说明          | 是否必选 | 注意事项                        |
|-------|---------------|------|-----------------------------|
| title | 推送标题（通知栏顶部显示） | 可选   | 若省略，仅显示内容；含特殊字符需 URL 编码     |
| body  | 推送内容（标题下方显示）  | 必选   | 长度建议不超过 200 字，含特殊字符需 URL 编码 |

### 2.3 示例

推送标题「服务器告警」，内容「CPU 使用率超过 90%」：

```plain text
https://api.day.app/abc123xyz/服务器告警/CPU使用率超过90%
```

效果：手机立即收到通知，显示标题和内容，点击可查看详情。

## 三、高级推送（URL 参数扩展）

在基础推送的基础上，通过 URL 参数添加高级功能，支持自定义铃声、跳转、分组等，适用于更灵活的场景。

### 3.1 语法格式（GET）

```plain text
https://api.day.app/<your-key>/<title>/<body>?param1=value1&param2=value2
```

### 3.2 所有支持的 URL 参数

| 参数名       | 参数说明                   | 可选值                                             | 示例                          |
|-----------|------------------------|-------------------------------------------------|-----------------------------|
| sound     | 自定义通知铃声                | 默认：default；支持系统铃声名称，或自定义铃声（需在 App 内导入）          | sound=alarm（系统闹钟铃声）         |
| url       | 点击通知跳转的链接              | 任意合法 URL（http/https 开头）                         | url=https://www.example.com |
| isArchive | 是否存档消息（App 内查看历史）      | 1（存档，默认）、0（不存档）                                 | isArchive=0                 |
| group     | 消息分组（App 内按分组管理）       | 自定义分组名称（如 server、task）                          | group=server-监控             |
| level     | 通知优先级（突破静音）            | active（普通，默认）、timeSensitive（时效性）、critical（重要警告） | level=critical              |
| copy      | 点击通知自动复制内容             | 任意文本（需 URL 编码）                                  | copy=服务器IP：192.168.1.1      |
| autoCopy  | 是否自动复制 body 内容（无需手动点击） | 1（自动复制）、0（不自动，默认）                               | autoCopy=1                  |

### 3.3 高级示例

推送「爬虫完成通知」，自定义铃声、点击跳转链接、自动复制结果：

```plain text
https://api.day.app/abc123xyz/爬虫完成/数据已抓取完毕?sound=success&url=https://scrapy.example.com&autoCopy=1&group=爬虫任务
```

## 四、POST 请求方式（推荐高级场景）

当参数较多（如加密内容、长文本）时，推荐使用 POST 请求，避免 URL 过长或参数泄露，请求体支持 JSON 格式。

### 4.1 语法格式

```plain text
请求地址：https://api.day.app/<your-key>
请求方法：POST
请求头：Content-Type: application/json
请求体：JSON 格式参数
```

### 4.2 支持的 JSON 参数

| 参数名       | 类型     | 说明                          | 可选值/示例                                |
|-----------|--------|-----------------------------|---------------------------------------|
| title     | string | 推送标题                        | "服务器告警"                               |
| body      | string | 推送内容                        | "CPU 超过 90%"                          |
| sound     | string | 铃声名称                        | "alarm"、"birdsong"                    |
| call      | int    | 持续响铃约 30 秒                  | 1（开启）                                 |
| isArchive | int    | 是否存档                        | 1（存档，默认）、0（不存档）                       |
| icon      | string | 自定义图标 URL                   | "https://example.com/icon.png"        |
| group     | string | 消息分组                        | "监控"、"任务"                             |
| level     | string | 通知优先级                       | active、timeSensitive、passive、critical |
| volume    | int    | 重要警告音量，仅 level=critical 时有效 | 0~10，默认 5                             |
| url       | string | 点击通知跳转链接                    | "https://www.example.com"             |
| copy      | string | 点击复制的内容                     | "192.168.1.1"                         |
| autoCopy  | int    | 自动复制 body 内容                | 1（自动复制）、0（默认）                         |
| badge     | int    | 角标数字                        | 42                                    |

> 加密推送时，以上参数作为明文 JSON 整体加密后发送，见第五章。

### 4.3 POST 示例（curl）

```bash
curl -X POST \
  https://api.day.app/abc123xyz \
  -H "Content-Type: application/json" \
  -d '{
    "title": "服务器告警",
    "body": "内存使用率超过 85%",
    "level": "critical",
    "url": "https://monitor.example.com",
    "group": "server-监控"
  }'
```

## 五、加密推送（隐私保护）

Bark 支持端到端加密，推送内容在客户端加密后发送，服务端无法读取，适合发送敏感信息（如验证码、隐私数据）。

> 参考文档：https://bark.day.app/#/encryption

### 5.1 支持的算法与模式

#### 密钥长度（Algorithm）

| 枚举值    | 密钥长度      | 说明            |
|--------|-----------|---------------|
| AES128 | **16 字节** | AES-128，默认值   |
| AES192 | **24 字节** | AES-192       |
| AES256 | **32 字节** | AES-256，安全性最高 |

#### 加密模式（Mode）

| 枚举值 | IV/Nonce        | 填充    | 说明                       |
|-----|-----------------|-------|--------------------------|
| CBC | IV **16 字节**    | PKCS7 | 默认模式，Bark 官方示例使用         |
| ECB | 无               | PKCS7 | 无状态，相同明文产生相同密文，不推荐用于敏感场景 |
| GCM | Nonce **12 字节** | 无需    | 自带消息认证标签（AEAD），安全性更高     |

> IV 和 Nonce 为空时 SDK 自动随机生成，并将实际值随密文一并发送。

### 5.2 加密流程

1. 在 Bark App 内，进入「设置 → 加密」，配置与 SDK 一致的密钥。
2. SDK 将**完整推送参数（JSON）** 序列化为明文，按选定的 Algorithm + Mode 加密。
3. 加密结果 Base64 编码，得到 `ciphertext`。
4. 发送请求时，**只传** `ciphertext` 和 `iv` 两个字段，不传明文内容。
5. Bark App 收到推送后，用本地密钥解密后显示。

### 5.3 加密请求体格式（POST JSON）

```json
{
  "ciphertext": "<Base64 编码的密文>",
  "iv": "<IV 原文（CBC）或 Nonce 原文（GCM）；ECB 模式无此字段>"
}
```

### 5.4 加密示例（Shell，AES128-CBC）

```bash
#!/usr/bin/env bash
# 参考：https://bark.day.app/#/encryption

deviceKey='your-device-key'
# 完整推送参数作为明文 JSON
json='{"body": "你的验证码是：123456", "sound": "birdsong"}'

# 密钥和 IV 必须各为 16 字节（AES128-CBC）
key='zxcvbnmlkjhgfdsa'
iv='zxcfdsaqwertyuio'

# OpenSSL 要求 Key 和 IV 使用十六进制编码传入
key_hex=$(printf $key | xxd -ps -c 200)
iv_hex=$(printf $iv | xxd -ps -c 200)

# AES-128-CBC 加密 + Base64 编码
ciphertext=$(echo -n $json | openssl enc -aes-128-cbc -K $key_hex -iv $iv_hex | base64)

# 发送：仅传 ciphertext 和 iv，不传明文
curl --data-urlencode "ciphertext=$ciphertext" \
     --data-urlencode "iv=$iv" \
     https://api.day.app/$deviceKey
```

### 5.5 各模式 OpenSSL 命令参考

```bash
# AES128-CBC（密钥 16 字节，IV 16 字节）
openssl enc -aes-128-cbc -K <key_hex> -iv <iv_hex>

# AES192-CBC（密钥 24 字节，IV 16 字节）
openssl enc -aes-192-cbc -K <key_hex> -iv <iv_hex>

# AES256-CBC（密钥 32 字节，IV 16 字节）
openssl enc -aes-256-cbc -K <key_hex> -iv <iv_hex>

# AES128-ECB（密钥 16 字节，无 IV）
openssl enc -aes-128-ecb -K <key_hex>

# AES256-ECB（密钥 32 字节，无 IV）
openssl enc -aes-256-ecb -K <key_hex>
```

> **注意事项**
> - 被加密的是推送参数的完整 JSON，而非单独的 body 字段。
> - ECB 模式无 IV，`iv` 字段为空，发送时省略该字段。
> - GCM 模式的 Nonce 为 12 字节，通过 `iv` 字段传递。
> - IV/Nonce 若随机生成，发送时需将实际使用的原文（非 hex）一并传入 `iv` 字段。

## 六、常见问题与注意事项

- Key 错误：提示 `invalid key`，请检查 App 内的 Key 是否正确，区分大小写。

- 推送失败：检查网络是否正常，官方服务端偶尔波动，可尝试重试；自建服务端检查部署是否正确。

- 特殊字符处理：标题、内容、参数中含中文、空格、特殊符号（如 !@#），需进行 URL 编码（GET 请求）或 JSON 转义（POST 请求）。

- 消息不存档：设置 `isArchive=0`，消息仅在通知栏显示，点击后消失，不进入 App 历史。

- 铃声不生效：确保手机铃声开启，自定义铃声需在 App 内导入，参数填写铃声名称（不含后缀）。

## 七、自建服务端 API 说明

自建服务端（bark-server）后，API 用法与官方服务端完全一致，仅需将请求地址替换为你的服务器地址即可。

自建服务端额外支持：批量推送、自定义域名、数据备份等功能，具体参考 bark-server 官方文档。
