import request from './request'

export function getDashboardSummary() {
  return request.get('/dashboard/summary')
}
