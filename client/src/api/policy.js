import request from './request'

export function getPolicies() {
  return request.get('/policies')
}

export function createPolicy(formData) {
  return request.post('/policies', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function updatePolicy(id, data) {
  return request.put(`/policies/${id}`, data)
}

export function replacePolicyFile(id, formData) {
  return request.put(`/policies/${id}/file`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function deletePolicy(id) {
  return request.delete(`/policies/${id}`)
}

export function getPolicyPreviewUrl(id) {
  return `/api/policies/${id}/preview`
}

export function getPolicyDownloadUrl(id) {
  return `/api/policies/${id}/download`
}
