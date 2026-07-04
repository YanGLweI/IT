import request from './request'

// 获取补丁更新记录列表
export function getPatchUpdates(params) {
  return request.get('/patch-updates', { params })
}

// 上传合规性报表
export function createPatchUpdate(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/patch-updates', formData, config)
}

// 更新补丁更新记录
export function updatePatchUpdate(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/patch-updates/${id}`, formData, config)
}

// 删除补丁更新记录
export function deletePatchUpdate(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.delete(`/patch-updates/${id}`, config)
}

// 上传修复报表
export function uploadPatchFixReport(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/patch-updates/${id}/fix`, formData, config)
}

// 删除修复报表
export function deletePatchFixReport(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.delete(`/patch-updates/${id}/fix`, config)
}

// 获取合规性报表预览URL
export function getPatchUpdatePreviewUrl(id) {
  return `/api/patch-updates/${id}/preview`
}

// 获取合规性报表下载URL
export function getPatchUpdateDownloadUrl(id) {
  return `/api/patch-updates/${id}/download`
}

// 获取修复报表预览URL
export function getPatchFixPreviewUrl(id) {
  return `/api/patch-updates/${id}/fix-preview`
}

// 获取修复报表下载URL
export function getPatchFixDownloadUrl(id) {
  return `/api/patch-updates/${id}/fix-download`
}
