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
 * @param fallback 降级文案（可选）
 */
export function t(key: string, fallback?: string): string {
  const message = messages[currentLocale.value]?.[key]
  if (message) {
    return message
  }
  
  // 如果当前语言没有，尝试中文
  if (currentLocale.value !== 'zh-CN') {
    const zhMessage = messages['zh-CN']?.[key]
    if (zhMessage) {
      return zhMessage
    }
  }
  
  // 都没有就返回 fallback 或 key
  return fallback || key
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

