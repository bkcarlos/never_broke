<template>
  <div class="page-container">
    <div class="page-content">
      <!-- 顶部：返回按钮 + 标题 -->
      <div class="flex items-center gap-3 mb-4">
        <button
          @click="$router.back()"
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <Icon name="arrow-left" size="20px" />
        </button>
        <h1 class="text-xl font-semibold">{{ t('timeline.title', '时间线') }}</h1>
      </div>

      <!-- 模式切换 + 区间选择 -->
      <div class="bg-white rounded-lg shadow p-4 mb-4">
        <!-- 模式切换 -->
        <div class="mb-4">
          <label class="block text-xs text-gray-600 mb-2">{{ t('timeline.mode', '查看模式') }}</label>
          <div class="flex gap-2">
            <button
              @click="mode = 'history'; loadData()"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                mode === 'history'
                  ? 'bg-indigo-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ t('timeline.history', '仅历史') }}
            </button>
            <button
              @click="mode = 'with_future'; loadData()"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                mode === 'with_future'
                  ? 'bg-indigo-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ t('timeline.withFuture', '含未来') }}
            </button>
          </div>
        </div>

        <!-- 区间选择 -->
        <div class="flex items-center justify-between">
          <button
            @click="changeMonth(-1)"
            class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
          >
            ← {{ t('reports.prevMonth') }}
          </button>
          <span class="text-sm font-medium">
            {{ currentYear }}{{ t('common.year') }} {{ currentMonth }}{{ t('common.month') }}
          </span>
          <button
            @click="changeMonth(1)"
            class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
          >
            {{ t('reports.nextMonth') }} →
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading && !timeline" class="text-center py-8">
        <p class="text-gray-500">{{ t('common.loading') }}</p>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="text-center py-8">
        <p class="text-red-500 mb-4">{{ t('common.error') }}: {{ error }}</p>
        <button
          @click="loadData()"
          class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          {{ t('common.retry') }}
        </button>
      </div>

      <!-- 日历视图 -->
      <div v-else-if="timeline" class="bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('timeline.calendar', '日历视图') }}</h2>
        <div class="grid grid-cols-7 gap-2">
          <!-- 星期标题 -->
          <div
            v-for="day in weekDays"
            :key="day"
            class="text-center text-xs text-gray-500 font-medium py-2"
          >
            {{ day }}
          </div>
          <!-- 日期格子 -->
          <div
            v-for="day in calendarDays"
            :key="day.date"
            @click="selectDate(day.date)"
            :class="[
              'p-2 border rounded-lg cursor-pointer transition-colors',
              day.isCurrentMonth
                ? day.date === selectedDate
                  ? 'bg-indigo-100 border-indigo-500'
                  : 'hover:bg-gray-50'
                : 'opacity-40',
              day.isToday ? 'ring-2 ring-blue-500' : ''
            ]"
          >
            <div class="text-xs text-gray-600 mb-1">{{ day.day }}</div>
            <div
              v-if="day.summary"
              class="text-xs font-medium"
              :class="day.summary.net >= 0 ? 'text-green-600' : 'text-red-600'"
            >
              {{ day.summary.net >= 0 ? '+' : '' }}{{ formatAmount(day.summary.net) }}
            </div>
            <div
              v-if="day.summary?.hasFuture"
              class="w-1 h-1 bg-blue-500 rounded-full mx-auto mt-1"
            ></div>
          </div>
        </div>
      </div>

      <!-- 当日明细列表 -->
      <div v-if="selectedDate && dailyDetail" class="bg-white rounded-lg shadow p-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">
          {{ selectedDate }} {{ t('timeline.detail', '明细') }}
        </h2>
        <div v-if="dailyDetail.transactions.length === 0" class="text-center py-8 text-gray-400">
          {{ t('detail.noData') }}
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="tx in dailyDetail.transactions"
            :key="tx.id"
            class="flex items-center justify-between p-3 border rounded-lg"
            :class="isProjected(tx) ? 'bg-blue-50 border-blue-200' : ''"
          >
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <span
                  :class="[
                    'px-2 py-1 rounded text-xs font-medium',
                    tx.type === 'income' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                  ]"
                >
                  {{ tx.type === 'income' ? t('detail.totalIncome') : t('detail.totalExpense') }}
                </span>
                <span v-if="isProjected(tx)" class="text-xs text-blue-600">
                  {{ t('timeline.projected', '预计') }}
                </span>
              </div>
              <p class="text-sm text-gray-600 mt-1">{{ tx.remark || '-' }}</p>
            </div>
            <div
              class="text-lg font-semibold"
              :class="tx.type === 'income' ? 'text-green-600' : 'text-red-600'"
            >
              {{ tx.type === 'income' ? '+' : '-' }}¥{{ formatAmount(tx.amount) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { fetchDailySummary, fetchDailyDetail } from '@/api/timeline'
import { t, formatAmount } from '@/i18n'
import Icon from '@/components/Icon.vue'
import type { DailyTimeline, TimelineMode, Transaction } from '@/types'

const mode = ref<TimelineMode>('history')
const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth() + 1)
const timeline = ref<DailyTimeline | null>(null)
const dailyDetail = ref<{ date: string; mode: TimelineMode; transactions: Transaction[] } | null>(null)
const selectedDate = ref<string | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

// 计算日历天数
const calendarDays = computed(() => {
  const year = currentYear.value
  const month = currentMonth.value
  const firstDay = new Date(year, month - 1, 1)
  const lastDay = new Date(year, month, 0)
  const firstDayWeek = firstDay.getDay()
  const daysInMonth = lastDay.getDate()

  const days: Array<{
    date: string
    day: number
    isCurrentMonth: boolean
    isToday: boolean
    summary?: { net: number; hasFuture: boolean }
  }> = []

  // 上月的日期
  const prevMonthLastDay = new Date(year, month - 1, 0).getDate()
  for (let i = firstDayWeek - 1; i >= 0; i--) {
    const day = prevMonthLastDay - i
    const date = new Date(year, month - 2, day)
    days.push({
      date: formatDate(date),
      day,
      isCurrentMonth: false,
      isToday: false,
    })
  }

  // 当月的日期
  const today = new Date()
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(year, month - 1, day)
    const dateStr = formatDate(date)
    const summary = timeline.value?.days.find(d => d.date === dateStr)
    days.push({
      date: dateStr,
      day,
      isCurrentMonth: true,
      isToday: date.getTime() === new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime(),
      summary: summary ? { net: summary.net, hasFuture: summary.hasFuture } : undefined,
    })
  }

  // 下月的日期（填满42个格子）
  const remaining = 42 - days.length
  for (let day = 1; day <= remaining; day++) {
    const date = new Date(year, month, day)
    days.push({
      date: formatDate(date),
      day,
      isCurrentMonth: false,
      isToday: false,
    })
  }

  return days
})

function formatDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function changeMonth(months: number) {
  const date = new Date(currentYear.value, currentMonth.value - 1 + months, 1)
  currentYear.value = date.getFullYear()
  currentMonth.value = date.getMonth() + 1
  loadData()
}

async function loadData() {
  loading.value = true
  error.value = null

  try {
    const year = currentYear.value
    const month = currentMonth.value
    const firstDay = new Date(year, month - 1, 1)
    const lastDay = new Date(year, month, 0)
    const fromDate = formatDate(firstDay)
    const toDate = formatDate(lastDay)

    timeline.value = await fetchDailySummary(fromDate, toDate, mode.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '未知错误'
  } finally {
    loading.value = false
  }
}

async function selectDate(date: string) {
  selectedDate.value = date
  loading.value = true
  try {
    dailyDetail.value = await fetchDailyDetail(date, mode.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '未知错误'
  } finally {
    loading.value = false
  }
}

function isProjected(tx: Transaction): boolean {
  return tx.remark?.includes('预计') || false
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: linear-gradient(to bottom, #6366f1, #8b5cf6);
  padding-bottom: 80px;
}

.page-content {
  padding: 1rem;
  max-width: 100%;
}
</style>
