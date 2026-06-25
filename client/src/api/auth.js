import request from './request'
import axios from 'axios'

// 登录接口不需要 request 实例的拦截器（因为未登录时拦截器会干扰）
export function login(username, password) {
  return axios.post('/api/login', { username, password })
}

export function getUserInfo() {
  return request.get('/user/info')
}
