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

export interface MonthOverview {
  monthIncome: number      // 本月收入
  monthExpense: number     // 本月支出
  status: string           // 收支状态：surplus(盈余) / deficit(赤字) / balanced(平衡)
  suggestion: string       // 消费建议
}

export interface AssetSummary {
  totalAssets: number        // 总资产（不含信用卡）
  availableBalance: number   // 可用余额（总资产+信用卡可用额度）
  totalLiabilities: number   // 总负债（信用卡待还款）
  netAssets: number          // 净资产（总资产-总负债）
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

export type PeriodType = 'monthly' | 'yearly'

export interface CategorySummary {
  categoryKey: string
  amount: number
  ratio: number
}

export interface TrendPoint {
  date: string // YYYY-MM-DD
  income: number
  expense: number
}

export interface CompareSummary {
  prevYear: number
  prevMonth: number
  expenseChange: number
  expenseChangeRate: number
}

export interface SummaryReport {
  period: PeriodType
  year: number
  month: number
  totalIncome: number
  totalExpense: number
  netAmount: number
  categories: CategorySummary[]
  trend: TrendPoint[]
  topCategories: CategorySummary[]
  compareWithPrevMonth?: CompareSummary
}

// 保留旧类型以兼容现有代码
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

export type InstallmentRepayMethod = 'equal_principal' | 'equal_payment'

export interface InstallmentPreviewItem {
  period: number
  dueDate: string
  principal: number
  interest: number
  total: number
  remainingPrincipal: number
}

export interface InstallmentPreviewResult {
  principal: number
  totalInterest: number
  totalAmount: number
  termMonths: number
  monthlyRate: number
  repayMethod: InstallmentRepayMethod
  schedule: InstallmentPreviewItem[]
}

export interface PreviewInstallmentPayload {
  principal: number
  termMonths: number
  monthlyRatePercent: number
  repayMethod: InstallmentRepayMethod
  firstDueDate: string
  currency: string
}

export interface InstallmentPlan {
  id: string
  accountId: string
  totalAmount: number
  principal: number
  totalInterest: number
  termMonths: number
  monthlyRate: number
  repayMethod: InstallmentRepayMethod
  currency: string
  firstDueDate: string
  status: InstallmentPlanStatus
  createdAt: string
}

export interface CreateExpenseInstallmentPayload {
  accountId: string
  principal: number
  termMonths: number
  monthlyRatePercent: number
  repayMethod: InstallmentRepayMethod
  firstDueDate: string
  currency: string
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
  | 'bimonthly'
  | 'quarterly'
  | 'semiannual'
  | 'weekly'
  | 'yearly'

export interface RecurringIncomePlan {
  id: string
  type: RecurringIncomeType
  label: string
  accountId: string
  amount: number // 工资：税后
  currency: string
  frequency: RecurringIncomeFrequency
  dayOfMonth?: number
  weekday?: number
  nextDueDate: string
  lastFiredDate?: string
  cityCode?: string // 工资用
  active: boolean
  createdAt: string

  // 薪资专用字段
  salaryPreTaxMonthly?: number
  salarySocialRate?: number
  salaryExtraDeductions?: number
  salaryPayDay?: number
  salaryStartDate?: string
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
  cityCode?: string // 工资用
  active: boolean
}

export interface RecurringIncomeRealizePayload {
  date?: string // YYYY-MM-DD，可选
}

export interface RecurringIncomeRealizeResult {
  plan: RecurringIncomePlan
  transaction: Transaction
}

// ========== 工资预览相关类型 ==========

export interface SalaryPreviewRequest {
  preTaxAmount: number
  cityCode: string
  startDate: string // YYYY-MM-DD
  payDay: number // 1-31
  frequency: RecurringIncomeFrequency
  currency: string
  socialRate?: number
  extraDeductions?: number
}

export interface SalaryPreviewItem {
  date: string
  preTax: number
  socialSecurity: number
  taxableIncome: number
  taxAmount: number
  postTax: number
  cumulativePreTax: number
  cumulativeTax: number
}

export interface SalaryPreviewResponse {
  cityCode: string
  year: number
  items: SalaryPreviewItem[]
  totalPreTax: number
  totalTax: number
  totalPostTax: number
}

// ========== 未来明细投射相关类型 ==========

export interface ProjectedTransaction {
  date: string
  accountId: string
  amount: number
  type: 'income'
  label: string
  planId: string
  isSalary: boolean
}

// ========== 时间线相关类型 ==========

export type TimelineMode = 'history' | 'with_future'

export interface DailySummary {
  date: string
  income: number
  expense: number
  net: number
  hasFuture: boolean
}

export interface DailyTimeline {
  fromDate: string
  toDate: string
  mode: TimelineMode
  days: DailySummary[]
}

// ========== 现金流预测相关类型 ==========

export type ForecastBase = 'total_assets' | 'account'
export type ForecastStatus = 'safe' | 'warning' | 'danger'

export interface DailyForecastPoint {
  date: string           // YYYY-MM-DD
  openingBalance: number
  inflow: number
  outflow: number
  closingBalance: number
}

export interface CashflowForecast {
  startDate: string
  endDate: string
  base: ForecastBase
  accountId?: string
  safeAmount: number
  points: DailyForecastPoint[]
  status: ForecastStatus
  dangerDates: string[]
  warningDates: string[]
}

// ========== 城市社保比例相关类型 ==========

export interface CitySocialRate {
  id: string
  cityCode: string
  cityName: string
  effectiveDate: string
  pensionRate: number
  medicalRate: number
  unemploymentRate: number
  injuryRate: number
  maternityRate: number
  housingFundRate: number
  totalRate: number
  isActive: boolean
  remark?: string
  createdAt: string
}

export interface SocialRateComponent {
  name: string
  rate: number
}

export interface CitySocialRateDetail {
  cityCode: string
  cityName: string
  totalRate: number
  components: SocialRateComponent[]
  effectiveDate: string
  remark?: string
}

// ========== 薪资方案配置相关类型 ==========

export interface SalaryPlanConfig {
  preTaxMonthly: number
  cityCode: string
  startDate: string
  payDay: number
  socialRate: number
  extraDeductions: number
  accountId: string
  currency: string
}

// ========== API 响应通用类型 ==========

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
