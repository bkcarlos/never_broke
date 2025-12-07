// ========== 首页相关类型 ==========

export interface TodayStats {
  todayExpense: number
  todayAvailable: number
  monthUsed: number
  monthBudget: number
  monthNet: number // 本月结余
}

export interface AssetSummary {
  totalAssets: number
  totalLiabilities: number
  netAssets: number
}

export interface CategoryStat {
  categoryKey: string
  amount: number
  percent: number
  rank: number
}

export interface HomeOverview {
  todayStats: TodayStats
  assetSummary: AssetSummary
  topCategories: CategoryStat[]
}

// ========== 明细页相关类型 ==========

export interface DailyTransaction {
  id: string
  type: 'income' | 'expense' | 'transfer'
  amount: number
  categoryKey: string
  remark?: string
  time: string // 'HH:mm'
}

export interface DailyTransactionResponse {
  date: string // 'YYYY-MM-DD'
  stats: {
    totalIncome: number
    totalExpense: number
  }
  list: DailyTransaction[]
}

// ========== 报表页相关类型 ==========

export interface CategoryReportItem {
  categoryKey: string
  amount: number
  percent: number
  rank: number
}

export interface DailyTrendPoint {
  date: string // 'YYYY-MM-DD'
  expense: number
  income: number
}

export interface MonthlyReport {
  month: string // 'YYYY-MM'
  totalIncome: number
  totalExpense: number
  netIncome: number
  categoryStats: CategoryReportItem[]
  dailyTrend: DailyTrendPoint[]
}

// ========== 用户设置相关类型 ==========

export type Locale = 'zh-CN' | 'en-US'

export interface UserSettings {
  locale: Locale
  showAmount: boolean
}

// ========== API 响应通用类型 ==========

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

