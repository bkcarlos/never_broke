import { request } from './http'
import type {
  RecurringIncomePlan,
  UpsertRecurringIncomePayload,
  RecurringIncomeRealizePayload,
  RecurringIncomeRealizeResult,
} from '@/types'

export async function listRecurringIncomePlans(): Promise<RecurringIncomePlan[]> {
  try {
    return await request<RecurringIncomePlan[]>('/recurring-incomes', {
      method: 'GET',
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch recurring incomes')
  }
}

export async function upsertRecurringIncomePlan(
  payload: UpsertRecurringIncomePayload,
): Promise<RecurringIncomePlan> {
  try {
    return await request<RecurringIncomePlan>('/recurring-incomes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to save recurring income plan')
  }
}

export async function listDueRecurringIncomePlans(
  date: string,
): Promise<RecurringIncomePlan[]> {
  try {
    return await request<RecurringIncomePlan[]>(
      `/recurring-incomes/due?date=${encodeURIComponent(date)}`,
      { method: 'GET' },
    )
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch due recurring incomes')
  }
}

export async function realizeRecurringIncome(
  id: string,
  payload: RecurringIncomeRealizePayload,
): Promise<RecurringIncomeRealizeResult> {
  try {
    return await request<RecurringIncomeRealizeResult>(
      `/recurring-incomes/${id}/realize`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      },
    )
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to realize recurring income')
  }
}
