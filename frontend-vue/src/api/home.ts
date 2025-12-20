import { get } from './http'
import type { 
  HomeOverview, 
  AssetSummary, 
  TodayStats, 
  PayCycleOverview, 
  MonthOverview,
  CategoryStat 
} from '@/types'

/**
 * 获取首页概览数据（完整数据，保留用于兼容）
 */
export function getHomeOverview(): Promise<HomeOverview> {
  return get<HomeOverview>('/home/overview')
}

/**
 * 获取资产总览
 */
export function getAssetSummary(): Promise<AssetSummary> {
  return get<AssetSummary>('/home/asset-summary')
}

/**
 * 获取本月预算统计
 */
export function getBudgetStats(): Promise<TodayStats> {
  return get<TodayStats>('/home/budget-stats')
}

/**
 * 获取支付周期信息
 */
export function getPayCycle(): Promise<PayCycleOverview> {
  return get<PayCycleOverview>('/home/pay-cycle')
}

/**
 * 获取本月概览（收支、TOP3大额支出、建议）
 */
export function getMonthOverview(): Promise<MonthOverview> {
  return get<MonthOverview>('/home/month-overview')
}

/**
 * 获取分类统计（本月消费分类 TOP3）
 */
export function getCategoryStats(): Promise<CategoryStat[]> {
  return get<CategoryStat[]>('/home/category-stats')
}
