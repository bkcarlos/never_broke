<template>
  <div class="profile-page p-4">
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
          <select 
            v-model="currentLocale" 
            @change="handleLocaleChange"
            class="px-3 py-1 border rounded"
          >
            <option value="zh-CN">简体中文</option>
            <option value="en-US">English</option>
          </select>
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

      <!-- 示例金额展示 -->
      <div class="example bg-blue-50 rounded-lg p-4">
        <p class="text-sm text-gray-600 mb-2">示例：</p>
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
        {{ saving ? '保存中...' : t('common.save') }}
      </button>

      <!-- 错误提示 -->
      <div v-if="error" class="p-4 bg-red-50 text-red-600 rounded-lg">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getUserSettings, updateUserSettings } from '@/api/user'
import { t, locale, setLocale } from '@/i18n'
import type { UserSettings, Locale } from '@/types'

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

function handleLocaleChange() {
  settings.value.locale = currentLocale.value
  setLocale(currentLocale.value)
}

function toggleShowAmount() {
  settings.value.showAmount = !settings.value.showAmount
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
    alert('设置已保存！')
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

