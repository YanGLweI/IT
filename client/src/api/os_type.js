import request from './request'

export function getOSTypes() {
  return request.get('/os-types')
}

export function createOSType(data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.post('/os-types', data, config)
}

export function updateOSType(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/os-types/${id}`, data, config)
}

export function deleteOSType(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/os-types/${id}`, config)
}
