import request from './request'

// 获取渗透测试报告列表
export function getPenetrationTests(params) {
  return request.get('/penetration-tests', { params })
}

// 创建渗透测试报告
export function createPenetrationTest(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/penetration-tests', formData, config)
}

// 更新渗透测试报告
export function updatePenetrationTest(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/penetration-tests/${id}`, formData, config)
}

// 删除渗透测试报告
export function deletePenetrationTest(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/penetration-tests/${id}`, config)
}

// 获取渗透测试报告预览URL
export function getPenetrationTestPreviewUrl(id) {
  return `/api/penetration-tests/${id}/preview`
}

// 获取渗透测试报告下载URL
export function getPenetrationTestDownloadUrl(id) {
  return `/api/penetration-tests/${id}/download`
}
