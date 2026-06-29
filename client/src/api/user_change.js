import request from './request'

// 获取用户变更记录历史列表
export function getUserChangeHistories(params) {
  return request.get('/user-change-histories', { params })
}

// 上传用户变更记录
export function createUserChangeHistory(formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/user-change-histories', formData, config)
}

// 删除用户变更记录
export function deleteUserChangeHistory(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/user-change-histories/${id}`, config)
}

// 更新用户变更记录（仅描述）
export function updateUserChangeHistory(id, formData, dualToken) {
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/user-change-histories/${id}`, formData, config)
}

// 获取预览URL
export function getUserChangePreviewUrl(id) {
  return `/api/user-change-histories/${id}/preview`
}

// 获取下载URL
export function getUserChangeDownloadUrl(id) {
  return `/api/user-change-histories/${id}/download`
}
