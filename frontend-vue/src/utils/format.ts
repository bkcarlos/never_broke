/**
 * 格式化金额
 * @param value 金额数值
 * @param currency 货币符号，默认 ¥
 * @returns 格式化后的字符串
 */
export function formatMoney(value: number | string, currency: string = '¥'): string {
  const num = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(num)) return `${currency}0.00`
  
  // 使用千分位分隔符
  const formatted = num.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
  
  return `${currency}${formatted}`
}

/**
 * 格式化百分比
 * @param value 小数形式的比例值 (如 0.12 表示 12%)
 * @returns 格式化后的百分比字符串
 */
export function formatPercent(value: number | string): string {
  const num = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(num)) return '0%'
  
  return `${(num * 100).toFixed(1)}%`
}

/**
 * 格式化数字
 * @param value 数值
 * @param decimals 小数位数
 * @returns 格式化后的字符串
 */
export function formatNumber(value: number | string, decimals: number = 2): string {
  const num = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(num)) return '0'
  
  return num.toLocaleString('zh-CN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  })
}
