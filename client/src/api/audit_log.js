import request from './request'

export function getLoginLogs(params) {
  return request.get('/login-logs', { params })
}

export function getOperationLogs(params) {
  return request.get('/operation-logs', { params })
}

export function getOperationLogDetails(id) {
  return request.get(`/operation-logs/${id}/details`)
}

export function logout() {
  return request.post('/logout')
}
