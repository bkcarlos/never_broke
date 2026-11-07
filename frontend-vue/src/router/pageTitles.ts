// Shared by desktop and mobile navigation so nested pages keep their context.
export const pageTitles: Record<string, string> = {
  '/login': '登录',
  '/': '财务总览',
  '/detail': '交易账本',
  '/reports': '分析报表',
  '/review': '核对工作台',
  '/plans': '账单与还款',
  '/accounts': '账户管理',
  '/accounts/new': '新增账户',
  '/insights': '订阅与预算',
  '/settings/import-bills': '导入与同步',
  '/record': '记一笔',
  '/actions': '快捷记账',
  '/transfer': '账户间转账',
  '/timeline': '收支日历',
  '/profile': '设置与偏好',
  '/cashflow': '每月收支预测',
  '/settings/budget': '月度预算',
  '/settings/salary': '收入计划',
  '/settings/recurring-income': '固定收入',
  '/me/salary': '薪资管理',
  '/tools/salary-calculator': '薪资计算器',
}

export const isSettingsPage = (path: string) => [
  '/profile', '/settings/budget', '/settings/salary',
  '/settings/recurring-income', '/me/salary', '/tools/salary-calculator',
].includes(path)
