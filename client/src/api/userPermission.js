import request from './request'

export function getUserPermissions(params) {
  return request.get('/user-permissions', { params })
}

export function getUserPermission(id) {
  return request.get(`/user-permissions/${id}`)
}

export function createUserPermission(data, dualToken) {
  return request.post('/user-permissions', data, withDual(null, dualToken))
}

export function updateUserPermission(id, data, dualToken) {
  return request.put(`/user-permissions/${id}`, data, withDual(null, dualToken))
}

export function deleteUserPermission(id, dualToken) {
  return request.delete(`/user-permissions/${id}`, withDual(null, dualToken))
}

export function exportDepartmentConfirmation(departmentId) {
  return request.get('/user-permissions/export-confirmation', {
    params: { department_id: departmentId },
    responseType: 'blob'
  })
}

function withDual(config, dualToken) {
  if (dualToken) {
    config = config || {}
    config.headers = { ...(config.headers || {}), 'X-Dual-Control-Token': dualToken }
  }
  return config
}
