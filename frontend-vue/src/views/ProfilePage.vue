<template>
  <div class="page-container">
    <div class="page-content px-4">
    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-8">
      <p class="text-gray-500">{{ t('common.loading') }}</p>
    </div>

    <!-- 设置列表 -->
    <div v-else class="settings space-y-4">
      <!-- 语言设置 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <div class="flex items-center justify-between">
          <span class="font-semibold">{{ t('profile.language') }}</span>
          <div class="flex gap-2">
            <button
              @click="handleLocaleSelect('zh-CN')"
              class="px-4 py-2 rounded-lg font-medium transition-all duration-200"
              :class="currentLocale === 'zh-CN' 
                ? 'bg-blue-500 text-white shadow-md' 
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'"
            >
              简体中文
            </button>
            <button
              @click="handleLocaleSelect('en-US')"
              class="px-4 py-2 rounded-lg font-medium transition-all duration-200"
              :class="currentLocale === 'en-US' 
                ? 'bg-blue-500 text-white shadow-md' 
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'"
            >
              English
            </button>
          </div>
        </div>
      </div>

      <!-- 金额显隐 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <div class="flex items-center justify-between">
          <span class="font-semibold">
            {{ settings.showAmount ? t('profile.showAmount') : t('profile.hideAmount') }}
          </span>
          <button 
            @click="toggleShowAmount"
            class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors"
            :class="settings.showAmount ? 'bg-blue-500' : 'bg-gray-300'"
          >
            <span 
              class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform"
              :class="settings.showAmount ? 'translate-x-6' : 'translate-x-1'"
            />
          </button>
        </div>
      </div>

      <!-- 账户管理 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <button 
          @click="handleAddAccount"
          class="w-full flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <Icon name="bank" size="24px" class="text-indigo-500" />
            <span class="font-semibold">{{ t('account.manage.title', '账户管理') }}</span>
          </div>
          <Icon name="add" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 设置预算 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <button 
          @click="handleSetBudget"
          class="w-full flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <Icon name="target" size="24px" class="text-purple-500" />
            <span class="font-semibold">{{ t('actions.setBudget') }}</span>
          </div>
          <Icon name="add" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 导入账单 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <button 
          @click="handleImport"
          class="w-full flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <Icon name="import" size="24px" class="text-orange-500" />
            <span class="font-semibold">{{ t('actions.import') }}</span>
          </div>
          <Icon name="add" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 周期收入配置 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <button 
          @click="handleRecurringIncome"
          class="w-full flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <Icon name="wallet" size="24px" class="text-green-500" />
            <span class="font-semibold">周期收入配置</span>
          </div>
          <Icon name="add" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 示例金额展示 -->
      <div class="example bg-blue-50 rounded-lg p-4">
        <p class="text-sm text-gray-600 mb-2">{{ t('profile.example') }}</p>
        <p class="text-2xl font-bold">
          {{ settings.showAmount ? '¥12,345.67' : '¥****' }}
        </p>
      </div>

      <!-- 保存按钮 -->
      <button 
        v-if="hasChanges"
        @click="saveSettings"
        :disabled="saving"
        class="w-full py-3 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
      >
        {{ saving ? t('profile.saving') : t('common.save') }}
      </button>

      <!-- 错误提示 -->
      <div v-if="error" class="p-4 bg-red-50 text-red-600 rounded-lg">
        {{ error }}
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getUserSettings, updateUserSettings } from '@/api/user'
import { t, locale, setLocale } from '@/i18n'
import type { UserSettings, Locale } from '@/types'
import Icon from '@/components/Icon.vue'

const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)

const settings = ref<UserSettings>({
  locale: 'zh-CN',
  showAmount: true
})

const originalSettings = ref<UserSettings>({
  locale: 'zh-CN',
  showAmount: true
})

const currentLocale = ref<Locale>(locale.value)

const hasChanges = computed(() => {
  return (
    settings.value.locale !== originalSettings.value.locale ||
    settings.value.showAmount !== originalSettings.value.showAmount
  )
})

async function loadSettings() {
  loading.value = true
  error.value = null
  
  try {
    const data = await getUserSettings()
    settings.value = { ...data }
    originalSettings.value = { ...data }
    currentLocale.value = data.locale
    setLocale(data.locale)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载设置失败'
  } finally {
    loading.value = false
  }
}

function handleLocaleSelect(locale: Locale) {
  currentLocale.value = locale
  settings.value.locale = locale
  setLocale(locale)
}

function toggleShowAmount() {
  settings.value.showAmount = !settings.value.showAmount
}

function handleAddAccount() {
  router.push('/accounts')
}

function handleSetBudget() {
  router.push('/settings/budget')
}

function handleImport() {
  router.push('/settings/import-bills')
}

function handleRecurringIncome() {
  router.push('/settings/recurring-income')
}

async function saveSettings() {
  saving.value = true
  error.value = null
  
  try {
    const updated = await updateUserSettings({
      locale: settings.value.locale,
      showAmount: settings.value.showAmount
    })
    settings.value = { ...updated }
    originalSettings.value = { ...updated }
    alert(t('profile.saveSuccess'))
  } catch (e) {
    error.value = e instanceof Error ? e.message : '保存设置失败'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
/* 页面样式已移至全局样式文件 (style.css) 的 .page-container 和 .page-content */

</style>

