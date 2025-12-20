import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { getUserSettings, updateUserSettings } from '@/api/user'
import type { UserSettings, Locale } from '@/types'
import { setLocale } from '@/i18n'

export const useUserSettingsStore = defineStore('userSettings', () => {
  // 状态
  const showAmount = ref<boolean>(true)
  const locale = ref<Locale>('zh-CN')
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)

  // 初始化：从 localStorage 读取
  function initFromLocalStorage() {
    const storedShowAmount = localStorage.getItem('showAmount')
    if (storedShowAmount !== null) {
      showAmount.value = storedShowAmount === 'true'
    }

    const storedLocale = localStorage.getItem('locale')
    if (storedLocale) {
      locale.value = storedLocale as Locale
      setLocale(locale.value)
    }
  }

  // 从服务器加载设置
  async function loadSettings() {
    loading.value = true
    error.value = null

    try {
      const data = await getUserSettings()
      showAmount.value = data.showAmount
      locale.value = data.locale
      setLocale(data.locale)

      // 同步到 localStorage
      localStorage.setItem('showAmount', String(data.showAmount))
      localStorage.setItem('locale', data.locale)
    } catch (e) {
      error.value = e instanceof Error ? e.message : '加载设置失败'
      console.error('加载用户设置失败:', e)
    } finally {
      loading.value = false
    }
  }

  // 更新金额显示设置
  async function updateShowAmount(value: boolean) {
    showAmount.value = value
    localStorage.setItem('showAmount', String(value))

    // 保存到服务器
    try {
      await updateUserSettings({
        locale: locale.value,
        showAmount: value
      })
    } catch (e) {
      error.value = e instanceof Error ? e.message : '保存金额显示设置失败'
      console.error('保存金额显示设置失败:', e)
    }
  }

  // 切换金额显示
  async function toggleShowAmount() {
    await updateShowAmount(!showAmount.value)
  }

  // 更新语言设置
  async function updateLocale(newLocale: Locale) {
    locale.value = newLocale
    setLocale(newLocale)
    localStorage.setItem('locale', newLocale)

    // 保存到服务器
    try {
      await updateUserSettings({
        locale: newLocale,
        showAmount: showAmount.value
      })
    } catch (e) {
      error.value = e instanceof Error ? e.message : '保存语言设置失败'
      console.error('保存语言设置失败:', e)
    }
  }

  // 监听 storage 事件，实现跨标签页同步
  if (typeof window !== 'undefined') {
    window.addEventListener('storage', (e) => {
      if (e.key === 'showAmount' && e.newValue !== null) {
        showAmount.value = e.newValue === 'true'
      }
      if (e.key === 'locale' && e.newValue !== null) {
        locale.value = e.newValue as Locale
        setLocale(locale.value)
      }
    })
  }

  return {
    showAmount,
    locale,
    loading,
    error,
    initFromLocalStorage,
    loadSettings,
    updateShowAmount,
    toggleShowAmount,
    updateLocale
  }
})
