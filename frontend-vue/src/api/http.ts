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
    
    // 创建不包含 headers 的 options，避免覆盖 headers
    const { headers: _ignoredHeaders, ...restOptions } = options || {}
    
    const response = await fetch(`${BASE_URL}${url}`, {
      ...restOptions,
      headers
    })


    if (!response.ok) {
      let message = `请求失败（${response.status}），请稍后重试`
      try {
        const body: { message?: string } = await response.json()
        if (body.message) message = body.message
      } catch {
        // A reverse proxy may return an empty body during a restart.
      }
      // 401未授权，跳转到登录页
      if (response.status === 401) {
        localStorage.removeItem('access_token')
        window.location.href = '/login'
      }
      throw new HttpError(
        response.status,
        message
      )
    }

    const result: ApiResponse<T> = await response.json()

    if (result.code !== 0) {
      throw new HttpError(result.code, result.message)
    }

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

