import { request } from './http'
import type { CashflowForecast, ForecastBase } from '@/types'

export interface CashflowForecastParams {
  startDate: string
  days: number
  base: ForecastBase
  accountId?: string
  safeAmount: number
}

export async function fetchCashflowForecast(
  params: CashflowForecastParams,
): Promise<CashflowForecast> {
  const query = new URLSearchParams()
  query.set('startDate', params.startDate)
  query.set('days', String(params.days))
  query.set('base', params.base)
  if (params.accountId) {
    query.set('accountId', params.accountId)
  }
  query.set('safeAmount', String(params.safeAmount))

  return request<CashflowForecast>(
    `/cashflow/forecast?${query.toString()}`,
    { method: 'GET' },
  )
}

