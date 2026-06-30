import request from './request'

// 获取系统加固检查记录列表
export function getSystemHardeningHistories(params) {
  return request.get('/system-hardening', { params })
}

// 导出系统加固检查表
export function getExportChecklistUrl() {
  return '/api/system-hardening/export-checklist'
}

// 上传系统加固检查记录
export function createSystemHardeningHistory(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/system-hardening', formData, config)
}

// 更新系统加固检查记录
export function updateSystemHardeningHistory(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/system-hardening/${id}`, formData, config)
}

// 删除系统加固检查记录
export function deleteSystemHardeningHistory(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/system-hardening/${id}`, config)
}

// 获取预览URL
export function getSystemHardeningPreviewUrl(id) {
  return `/api/system-hardening/${id}/preview`
}

// 获取下载URL
export function getSystemHardeningDownloadUrl(id) {
  return `/api/system-hardening/${id}/download`
}
