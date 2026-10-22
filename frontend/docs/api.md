# MC Launcher API 文档

## 基础信息

- **Base URL**: `http://localhost:9876/api`
- **Content-Type**: `application/json`
- **响应格式**: JSON

## 通用响应格式

### 成功响应
```json
{
  "ok": true,
  "data": { ... }
}
```

### 错误响应
```json
{
  "ok": false,
  "error": "错误信息"
}
```

---

## 接口列表

### 1. 健康检查

**GET** `/api/ping`

返回启动器状态。

```json
{
  "ok": true,
  "data": {
    "status": "ok",
    "version": "2.0.0"
  }
}
```

---

### 2. 启动器信息

**GET** `/api/info`

返回启动器基本信息。

```json
{
  "ok": true,
  "data": {
    "version": "2.0.0",
    "baseDir": "/path/to/.mc-launcher",
    "language": "zh",
    "goos": "linux",
    "goarch": "amd64",
    "accounts": 2,
    "versions": ["1.20.1", "1.19.4"]
  }
}
```

---

### 3. 版本管理

#### 获取已安装版本列表

**GET** `/api/versions`

```json
{
  "ok": true,
  "data": ["1.20.1", "1.19.4", "1.18.2"]
}
```

#### 安装版本

**POST** `/api/versions/install`

请求体：
```json
{
  "version": "1.20.1"
}
```

响应：
```json
{
  "ok": true,
  "data": null
}
```

#### 删除版本

**POST** `/api/versions/remove`

请求体：
```json
{
  "version": "1.20.1"
}
```

#### 验证版本

**POST** `/api/versions/verify`

请求体：
```json
{
  "version": "1.20.1"
}
```

响应：
```json
{
  "ok": true,
  "data": {
    "valid": true,
    "message": "版本验证通过"
  }
}
```

---

### 4. 启动游戏

**POST** `/api/launch`

请求体：
```json
{
  "version": "1.20.1",
  "username": "Player"
}
```

响应：
```json
{
  "ok": true,
  "data": null
}
```

---

### 5. 账号管理

#### 获取账号列表

**GET** `/api/accounts`

```json
{
  "ok": true,
  "data": [
    {
      "type": "offline",
      "name": "Player",
      "uuid": "xxxx-xxxx",
      "accessToken": "xxxx"
    }
  ]
}
```

#### 离线登录

**POST** `/api/accounts/offline`

请求体：
```json
{
  "name": "Player"
}
```

#### Microsoft 登录

**POST** `/api/accounts/microsoft`

需要浏览器授权，返回设备码和授权链接。

---

### 6. Java 管理

#### 获取 Java 列表

**GET** `/api/java`

```json
{
  "ok": true,
  "data": [
    {
      "path": "/usr/lib/jvm/java-17/bin/java",
      "major": 17,
      "version": "17.0.8"
    }
  ]
}
```

#### 安装 Java

**POST** `/api/java/install`

请求体：
```json
{
  "major": 17
}
```

#### 设置 Java 路径

**PUT** `/api/java/set`

请求体：
```json
{
  "path": "/path/to/java"
}
```

---

### 7. Modrinth API

#### 搜索项目

**GET** `/api/modrinth/search?q=optifine&type=mod&page=1`

查询参数：
- `q`: 搜索关键词
- `type`: 项目类型 (mod, modpack, shader, resourcepack, datapack)
- `page`: 页码

响应：
```json
{
  "ok": true,
  "data": {
    "hits": [...],
    "total": 100,
    "page": 1
  }
}
```

#### 获取项目详情

**GET** `/api/modrinth/info?slug=optifine`

#### 安装 Modrinth 内容

**POST** `/api/modrinth/install`

请求体：
```json
{
  "slug": "optifine",
  "mc_version": "1.20.1",
  "type": "mod"
}
```

type 可选值：mod, modpack, shader, resourcepack, datapack

---

### 8. 配置管理

#### 获取配置

**GET** `/api/config`

```json
{
  "ok": true,
  "data": {
    "maxMemory": 4096,
    "minMemory": 2048,
    "javaPath": "",
    "jvmArgs": [],
    "gameArgs": [],
    "source": "official",
    "width": 854,
    "height": 480,
    "versionIsolation": false,
    "threads": 8,
    "listenAddr": "127.0.0.1:9876"
  }
}
```

#### 更新配置

**PUT** `/api/config`

请求体：
```json
{
  "maxMemory": 4096,
  "minMemory": 2048,
  "width": 1280,
  "height": 720,
  "threads": 8
}
```

---

### 9. 服务器管理

#### 获取服务器列表

**GET** `/api/servers`

```json
{
  "ok": true,
  "data": [
    {"address": "mc.example.com", "name": "Example Server"}
  ]
}
```

#### 添加服务器

**POST** `/api/servers`

请求体：
```json
{
  "address": "mc.example.com",
  "name": "Example Server"
}
```

#### 删除服务器

**DELETE** `/api/servers/{address}`

---

### 10. 崩溃报告

#### 获取崩溃列表

**GET** `/api/crashes/{version}`

---

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
