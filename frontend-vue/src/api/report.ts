import { get } from './http'
import type { MonthlyReport } from '@/types'

/**
 * 获取指定月份的报表数据
 * @param month 月份，格式 YYYY-MM
 */
export function getMonthlyReport(month: string): Promise<MonthlyReport> {
  return get<MonthlyReport>(`/reports/monthly?month=${month}`)
}

