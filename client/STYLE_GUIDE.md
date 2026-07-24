# 表单发布模块 — 设计规范

> 提取自 `FormVault.vue`（管理端）/ `FormDownload.vue`（公共端），作为后续模块开发的统一设计参考。

---

## 一、色彩体系

### 功能色

| 用途 | 色值 | Hover |
|------|------|-------|
| 主色（操作、焦点、高亮） | `#3b82f6` | `#2563eb` |
| 危险（删除、错误） | `#ef4444` | `#dc2626` |
| 成功（已发布） | `#10b981` | — |
| 警告（取消发布） | `#f59e0b` | — |
| 公共端主色 | `#409EFF` | `#2563EB` |

### 中性色

| 层级 | 色值 | 使用场景 |
|------|------|----------|
| 文字-主 | `#1e293b` | 页面标题、表格正文、弹窗标题 |
| 文字-次 | `#64748b` | 副标题、表单 label、按钮默认文字 |
| 文字-辅 | `#94a3b8` | placeholder、空状态、上传提示 |
| 边框 | `#e2e8f0` | 卡片边框、输入框边框、表格线 |
| 分割线 | `#f1f5f9` | 弹窗 header/footer 分隔、表格行分隔 |
| 背景-表头 | `#f8fafc` | 表格容器、表头 |
| 背景-Hover | `#f1f5f9` | 表格行 hover |

### 状态标签（管理端 el-tag，胶囊形 `border-radius: 20px`）

| 状态 | 背景 | 文字 |
|------|------|------|
| 已发布 | `rgba(16,185,129,0.1)` | `#10b981` |
| 未发布 | `rgba(148,163,184,0.1)` | `#64748b` |
| 来源-上传 | `rgba(59,130,246,0.1)` | `#3b82f6` |
| 来源-静态 | `rgba(245,158,11,0.1)` | `#f59e0b` |
| 来源-动态 | `rgba(239,68,68,0.1)` | `#ef4444` |

> 规则：**10% 透明度同色背景 + 纯色文字**

### 公共端卡片专用色

| 元素 | 背景 | 文字/图标 |
|------|------|-----------|
| 文件图标-PDF | `#FEF2F2` | `#DC2626` |
| 文件图标-DOC | `#EFF6FF` | `#2563EB` |
| 文件图标-XLS | `#F0FDF4` | `#16A34A` |
| 来源标签-文件 | `#EFF6FF` | `#2563EB` |
| 来源标签-模板 | `#F5F3FF` | `#7C3AED` |
| 来源标签-动态 | `#FFF7ED` | `#D97706` |
| 扩展名标签 | `#F1F5F9` | `#64748B`（等宽字体） |

---

## 二、圆角体系

| 数值 | 应用 |
|------|------|
| `16px` | 弹窗 el-dialog |
| `14px` | 页面容器、表格容器 `.table-card` |
| `12px` | 公共端卡片、卡片内文件图标区 |
| `10px` | 按钮、输入框、下拉 |
| `8px` | 表格内行内操作按钮 |
| `6px` | 公共端下载链接 |
| `4px` | 公共端来源标签、扩展名标签 |
| `20px` | 管理端状态/来源 el-tag（胶囊形） |

---

## 三、间距体系

### 管理端（FormVault）

| 位置 | 数值 |
|------|------|
| 页面容器 margin / padding | `20px` / `24px` |
| 页面容器高度 | `calc(100% - 85px)` |
| 头部 → 工具栏 | `20px` |
| 工具栏 → 表格 | `16px` |
| 表格 → 分页 | `16px` |
| 工具栏内部 gap | `16px`，`flex-wrap: wrap`（**已提取为全局 `filter-bar.css`**） |
| 表格单元格 padding | `12px 14px` |
| 操作按钮间距 | `6px` |
| 弹窗 header padding | `20px 24px 16px` |
| 弹窗 body padding / max-height | `20px 24px` / `65vh` |
| 弹窗 footer padding | `16px 24px` |
| 分割线分组间距 | `margin: 20px 0 12px` |

### 公共端（FormDownload）

| 位置 | 数值 |
|------|------|
| 页面 max-width | `1200px` |
| 标题区底部 | `28px` |
| 筛选栏底部 | `28px`，gap `12px` |
| 卡片网格 gap | `20px`，`minmax(280px, 1fr)` |
| 卡片 padding | `20px` |
| 卡片图标尺寸/底部间距 | `56×56` / `14px` |
| 卡片 footer 分隔线上间距 | `margin-top: 16px` + `padding-top: 14px` |
| 空状态 padding | `80px 20px` |

---

## 四、按钮规范

### 管理端主按钮

- 背景 `#3b82f6`，白字，无边框，圆角 `10px`，padding `9px 18px`，`13px/500`
- Hover：`#2563eb`

### 管理端次要按钮

- 透明背景 + `1px solid #e2e8f0`，圆角 `10px`，文字 `#64748b`
- Hover：边框 `#94a3b8`，文字 `#1e293b`

### 筛选栏搜索按钮（`.filter-bar .el-button--primary`）

- 背景 `#3b82f6`，白字，无边框，圆角 `10px`
- Hover：背景 `#2563eb`，白字
- 需在组件 scoped 中显式声明 `color: #fff`（含 hover），防止被全局 `.filter-bar .el-button` 规则覆盖

### 内容区操作按钮（如模板信息区、卡片内按钮）

- 统一圆角 `10px`（覆盖 Element UI 默认的 `3px`/`4px`）
- 主按钮：背景 `#3b82f6`，白字
- 次要按钮：透明 + 边框 `#e2e8f0`，文字 `#64748b`

### 表格行内按钮（全局 `.table-card` 控制）

- 透明背景 + `1px solid #e2e8f0`，圆角 `8px`，padding `5px 12px`，`12px`
- 默认文字 `#64748b`，Hover 变主色
- **重要操作默认态突出**：success/warning/danger 使用 10% 透明背景 + 30% 透明边框 + 纯色文字，Hover 变纯色白字

### 弹窗底部按钮

- 取消：透明 + 边框 `#e2e8f0`，hover 加深
- 确认：与主按钮一致
- 禁用：`opacity: 0.6` + `cursor: not-allowed`

### 公共端下载按钮

- 背景 `#EFF6FF`，文字 `#409EFF`，圆角 `6px`，padding `6px 14px`
- Hover：背景 `#DBEAFE`，文字 `#2563EB`

---

## 五、字体规范

| 场景 | 字号 | 字重 |
|------|------|------|
| 管理端页面标题 | `20px` | `600` |
| 公共端页面标题 | `24px` | `700` |
| 弹窗标题 | `16px` | `600` |
| 公共端卡片标题 | `15px` | `600` |
| 正文/表格 | `13px` | `400` |
| 公共端描述 | `13px`~`14px` | `400`，行高 `1.5` |
| 表单 label | `12px` | `500` |
| 表头 | `12px` | `500` |
| 标签/徽章 | `11px` | `500` |
| 公共端扩展名 | `11px` | `500`，等宽字体 |

---

## 六、交互反馈

| 交互 | 表现 |
|------|------|
| 输入框 focus | 边框 → `#3b82f6` |
| 表格行 hover | 背景 `#f1f5f9` + 首列 `inset 3px` 蓝色指示条 |
| 表格选中行 | 背景 `rgba(59,130,246,0.06)` |
| 上传拖拽区 hover | 边框 → `#3b82f6` |
| 过渡动画 | 统一 `transition: 0.2s` |
| 公共端卡片 hover | `translateY(-3px)` + `box-shadow: 0 8px 24px rgba(0,0,0,0.08), 0 2px 8px rgba(64,158,255,0.06)` + 边框 `#BFDBFE` |
| 公共端卡片入场 | `cardFadeInUp`（`translateY(20px)` → `0`），每张延迟 `0.06s` |

---

## 七、布局模式

### 管理端

```
┌─ 页面容器 (白底, 14px圆角, 1px边框, margin 20px) ─────────┐
│  [头部] 标题+副标题 ←→ 按钮组 (flex, gap 10px)              │
│  [工具栏] 搜索框 + 筛选下拉 (flex, gap 16px, wrap)          │
│  [表格] .table-card > .table-wrapper > el-table              │
│  [分页] .pagination-wrap 右对齐                              │
└──────────────────────────────────────────────────────────────┘
```

### 管理端表格高度规范（重要）

**核心原则：页面不滚动，表格内容撑开 + 动态最大高度 + 表头固定 + 内部滚动**

#### 行为要求

| 场景 | 行为 |
|------|------|
| 数据少（如 10 行） | 表格高度由内容自然撑开，无滚动条 |
| 数据多（超出可用空间） | 表格 body 内部出现垂直滚动条，表头固定 |
| 窗口缩小 | `tableMaxHeight` 自动减小，表格收缩，body 内部滚动 |
| 页面级 | 容器 `overflow: hidden`，禁止页面滚动 |

#### 实现方式（全局 mixin + 全局 CSS，无需组件内重复）

1. **引入 mixin**：

```js
import tableHeightMixin from '@/mixins/table-height'
export default {
  mixins: [tableHeightMixin]
}
```

2. **模板结构**：

```html
<div class="table-card" ref="tableCard">
  <div class="table-wrapper">
    <el-table :data="list" :max-height="tableMaxHeight">...</el-table>
  </div>
</div>
```

3. **fetchData 的 `finally` 中调用**：

```js
finally {
  this.loading = false
  this.$nextTick(() => this.calcTableHeight())
}
```

4. **全局 CSS 已内置**（`styles/table-theme.css`），无需组件内写任何表格高度相关样式

> mixin 自动处理：`tableMaxHeight` 数据、`calcTableHeight` 计算、`resize` 监听、`beforeDestroy` 清理

#### 常见错误

| 错误 | 后果 | 正确做法 |
|------|------|----------|
| 使用 `container.clientHeight` 不减 padding | tableMaxHeight 偏大，底部被裁剪 | 减去 `paddingTop + paddingBottom` |
| 只减元素 `offsetHeight` 不含 margin | 计算偏大，表格溢出 | 加上对应的 margin 值 |
| 用固定值（如 600px）不做动态计算 | 窗口缩小时表格溢出不可见 | 必须动态计算 + resize 监听 |
| 不用 el-table `:max-height`，只靠外部 CSS | 表头无法固定 | 必须用 `:max-height` 属性 |
| `.table-card` 使用 `flex: 1` | 卡片撑满容器，表格无法收缩 | 不使用 flex: 1 |

### 管理端弹窗（class="vault-dialog"）

```
┌─ 弹窗 (16px圆角) ──────────────────────────────────────────┐
│  [Header] 标题 (下划线 #f1f5f9)                              │
│  [Body] 可滚动 max-height: 65vh                              │
│    表单 / 上传拖拽区 / 弹窗内表格 / 分割线分组               │
│  [Footer] 取消 + 确认 (上划线 #f1f5f9)                       │
└──────────────────────────────────────────────────────────────┘
```

### 预览类弹窗（class="preview-dialog"）

```
─ 弹窗 (16px圆角, 更高, 便于查看文件内容) ────────────────────┐
│  [Header] 标题 (下划线 #f1f5f9)                              │
│  [Body] 可滚动 max-height: 82vh (普通弹窗 65vh)              │
│    iframe / 图片预览 / docx-preview                          │
│  [Footer] 下载按钮                                           │
└──────────────────────────────────────────────────────────────
```

> 用法：`el-dialog` 同时加 `class="vault-dialog preview-dialog"`，与 `vault-dialog` 叠加使用

### 公共端

```
┌─ 页面 (max-width: 1200px) ─────────────────────────────────┐
│  [标题] SVG图标 + 标题 (24px/700) + 描述                     │
│  [筛选] 搜索框(flex:1) + 分类下拉(160px)                     │
│  [卡片网格] grid auto-fill minmax(280px,1fr) gap 20px        │
│    卡片: 图标 → 标题 → 描述(2行省略) → meta标签 → 下载按钮   │
│  [空状态] 图标 + 文字                                        │
└──────────────────────────────────────────────────────────────┘
```

---

## 八、关键约束

1. 弹窗统一 `class="vault-dialog"` 覆盖全局样式；预览类弹窗额外加 `class="preview-dialog"`（body max-height 82vh，更大可视区域）
2. 表格必须 `.table-card > .table-wrapper > el-table` 三层包裹，分页必须 `.pagination-wrap` 包裹
3. **表格高度必须动态计算**（`calcTableHeight`），禁止使用固定 max-height 值；必须监听 `resize` 事件；el-table 必须使用 `:max-height` 属性保证表头固定
4. 筛选栏使用全局 `.filter-bar`（`styles/filter-bar.css`），输入框 `36px` 高、`10px` 圆角、focus 蓝色边框，无需在组件内重复定义
5. 操作列 `fixed="right"` 实现 sticky
6. 所有写操作需通过 `DualControlDialog` 双控验证
7. 文件预览：PDF(iframe)、DOCX(docx-preview)、XLSX(xlsx→html)、图片(img)
8. 公共端 `@media (max-width: 640px)` 卡片单列，标题缩至 `20px`
