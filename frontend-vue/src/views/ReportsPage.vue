<template>
  <div class="reports-page p-4">
    <!-- 月份选择 -->
    <div class="month-selector bg-white rounded-lg shadow p-4 mb-4">
      <div class="flex items-center justify-between">
        <button 
          @click="changeMonth(-1)"
          class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
        >
          ← 上月
        </button>
        <input 
          v-model="selectedMonth" 
          type="month" 
          @change="loadData"
          class="px-3 py-1 border rounded"
        />
        <button 
          @click="changeMonth(1)"
          class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
        >
          下月 →
        </button>
      </div>
    </div>

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
    <div v-else-if="report">
      <!-- 月度汇总 -->
      <div class="summary bg-white rounded-lg shadow p-4 mb-4">
        <div class="grid grid-cols-3 gap-3">
          <div>
            <p class="text-xs text-gray-500">{{ t('reports.totalIncome') }}</p>
            <p class="text-lg font-bold text-green-500">
              ¥{{ formatAmount(report.totalIncome, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t('reports.totalExpense') }}</p>
            <p class="text-lg font-bold text-red-500">
              ¥{{ formatAmount(report.totalExpense, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t('reports.netIncome') }}</p>
            <p class="text-lg font-bold" :class="report.netIncome >= 0 ? 'text-green-500' : 'text-red-500'">
              ¥{{ formatAmount(report.netIncome, showAmount) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 分类统计 -->
      <div class="category-stats bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-lg font-semibold mb-3">{{ t('reports.categoryStats') }}</h2>
        <div v-if="report.categoryStats.length > 0" class="space-y-2">
          <div 
            v-for="cat in report.categoryStats" 
            :key="cat.categoryKey"
            class="flex items-center justify-between py-2 border-b last:border-b-0"
          >
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500">#{{ cat.rank }}</span>
              <span>{{ t(`category.${cat.categoryKey}`) }}</span>
            </div>
            <div class="flex items-center gap-3">
              <div class="w-20 bg-gray-200 rounded-full h-2">
                <div 
                  class="bg-blue-500 h-2 rounded-full" 
                  :style="{ width: `${cat.percent}%` }"
                ></div>
              </div>
              <span class="text-sm text-gray-500 w-12 text-right">{{ cat.percent }}%</span>
              <span class="font-semibold w-24 text-right">
                ¥{{ formatAmount(cat.amount, showAmount) }}
              </span>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-4 text-gray-500">
          {{ t('detail.noData') }}
        </div>
      </div>

      <!-- 每日趋势 -->
      <div class="daily-trend bg-white rounded-lg shadow p-4">
        <h2 class="text-lg font-semibold mb-3">{{ t('reports.dailyTrend') }}</h2>
        <div v-if="report.dailyTrend.length > 0" class="space-y-2">
          <div 
            v-for="day in report.dailyTrend" 
            :key="day.date"
            class="flex items-center justify-between py-2 border-b last:border-b-0"
          >
            <span class="text-sm text-gray-600">{{ day.date }}</span>
            <div class="flex gap-4">
              <span class="text-sm text-green-600">
                收 ¥{{ formatAmount(day.income, showAmount) }}
              </span>
              <span class="text-sm text-red-600">
                支 ¥{{ formatAmount(day.expense, showAmount) }}
              </span>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-4 text-gray-500">
          {{ t('detail.noData') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getMonthlyReport } from '@/api/report'
import { t, formatAmount } from '@/i18n'
import type { MonthlyReport } from '@/types'

const loading = ref(false)
const error = ref<string | null>(null)
const report = ref<MonthlyReport | null>(null)
const showAmount = ref(true)

// 默认当前月
const now = new Date()
const currentMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
const selectedMonth = ref(currentMonth)

async function loadData() {
  loading.value = true
  error.value = null
  
  try {
    report.value = await getMonthlyReport(selectedMonth.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '未知错误'
  } finally {
    loading.value = false
  }
}

function changeMonth(months: number) {
  const [year, month] = selectedMonth.value.split('-').map(Number)
  const date = new Date(year, month - 1 + months, 1)
  selectedMonth.value = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

