<p align="center">
  <img src="./assets/readme/hero.svg" width="100%" alt="IT 运维管理平台 — 资产管理 · 安全合规 · 权限治理 · 日程协作">
</p>

一站式企业 IT 运维管理平台。将资产台账、安全合规、权限治理、变更审批和团队协作整合到同一个系统中，通过 LDAP 统一认证和双控审批机制确保每一次关键操作可追溯、可控制。

## 核心能力

| 模块 | 说明 |
|------|------|
| 资产管理 | IT 资产全生命周期管理，关联区域 / OS 类型，数据看板汇总关键指标 |
| 安全合规 | 漏洞扫描、渗透测试、防火墙检查、系统加固、补丁更新、备份管理，形成闭环台账 |
| 权限治理 | LDAP 域控认证、岗位权限配置、月度 / 季度检查、用户变更审计 |
| 变更管理 | 变更创建 / 审批 / 归档，自定义变更类型，模板与扫描件管理 |
| 双控审批 | 关键写操作需第二人验证，防止单人误操作 |
| 协作与知识 | 多视图日历、步骤式 IT 指南（图文 / 视频）、表单发布、AES 加密密码本 |
| 审计日志 | 登录日志 + 操作日志，全链路可追溯 |

## 安全机制

- **JWT 双 Token** — Access Token (短期) + Refresh Token (长期)，无感刷新
- **RSA 加密传输** — 登录密码公钥加密，后端私钥解密
- **双控审批** — 关键写操作需第二人验证
- **AES 密码本** — 敏感密码加密存储，按需解锁，查看留痕
- **HTTPS** — 全站 TLS 加密通信
- **LDAP 集成** — 企业 AD 域控单点登录

## 快速开始

**环境要求：** Go 1.25+ · Node.js + npm · MariaDB 11.8+

```bash
# 1. 后端
cd server
vim config.yml          # 配置数据库、LDAP、TLS
go mod download
go build -o it-server . && ./it-server

# 2. 前端
cd client
npm install
npm run serve           # 开发模式
npm run build           # 生产构建 → client/dist/
```

**部署：** 后端交叉编译为 Linux 二进制 `GOOS=linux GOARCH=amd64 go build -o it-server-linux .`，前端产物部署至静态服务器或与后端合并。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 2.7 · Element UI · Vue Router · Axios · ECharts · anime.js |
| 后端 | Go 1.25 · Gin · GORM · JWT · LDAP · RSA/AES |
| 数据库 | MySQL / MariaDB |
| 部署 | HTTPS (TLS) · 交叉编译 Linux 二进制 |

## 项目结构

```
IT/
├── client/                  # 前端 (Vue 2)
│   ├── src/
│   │   ├── api/             # API 请求模块
│   │   ├── components/      # 公共组件
│   │   ├── router/          # 路由配置
│   │   ├── styles/          # 全局样式
│   │   └── views/           # 页面视图
│   └── public/              # 静态资源
├── server/                  # 后端 (Go/Gin)
│   ├── config/              # 配置加载
│   ├── database/            # 数据库初始化 & 种子数据
│   ├── handlers/            # HTTP 处理器
│   ├── middleware/          # 中间件 (CORS / JWT / 双控)
│   ├── models/              # 数据模型
│   ├── routes/              # 路由注册
│   ├── services/            # 业务服务层
│   ├── uploads/             # 文件上传存储
│   └── main.go              # 入口
└── docs/                    # 设计文档
```

## 功能清单

<details>
<summary>基础设施管理</summary>

- 区域管理 — 数据中心 / 办公区域
- 操作系统管理 — OS 类型字典维护
- 资产管理 — CRUD + 关联区域 / OS 类型
- 数据看板 — 资产、安全、权限关键指标汇总
</details>

<details>
<summary>安全合规</summary>

- 漏洞扫描 — 扫描记录、整改报告上传、闭环追踪
- 渗透测试 — 报告上传与管理
- 防火墙检查 — 检查记录与整改报告
- 系统加固 — 加固检查表导出与历史记录
- 补丁更新 — 补丁记录与合规性报表
- 备份管理 — 备份申请、恢复还原、模板管理
</details>

<details>
<summary>身份与权限治理</summary>

- LDAP 统一认证 — 企业 AD 域控单点登录
- 双 Token 无感刷新 — Access + Refresh Token
- 双控审批 — 关键写操作第二人验证
- 岗位权限设置 — 基于岗位的权限规则配置
- 用户权限一览 — 全量权限查看与导出
- 月度 / 季度检查历史 — 权限定期确认
- 用户变更记录 — 权限变更审计追踪
</details>

<details>
<summary>变更管理</summary>

- 变更记录创建 / 审批 / 归档
- 变更类型自定义与排序
- 变更模板管理（上传 / 下载 / 预览）
- 变更记录扫描件管理
</details>

<details>
<summary>协作与知识</summary>

- 日程管理 — 月 / 周 / 日多视图、重复日程、冲突检测
- IT 指南 — 步骤式操作指南，图文 / 视频，公开访问
- 表单发布 — 上传 / 发布 / 公开下载，跨模块引用
- 密码本 — AES 加密、分类管理、收藏排序、查看审计
</details>

<details>
<summary>其他</summary>

- SFTP 服务器 / 账号管理
- 核准软件目录与资产对应表
- IT 政策文档管理
- 网络拓扑图管理
- 审计日志（登录 + 操作）
</details>

---

Private — Internal Use Only
