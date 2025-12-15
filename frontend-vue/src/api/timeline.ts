import { request } from './http'
import type { DailyTimeline, TimelineMode, Transaction } from '@/types'

export async function fetchDailySummary(
  from: string,
  to: string,
  mode: TimelineMode,
): Promise<DailyTimeline> {
  const query = new URLSearchParams()
  query.set('from', from)
  query.set('to', to)
  query.set('mode', mode)

  return request<DailyTimeline>(
    `/timeline/daily-summary?${query.toString()}`,
    { method: 'GET' },
  )
}

export async function fetchDailyDetail(
  date: string,
  mode: TimelineMode,
): Promise<{ date: string; mode: TimelineMode; transactions: Transaction[] }> {
  const query = new URLSearchParams()
  query.set('date', date)
  query.set('mode', mode)

  return request<{
    date: string
    mode: TimelineMode
    transactions: Transaction[]
  }>(`/timeline/daily-detail?${query.toString()}`, { method: 'GET' })
}
