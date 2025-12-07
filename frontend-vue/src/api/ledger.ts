import { get } from './http'
import type { DailyTransactionResponse } from '@/types'

/**
 * 获取指定日期的账单明细
 * @param date 日期，格式 YYYY-MM-DD
 */
export function getDailyTransactions(date: string): Promise<DailyTransactionResponse> {
  return get<DailyTransactionResponse>(`/transactions/daily?date=${date}`)
}

