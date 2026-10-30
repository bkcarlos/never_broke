import { request } from './http'
import type { SummaryReport } from '@/types'

export async function fetchMonthlySummary(
  year: number,
  month: number,
  currency = 'CNY',
): Promise<SummaryReport> {
  const query = new URLSearchParams()
  query.set('year', String(year))
  query.set('month', String(month))
  query.set('currency', currency)

  return request<SummaryReport>(
    `/reports/monthly-summary?${query.toString()}`,
    { method: 'GET' },
  )
}
