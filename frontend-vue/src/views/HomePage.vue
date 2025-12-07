<template>
  <div class="flex flex-col min-h-screen bg-slate-50">
    <!-- 顶部渐变 Hero 区 -->
    <div class="relative bg-gradient-to-br from-indigo-500 via-purple-500 to-sky-400 text-white h-52 px-4 pt-6 pb-8">
      <!-- 右上角操作按钮 -->
      <div class="absolute top-6 right-4 flex gap-3">
        <button 
          @click="toggleShowAmount" 
          class="hover:scale-110 transition-transform"
          :title="showAmount ? t('profile.hideAmount') : t('profile.showAmount')"
        >
          <Icon :name="showAmount ? 'eye' : 'eye-close'" size="24px" class="text-white" />
        </button>
      </div>

      <!-- 问候语和标题 -->
      <div class="mt-4">
        <p class="text-sm text-white/80 mb-1">{{ greeting }}</p>
        <div class="flex items-center gap-3 mb-0">
          <h1 class="text-3xl font-bold">NeverBroke</h1>
          <div class="inline-block bg-white/20 backdrop-blur-sm px-3 py-1 rounded-full">
            <p class="text-xs text-white/90">{{ t('home.keepRecording', { days: String(recordingDays) }) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <p class="text-gray-500">{{ t('common.loading') }}</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="flex-1 flex flex-col items-center justify-center px-4">
      <p class="text-red-500 mb-4">{{ t('common.error') }}: {{ error }}</p>
      <button 
        @click="loadData" 
        class="px-6 py-2 bg-indigo-500 text-white rounded-lg hover:bg-indigo-600 transition-colors"
      >
        {{ t('common.retry') }}
      </button>
    </div>

    <!-- 主内容区 -->
    <div v-else-if="data" class="flex-1 px-4 pb-4">
      <!-- 卡片 1: 资产总览 -->
      <div class="bg-white rounded-2xl shadow-md p-6 -mt-10 mb-4 relative z-10">
        <h2 class="text-lg font-semibold text-gray-800 mb-4">{{ t('home.assetSummary') }}</h2>
        <div class="grid grid-cols-3 gap-4">
          <div class="text-center">
            <p class="text-xs text-gray-500 mb-1">{{ t('home.totalAssets') }}</p>
            <p class="text-lg font-bold text-blue-600">
              ¥{{ formatMoney(data.assetSummary.totalAssets) }}
            </p>
          </div>
          <div class="text-center">
            <p class="text-xs text-gray-500 mb-1">{{ t('home.totalLiabilities') }}</p>
            <p class="text-lg font-bold text-orange-600">
              ¥{{ formatMoney(data.assetSummary.totalLiabilities) }}
            </p>
          </div>
          <div class="text-center">
            <p class="text-xs text-gray-500 mb-1">{{ t('home.netAssets') }}</p>
            <p class="text-lg font-bold text-green-600">
              ¥{{ formatMoney(data.assetSummary.netAssets) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 卡片 2: 本月预算 -->
      <div class="bg-white rounded-2xl shadow-md p-6 mb-4">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-gray-800">{{ t('home.monthBudget') }}</h2>
          <button 
            @click="$router.push('/reports')" 
            class="text-sm text-indigo-500 hover:text-indigo-600"
          >
            {{ t('home.adjustBudget') }} →
          </button>
        </div>
        <div class="flex items-center gap-6">
          <!-- 圆形进度 -->
          <div class="relative w-24 h-24 flex-shrink-0">
            <svg class="w-full h-full transform -rotate-90">
              <circle 
                cx="48" cy="48" r="40" 
                stroke="#e5e7eb" 
                stroke-width="8" 
                fill="none"
              />
              <circle 
                cx="48" cy="48" r="40" 
                :stroke="budgetColor" 
                stroke-width="8" 
                fill="none"
                :stroke-dasharray="`${budgetUsedPercent * 2.51} 251`"
                class="transition-all duration-500"
              />
            </svg>
            <div class="absolute inset-0 flex items-center justify-center">
              <span class="text-sm font-bold text-gray-700">
                {{ Math.round(budgetUsedPercent) }}%
              </span>
            </div>
          </div>
          <!-- 预算详情 -->
          <div class="flex-1 space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">{{ t('home.monthBudget') }}</span>
              <span class="font-semibold">¥{{ formatMoney(data.todayStats.monthBudget) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">{{ t('home.monthUsed') }}</span>
              <span class="font-semibold text-red-500">¥{{ formatMoney(data.todayStats.monthUsed) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">{{ t('home.remaining') }}</span>
              <span class="font-semibold text-green-500">
                ¥{{ formatMoney(data.todayStats.monthBudget - data.todayStats.monthUsed) }}
              </span>
            </div>
          </div>
        </div>
        <p class="text-xs text-gray-400 mt-3">
          {{ t('home.budgetProgress', { 
            dayPercent: String(Math.round(dayOfMonth / daysInMonth * 100)),
            usedPercent: String(Math.round(budgetUsedPercent))
          }) }}
        </p>
      </div>

      <!-- 卡片 3: 本周期可自由支配 -->
      <div class="bg-white rounded-2xl shadow-md p-6 mb-4">
        <div class="flex items-center justify-between">
          <div class="flex-1">
            <p class="text-sm text-gray-500 mb-1">{{ t('home.freeToSpend') }}</p>
            <p class="text-3xl font-bold text-indigo-600 mb-2">
              ¥{{ formatMoney(data.payCycle.freeToSpend) }}
            </p>
            <p class="text-xs text-gray-400">
              {{ t('home.daysToPayday', { days: String(data.payCycle.daysToNextPayday) }) }}
            </p>
          </div>
          <div class="flex-shrink-0 ml-4">
            <!-- 简单的环形进度 -->
            <div class="relative w-20 h-20">
              <svg class="w-full h-full transform -rotate-90">
                <circle 
                  cx="40" cy="40" r="32" 
                  stroke="#e5e7eb" 
                  stroke-width="6" 
                  fill="none"
                />
                <circle 
                  cx="40" cy="40" r="32" 
                  stroke="#6366f1" 
                  stroke-width="6" 
                  fill="none"
                  :stroke-dasharray="`${data.payCycle.cycleUsedPercent * 2.01} 201`"
                  class="transition-all duration-500"
                />
              </svg>
              <div class="absolute inset-0 flex items-center justify-center">
                <span class="text-xs font-semibold text-gray-700">
                  {{ Math.round(data.payCycle.cycleUsedPercent) }}%
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 模块 4: 本月消费概览 -->
      <div class="bg-white rounded-2xl shadow-md p-6 mb-4">
        <h2 class="text-lg font-semibold text-gray-800 mb-4">{{ t('home.spendOverview') }}</h2>
        <!-- TOP3 列表 -->
        <div class="space-y-3">
            <div 
              v-for="(cat, index) in data.categoryTop" 
              :key="cat.categoryKey"
              class="flex items-center justify-between"
            >
              <div class="flex items-center gap-2">
                <div 
                  class="w-3 h-3 rounded-full"
                  :class="getCategoryDotColor(index)"
                ></div>
                <span class="text-sm font-medium text-gray-700">
                  {{ t(`category.${cat.categoryKey}`) }}
                </span>
              </div>
              <div class="text-right">
                <p class="text-sm font-semibold text-gray-800">
                  ¥{{ formatMoney(cat.amount) }}
                </p>
                <p class="text-xs text-gray-400">{{ cat.percent }}%</p>
              </div>
            </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { HomeOverview } from '@/types'
import { getHomeOverview } from '@/api/home'
import { t } from '@/i18n'
import Icon from '@/components/Icon.vue'

const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const data = ref<HomeOverview | null>(null)

// 金额显隐，从 localStorage 读取
const showAmount = ref<boolean>(true)

// 记账天数
const recordingDays = computed(() => {
  // 从 localStorage 获取首次记账日期
  const firstRecordDate = localStorage.getItem('firstRecordDate')
  if (!firstRecordDate) {
    // 如果没有记录，设置当前日期为首次记账日期
    const today = new Date().toISOString().split('T')[0]
    localStorage.setItem('firstRecordDate', today)
    return 1
  }
  
  // 计算天数差
  const first = new Date(firstRecordDate)
  const today = new Date()
  const diffTime = Math.abs(today.getTime() - first.getTime())
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  return diffDays + 1
})

// 问候语
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return t('home.greeting.night')
  if (hour < 12) return t('home.greeting.morning')
  if (hour < 18) return t('home.greeting.afternoon')
  return t('home.greeting.evening')
})

// 当前月份
const currentMonth = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}${t('common.year')}${now.getMonth() + 1}${t('common.month')}`
})

// 预算使用百分比
const budgetUsedPercent = computed(() => {
  if (!data.value) return 0
  return (data.value.todayStats.monthUsed / data.value.todayStats.monthBudget) * 100
})

// 预算颜色
const budgetColor = computed(() => {
  const percent = budgetUsedPercent.value
  if (percent >= 90) return '#ef4444' // red
  if (percent >= 70) return '#f59e0b' // orange
  return '#10b981' // green
})

// 当前日期信息
const dayOfMonth = computed(() => new Date().getDate())
const daysInMonth = computed(() => {
  const now = new Date()
  return new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate()
})

// 格式化金额
function formatMoney(value: number): string {
  if (!showAmount.value) return '***'
  return value.toFixed(2)
}

// 切换金额显隐
function toggleShowAmount() {
  showAmount.value = !showAmount.value
  localStorage.setItem('showAmount', String(showAmount.value))
}

function getCategoryDotColor(index: number): string {
  const colors = ['bg-indigo-500', 'bg-purple-500', 'bg-pink-500']
  return colors[index] || 'bg-gray-500'
}

// 加载数据
async function loadData() {
  loading.value = true
  error.value = null
  
  console.log('🔍 开始加载首页数据...')
  
  try {
    const result = await getHomeOverview()
    console.log('✅ 数据加载成功:', result)
    data.value = result
  } catch (e) {
    console.error('❌ 数据加载失败:', e)
    error.value = e instanceof Error ? e.message : t('common.error')
  } finally {
    loading.value = false
    console.log('📊 当前状态 - loading:', loading.value, 'error:', error.value, 'hasData:', !!data.value)
  }
}

// 初始化
onMounted(() => {
  // 读取金额显隐设置
  const stored = localStorage.getItem('showAmount')
  if (stored !== null) {
    showAmount.value = stored === 'true'
  }
  
  // 加载数据
  loadData()
})
</script>

<style scoped>
/* 可选：添加一些自定义动画 */
</style>
