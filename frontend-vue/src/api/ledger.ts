import { get, post } from './http'
import type { DailyTransactionResponse, CreateTransactionPayload, Transaction } from '@/types'

/**
 * 获取指定日期的账单明细
 * @param date 日期，格式 YYYY-MM-DD
 */
export async function getDailyTransactions(date: string, currency = 'CNY'): Promise<DailyTransactionResponse> {
  const result = await get<Omit<DailyTransactionResponse, 'list'> & {
    list?: DailyTransactionResponse['list'] | null
  }>(`/transactions/daily?date=${date}&currency=${currency}`)
  // Older servers serialize an empty Go slice as null. Keep the UI contract
  // stable during deployment and when displaying dates without transactions.
  return { ...result, list: result.list ?? [] }
}

/**
 * 创建交易（支出/收入/转账）
 */
export async function createTransaction(payload: CreateTransactionPayload): Promise<void> {
  try {
    await post<Transaction>('/transactions', payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to create transaction')
  }
}

