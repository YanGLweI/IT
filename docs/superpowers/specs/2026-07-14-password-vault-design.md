# 密码本模块需求设计文档

## 1. 模块概述

密码本模块为 IT 合规管理平台提供安全的账号密码集中管理功能。用户可记录各类设备、应用、工具的账号密码，通过 AES-256-GCM 加密存储、LDAP 身份验证 + 逐条授权查看机制保障密码安全。

### 1.1 核心能力

- 密码条目的增删改查（CRUD）
- 分类管理（预设 + 自定义）
- 密码查看授权（创建时指定可查看用户列表）
- 密码查看验证（查看时需输入 LDAP 密码验证身份）
- 密码加密存储（AES-256-GCM）
- 密码生成器（前端随机强密码生成）
- 收藏/置顶功能
- 查看日志审计

---

## 2. 数据模型

### 2.1 密码分类表 `password_categories`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK, auto_increment | 自增主键 |
| name | varchar(50) | NOT NULL, UNIQUE | 分类名称 |
| icon | varchar(50) | NOT NULL | 分类图标标识 |
| is_preset | tinyint(1) | DEFAULT 0 | 是否预设分类（预设不可删除） |
| sort_order | int | DEFAULT 0 | 排序权重（升序） |
| created_at | datetime | | 创建时间 |
| updated_at | datetime | | 更新时间 |

**预设分类**（seed 数据）：

| 名称 | 图标 | sort_order |
|------|------|------------|
| 服务器 | server | 1 |
| 网络设备 | router | 2 |
| 应用系统 | app | 3 |
| 数据库 | database | 4 |
| 管理工具 | tool | 5 |
| 云服务 | cloud | 6 |

### 2.2 密码条目表 `password_entries`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK, auto_increment | 自增主键 |
| category_id | uint | FK → password_categories.id | 所属分类 |
| icon | varchar(50) | NOT NULL | 条目图标（从预设图标库选择） |
| name | varchar(200) | NOT NULL | 条目名称 |
| username | varchar(200) | NOT NULL | 账号/用户名 |
| encrypted_password | text | NOT NULL | AES-256-GCM 加密密文（格式：`iv_hex:ciphertext_hex`） |
| url | varchar(500) | | URL / IP 地址（可空） |
| port | int | | 端口号（可空） |
| notes | text | | 备注（可空） |
| is_starred | tinyint(1) | DEFAULT 0 | 是否收藏/置顶 |
| created_by | varchar(100) | NOT NULL | 创建人（LDAP username） |
| updated_by | varchar(100) | NOT NULL | 最后修改人 |
| created_at | datetime | | 创建时间 |
| updated_at | datetime | | 更新时间 |

### 2.3 密码查看授权表 `password_entry_viewers`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK, auto_increment | 自增主键 |
| entry_id | uint | FK → password_entries.id, NOT NULL | 关联密码条目 |
| username | varchar(100) | NOT NULL | 被授权查看的用户（LDAP username） |

**唯一约束**：`(entry_id, username)` 联合唯一

### 2.4 密码查看日志表 `password_view_logs`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK, auto_increment | 自增主键 |
| entry_id | uint | NOT NULL | 查看的密码条目 ID |
| entry_name | varchar(200) | NOT NULL | 条目名称（冗余，便于审计查询） |
| viewer | varchar(100) | NOT NULL | 查看人（LDAP username） |
| viewed_at | datetime | NOT NULL | 查看时间 |

---

## 3. 安全设计

### 3.1 加密方案

**传输层**：复用现有 RSA 加密机制
- 前端从 `/api/public-key` 获取 RSA 公钥
- 密码字段在前端用 RSA-OAEP/SHA-256 加密后传输
- 后端用 RSA 私钥解密获取明文

**存储层**：AES-256-GCM 对称加密
- AES 密钥配置在 `config.yml` 的 `password_vault.aes_key` 字段
- 密钥为 32 字节（256 位），以 hex 字符串存储
- 每条密码独立生成随机 12 字节 IV（Nonce）
- 存储格式：`iv_hex:ciphertext_hex`（冒号分隔）
- GCM 模式自带认证标签（16 字节），附加在密文后

**密钥管理**：
- 服务启动时从配置文件加载 AES 密钥
- 密钥缺失或长度不合法时服务启动失败并报错
- 密钥变更需手动重新加密所有密码条目（后续可考虑密钥轮换功能）

### 3.2 密码查看验证流程

```
1. 用户点击"查看密码"按钮
2. 前端弹出验证对话框，要求输入当前用户的 LDAP 密码
3. 前端用 RSA 公钥加密 LDAP 密码
4. POST /api/password-entries/:id/unlock
   Body: { ldap_password: "<RSA加密后的LDAP密码>" }
5. 后端处理：
   a. RSA 解密获取用户 LDAP 明文密码
   b. LDAP 认证：验证 username + password 是否匹配
   c. 权限检查：查询 password_entry_viewers 表，确认该 username 在授权列表中
   d. 若全部通过：
      - 从数据库读取 encrypted_password 字段
      - AES-256-GCM 解密，返回密码明文
      - 写入 password_view_logs 审计日志
   e. 若认证失败或无权限：返回 403
6. 前端显示密码明文，提供复制按钮
7. 用户点击"复制" → 复制到剪贴板 → 立即隐藏密码
8. 若用户不操作，30 秒后自动隐藏密码
```

### 3.3 密码创建/编辑流程

```
1. 用户填写表单：名称、账号、密码、分类、URL、端口、备注、图标、授权查看用户
2. 密码字段：前端用 RSA 公钥加密
3. 授权查看用户：通过 el-select 远程搜索 LDAP IT 组用户，支持多选
4. 点击"保存" → 触发双控验证弹窗
5. 双控验证通过后，POST/PUT 提交
6. 后端处理：
   a. RSA 解密密码明文
   b. 生成随机 IV
   c. AES-256-GCM 加密密码
   d. 存储 iv_hex:ciphertext_hex 到 encrypted_password 字段
   e. 保存 password_entry_viewers 授权记录
   f. 记录操作日志
```

### 3.4 路由认证策略

| 操作类型 | 认证要求 | 说明 |
|----------|----------|------|
| 读取（GET） | JWT | 列表查看只需登录态 |
| 写入（POST/PUT/DELETE） | JWT + 双控 | 增删改需双控验证 |
| 密码查看（POST /unlock） | JWT | 独立 LDAP 密码验证，不需要双控 |

---

## 4. API 接口设计

### 4.1 分类管理

#### GET /api/password-categories
获取所有分类列表，返回每个分类及其条目数量。

**响应**：
```json
{
  "code": 200,
  "data": [
    { "id": 1, "name": "服务器", "icon": "server", "is_preset": true, "sort_order": 1, "entry_count": 15 },
    { "id": 7, "name": "自定义分类", "icon": "folder", "is_preset": false, "sort_order": 10, "entry_count": 3 }
  ]
}
```

#### POST /api/password-categories（需双控）
创建自定义分类。

**请求**：`{ "name": "新分类", "icon": "folder" }`

#### PUT /api/password-categories/:id（需双控）
修改分类名称或图标。预设分类可修改图标但不可删除。

#### DELETE /api/password-categories/:id（需双控）
删除自定义分类。预设分类不可删除。分类下有条目时拒绝删除。

### 4.2 密码条目管理

#### GET /api/password-entries
获取密码条目列表（不含密码明文）。

**查询参数**：
- `category_id`：按分类筛选（可选）
- `is_starred`：是否只看收藏（可选）
- `keyword`：搜索名称/账号/URL（可选）
- `page`、`page_size`：分页

**响应**：
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "category_id": 1,
      "category_name": "服务器",
      "icon": "server",
      "name": "生产数据库主库",
      "username": "root",
      "url": "192.168.1.100",
      "port": 3306,
      "notes": "MySQL 8.0",
      "is_starred": true,
      "created_by": "admin",
      "updated_by": "admin",
      "created_at": "2026-07-14T10:00:00+08:00",
      "updated_at": "2026-07-14T10:00:00+08:00",
      "viewers": ["admin", "zhangsan"]
    }
  ],
  "total": 25,
  "page_size": 20
}
```

#### POST /api/password-entries（需双控）
创建密码条目。

**请求**：
```json
{
  "category_id": 1,
  "icon": "server",
  "name": "生产数据库主库",
  "username": "root",
  "encrypted_password": "<RSA加密后的密码>",
  "url": "192.168.1.100",
  "port": 3306,
  "notes": "MySQL 8.0",
  "viewers": ["admin", "zhangsan"]
}
```

#### PUT /api/password-entries/:id（需双控）
修改密码条目。密码字段可选，不传则不更新密码。

#### DELETE /api/password-entries/:id（需双控）
删除密码条目，同时删除关联的 viewers 和 view_logs。

### 4.3 密码查看

#### POST /api/password-entries/:id/unlock
验证身份并查看密码明文。

**请求**：`{ "ldap_password": "<RSA加密后的LDAP密码>" }`

**成功响应**：
```json
{
  "code": 200,
  "data": { "password": "明文密码" }
}
```

**失败响应**：
```json
{ "code": 403, "message": "身份验证失败" }
{ "code": 403, "message": "您无权查看此密码" }
```

### 4.4 收藏管理

#### PUT /api/password-entries/:id/star
切换收藏状态。

**请求**：`{ "is_starred": true }`

### 4.5 审计日志

#### GET /api/password-view-logs
获取密码查看审计日志。

**查询参数**：
- `entry_id`：按条目筛选（可选）
- `viewer`：按查看人筛选（可选）
- `start_date`、`end_date`：时间范围（可选）
- `page`、`page_size`：分页

### 4.6 图标列表

#### GET /api/password-icons
返回预设图标列表（前端也可内置，此接口作为备用）。

---

## 5. 前端页面设计

### 5.1 页面路由

- 路径：`/password-vault`
- 菜单位置：与现有模块平级，在 Layout 下新增路由
- 组件：`views/password-vault/index.vue`

### 5.2 整体布局

左右分栏结构：

**左侧栏（宽度 240px）**：
- 搜索框：快速搜索分类名称
- "全部" 按钮：显示所有条目
- "收藏" 按钮：只显示已收藏条目
- 分割线
- 分类列表：每项显示 `图标 + 名称 + 条目数量角标`，点击切换筛选
- 底部 "添加分类" 按钮

**右侧主区域**：
- 顶部工具栏：搜索框 + 分类筛选下拉 + 排序方式
- 密码条目表格
- 底部分页器

### 5.3 表格设计

| 列 | 宽度 | 说明 |
|----|------|------|
| 图标 | 50px | 条目选择的图标 |
| 名称 | 200px | 条目名称 |
| 账号 | 150px | 部分遮掩显示（如 `ad***in`），旁有复制按钮 |
| 分类 | 120px | 分类标签（el-tag） |
| URL/端口 | 180px | URL:Port 格式 |
| 收藏 | 50px | 星标图标，点击切换 |
| 操作 | 160px | 查看密码、编辑、删除 |

**操作列按钮**：
- 🔒 查看密码：点击弹出验证对话框
- ✏️ 编辑：弹出编辑对话框（需双控）
- 🗑 删除：确认对话框（需双控）

### 5.4 查看密码对话框

- 标题：显示条目名称
- 内容：输入当前用户 LDAP 密码的输入框
- 验证通过后：
  - 显示密码明文（monospace 字体）
  - 提供"复制"按钮
  - 显示倒计时（30 秒）
  - **点击"复制" → 复制到剪贴板 → 立即隐藏密码 → 显示"已复制"提示**
  - 不操作则 30 秒后自动隐藏

### 5.5 新增/编辑对话框

两列布局：

**左列 — 图标选择器**：
- 网格展示所有预设图标
- 点击选中，高亮显示
- 图标分组：服务器类、网络类、应用类、工具类、通用类

**右列 — 表单字段**：
- 名称（必填）
- 账号（必填）
- 密码（必填，带"生成"按钮和密码可见切换）
- 分类（下拉选择，必填）
- URL / IP 地址
- 端口
- 备注（textarea）
- 授权查看用户（el-select 多选，远程搜索 LDAP IT 组用户）

**底部**：保存 / 取消

### 5.6 密码生成器

前端组件，嵌入在密码输入框旁：
- 点击"生成密码"按钮弹出配置面板
- 配置项：
  - 密码长度：滑块，范围 8-64，默认 16
  - 字符类型（复选框）：大写字母、小写字母、数字、特殊符号（至少勾选两项）
- "生成"按钮：生成并填入密码框
- 生成逻辑在前端完成（`crypto.getRandomValues`），不经过后端

### 5.7 审计日志页面

可作为密码本页面的子 Tab 或独立路由 `/password-vault/logs`：
- 表格列：条目名称、查看人、查看时间
- 筛选：时间范围选择器 + 条目下拉 + 查看人下拉
- 分页

### 5.8 预设图标列表

前端内置 SVG 图标组件：

| 分组 | 图标标识 | 描述 |
|------|----------|------|
| 服务器类 | server | 服务器 |
| 服务器类 | database | 数据库 |
| 服务器类 | cloud | 云服务 |
| 网络类 | router | 路由器 |
| 网络类 | switch | 交换机 |
| 网络类 | firewall | 防火墙 |
| 应用类 | app | 应用系统 |
| 应用类 | browser | 网页/浏览器 |
| 应用类 | terminal | 终端/命令行 |
| 工具类 | tool | 工具 |
| 工具类 | key | 密钥/证书 |
| 工具类 | monitor | 监控 |
| 通用类 | desktop | 台式机 |
| 通用类 | laptop | 笔记本 |
| 通用类 | printer | 打印机 |
| 通用类 | wifi | 无线网络 |

---

## 6. 后端文件结构

```
server/
├── models/
│   └── password_vault.go          # 4 个 model 定义
├── handlers/
│   └── password_vault.go          # 所有 handler 函数
├── routes/
│   └── router.go                  # 新增密码本路由组
├── config/
│   └── config.go                  # 新增 PasswordVault 配置段
└── config.yml                     # 新增 password_vault.aes_key 配置
```

## 7. 前端文件结构

```
client/src/
├── api/
│   └── password_vault.js          # API 请求函数
├── views/
│   └── password-vault/
│       ├── index.vue              # 主页面（左右分栏）
│       ├── PasswordEntryDialog.vue # 新增/编辑对话框
│       ├── UnlockDialog.vue       # 密码查看验证对话框
│       ├── CategoryDialog.vue     # 分类管理对话框
│       ├── IconPicker.vue         # 图标选择器组件
│       └── PasswordGenerator.vue  # 密码生成器组件
└── router/
    └── index.js                   # 新增路由
```

---

## 8. 配置项

`config.yml` 新增：

```yaml
password_vault:
  aes_key: "0123456789abcdef0123456789abcdef"  # 32字节hex，64个字符
```

---

## 9. 权限与菜单

- 菜单名称：密码本
- 英文标题：Password Vault
- 路由路径：`/password-vault`
- 菜单图标：`lock`
- 权限：与现有模块一致，通过 JWT 认证控制访问
