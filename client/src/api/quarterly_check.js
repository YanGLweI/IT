import request from './request'

// 获取季度检查历史列表
export function getQuarterlyChecks(params) {
  return request.get('/quarterly-checks', { params })
}

// 上传季度检查记录
export function createQuarterlyCheck(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/quarterly-checks', formData, config)
}

// 删除季度检查记录
export function deleteQuarterlyCheck(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/quarterly-checks/${id}`, config)
}

// 更新季度检查记录（仅描述）
export function updateQuarterlyCheck(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/quarterly-checks/${id}`, formData, config)
}

// 获取预览URL
export function getQuarterlyCheckPreviewUrl(id) {
  return `/api/quarterly-checks/${id}/preview`
}

// 获取下载URL
export function getQuarterlyCheckDownloadUrl(id) {
  return `/api/quarterly-checks/${id}/download`
}
