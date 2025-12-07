<template>
  <div class="home-page p-4">
    <h1 class="text-2xl font-bold mb-6">{{ t('home.title') }}</h1>

    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-8">
      <p class="text-gray-500">{{ t('common.loading') }}</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="text-center py-8">
      <p class="text-red-500 mb-4">{{ t('common.error') }}: {{ error }}</p>
      <button 
        @click="loadData" 
        class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        {{ t('common.retry') }}
      </button>
    </div>

    <!-- 数据展示 -->
    <div v-else-if="overview">
      <!-- 今日数据卡片 -->
      <div class="card bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-lg font-semibold mb-3">今日概览</h2>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-500">{{ t('home.todayExpense') }}</p>
            <p class="text-xl font-bold text-red-500">
              ¥{{ formatAmount(overview.todayStats.todayExpense, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-sm text-gray-500">{{ t('home.todayAvailable') }}</p>
            <p class="text-xl font-bold text-green-500">
              ¥{{ formatAmount(overview.todayStats.todayAvailable, showAmount) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 本月数据卡片 -->
      <div class="card bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-lg font-semibold mb-3">本月概览</h2>
        <div class="grid grid-cols-3 gap-3">
          <div>
            <p class="text-xs text-gray-500">{{ t('home.monthUsed') }}</p>
            <p class="text-lg font-bold">
              ¥{{ formatAmount(overview.todayStats.monthUsed, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t('home.monthBudget') }}</p>
            <p class="text-lg font-bold">
              ¥{{ formatAmount(overview.todayStats.monthBudget, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t('home.monthNet') }}</p>
            <p class="text-lg font-bold" :class="overview.todayStats.monthNet >= 0 ? 'text-green-500' : 'text-red-500'">
              ¥{{ formatAmount(overview.todayStats.monthNet, showAmount) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 资产总览卡片 -->
      <div class="card bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-lg font-semibold mb-3">{{ t('home.assetSummary') }}</h2>
        <div class="space-y-2">
          <div class="flex justify-between">
            <span class="text-gray-600">{{ t('home.totalAssets') }}</span>
            <span class="font-bold text-blue-600">
              ¥{{ formatAmount(overview.assetSummary.totalAssets, showAmount) }}
            </span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">{{ t('home.totalLiabilities') }}</span>
            <span class="font-bold text-orange-600">
              ¥{{ formatAmount(overview.assetSummary.totalLiabilities, showAmount) }}
            </span>
          </div>
          <div class="flex justify-between pt-2 border-t">
            <span class="text-gray-800 font-semibold">{{ t('home.netAssets') }}</span>
            <span class="font-bold text-green-600 text-lg">
              ¥{{ formatAmount(overview.assetSummary.netAssets, showAmount) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 支出分类 TOP -->
      <div v-if="overview.topCategories.length > 0" class="card bg-white rounded-lg shadow p-4">
        <h2 class="text-lg font-semibold mb-3">{{ t('home.topCategories') }}</h2>
        <div class="space-y-2">
          <div 
            v-for="cat in overview.topCategories" 
            :key="cat.categoryKey"
            class="flex items-center justify-between"
          >
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500">#{{ cat.rank }}</span>
              <span>{{ t(`category.${cat.categoryKey}`) }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500">{{ cat.percent }}%</span>
              <span class="font-semibold">
                ¥{{ formatAmount(cat.amount, showAmount) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getHomeOverview } from '@/api/home'
import { t, formatAmount } from '@/i18n'
import type { HomeOverview } from '@/types'

const loading = ref(false)
const error = ref<string | null>(null)
const overview = ref<HomeOverview | null>(null)
const showAmount = ref(true) // 后续可以从用户设置读取

async function loadData() {
  loading.value = true
  error.value = null
  
  try {
    overview.value = await getHomeOverview()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '未知错误'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

