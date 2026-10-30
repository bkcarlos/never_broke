<template>
  <div class="page-container">
    <div class="page-content px-4">
      <!-- 顶部：返回按钮 + 标题 -->
      <div class="flex items-center gap-3 mb-4">
        <button
          @click="$router.back()"
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <Icon name="arrow-left" size="20px" />
        </button>
        <h1 class="text-xl font-semibold">{{ t('cashflow.title') }}</h1>
      </div>

      <!-- Section 1: 过滤条件 -->
      <div class="bg-white rounded-lg shadow p-4 mb-4">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('cashflow.filterTitle') }}</h2>
        
        <!-- 时间范围选择 -->
        <div class="mb-4">
          <label class="block text-xs text-gray-600 mb-2">{{ t('cashflow.timeRange') }}</label>
          <div class="flex gap-2">
            <button
              v-for="daysOption in daysOptions"
              :key="daysOption"
              @click="selectedDays = daysOption; loadForecast()"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                selectedDays === daysOption
                  ? 'bg-indigo-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ daysOption }}{{ t('cashflow.days') }}
            </button>
          </div>
        </div>

        <!-- 基准选择 -->
        <div class="mb-4">
          <label class="block text-xs text-gray-600 mb-2">{{ t('cashflow.base') }}</label>
          <div class="flex gap-2 mb-2">
            <button
              @click="base = 'total_assets'; accountId = undefined; loadForecast()"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                base === 'total_assets'
                  ? 'bg-indigo-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ t('cashflow.totalAssets') }}
            </button>
            <button
              @click="base = 'account'"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                base === 'account'
                  ? 'bg-indigo-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ t('cashflow.singleAccount') }}
            </button>
          </div>
          <select
            v-if="base === 'account'"
            v-model="accountId"
            @change="loadForecast()"
            class="w-full px-3 py-2 border rounded-lg text-sm"
          >
            <option value="">{{ t('cashflow.pleaseSelect') }}</option>
            <option
              v-for="acc in accounts"
              :key="acc.id"
              :value="acc.id"
            >
              {{ formatAccountDisplayName(acc) }}
            </option>
          </select>
        </div>

        <!-- 安全垫金额 -->
        <div class="mb-4">
          <label class="block text-xs text-gray-600 mb-2">{{ t('cashflow.safeAmount') }}</label>
          <input
            v-model.number="safeAmount"
            type="number"
            :placeholder="t('cashflow.safeAmountPlaceholder')"
            class="w-full px-3 py-2 border rounded-lg text-sm"
            @change="loadForecast()"
          />
        </div>

        <!-- 刷新按钮 -->
        <button
          @click="loadForecast()"
          :disabled="loading"
          class="w-full px-4 py-2 bg-indigo-500 text-white rounded-lg hover:bg-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? t('common.loading') : t('cashflow.refresh') }}
        </button>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading && !forecast" class="text-center py-8">
        <p class="text-gray-500">{{ t('common.loading') }}</p>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="text-center py-8">
        <p class="text-red-500 mb-4">{{ t('common.error') }}: {{ error }}</p>
        <button
          @click="loadForecast()"
          class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          {{ t('common.retry') }}
        </button>
      </div>

      <!-- 数据展示 -->
      <div v-else-if="forecast&&!forecast.balanceKnown" class="notice">余额尚未独立核验，或还款事项存在未知金额，暂不推算余额和资金风险。<RouterLink to="/accounts">核验账户</RouterLink> · <RouterLink to="/plans">查看已知还款安排</RouterLink></div>
      <div v-else-if="forecast"><p class="notice">根据已启用的周期收入和有效还款计划预测；日常消费按本月剩余预算每天 {{formatAmount(forecast.dailyBudget)}} 元估算。未设置预算时不额外预估日常消费。</p>
        <!-- Section 2: 概览卡片 -->
        <div class="bg-white rounded-lg shadow p-4 mb-4">
          <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('cashflow.overview') }}</h2>
          
          <!-- 预测状态 -->
          <div class="mb-4">
            <div
              :class="[
                'px-4 py-3 rounded-lg text-center font-semibold',
                statusClass
              ]"
            >
              <p class="text-lg">{{ statusText }}</p>
            </div>
          </div>

          <!-- 最低余额、危险天数、预警天数 -->
          <div class="grid grid-cols-3 gap-3">
            <div>
              <p class="text-xs text-gray-500 mb-1">{{ t('cashflow.minBalance') }}</p>
              <p class="text-sm font-semibold" :class="minBalanceClass">
                ¥{{ formatAmount(minBalance) }}
              </p>
              <p class="text-xs text-gray-400 mt-1">{{ minBalanceDate }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 mb-1">{{ t('cashflow.dangerDays') }}</p>
              <p class="text-sm font-semibold text-red-500">
                {{ forecast.dangerDates.length }}
              </p>
            </div>
            <div>
              <p class="text-xs text-gray-500 mb-1">{{ t('cashflow.warningDays') }}</p>
              <p class="text-sm font-semibold text-orange-500">
                {{ forecast.warningDates.length }}
              </p>
            </div>
          </div>
        </div>

        <!-- Section 3: 折线图 -->
        <div class="bg-white rounded-lg shadow p-4 mb-4">
          <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('cashflow.balanceTrend') }}</h2>
          <div class="h-64 relative">
            <canvas ref="chartCanvas" class="w-full h-full"></canvas>
          </div>
        </div>

        <!-- Section 4: 资金压力日列表 -->
        <div class="bg-white rounded-lg shadow p-4">
          <h2 class="text-sm font-semibold text-gray-700 mb-3">{{ t('cashflow.pressureDays') }}</h2>
          <div v-if="pressureDays.length === 0" class="text-center py-8 text-gray-400">
            {{ t('detail.noData') }}
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="day in pressureDays"
              :key="day.date"
              class="flex items-center justify-between p-3 border rounded-lg"
              :class="day.status === 'danger' ? 'border-red-200 bg-red-50' : 'border-orange-200 bg-orange-50'"
            >
              <div>
                <p class="text-sm font-medium">{{ day.date }}</p>
                <p class="text-xs text-gray-500 mt-1">
                  {{ t('cashflow.expectedBalance') }}: ¥{{ formatAmount(day.balance) }}
                </p>
              </div>
              <div
                :class="[
                  'px-3 py-1 rounded-full text-xs font-medium',
                  day.status === 'danger' ? 'bg-red-500 text-white' : 'bg-orange-500 text-white'
                ]"
              >
                {{ day.status === 'danger' ? t('cashflow.danger') : t('cashflow.warning') }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { fetchCashflowForecast } from '@/api/cashflow'
import { getAccounts } from '@/api/accounts'
import { t, formatAmount } from '@/i18n'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import Icon from '@/components/Icon.vue'
import type { CashflowForecast, ForecastBase, Account } from '@/types'


const daysOptions: number[] = [30, 60, 90]
const selectedDays = ref<number>(30)
const base = ref<ForecastBase>('total_assets')
const accountId = ref<string | undefined>(undefined)
const safeAmount = ref<number>(2000)
const forecast = ref<CashflowForecast | null>(null)
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const accounts = ref<Account[]>([])
const chartCanvas = ref<HTMLCanvasElement | null>(null)

// 计算今日日期
const today = new Date()
const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`

// 状态相关计算
const statusClass = computed(() => {
  if (!forecast.value) return ''
  switch (forecast.value.status) {
    case 'safe':
      return 'bg-green-100 text-green-700'
    case 'warning':
      return 'bg-orange-100 text-orange-700'
    case 'danger':
      return 'bg-red-100 text-red-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
})

const statusText = computed(() => {
  if (!forecast.value) return ''
  switch (forecast.value.status) {
    case 'safe':
      return t('cashflow.statusSafe')
    case 'warning':
      return t('cashflow.statusWarning')
    case 'danger':
      return t('cashflow.statusDanger')
    default:
      return ''
  }
})

const minBalance = computed(() => {
  if (!forecast.value || forecast.value.points.length === 0) return 0
  return Math.min(...forecast.value.points.map(p => (p.closingBalance??0)))
})

const minBalanceDate = computed(() => {
  if (!forecast.value || forecast.value.points.length === 0) return ''
  const minPoint = forecast.value.points.reduce((min, p) => 
    (p.closingBalance??0) < (min.closingBalance??0) ? p : min
  )
  return minPoint.date
})

const minBalanceClass = computed(() => {
  if (minBalance.value < 0) return 'text-red-500'
  if (minBalance.value < safeAmount.value) return 'text-orange-500'
  return 'text-green-500'
})

const pressureDays = computed(() => {
  if (!forecast.value) return []
  const allDates = [...forecast.value.dangerDates, ...forecast.value.warningDates]
  return allDates.map(date => {
    const point = forecast.value!.points.find(p => p.date === date)
    return {
      date,
      balance: point?.closingBalance || 0,
      status: forecast.value!.dangerDates.includes(date) ? 'danger' as const : 'warning' as const
    }
  }).sort((a, b) => a.date.localeCompare(b.date))
})

// 加载账户列表
async function loadAccounts() {
  try {
    accounts.value = await getAccounts()
  } catch (e) {
    console.error('Failed to load accounts:', e)
  }
}

// 加载预测数据
async function loadForecast() {
  if (base.value === 'account' && !accountId.value) {
    error.value = t('cashflow.pleaseSelectAccount')
    return
  }

  loading.value = true
  error.value = null

  try {
    forecast.value = await fetchCashflowForecast({
      startDate: todayStr,
      days: selectedDays.value,
      base: base.value,
      accountId: accountId.value,
      safeAmount: safeAmount.value,
    })
    
    // 绘制图表
    await nextTick()
    drawChart()
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('common.unknownError', '未知错误')
  } finally {
    loading.value = false
  }
}

// 绘制折线图
function drawChart() {
  if (!chartCanvas.value || !forecast.value) return

  const canvas = chartCanvas.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const width = canvas.width = canvas.offsetWidth
  const height = canvas.height = canvas.offsetHeight
  const padding = { top: 20, right: 20, bottom: 40, left: 60 }

  ctx.clearRect(0, 0, width, height)

  const points = forecast.value.points
  if (points.length === 0) return

  // 计算数据范围
  const balances = points.map(p => (p.closingBalance??0))
  const minBalance = Math.min(...balances, 0, forecast.value.safeAmount)
  const maxBalance = Math.max(...balances, forecast.value.safeAmount)
  const range = maxBalance - minBalance || 1

  // 计算坐标
  const chartWidth = width - padding.left - padding.right
  const chartHeight = height - padding.top - padding.bottom

  // 绘制安全线
  const safeY = padding.top + chartHeight - ((forecast.value.safeAmount - minBalance) / range) * chartHeight
  ctx.strokeStyle = '#f59e0b'
  ctx.lineWidth = 2
  ctx.setLineDash([5, 5])
  ctx.beginPath()
  ctx.moveTo(padding.left, safeY)
  ctx.lineTo(width - padding.right, safeY)
  ctx.stroke()
  ctx.setLineDash([])

  // 绘制安全线标签
  ctx.fillStyle = '#f59e0b'
  ctx.font = '12px sans-serif'
  ctx.fillText(t('cashflow.safeLine'), padding.left + 5, safeY - 5)

  // 绘制余额折线
  ctx.strokeStyle = '#3b82f6'
  ctx.lineWidth = 2
  ctx.beginPath()
  points.forEach((point, index) => {
    const x = padding.left + (index / (points.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (((point.closingBalance??0) - minBalance) / range) * chartHeight
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })
  ctx.stroke()

  // 绘制数据点
  ctx.fillStyle = '#3b82f6'
  points.forEach((point, index) => {
    const x = padding.left + (index / (points.length - 1)) * chartWidth
    const y = padding.top + chartHeight - (((point.closingBalance??0) - minBalance) / range) * chartHeight
    ctx.beginPath()
    ctx.arc(x, y, 3, 0, Math.PI * 2)
    ctx.fill()
  })

  // 绘制危险区域
  forecast.value.dangerDates.forEach(date => {
    const index = points.findIndex(p => p.date === date)
    if (index >= 0) {
      const x = padding.left + (index / (points.length - 1)) * chartWidth
      ctx.fillStyle = 'rgba(239, 68, 68, 0.2)'
      ctx.fillRect(x - 5, padding.top, 10, chartHeight)
    }
  })

  // 绘制坐标轴
  ctx.strokeStyle = '#e5e7eb'
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(padding.left, padding.top)
  ctx.lineTo(padding.left, height - padding.bottom)
  ctx.lineTo(width - padding.right, height - padding.bottom)
  ctx.stroke()

  // 绘制 Y 轴标签
  ctx.fillStyle = '#6b7280'
  ctx.font = '10px sans-serif'
  ctx.textAlign = 'right'
  for (let i = 0; i <= 5; i++) {
    const value = minBalance + (range / 5) * i
    const y = padding.top + chartHeight - (i / 5) * chartHeight
    ctx.fillText(`¥${formatAmount(value)}`, padding.left - 10, y + 3)
  }

  // 绘制 X 轴标签（只显示部分日期）
  ctx.textAlign = 'center'
  const labelCount = Math.min(5, points.length)
  for (let i = 0; i < labelCount; i++) {
    const index = Math.floor((i / (labelCount - 1)) * (points.length - 1))
    const x = padding.left + (index / (points.length - 1)) * chartWidth
    const date = points[index].date
    ctx.fillText(date.substring(5), x, height - padding.bottom + 15)
  }
}

// 监听窗口大小变化
watch(() => forecast.value, () => {
  if (forecast.value) {
    nextTick(() => drawChart())
  }
})

onMounted(async () => {
  await loadAccounts()
  await loadForecast()
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
