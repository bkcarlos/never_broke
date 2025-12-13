import { get, post } from './http'
import type { IncomePlan, CreateIncomePlanPayload } from '@/types'

/**
 * 获取收入计划列表
 */
export async function listIncomePlans(): Promise<IncomePlan[]> {
  try {
    return await get<IncomePlan[]>('/income-plans')
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch income plans')
  }
}

/**
 * 创建收入计划
 */
export async function createIncomePlan(payload: CreateIncomePlanPayload): Promise<IncomePlan> {
  try {
    return await post<IncomePlan>('/income-plans', payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to create income plan')
  }
}


