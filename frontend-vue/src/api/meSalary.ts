import { request } from './http'
import type {
  RecurringIncomePlan,
  SalaryPlanConfig,
  SalaryPreviewResponse,
} from '@/types'

export interface UpsertSalaryPlanResponse {
  plan: RecurringIncomePlan
  preview: SalaryPreviewResponse
}

export interface GetSalaryPlanResponse {
  plan?: RecurringIncomePlan
}

// 获取当前用户的工资方案
export async function getSalaryPlan(): Promise<GetSalaryPlanResponse> {
  try {
    return await request<GetSalaryPlanResponse>('/me/salary-plan', {
      method: 'GET',
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to load salary plan')
  }
}

// 保存/更新工资方案
export async function upsertSalaryPlan(
  payload: SalaryPlanConfig,
): Promise<UpsertSalaryPlanResponse> {
  try {
    return await request<UpsertSalaryPlanResponse>('/me/salary-plan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to save salary plan')
  }
}

// 删除工资方案
export async function deleteSalaryPlan(planId: string): Promise<void> {
  try {
    await request<null>(`/me/salary-plan/${encodeURIComponent(planId)}`, {
      method: 'DELETE',
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to delete salary plan')
  }
}
