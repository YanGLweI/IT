import request from './request'

// ============================================================
// 变更类型管理
// ============================================================

// 获取变更类型列表
export function getChangeTypes() {
  return request.get('/change-types')
}

// 创建变更类型
export function createChangeType(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/change-types', data, config)
}

// 更新变更类型
export function updateChangeType(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/change-types/${id}`, data, config)
}

// 删除变更类型
export function deleteChangeType(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/change-types/${id}`, config)
}

// ============================================================
// 模板管理
// ============================================================

// 获取模板历史版本列表
export function getChangeRecordTemplates() {
  return request.get('/change-record-templates')
}

// 获取当前版本模板
export function getCurrentChangeRecordTemplate() {
  return request.get('/change-record-templates/current')
}

// 上传新版本模板
export function uploadChangeRecordTemplate(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/change-record-templates', formData, config)
}

// 删除历史版本模板
export function deleteChangeRecordTemplate(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/change-record-templates/${id}`, config)
}

// 获取模板下载URL
export function getChangeRecordTemplateDownloadUrl(id) {
  return `/api/change-record-templates/${id}/download`
}

// 获取模板预览URL
export function getChangeRecordTemplatePreviewUrl(id) {
  return `/api/change-record-templates/${id}/preview`
}

// ============================================================
// 扫描件存档
// ============================================================

// 获取变更记录扫描件列表
export function getChangeRecords(params) {
  return request.get('/change-records', { params })
}

// 上传变更记录扫描件
export function createChangeRecord(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/change-records', formData, config)
}

// 更新变更记录扫描件
export function updateChangeRecord(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/change-records/${id}`, formData, config)
}

// 删除变更记录扫描件
export function deleteChangeRecord(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/change-records/${id}`, config)
}

// 获取扫描件预览URL
export function getChangeRecordPreviewUrl(id) {
  return `/api/change-records/${id}/preview`
}

// 获取扫描件下载URL
export function getChangeRecordDownloadUrl(id) {
  return `/api/change-records/${id}/download`
}
