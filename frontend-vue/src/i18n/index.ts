import { ref, computed } from 'vue'
import type { Locale } from '@/types'
import zhCN from './zh-CN'
import enUS from './en-US'

const messages: Record<Locale, Record<string, string>> = {
  'zh-CN': zhCN,
  'en-US': enUS
}

// 从 localStorage 读取用户语言偏好，默认中文
const getStoredLocale = (): Locale => {
  const stored = localStorage.getItem('locale')
  if (stored === 'zh-CN' || stored === 'en-US') {
    return stored
  }
  return 'zh-CN'
}

const currentLocale = ref<Locale>(getStoredLocale())

/**
 * 获取当前语言
 */
export const locale = computed(() => currentLocale.value)

/**
 * 设置语言
 */
export function setLocale(newLocale: Locale) {
  currentLocale.value = newLocale
  localStorage.setItem('locale', newLocale)
}

/**
 * 翻译函数
 * @param key 翻译键
 * @param params 参数对象（可选）
 * @param fallback 降级文案（可选）
 */
export function t(key: string, params?: Record<string, string | number> | string, fallback?: string): string {
  // 兼容旧的调用方式：t(key, fallback)
  if (typeof params === 'string') {
    fallback = params
    params = undefined
  }

  let message = messages[currentLocale.value]?.[key]
  
  // 如果当前语言没有，尝试中文
  if (!message && currentLocale.value !== 'zh-CN') {
    message = messages['zh-CN']?.[key]
  }
  
  // 都没有就返回 fallback 或 key
  if (!message) {
    return fallback || key
  }

  // 替换参数
  if (params && typeof params === 'object') {
    const values = params
    return message.replace(/\{(\w+)\}/g, (match, key: string) => {
      return values[key] !== undefined ? String(values[key]) : match
    })
  }

  return message
}

/**
 * 格式化金额
 */
export function formatAmount(amount: number, showAmount: boolean = true): string {
  if (!showAmount) {
    return '****'
  }
  return amount.toFixed(2)
}

/**
 * 格式化日期
 */
export function formatDate(date: string): string {
  // 简单实现，可以后续用 dayjs 等库优化
  return date
}

