import { get, request } from './http'
import type { MonthlyBudget } from '@/types'

/**
 * 获取月度预算
 */
export async function getMonthlyBudget(month: string): Promise<MonthlyBudget> {
  try {
    const resp = await request<MonthlyBudget>(`/budget/monthly?month=${encodeURIComponent(month)}`, {
      method: 'GET'
    })
    return resp
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch monthly budget')
  }
}

/**
 * 保存月度预算
 */
export async function saveMonthlyBudget(payload: MonthlyBudget): Promise<MonthlyBudget> {
  try {
    const resp = await request<MonthlyBudget>('/budget/monthly', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    return resp
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to save monthly budget')
  }
}



