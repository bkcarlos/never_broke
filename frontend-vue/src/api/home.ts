import { get } from './http'
import type { HomeOverview } from '@/types'

/**
 * 获取首页概览数据
 */
export function getHomeOverview(): Promise<HomeOverview> {
  return get<HomeOverview>('/home/overview')
}
