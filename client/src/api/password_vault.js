import request from './request'

// ============ 分类管理 ============

// 获取分类列表
export function getPasswordCategories() {
  return request.get('/password-categories')
}

// 创建分类
export function createPasswordCategory(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/password-categories', data, config)
}

// 更新分类
export function updatePasswordCategory(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/password-categories/${id}`, data, config)
}

// 删除分类
export function deletePasswordCategory(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/password-categories/${id}`, config)
}

// 调整分类排序
export function sortPasswordCategory(id, direction, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/password-categories/${id}/sort`, { direction }, config)
}

// ============ 密码条目管理 ============

// 获取密码条目列表
export function getPasswordEntries(params) {
  return request.get('/password-entries', { params })
}

// 创建密码条目
export function createPasswordEntry(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/password-entries', data, config)
}

// 更新密码条目
export function updatePasswordEntry(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/password-entries/${id}`, data, config)
}

// 删除密码条目
export function deletePasswordEntry(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/password-entries/${id}`, config)
}

// 切换收藏状态
export function togglePasswordEntryStar(id, isStarred, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/password-entries/${id}/star`, { is_starred: isStarred }, config)
}

// ============ 密码查看 ============

// 验证并查看密码
export function unlockPasswordEntry(id, ldapPasswordEncrypted) {
  return request.post(`/password-entries/${id}/unlock`, {
    ldap_password: ldapPasswordEncrypted
  })
}

// ============ 审计日志 ============

// 获取查看日志
export function getPasswordViewLogs(params) {
  return request.get('/password-view-logs', { params })
}
