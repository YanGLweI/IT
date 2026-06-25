import request from './request'

export function getPolicies() {
  return request.get('/policies')
}

export function createPolicy(formData) {
  return request.post('/policies', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function updatePolicy(id, data, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.put(`/policies/${id}`, data, config)
}

export function replacePolicyFile(id, formData, dualToken) {
  const headers = { 'Content-Type': 'multipart/form-data' }
  if (dualToken) headers['X-Dual-Control-Token'] = dualToken
  return request.put(`/policies/${id}/file`, formData, { headers })
}

export function deletePolicy(id, dualToken) {
  const config = {}
  if (dualToken) config.headers = { 'X-Dual-Control-Token': dualToken }
  return request.delete(`/policies/${id}`, config)
}

export function getPolicyPreviewUrl(id) {
  return `/api/policies/${id}/preview`
}

export function getPolicyDownloadUrl(id) {
  return `/api/policies/${id}/download`
}
