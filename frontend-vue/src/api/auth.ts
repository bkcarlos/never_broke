// 自定义错误类，包含 HTTP 状态码
export class HttpError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'HttpError'
  }
}

// 生成设备ID（存储在localStorage）
function getOrCreateDeviceId(): string {
  const key = 'device_id'
  let deviceId = localStorage.getItem(key)
  
  if (!deviceId) {
    // 生成UUID v4
    deviceId = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = Math.random() * 16 | 0
      const v = c === 'x' ? r : (r & 0x3 | 0x8)
      return v.toString(16)
    })
    localStorage.setItem(key, deviceId)
  }
  
  return deviceId
}

// 注册请求
export interface RegisterRequest {
  email: string
  password: string
  nickname: string
}

// 登录请求
export interface LoginRequest {
  email: string
  password: string
}

// 认证响应
export interface AuthResponse {
  access_token: string
  expires_in: number
  token_type: string
  user: {
    email: string
    nickname: string
  }
}

// 用户信息
export interface UserInfo {
  email: string
  nickname: string
}

/**
 * 注册
 */
export async function register(payload: RegisterRequest): Promise<AuthResponse> {
  const deviceId = getOrCreateDeviceId()
  
  // 直接发送请求，避免循环依赖
  const response = await fetch('/api/auth/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      ...payload,
      device_id: deviceId
    })
  })
  
  if (!response.ok) {
    const error = await response.json()
    throw new HttpError(response.status, error.message || 'Registration failed')
  }
  
  const result = await response.json()
  const data = result.data || result
  
  // 保存access token
  if (data.access_token) {
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('user_email', data.user.email)
    localStorage.setItem('user_nickname', data.user.nickname)
  }
  
  return data
}

/**
 * 登录
 */
export async function login(payload: LoginRequest): Promise<AuthResponse> {
  const deviceId = getOrCreateDeviceId()
  
  // 直接发送请求，避免循环依赖
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      ...payload,
      device_id: deviceId
    })
  })
  
  if (!response.ok) {
    const error = await response.json()
    throw new HttpError(response.status, error.message || 'Login failed')
  }
  
  const result = await response.json()
  const data = result.data || result
  
  // 保存access token
  if (data.access_token) {
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('user_email', data.user.email)
    localStorage.setItem('user_nickname', data.user.nickname)
  }
  
  return data
}

/**
 * 登出
 */
export async function logout(): Promise<void> {
  const deviceId = getOrCreateDeviceId()
  const token = getAccessToken()
  
  try {
    await fetch('/api/auth/logout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        device_id: deviceId
      })
    })
  } finally {
    // 清除本地存储
    localStorage.removeItem('access_token')
    localStorage.removeItem('user_email')
    localStorage.removeItem('user_nickname')
  }
}

/**
 * 获取当前用户信息
 */
export async function getCurrentUser(): Promise<UserInfo> {
  const token = getAccessToken()
  
  const response = await fetch('/api/me', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  })
  
  if (!response.ok) {
    throw new Error('Failed to get user info')
  }
  
  const result = await response.json()
  return result.data || result
}

/**
 * 检查是否已登录
 */
export function isAuthenticated(): boolean {
  return !!localStorage.getItem('access_token')
}

/**
 * 获取Access Token
 */
export function getAccessToken(): string | null {
  return localStorage.getItem('access_token')
}

/**
 * 获取当前用户信息（从本地存储）
 */
export function getLocalUserInfo(): UserInfo | null {
  const email = localStorage.getItem('user_email')
  const nickname = localStorage.getItem('user_nickname')
  
  if (email && nickname) {
    return { email, nickname }
  }
  
  return null
}
