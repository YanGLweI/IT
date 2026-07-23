# 专线信息模块 — 设计规范

> 提取自 `DedicatedLineList.vue` / `DedicatedLineDialog.vue`，作为后续模块开发的统一设计参考。

---

## 一、色彩体系

### 功能色

| 用途 | 色值 | 预览 | Hover 态 | 预览 |
|------|------|------|-----------|------|
| 主色（操作、焦点、高亮） | `#3b82f6` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#3b82f6;vertical-align:middle"></span> | `#2563eb` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#2563eb;vertical-align:middle"></span> |
| 危险（删除、错误） | `#ef4444` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#ef4444;vertical-align:middle"></span> | `#dc2626` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#dc2626;vertical-align:middle"></span> |
| 成功/移动运营商 | `#10b981` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#10b981;vertical-align:middle"></span> | — | — |
| 警告/广电运营商 | `#f59e0b` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#f59e0b;vertical-align:middle"></span> | — | — |

### 中性色（文字 & 边框 & 背景）

| 层级 | 色值 | 预览 | 使用场景 |
|------|------|------|----------|
| 文字-标题/主内容 | `#1e293b` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#1e293b;vertical-align:middle"></span> | 页面标题、表格正文、弹窗标题 |
| 文字-次要 | `#64748b` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#64748b;vertical-align:middle"></span> | 副标题、表单 label、按钮默认文字 |
| 文字-辅助/占位 | `#94a3b8` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#94a3b8;vertical-align:middle"></span> | placeholder、空状态、详情 label |
| 文字-备注正文 | `#475569` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#475569;vertical-align:middle"></span> | 备注段落 |
| 边框 | `#e2e8f0` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#e2e8f0;vertical-align:middle;border:1px solid #ccc"></span> | 卡片边框、输入框边框、表格线 |
| 分割线 | `#f1f5f9` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#f1f5f9;vertical-align:middle;border:1px solid #ddd"></span> | 分组下划线、表格行分隔 |
| 背景-卡片/表头 | `#f8fafc` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#f8fafc;vertical-align:middle;border:1px solid #ddd"></span> | 表格容器、表头、偶数行 |
| 背景-页面/主体 | `#ffffff` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#ffffff;vertical-align:middle;border:1px solid #ddd"></span> | 页面容器、弹窗、奇数行 |
| 背景-Hover | `#f1f5f9` | <span style="display:inline-block;width:14px;height:14px;border-radius:3px;background:#f1f5f9;vertical-align:middle;border:1px solid #ddd"></span> | 表格行 hover |

### 运营商徽章配色

| 运营商 | 背景（10% 透明） | 文字 | 预览 |
|--------|------------------|------|------|
| 电信 | `rgba(59,130,246,0.1)` | `#3b82f6` | <span style="display:inline-block;padding:2px 10px;border-radius:20px;font-size:11px;background:rgba(59,130,246,0.1);color:#3b82f6">电信</span> |
| 联通 | `rgba(239,68,68,0.1)` | `#ef4444` | <span style="display:inline-block;padding:2px 10px;border-radius:20px;font-size:11px;background:rgba(239,68,68,0.1);color:#ef4444">联通</span> |
| 移动 | `rgba(16,185,129,0.1)` | `#10b981` | <span style="display:inline-block;padding:2px 10px;border-radius:20px;font-size:11px;background:rgba(16,185,129,0.1);color:#10b981">移动</span> |
| 广电 | `rgba(245,158,11,0.1)` | `#f59e0b` | <span style="display:inline-block;padding:2px 10px;border-radius:20px;font-size:11px;background:rgba(245,158,11,0.1);color:#f59e0b">广电</span> |

> 徽章规则：**10% 透明度同色背景 + 纯色文字**，胶囊形 `border-radius: 20px`。

---

## 二、圆角体系

| 层级 | 数值 | 应用 |
|------|------|------|
| 大容器 | `14px` | 页面卡片、表格容器 |
| 弹窗 | `16px` | el-dialog |
| 按钮/输入框/下拉 | `10px` | 所有可交互控件 |
| 小按钮/缩略图 | `6px ~ 8px` | 行内操作按钮、图片 |
| 徽章/标签 | `20px` | 胶囊形全圆角 |

---

## 三、间距体系

### 页面级

| 位置 | 数值 |
|------|------|
| 页面容器外边距（与 Layout 之间） | `20px` |
| 页面容器内边距 | `24px` |
| 头部与工具栏间距 | `20px` |
| 工具栏与表格间距 | `16px` |
| 表格与分页间距 | `16px` |

### 工具栏内部

| 位置 | 数值 |
|------|------|
| 搜索框与筛选下拉间距 | `16px` |
| 筛选下拉之间间距 | `16px` |

### 表格内部

| 位置 | 数值 |
|------|------|
| 单元格 padding | `12px 14px` |
| 操作按钮之间间距 | `6px` |

### 弹窗/表单内部

| 位置 | 数值 |
|------|------|
| 弹窗 body padding | `20px 24px` |
| 分组标题上方间距 | `20px`（首个为 0） |
| 分组标题下方间距 | `12px` |
| 表单行间距（grid gap） | `16px` |
| 表单行底部 margin | `16px` |
| label 与输入框间距 | `6px` |
| 底部按钮区 padding | `16px 24px` |
| 按钮之间间距 | `10px` |

### 详情弹窗

| 位置 | 数值 |
|------|------|
| 分组之间间距 | `20px` |
| 标题与内容间距 | `10px` |
| 字段网格 gap | `10px 20px`（行 列） |

---

## 四、按钮设计规范

### 主按钮

- 背景 `#3b82f6`，白色文字，无边框
- 圆角 `10px`，padding `9px 18px`
- 字号 `13px`，字重 `500`
- Hover 加深至 `#2563eb`
- 图标 + 文字用 `flex + gap: 6px`
- 小尺寸变体：padding `7px 14px`，字号 `12px`

### 行内操作按钮（表格内）

- 透明背景 + `1px solid #e2e8f0` 边框
- 圆角 `8px`，padding `5px 12px`，字号 `12px`
- 默认文字 `#64748b`
- Hover：边框和文字变为主色 `#3b82f6`
- 危险变体 Hover：变为 `#ef4444`

### 弹窗按钮

- 取消：透明背景 + 边框，hover 边框加深
- 保存：与主按钮一致
- 禁用态：`opacity: 0.6` + `cursor: not-allowed`

---

## 五、字体规范

| 场景 | 字号 | 字重 | 字体 |
|------|------|------|------|
| 页面标题 | `20px` | `600` | 系统字体 |
| 弹窗标题 | `16px` | `600` | 系统字体 |
| 正文/表格内容 | `13px` | `400` | 系统字体 |
| 副标题/label | `13px` / `12px` | `400` / `500` | 系统字体 |
| 表头 | `12px` | `500` | 系统字体 |
| 徽章 | `11px` | `500` | 系统字体 |
| 详情 label | `11px` | `400` | 系统字体 |
| IP/地址等数据 | `12px` | `400` | `'SF Mono', 'Fira Code', 'Consolas', monospace` |

> 等宽字体用于所有网络地址（IP、掩码、网关、DNS），颜色统一为主色 `#3b82f6`。

---

## 六、交互反馈规范

| 交互 | 表现 |
|------|------|
| 输入框 focus | 边框变为 `#3b82f6` |
| 表格行 hover | 背景 `#f1f5f9` + 首列左侧 `inset 3px` 蓝色指示条 |
| 图片 hover | `transform: scale(1.05)` |
| 缩略图删除按钮 | 默认 `opacity: 0`，父元素 hover 时 `opacity: 1` |
| 按钮/输入框 | 统一 `transition: 0.2s` |

---

## 七、布局模式

### 页面结构

```
┌─ 页面容器 (白底, 14px圆角, 1px边框) ─────────────┐
│  [头部] 标题+副标题 ←→ 主按钮                      │
│  [工具栏] 搜索框 + 筛选下拉 (flex, gap 16px)       │
│  [表格卡片] (浅灰底, 14px圆角, 横向可滚动)         │
│  [分页] 右对齐                                     │
└───────────────────────────────────────────────────┘
```

### 表单弹窗

```
┌─ 弹窗 (16px圆角) ────────────────────────────────┐
│  [Header] 标题                                     │
│  [Body] 可滚动区 max-height: 65vh                  │
│    ┌ 分组标题 (下划线分割) ┐                       │
│    │  双列 Grid (1fr 1fr, gap 16px)               │
│    │  全宽行 (1fr)                                │
│    └──────────────────────┘                       │
│  [Footer] 取消 + 保存 (flex, 右对齐)               │
└───────────────────────────────────────────────────┘
```

### 详情弹窗

```
分组标题 → 双列 Grid 字段（label 11px 弱色 + value 13px 主色）
```

---

## 八、关键设计约束

1. **带宽单位**：禁止 `el-input` 的 `slot="append"`，必须用 `flex + 独立 <span>` 显示单位
2. **横向滚动**：表格 `min-width: 900px` + 父容器 `overflow-x: auto`，操作列 `position: sticky; right: 0`
3. **下拉组件**：全部使用 `el-select`，禁止原生 `<select>`
4. **工具栏防重叠**：各控件间距 `16px`，搜索框 `z-index: 1`
5. **弹窗滚动**：body 区域 `max-height: 65vh; overflow-y: auto`
