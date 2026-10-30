<template>
  <div class="page-container">
    <div class="page-content">
    <label class="block mb-3">统计币种 <select v-model="currency" @change="loadData" class="border rounded p-1"><option>CNY</option><option>USD</option></select> <span class="text-xs text-gray-500">退款冲减支出，不计入收入</span></label>
    <!-- 月份选择 -->
    <div class="month-selector bg-white rounded-lg shadow p-4 mb-4">
      <div class="flex items-center justify-between">
        <button
          @click="changeMonth(-1)"
          class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
        >
          ← {{ t('reports.prevMonth') }}
        </button>
        <input
          :value="selectedMonthStr"
          type="month"
          @change="handleMonthChange"
          class="px-3 py-1 border rounded month-input"
        />
        <button
          @click="changeMonth(1)"
          class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
        >
          {{ t('reports.nextMonth') }} →
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
      <!-- Section 1: 数字概览 -->
      <div class="summary bg-white rounded-lg shadow p-4 mb-4">
        <div class="grid grid-cols-3 gap-3">
          <div>
            <p class="text-xs text-gray-500">{{ t('reports.totalIncome') }}</p>
            <p class="text-lg font-bold text-green-500">
              {{ currency === 'USD' ? 'US$' : '¥' }}{{ formatAmount(report.totalIncome, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t('reports.totalExpense') }}</p>
            <p class="text-lg font-bold text-red-500">
              {{ currency === 'USD' ? 'US$' : '¥' }}{{ formatAmount(report.totalExpense, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500">{{ t('reports.netIncome') }}</p>
            <p class="text-lg font-bold" :class="report.netAmount >= 0 ? 'text-green-500' : 'text-red-500'">
              {{ currency === 'USD' ? 'US$' : '¥' }}{{ formatAmount(report.netAmount, showAmount) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Section 2: 与上月对比 -->
      <div v-if="report.compareWithPrevMonth" class="bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-2">{{ t('reports.compareWithPrevMonth', '与上月对比') }}</h2>
        <p class="text-sm" :class="report.compareWithPrevMonth.expenseChange >= 0 ? 'text-red-500' : 'text-green-500'">
          {{ compareText }}
        </p>
      </div>

      <!-- Section 3: 分类饼图 -->
      <div class="bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('reports.categoryStats') }}（本月全部 {{ report.categories?.length || 0 }} 类）</h2><p class="text-xs text-gray-500">表格含退款冲减；饼图仅展示净支出为正的分类。</p>
        <div v-if="report.categories && report.categories.length > 0" class="flex flex-col items-center">
          <canvas ref="pieCanvas" class="w-full h-64 mb-4"></canvas>
          <div class="w-full space-y-2">
            <div
              v-for="(cat, index) in report.categories"
              :key="cat.categoryKey"
              class="flex items-center justify-between py-2 border-b last:border-b-0"
            >
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-500">#{{ index + 1 }}</span>
                <span>{{ t(`category.${cat.categoryKey}`, cat.categoryKey) }}</span>
              </div>
              <div class="flex items-center gap-3">
                <div class="w-20 bg-gray-200 rounded-full h-2">
                  <div
                    class="bg-blue-500 h-2 rounded-full"
                    :style="{ width: `${Math.max(0, Math.min(100, cat.ratio * 100))}%` }"
                  ></div>
                </div>
                <span class="text-sm text-gray-500 w-12 text-right">{{ (cat.ratio * 100).toFixed(1) }}%</span>
                <span class="font-semibold w-24 text-right">
                  {{ currency === 'USD' ? 'US$' : '¥' }}{{ formatAmount(cat.amount, showAmount) }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-4 text-gray-500">
          {{ t('detail.noData') }}
        </div>
      </div>

      <!-- Section 4: 趋势折线图 -->
      <div class="bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('reports.dailyTrend') }}</h2>
        <canvas ref="chartCanvas" class="w-full h-64"></canvas>
      </div>

      <!-- Section 5: 文案总结 -->
      <div class="bg-white rounded-lg shadow p-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('reports.summary', '支出总结') }}</h2>
        <div v-if="report.topCategories && report.topCategories.length > 0" class="space-y-2">
          <p class="text-sm text-gray-600 mb-2">
            {{ t('reports.topCategoriesDesc', '本月前三大支出分类：') }}
          </p>
          <ul class="space-y-1">
            <li
              v-for="(cat, index) in report.topCategories"
              :key="cat.categoryKey"
              class="text-sm text-gray-700"
            >
              {{ index + 1 }}. {{ t(`category.${cat.categoryKey}`, cat.categoryKey) }} -
              {{ currency === 'USD' ? 'US$' : '¥' }}{{ formatAmount(cat.amount, showAmount) }}
              ({{ (cat.ratio * 100).toFixed(1) }}%)
            </li>
          </ul>
          <p class="text-sm text-gray-500 mt-3">
            {{ t('reports.suggestion', '建议：合理控制支出，保持收支平衡。') }}
          </p>
        </div>
        <div v-else class="text-center py-4 text-gray-500">
          {{ t('detail.noData') }}
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { fetchMonthlySummary } from '@/api/reports'
import { t, formatAmount } from '@/i18n'
import type { SummaryReport } from '@/types'
import { useUserSettingsStore } from '@/stores/userSettings'

const userSettingsStore = useUserSettingsStore()
const loading = ref(false)
const error = ref<string | null>(null)
const report = ref<SummaryReport | null>(null)
const showAmount = computed(() => userSettingsStore.showAmount)
const chartCanvas = ref<HTMLCanvasElement | null>(null)
const pieCanvas = ref<HTMLCanvasElement | null>(null)

// 默认当前月
const now = new Date()
const currency = ref('CNY')
const selectedYear = ref(now.getFullYear())
const selectedMonth = ref(now.getMonth() + 1)

// 计算月份字符串用于显示
const selectedMonthStr = computed(() => {
  return `${selectedYear.value}-${String(selectedMonth.value).padStart(2, '0')}`
})

// 上月对比文案
const compareText = computed(() => {
  if (!report.value?.compareWithPrevMonth) return ''
  const comp = report.value.compareWithPrevMonth
  const changePercent = (comp.expenseChangeRate * 100).toFixed(1)
  if (comp.expenseChange > 0) {
    return t('reports.expenseUp', { percent: changePercent }, `比上月增加 ${changePercent}%`)
  } else if (comp.expenseChange < 0) {
    return t('reports.expenseDown', { percent: Math.abs(parseFloat(changePercent)).toFixed(1) }, `比上月减少 ${Math.abs(comp.expenseChangeRate * 100).toFixed(1)}%`)
  } else {
    return t('reports.expenseSame', '与上月持平')
  }
})

async function loadData() {
  loading.value = true
  error.value = null

  try {
    report.value = await fetchMonthlySummary(selectedYear.value, selectedMonth.value, currency.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('common.unknownError', '未知错误')
  } finally {
    loading.value = false
  }
  // Canvas elements mount only after the loading branch is removed.
  if (!error.value) {
    await nextTick()
    drawTrendChart()
    drawPieChart()
  }
}

function handleMonthChange(event: Event) {
  const target = event.target as HTMLInputElement
  const [year, month] = target.value.split('-').map(Number)
  selectedYear.value = year
  selectedMonth.value = month
  loadData()
}

function changeMonth(months: number) {
  const date = new Date(selectedYear.value, selectedMonth.value - 1 + months, 1)
  selectedYear.value = date.getFullYear()
  selectedMonth.value = date.getMonth() + 1
  loadData()
}

// 绘制趋势折线图
function drawTrendChart() {
  if (!chartCanvas.value || !report.value || !report.value.trend || report.value.trend.length === 0) return

  const canvas = chartCanvas.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const width = canvas.width = canvas.offsetWidth
  const height = canvas.height = canvas.offsetHeight
  const padding = { top: 20, right: 20, bottom: 40, left: 60 }

  ctx.clearRect(0, 0, width, height)

  const trend = report.value.trend
  const maxValue = Math.max(
    ...trend.map(t => Math.max(t.income, t.expense)),
    1
  )

  const minValue = Math.min(0, ...trend.map(t => Math.min(t.income, t.expense)))
  const span = maxValue - minValue
  const chartWidth = width - padding.left - padding.right
  const chartHeight = height - padding.top - padding.bottom

  // 绘制收入折线
  ctx.strokeStyle = '#10b981'
  ctx.lineWidth = 2
  ctx.beginPath()
  trend.forEach((point, index) => {
    const x = padding.left + (index / Math.max(trend.length - 1, 1)) * chartWidth
    const y = padding.top + chartHeight - ((point.income - minValue) / span) * chartHeight
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })
  ctx.stroke()

  // 绘制支出折线
  ctx.strokeStyle = '#ef4444'
  ctx.beginPath()
  trend.forEach((point, index) => {
    const x = padding.left + (index / Math.max(trend.length - 1, 1)) * chartWidth
    const y = padding.top + chartHeight - ((point.expense - minValue) / span) * chartHeight
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })
  ctx.stroke()

  // 绘制坐标轴
  ctx.strokeStyle = '#e5e7eb'
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(padding.left, padding.top)
  ctx.lineTo(padding.left, height - padding.bottom)
  ctx.lineTo(width - padding.right, height - padding.bottom)
  ctx.stroke()
}

// 绘制分类饼图
function drawPieChart() {
  if (!pieCanvas.value || !report.value || !report.value.categories || report.value.categories.length === 0) return

  const canvas = pieCanvas.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const width = canvas.width = canvas.offsetWidth
  const height = canvas.height = canvas.offsetHeight
  const centerX = width / 2
  const centerY = height / 2
  const radius = Math.min(width, height) / 2 - 20

  ctx.clearRect(0, 0, width, height)

  const categories = report.value.categories.filter(c => c.amount > 0)
  const total = categories.reduce((sum, cat) => sum + cat.amount, 0)
  if (total === 0) return

  const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
  let currentAngle = -Math.PI / 2

  categories.forEach((cat, index) => {
    const sliceAngle = (cat.amount / total) * 2 * Math.PI
    ctx.beginPath()
    ctx.moveTo(centerX, centerY)
    ctx.arc(centerX, centerY, radius, currentAngle, currentAngle + sliceAngle)
    ctx.closePath()
    ctx.fillStyle = colors[index % colors.length]
    ctx.fill()
    currentAngle += sliceAngle
  })
}

function redrawCharts() {
  if (report.value && !loading.value) {
    nextTick(() => {
      drawTrendChart()
      drawPieChart()
    })
  }
}
onMounted(() => {
  loadData()
  window.addEventListener('resize', redrawCharts)
})
onUnmounted(() => window.removeEventListener('resize', redrawCharts))
</script>

<style scoped>
/* 页面样式已移至全局样式文件 (style.css) 的 .page-container 和 .page-content */

/* 确保月份选择器的下拉选项不影响导航栏 */
.month-input {
  z-index: 1;
  position: relative;
}

/* 在移动端，month input的下拉选项可能超出容器，需要特殊处理 */
@media (max-width: 768px) {
  .month-input:focus {
    z-index: 999;
  }
}
</style>

