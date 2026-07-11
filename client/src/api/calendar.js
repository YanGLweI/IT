import request from './request'

export function getCalendars(params) {
  return request.get('/calendars', { params })
}

export function getCalendar(id) {
  return request.get(`/calendars/${id}`)
}

export function createCalendar(data) {
  return request.post('/calendars', data)
}

export function updateCalendar(id, data) {
  return request.put(`/calendars/${id}`, data)
}

export function deleteCalendar(id) {
  return request.delete(`/calendars/${id}`)
}

export function getLDAPUsers() {
  return request.get('/ldap/users')
}

export function getTodayNotifications() {
  return request.get('/calendars/today-notifications')
}

export function getUnreadCount() {
  return request.get('/calendars/unread-count')
}

export function getPendingNotifications() {
  return request.get('/calendars/pending-notifications')
}

export function markNotificationRead(id) {
  return request.put(`/calendars/notifications/${id}/read`)
}

export function markNotificationPopupShown(id) {
  return request.put(`/calendars/notifications/${id}/popup-shown`)
}

export function checkConflict(data) {
  return request.post('/calendars/check-conflict', data)
}
