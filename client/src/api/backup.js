import request from './request'

// 获取备份记录列表
export function getBackups(params) {
  return request.get('/backups', { params })
}

// 创建备份记录
export function createBackup(data, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.post('/backups', data, config)
}

// 更新备份记录
export function updateBackup(id, data, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.put(`/backups/${id}`, data, config)
}

// 删除备份记录
export function deleteBackup(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.delete(`/backups/${id}`, config)
}

// 获取备份申请表预览URL
export function getBackupPreviewUrl(id) {
  return `/api/backups/${id}/preview`
}

// 获取备份申请表下载URL
export function getBackupDownloadUrl(id) {
  return `/api/backups/${id}/download`
}

// === 恢复还原记录 ===

// 创建恢复还原记录
export function createBackupRecovery(backupId, data, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.post(`/backups/${backupId}/recoveries`, data, config)
}

// 更新恢复还原记录
export function updateBackupRecovery(id, data, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.put(`/backup-recoveries/${id}`, data, config)
}

// 删除恢复还原记录
export function deleteBackupRecovery(id, dualToken) {
  const config = {}
  if (dualToken) {
    config.headers = { 'X-Dual-Control-Token': dualToken }
  }
  return request.delete(`/backup-recoveries/${id}`, config)
}

// 获取恢复还原记录预览URL
export function getBackupRecoveryPreviewUrl(id) {
  return `/api/backup-recoveries/${id}/preview`
}

// 获取恢复还原记录下载URL
export function getBackupRecoveryDownloadUrl(id) {
  return `/api/backup-recoveries/${id}/download`
}
