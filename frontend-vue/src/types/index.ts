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

export type AccountInstitution =
  | 'icbc'   // 工商银行
  | 'ccb'    // 建设银行
  | 'abc'    // 农业银行
  | 'boc'    // 中国银行
  | 'bocom'  // 交通银行
  | 'cmb'    // 招商银行
  | 'spdb'   // 浦发银行
  | 'citic'  // 中信银行
  | 'ceb'    // 光大银行
  | 'hxb'    // 华夏银行
  | 'cmbc'   // 民生银行
  | 'cgb'    // 广发银行
  | 'pab'    // 平安银行
  | 'cib'    // 兴业银行
  | 'psbc'   // 邮储银行
  | 'wechat' // 微信
  | 'alipay' // 支付宝
  | 'other'  // 其他

export interface Account {
  id: string
  type: AccountType
  institution?: AccountInstitution
  cardLast4?: string
  currency: string
  balance: number
  creditLimit?: number
  tempLimit?: number
  archived: boolean
}

export interface AccountOption {
  id: string
  name: string
}

export interface CreateAccountPayload {
  type: AccountType
  institution?: AccountInstitution
  cardLast4?: string
  currency: string
  balance: number
  creditLimit?: number
  tempLimit?: number
}

export interface UpdateCreditLimitPayload {
  creditLimit?: number
  tempLimit?: number
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
  // 新增：分期关联信息（如果是分期每一期的还款）
  installmentPlanId?: string
  installmentPeriod?: number
  installmentTotal?: number
}

// ========== 预算相关类型 ==========

export interface MonthlyBudget {
  month: string       // 'YYYY-MM'
  totalBudget: number
}

// ========== 导入相关类型 ==========

export interface ImportPreviewItem {
  date: string
  categoryGuess: string
  amount: number
  remark: string
}

export interface ImportPreviewResult {
  count: number
  items: ImportPreviewItem[]
}

// ========== 收入计划相关类型 ==========

export type IncomeFrequency = 'monthly' | 'weekly' | 'biweekly' | 'yearly'

export interface IncomePlan {
  id: string
  name: string
  amount: number
  frequency: IncomeFrequency
  accountId: string
  firstPayDate: string // YYYY-MM-DD
  endDate?: string
  active: boolean
}

export interface CreateIncomePlanPayload {
  name: string
  amount: number
  frequency: IncomeFrequency
  accountId: string
  firstPayDate: string
  endDate?: string
}

// ========== 分期相关类型 ==========

export type InstallmentPlanStatus = 'active' | 'finished' | 'canceled'

export interface InstallmentPlan {
  id: string
  accountId: string
  totalAmount: number
  principal: number
  feeTotal: number
  termMonths: number
  monthlyAmount: number
  currency: string
  firstDueDate: string
  status: InstallmentPlanStatus
  createdAt: string
}

export interface CreateExpenseInstallmentPayload {
  accountId: string
  totalAmount: number
  principal: number
  feeTotal: number
  termMonths: number
  currency: string
  firstDueDate: string
  categoryKey?: string
  remark?: string
  purchaseDate: string
}

export interface CreateExpenseInstallmentResult {
  plan: InstallmentPlan
  transactions: Transaction[]
}

// ========== 周期收入配置相关类型 ==========

export type RecurringIncomeType = 'salary' | 'rent' | 'side' | 'other'

export type RecurringIncomeFrequency =
  | 'monthly'
  | 'weekly'
  | 'biweekly'
  | 'yearly'

export interface RecurringIncomePlan {
  id: string
  type: RecurringIncomeType
  label: string
  accountId: string
  amount: number
  currency: string
  frequency: RecurringIncomeFrequency
  dayOfMonth?: number
  weekday?: number
  nextDueDate: string
  lastFiredDate?: string
  active: boolean
  createdAt: string
}

export interface UpsertRecurringIncomePayload {
  id?: string
  type: RecurringIncomeType
  label: string
  accountId: string
  amount: number
  currency: string
  frequency: RecurringIncomeFrequency
  dayOfMonth?: number
  weekday?: number
  nextDueDate: string
  active: boolean
}

export interface RecurringIncomeRealizePayload {
  date?: string // YYYY-MM-DD，可选
}

export interface RecurringIncomeRealizeResult {
  plan: RecurringIncomePlan
  transaction: Transaction
}

// ========== API 响应通用类型 ==========

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
