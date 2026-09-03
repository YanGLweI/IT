# 例外管理功能部署说明

## 功能概述

在系统安全模块下新增"例外管理"子模块，用于留档因生产要求必须接入网络但不支持升级补丁的电脑例外授权文件（PDF）。

### 主要特性
- **申请日期**：选择申请授权的日期
- **申请人**：填写申请人姓名
- **例外情况说明**：描述具体原因和审批情况
- **持续时间**：选择授权有效的截止日期
- **扫描件上传**：领导签字的授权单 PDF 文件上传和存档

## 部署步骤

### 1. 后端部署

#### 1.1 数据库迁移
启动服务器后，数据库会自动创建 `exception_managements` 表结构：
```sql
CREATE TABLE exception_managements (
  id int PRIMARY KEY AUTO_INCREMENT,
  created_at datetime DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at datetime DEFAULT NULL,
  
  apply_date varchar(20),
  applicant varchar(50) NOT NULL,
  reason text,
  end_date varchar(20),
  
  file_name varchar(300) NOT NULL,
  file_path varchar(500) NOT NULL,
  file_size bigint,
  file_type varchar(255)
);
```

#### 1.2 配置文件
确保 `server/config.yml` 中包含以下配置：
```yaml
upload:
  exception_management_path: "./uploads/exception_managements"
```

#### 1.3 编译并替换二进制文件
```bash
cd server
go build -o it-server main.go
# 将生成的 it-server 文件复制到生产服务器
scp it-server user@production-server:/path/to/it-platform/server/
# 重启服务
systemctl restart it-platform
```

### 2. 前端部署

#### 2.1 构建前端项目
```bash
cd client
npm install
npm run build
```

#### 2.2 部署到服务器
```bash
# 将 dist 目录下的文件复制到生产服务器
scp -r dist/* user@production-server:/path/to/it-platform/client/dist/
# 或者通过 nginx 配置静态文件路径
```

### 3. 权限验证

该功能使用了双控验证机制，所有写操作（创建、更新、删除）都需要经过双重认证确认。

### 4. 文件存储

上传的 PDF 文件会存储在：
```
server/uploads/exception_managements/
├── [timestamp]_[applicants].pdf
└── ...
```

## API 接口说明

### 4.1 查询列表（不需要双控）
```
GET /api/exception-managements?page=1&page_size=10&keyword=xxx
```

参数：
- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 10）
- `apply_date`: 申请日期筛选
- `end_date`: 结束日期筛选
- `applicant`: 申请人筛选
- `keyword`: 关键词搜索

### 4.2 创建记录（需要双控）
```
POST /api/exception-managements
Content-Type: multipart/form-data

表单字段：
- apply_date: string (必填) - 申请日期 YYYY-MM-DD
- applicant: string (必填) - 申请人
- reason: string (必填) - 例外说明
- endDate: string (必填) - 截止时间 YYYY-MM-DD
- file: file (必填) - PDF 文件
```

### 4.3 更新记录（需要双控）
```
PUT /api/exception-managements/:id
Content-Type: multipart/form-data

表单字段（可选）：
- apply_date: string (必填)
- applicant: string (必填)
- reason: string (必填)
- endDate: string (必填)
- file: file (可选，不传则保持原文件)
```

### 4.4 删除记录（需要双控）
```
DELETE /api/exception-managements/:id
```

### 4.5 预览 PDF（不需要双控）
```
GET /api/exception-managements/:id/preview
```

返回示例：
```json
{
  "code": 200,
  "data": "https://localhost:9080/api/exception-managements/1/download"
}
```

### 4.6 下载 PDF（不需要双控）
```
GET /api/exception-managements/:id/download
```

## 界面说明

访问路径：`http://your-server/#/exception-management`

页面布局：
- **顶部**：标题和操作按钮（上传授权文件、刷新）
- **筛选区**：关键词搜索框
- **数据表格**：
  | 序号 | 申请日期 | 申请人 | 例外说明 | 持续到 | 操作 |
  |------|----------|--------|----------|--------|------|
- **分页**：支持切换每页数量

操作按钮：
- 预览：使用 FileViewer 在线预览 PDF
- 下载：直接下载 PDF 文件
- 编辑：修改记录信息（不能更换文件）
- 删除：删除记录和关联文件

## 测试建议

### 功能测试
1. 创建例外管理记录
   - 上传 PDF 文件
   - 填写完整信息
   - 验证双控验证流程
   
2. 查看列表
   - 验证搜索功能
   - 验证分页功能
   
3. 预览功能
   - 验证 PDF 在线预览
   - 验证文件加载准确性
   
4. 下载功能
   - 验证文件完整性
   - 验证文件名正确性
   
5. 编辑功能
   - 验证信息修改
   - 验证文件不可更换特性
   
6. 删除功能
   - 验证记录删除
   - 验证关联文件删除

### 安全测试
1. 未登录用户无法访问
2. 未经授权的用户无法执行写操作
3. 双控验证流程正常触发
4. 文件上传类型限制（仅 PDF）
5. SQL 注入防护验证

## 后续优化建议

1. 添加批量删除功能
2. 导出 CSV 报表
3. 统计图表展示（按月份、申请人等维度）
4. 邮件提醒功能（授权即将到期时通知）
5. 授权文件 OCR 识别提取关键信息
6. 电子签名集成（支持在线审批流程）

## 常见问题

**Q: 上传 PDF 失败？**
A: 检查文件大小限制，确认文件格式为 PDF，确保有写入 uploads 目录的权限。

**Q: 双控验证不显示？**
A: 检查浏览器控制台是否有 JavaScript 错误，确认 DualControlDialog 组件已正确注册。

**Q: 预览功能无法打开？**
A: 检查 FileViewer 配置，确认 fv-feature.js 中 enabled 设置为 true。

**Q: 中文乱码问题？**
A: 确保数据库连接使用 utf8mb4 编码，确认请求头 Content-Type 包含 charset=utf-8。

## 联系人

如有问题请联系开发团队或提交 Issue。
