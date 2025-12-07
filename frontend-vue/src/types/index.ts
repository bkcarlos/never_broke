// ========== 首页相关类型 ==========

export interface TodayStats {
  todayExpense: number
  todayAvailable: number
  monthUsed: number
  monthBudget: number
  monthNet: number
}

export interface PayCycleOverview {
  cycleName: string
  daysToNextPayday: number
  freeToSpend: number
  cycleUsedPercent: number
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

export type HintType = 'info' | 'warning' | 'success'

export interface HomeHint {
  type: HintType
  textKey: string
}

export interface HomeOverview {
  todayStats: TodayStats
  payCycle: PayCycleOverview
  assetSummary: AssetSummary
  categoryTop: CategoryStat[]
  hint: HomeHint
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

// ========== 账户相关类型 ==========

export type AccountType = 'cash' | 'bank' | 'wallet' | 'credit'

export interface Account {
  id: string
  name: string
  type: AccountType
  currency: string
  balance: number
  archived: boolean
}

export interface AccountOption {
  id: string
  name: string
}

// ========== 交易相关类型 ==========

export type TransactionType = 'expense' | 'income' | 'transfer'

export interface CreateTransactionPayload {
  type: TransactionType
  amount: number
  categoryKey?: string
  remark?: string
  occurredAt: string
  accountId?: string // expense/income
  fromAccountId?: string // transfer
  toAccountId?: string // transfer
}

export interface Transaction {
  id: string
  type: TransactionType
  amount: number
  categoryKey?: string
  remark?: string
  occurredAt: string
  accountId?: string
  fromAccountId?: string
  toAccountId?: string
}

// ========== API 响应通用类型 ==========

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
