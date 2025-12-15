import { request } from './http'
import type { SummaryReport } from '@/types'

export async function fetchMonthlySummary(
  year: number,
  month: number,
): Promise<SummaryReport> {
  const query = new URLSearchParams()
  query.set('year', String(year))
  query.set('month', String(month))

  return request<SummaryReport>(
    `/reports/monthly-summary?${query.toString()}`,
    { method: 'GET' },
  )
}
