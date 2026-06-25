import request from './request'

export function getPermissionRules() {
  return request.get('/permission-rules')
}

// 获取特定岗位的权限规则
export function getPositionPermissions(positionName) {
  return request.get('/permission-rules/position', { params: { position_name: positionName } })
}

function withDual(config, dualToken) {
  if (dualToken) {
    config = config || {}
    config.headers = { ...(config.headers || {}), 'X-Dual-Control-Token': dualToken }
  }
  return config
}

export function createPermissionRule(data, dualToken) {
  return request.post('/permission-rules', data, withDual(null, dualToken))
}

export function addSystemToPermissions(data, dualToken) {
  return request.post('/permission-rules/systems', data, withDual(null, dualToken))
}

export function updatePermissionRule(id, data, dualToken) {
  return request.put(`/permission-rules/${id}`, data, withDual(null, dualToken))
}

export function deletePermissionRule(id, dualToken) {
  return request.delete(`/permission-rules/${id}`, withDual(null, dualToken))
}

export function removeSystemFromPermissions(data, dualToken) {
  return request.delete('/permission-rules/systems', withDual({ data }, dualToken))
}

export function renameSystemInPermissions(data, dualToken) {
  return request.put('/permission-rules/systems/rename', data, withDual(null, dualToken))
}

export function manageRolesInSystem(data, dualToken) {
  return request.post('/permission-rules/systems/roles', data, withDual(null, dualToken))
}

export function reorderPermissionRule(data, dualToken) {
  return request.post('/permission-rules/reorder', data, withDual(null, dualToken))
}

export function reorderSystemInPermissions(data, dualToken) {
  return request.put('/permission-rules/systems/reorder', data, withDual(null, dualToken))
}
