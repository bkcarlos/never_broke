import { post } from './http'
import type { CreateExpenseInstallmentPayload, CreateExpenseInstallmentResult } from '@/types'

/**
 * 创建支出分期
 */
export async function createExpenseInstallment(
  payload: CreateExpenseInstallmentPayload
): Promise<CreateExpenseInstallmentResult> {
  try {
    return await post<CreateExpenseInstallmentResult>('/installments/expense', payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to create installment expense')
  }
}
