import { get, post } from './http'
import type { DailyTransactionResponse, CreateTransactionPayload, Transaction } from '@/types'

/**
 * 获取指定日期的账单明细
 * @param date 日期，格式 YYYY-MM-DD
 */
export function getDailyTransactions(date: string): Promise<DailyTransactionResponse> {
  return get<DailyTransactionResponse>(`/transactions/daily?date=${date}`)
}

/**
 * 创建交易（支出/收入/转账）
 */
export function createTransaction(payload: CreateTransactionPayload): Promise<Transaction> {
  return post<Transaction>('/transactions', payload)
}

