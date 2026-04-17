import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const service = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 30000
})

service.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = token
    }
    if (!config.headers['Content-Type'] && config.method !== 'GET') {
      config.headers['Content-Type'] = 'application/json'
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

service.interceptors.response.use(
  response => {
    const res = response.data
    
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.msg || '请求失败')
      
      if (res.code === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        localStorage.removeItem('isLoggedIn')
        router.push('/login')
      }
      
      return Promise.reject(res)
    }
    
    return res
  },
  error => {
    const msg = error.response?.data?.msg || error.message || '网络错误'
    ElMessage.error(msg)
    
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('isLoggedIn')
      router.push('/login')
    }
    
    return Promise.reject(error)
  }
)

export default service