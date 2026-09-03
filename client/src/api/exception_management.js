import request from './request'

// ============================================================
// 例外管理
// ============================================================

// 获取例外管理列表
export function getExceptionManagementList(params = {}) {
  return request.get('/exception-managements', { params })
}

// 创建例外管理记录（上传 PDF）
export function createExceptionManagement(formData, dualToken) {
  const config = {
    headers: { 
      'Content-Type': 'multipart/form-data'
    }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.post('/exception-managements', formData, config)
}

// 更新例外管理记录（替换 PDF）
export function updateExceptionManagement(id, formData, dualToken) {
  const config = {
    headers: { 
      'Content-Type': 'multipart/form-data'
    }
  }
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.put(`/exception-managements/${id}`, formData, config)
}

// 删除例外管理记录
export function deleteExceptionManagement(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers['X-Dual-Control-Token'] = dualToken
  }
  return request.delete(`/exception-managements/${id}`, config)
}

// 预览 PDF
export function previewExceptionManagement(id) {
  return request.get(`/exception-managements/${id}/preview`)
}

// 获取预览 URL
export function previewExceptionManagementUrl(id) {
  return `/api/exception-managements/${id}/preview`
}

// 下载 PDF URL
export function downloadExceptionManagementUrl(id) {
  return `/api/exception-managements/${id}/download`
}
