import request from './request'

export function getPermissionRules() {
  return request.get('/permission-rules')
}

export function updatePermissionRule(id, data) {
  return request.put(`/permission-rules/${id}`, data)
}
