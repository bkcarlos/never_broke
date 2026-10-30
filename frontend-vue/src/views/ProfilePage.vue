<template>
  <div class="page-container">
    <div class="page-content">
    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-8">
      <p class="text-gray-500">{{ t('common.loading') }}</p>
    </div>

    <!-- 设置列表 -->
    <div v-else class="settings space-y-4">
      <!-- 语言设置 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Icon name="language" size="24px" class="text-blue-500" />
            <span class="font-semibold">{{ t('profile.language') }}</span>
          </div>
          <div class="flex gap-2">
            <button
              @click="handleLocaleSelect('zh-CN')"
              class="px-4 py-2 rounded-lg font-medium transition-all duration-200"
              :class="userSettingsStore.locale === 'zh-CN' 
                ? 'bg-blue-500 text-white shadow-md' 
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'"
            >
              简体中文
            </button>
            <button
              @click="handleLocaleSelect('en-US')"
              class="px-4 py-2 rounded-lg font-medium transition-all duration-200"
              :class="userSettingsStore.locale === 'en-US' 
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
          <div class="flex items-center gap-3">
            <Icon :name="userSettingsStore.showAmount ? 'eye' : 'eye-close'" size="24px" class="text-blue-500" />
            <span class="font-semibold">
              {{ userSettingsStore.showAmount ? t('profile.showAmount') : t('profile.hideAmount') }}
            </span>
          </div>
          <button 
            @click="toggleShowAmount"
            class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors"
            :class="userSettingsStore.showAmount ? 'bg-blue-500' : 'bg-gray-300'"
          >
            <span 
              class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform"
              :class="userSettingsStore.showAmount ? 'translate-x-6' : 'translate-x-1'"
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
          <Icon name="chevron-right" size="20px" class="text-gray-400" />
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
          <Icon name="chevron-right" size="20px" class="text-gray-400" />
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
          <Icon name="chevron-right" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 薪资管理 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <button 
          @click="handleSalaryControl"
          class="w-full flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <Icon name="money" size="24px" class="text-emerald-500" />
            <div class="text-left">
              <div class="font-semibold block">
                {{ t('me.salary.title', '薪资管理') }}
              </div>
              <div class="text-xs text-slate-500">
                {{ t('me.salary.subtitle', '配置工资方案并预览税后收入') }}
              </div>
            </div>
          </div>
          <Icon name="chevron-right" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 固定收入 -->
      <div class="setting-item bg-white rounded-lg shadow p-4">
        <button 
          @click="handleRecurringIncome"
          class="w-full flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <Icon name="refresh" size="24px" class="text-teal-500" />
            <div class="text-left">
              <span class="font-semibold block">{{ t('recurring.title', '固定收入') }}</span>
              <p class="text-xs text-gray-500 mt-0.5">{{ t('recurring.subtitle', '管理您的周期性固定收入') }}</p>
            </div>
          </div>
          <Icon name="chevron-right" size="20px" class="text-gray-400" />
        </button>
      </div>

      <!-- 错误提示 -->
      <div v-if="error" class="p-4 bg-red-50 text-red-600 rounded-lg">
        {{ error }}
      </div>

      <!-- 退出登录 - 独立栏 -->
      <div class="mt-8 pt-4 border-t border-gray-200">
        <button
          @click="handleLogout"
          class="w-full py-4 bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow flex items-center justify-center gap-2 text-red-600 hover:text-red-700 font-semibold"
        >
          <Icon name="logout" size="20px" />
          <span>{{ t('auth.logout') }}</span>
        </button>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { logout } from '@/api/auth'
import { t } from '@/i18n'
import type { Locale } from '@/types'
import Icon from '@/components/Icon.vue'
import { useUserSettingsStore } from '@/stores/userSettings'

const router = useRouter()
const userSettingsStore = useUserSettingsStore()

const error = computed(() => userSettingsStore.error)
const loading = computed(() => userSettingsStore.loading)

async function handleLocaleSelect(locale: Locale) {
  await userSettingsStore.updateLocale(locale)
}

async function toggleShowAmount() {
  await userSettingsStore.toggleShowAmount()
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

function handleSalaryControl() {
  router.push('/me/salary')
}

function handleRecurringIncome() {
  router.push('/settings/recurring-income')
}

// 退出登录
async function handleLogout() {
  if (confirm(t('auth.logoutConfirm'))) {
    try {
      await logout()
      router.push('/login')
    } catch (error) {
      console.error('Logout error:', error)
      // 即使失败也清除本地数据并跳转
      router.push('/login')
    }
  }
}

onMounted(() => {
  userSettingsStore.loadSettings()
})
</script>

<style scoped>
/* 页面样式已移至全局样式文件 (style.css) 的 .page-container 和 .page-content */

</style>

