import request from './request'

export function getPermissionRules() {
  return request.get('/permission-rules')
}

export function createPermissionRule(data) {
  return request.post('/permission-rules', data)
}

export function addSystemToPermissions(data) {
  return request.post('/permission-rules/systems', data)
}

export function updatePermissionRule(id, data) {
  return request.put(`/permission-rules/${id}`, data)
}

export function deletePermissionRule(id) {
  return request.delete(`/permission-rules/${id}`)
}

export function removeSystemFromPermissions(data) {
  return request.delete('/permission-rules/systems', { data })
}

export function renameSystemInPermissions(data) {
  return request.put('/permission-rules/systems/rename', data)
}

export function manageRolesInSystem(data) {
  return request.post('/permission-rules/systems/roles', data)
}

export function reorderPermissionRule(data) {
  return request.post('/permission-rules/reorder', data)
}

export function reorderSystemInPermissions(data) {
  return request.put('/permission-rules/systems/reorder', data)
}
