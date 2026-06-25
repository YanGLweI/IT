import request from './request'

export function getDepartments() {
  return request.get('/departments')
}

export function createDepartment(data, dualToken) {
  return request.post('/departments', data, withDual(null, dualToken))
}

export function updateDepartment(id, data, dualToken) {
  return request.put(`/departments/${id}`, data, withDual(null, dualToken))
}

export function deleteDepartment(id, dualToken) {
  return request.delete(`/departments/${id}`, withDual(null, dualToken))
}

export function getDepartmentPositions(id) {
  return request.get(`/departments/${id}/positions`)
}

export function addDepartmentPosition(id, data, dualToken) {
  return request.post(`/departments/${id}/positions`, data, withDual(null, dualToken))
}

export function removeDepartmentPosition(id, pid, dualToken) {
  return request.delete(`/departments/${id}/positions/${pid}`, withDual(null, dualToken))
}

function withDual(config, dualToken) {
  if (dualToken) {
    config = config || {}
    config.headers = { ...(config.headers || {}), 'X-Dual-Control-Token': dualToken }
  }
  return config
}
