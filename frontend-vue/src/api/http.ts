import type { ApiResponse } from '@/types'

const BASE_URL = '/api'

export class HttpError extends Error {
  constructor(
    public code: number,
    message: string
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

export async function request<T>(
  url: string,
  options?: RequestInit
): Promise<T> {
  try {
    console.log(`🌐 请求: ${BASE_URL}${url}`)
    
    // 添加认证头
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options?.headers as Record<string, string> || {})
    }
    
    // 直接从 localStorage 获取 token，避免循环依赖
    const token = localStorage.getItem('access_token')
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    
    const response = await fetch(`${BASE_URL}${url}`, {
      headers,
      ...options
    })

    console.log(`📡 响应状态: ${response.status} ${response.statusText}`)

    if (!response.ok) {
      // 401未授权，跳转到登录页
      if (response.status === 401) {
        localStorage.removeItem('access_token')
        window.location.href = '/login'
      }
      throw new HttpError(
        response.status,
        `HTTP Error: ${response.status} ${response.statusText}`
      )
    }

    const result: ApiResponse<T> = await response.json()
    console.log(`📦 原始响应数据:`, result)

    if (result.code !== 0) {
      throw new HttpError(result.code, result.message)
    }

    console.log(`✨ 解析后的数据:`, result.data)
    return result.data
  } catch (error) {
    console.error(`❌ 请求失败:`, error)
    if (error instanceof HttpError) {
      throw error
    }
    throw new HttpError(-1, error instanceof Error ? error.message : 'Unknown error')
  }
}

export function get<T>(url: string): Promise<T> {
  return request<T>(url, { method: 'GET' })
}

export function post<T>(url: string, data?: unknown): Promise<T> {
  return request<T>(url, {
    method: 'POST',
    body: data ? JSON.stringify(data) : undefined
  })
}

export function patch<T>(url: string, data?: unknown): Promise<T> {
  return request<T>(url, {
    method: 'PATCH',
    body: data ? JSON.stringify(data) : undefined
  })
}

export function del<T>(url: string): Promise<T> {
  return request<T>(url, { method: 'DELETE' })
}

