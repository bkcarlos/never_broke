import type { ApiResponse } from '@/types'

const BASE_URL = '/api/v1'

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
    const response = await fetch(`${BASE_URL}${url}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers
      },
      ...options
    })

    if (!response.ok) {
      throw new HttpError(
        response.status,
        `HTTP Error: ${response.status} ${response.statusText}`
      )
    }

    const result: ApiResponse<T> = await response.json()

    if (result.code !== 0) {
      throw new HttpError(result.code, result.message)
    }

    return result.data
  } catch (error) {
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

