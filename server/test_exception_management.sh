// 测试例外管理功能的简单脚本
# 后端测试
echo "=== 测试例外管理 API ==="

# 1. 创建授权文件（需要在浏览器中通过双控验证）
curl -X POST "https://localhost:9080/api/exception-managements" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -F "apply_date=2024-01-15" \
  -F "applicant=张三" \
  -F "reason=电脑型号较老，不支持最新补丁更新，但需接入网络进行生产工作，已获领导审批" \
  -F "endDate=2024-12-31" \
  -F "file=@/path/to/授权书.pdf" \
  | jq .

# 2. 查询列表
curl -X GET "https://localhost:9080/api/exception-managements?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  | jq .

# 3. 预览 PDF
curl -X GET "https://localhost:9080/api/exception-managements/1/preview" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  | jq .

# 4. 下载 PDF
curl -X GET "https://localhost:9080/api/exception-managements/1/download" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -o downloaded_authorization.pdf \
  && echo "下载成功：downloaded_authorization.pdf"

# 5. 删除记录
curl -X DELETE "https://localhost:9080/api/exception-managements/1" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  | jq .
