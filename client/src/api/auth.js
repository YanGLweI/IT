import request from './request'
import axios from 'axios'
import { encryptPassword } from '@/utils/rsa'

// 登录接口不需要 request 实例的拦截器（因为未登录时拦截器会干扰）
export async function login(username, password) {
  const encryptedPassword = await encryptPassword(password)
  return axios.post('/api/login', { username, password: encryptedPassword })
}

export function getUserInfo() {
  return request.get('/user/info')
}
