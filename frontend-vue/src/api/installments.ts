import { post } from './http'
import type {
  PreviewInstallmentPayload,
  InstallmentPreviewResult,
  CreateExpenseInstallmentPayload,
  CreateExpenseInstallmentResult
} from '@/types'

/**
 * 预览分期
 */
export async function previewInstallment(
  payload: PreviewInstallmentPayload
): Promise<InstallmentPreviewResult> {
  try {
    return await post<InstallmentPreviewResult>('/installments/preview', payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to preview installment')
  }
}

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
