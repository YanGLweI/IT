import request from './request'

export function getOSTypes() {
  return request.get('/os-types')
}

export function createOSType(data) {
  return request.post('/os-types', data)
}

export function updateOSType(id, data) {
  return request.put(`/os-types/${id}`, data)
}

export function deleteOSType(id) {
  return request.delete(`/os-types/${id}`)
}
