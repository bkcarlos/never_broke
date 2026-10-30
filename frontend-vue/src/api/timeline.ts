import { request } from './http'
import type { DailyTimeline, TimelineMode, Transaction } from '@/types'

export async function fetchDailySummary(
  from: string,
  to: string,
  mode: TimelineMode,
  currency = 'CNY',
): Promise<DailyTimeline> {
  const query = new URLSearchParams()
  query.set('from', from)
  query.set('to', to)
  query.set('mode', mode)
  query.set('currency', currency)

  return request<DailyTimeline>(
    `/timeline/daily-summary?${query.toString()}`,
    { method: 'GET' },
  )
}

export async function fetchDailyDetail(
  date: string,
  mode: TimelineMode,
  currency = 'CNY',
): Promise<{ date: string; mode: TimelineMode; transactions: Transaction[] }> {
  const query = new URLSearchParams()
  query.set('date', date)
  query.set('mode', mode)
  query.set('currency', currency)

  return request<{
    date: string
    mode: TimelineMode
    transactions: Transaction[]
  }>(`/timeline/daily-detail?${query.toString()}`, { method: 'GET' })
}
