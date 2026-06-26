import request from './request'

// 获取月度检查历史列表
export function getMonthlyChecks(params) {
  return request.get('/monthly-checks', { params })
}

// 上传月度检查记录
export function createMonthlyCheck(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/monthly-checks', formData, config)
}

// 删除月度检查记录
export function deleteMonthlyCheck(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/monthly-checks/${id}`, config)
}

// 更新月度检查记录（仅描述）
export function updateMonthlyCheck(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/monthly-checks/${id}`, formData, config)
}

// 获取预览URL
export function getMonthlyCheckPreviewUrl(id) {
  return `/api/monthly-checks/${id}/preview`
}

// 获取下载URL
export function getMonthlyCheckDownloadUrl(id) {
  return `/api/monthly-checks/${id}/download`
}
