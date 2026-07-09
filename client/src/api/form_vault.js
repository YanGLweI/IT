import request from './request'

// 保管区列表
export function getFormVaultItems(params) {
  return request.get('/form-vault', { params })
}

// 上传表单文件
export function uploadFormVaultItem(formData, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post('/form-vault', formData, config)
}

// 编辑表单
export function updateFormVaultItem(id, data, dualToken) {
  const config = { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.put(`/form-vault/${id}`, data, config)
}

// 删除表单
export function deleteFormVaultItem(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/form-vault/${id}`, config)
}

// 发布
export function publishFormVaultItem(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/form-vault/${id}/publish`, {}, config)
}

// 取消发布
export function unpublishFormVaultItem(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/form-vault/${id}/unpublish`, {}, config)
}

// 跨模块引用源列表
export function getCrossModuleSources() {
  return request.get('/form-vault/cross-module-sources')
}

// 模块内文件列表
export function getCrossModuleFiles(moduleKey) {
  return request.get(`/form-vault/cross-module-sources/${moduleKey}/files`)
}

// 创建跨模块引用
export function createCrossModuleRef(data, dualToken) {
  const config = { headers: { 'Content-Type': 'multipart/form-data' } }
  if (dualToken) config.headers['X-Dual-Control-Token'] = dualToken
  return request.post('/form-vault/cross-module', data, config)
}

// 预览/下载 URL
export function getFormVaultPreviewUrl(id) {
  return `/api/form-vault/${id}/preview`
}

export function getFormVaultDownloadUrl(id) {
  return `/api/form-vault/${id}/download`
}
