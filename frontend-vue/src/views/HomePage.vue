<template>
  <div class="home-page w-full">
    <!-- PC端欢迎头部 -->
    <div class="welcome-header desktop-only">
      <div>
        <div class="welcome-title">
          <h1>{{ t('app.title') }}</h1>
          <div class="badge">
            <span>{{ t('home.keepRecording', { days: String(recordingDays) }) }}</span>
          </div>
        </div>
        <p class="greeting-text">{{ greeting }}</p>
      </div>
    </div>

    <!-- 移动端头部 -->
    <div class="mobile-header mobile-only">
      <div class="mobile-header-content">
        <div class="flex items-center gap-3 mb-1">
          <h1 class="text-3xl font-bold">NeverBroke</h1>
          <div class="inline-block bg-white/20 backdrop-blur-sm px-3 py-1 rounded-full">
            <p class="text-xs text-white/90">{{ t('home.keepRecording', { days: String(recordingDays) }) }}</p>
          </div>
        </div>
        <p class="text-sm text-white/80">{{ greeting }}</p>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex-1 flex items-center justify-center page-content px-4">
      <p class="text-white/90">{{ t('common.loading') }}</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="flex-1 flex flex-col items-center justify-center page-content px-4">
      <p class="text-white mb-4">{{ t('common.error') }}: {{ error }}</p>
      <button 
        @click="loadData" 
        class="px-6 py-2 bg-white text-indigo-600 rounded-lg hover:bg-white/90 transition-colors font-semibold"
      >
        {{ t('common.retry') }}
      </button>
    </div>

    <!-- 主内容区 -->
    <div v-else class="content-grid">
      <!-- PC端：第一行三列布局 -->
      <div class="stats-row">
        <!-- 卡片 1: 资产总览 -->
        <div class="card asset-card">
          <h2 class="card-title">{{ t('home.assetSummary') }}</h2>
          <div class="asset-stats">
            <div class="stat-item">
              <p class="stat-label">💰 {{ t('home.totalAssets') }}</p>
              <p class="stat-value text-blue-600">
                ¥{{ formatMoney(assetSummary.totalAssets) }}
              </p>
            </div>
            <div class="stat-item">
              <p class="stat-label">💳 可用余额</p>
              <p class="stat-value text-indigo-600">
                ¥{{ formatMoney(assetSummary.availableBalance) }}
              </p>
            </div>
            <div class="stat-item">
              <p class="stat-label">💸 {{ t('home.totalLiabilities') }}</p>
              <p class="stat-value text-orange-600">
                ¥{{ formatMoney(assetSummary.totalLiabilities) }}
              </p>
            </div>
            <div class="stat-item">
              <p class="stat-label">💎 {{ t('home.netAssets') }}</p>
              <p class="stat-value text-green-600">
                ¥{{ formatMoney(assetSummary.netAssets) }}
              </p>
            </div>
          </div>
        </div>

        <!-- 卡片 2: 本月预算 -->
        <div class="card budget-card">
          <div class="card-header">
            <h2 class="card-title">{{ t('home.monthBudget') }}</h2>
            <button 
              @click="$router.push('/reports')" 
              class="text-sm text-indigo-500 hover:text-indigo-600 font-medium"
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
              <span class="font-semibold">¥{{ formatMoney(budgetStats.monthBudget) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">{{ t('home.monthUsed') }}</span>
              <span class="font-semibold text-red-500">¥{{ formatMoney(budgetStats.monthUsed) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500">{{ t('home.remaining') }}</span>
              <span class="font-semibold text-green-500">
                ¥{{ formatMoney(budgetStats.monthBudget - budgetStats.monthUsed) }}
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

        <!-- 卡片 3: 本月概览 -->
        <div class="card month-overview-card">
          <div class="mb-4">
            <div class="flex items-center justify-between mb-2">
              <h2 class="card-title mb-0">本月概览</h2>
              <span 
                class="px-3 py-1 rounded-full text-xs font-medium"
                :class="{
                  'bg-green-100 text-green-700': monthOverview.status === 'surplus',
                  'bg-red-100 text-red-700': monthOverview.status === 'deficit',
                  'bg-gray-100 text-gray-700': monthOverview.status === 'balanced'
                }"
              >
                {{ getStatusText(monthOverview.status) }}
              </span>
            </div>
            <!-- 副标题：消费建议 -->
            <p class="text-xs text-gray-500 mt-1">{{ monthOverview.suggestion }}</p>
          </div>
          
          <div class="space-y-3">
            <!-- 收支统计 -->
            <div class="flex justify-between items-center py-2 border-b border-gray-100">
              <span class="text-sm text-gray-500">📈 本月收入</span>
              <span class="font-semibold text-green-600">
                ¥{{ formatMoney(monthOverview.monthIncome) }}
              </span>
            </div>
            
            <div class="flex justify-between items-center py-2 border-b border-gray-100">
              <span class="text-sm text-gray-500">📉 本月支出</span>
              <span class="font-semibold text-red-500">
                ¥{{ formatMoney(monthOverview.monthExpense) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- PC端：第二行消费概览（全宽） -->
      <div class="spending-overview">
        <div class="card spending-card">
          <h2 class="card-title">{{ t('home.spendOverview') }}</h2>
          
          <div class="spending-content">
            <!-- 左侧：扇形图 -->
            <div class="chart-container">
              <div ref="pieChartRef" class="pie-chart"></div>
            </div>

            <!-- 右侧：TOP3 + 其他 列表 -->
            <div class="category-list">
              <div 
                v-for="(item, index) in pieChartData" 
                :key="item.name"
                class="category-item"
              >
                <div class="category-info">
                  <div 
                    class="category-dot"
                    :style="{ backgroundColor: item.color }"
                  ></div>
                  <span class="category-name">{{ item.label }}</span>
                </div>
                <div class="category-stats">
                  <p class="category-amount">¥{{ formatMoney(item.value) }}</p>
                  <p class="category-percent">{{ item.percent.toFixed(1) }}%</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import type { AssetSummary, TodayStats, MonthOverview, CategoryStat } from '@/types'
import { getAssetSummary, getBudgetStats, getMonthOverview, getCategoryStats } from '@/api/home'
import { t } from '@/i18n'
import { useUserSettingsStore } from '@/stores/userSettings'

const userSettingsStore = useUserSettingsStore()
const loading = ref<boolean>(false)
const error = ref<string | null>(null)

// 独立的数据状态
const assetSummary = ref<AssetSummary>({
  totalAssets: 0,
  availableBalance: 0,
  totalLiabilities: 0,
  netAssets: 0
})
const budgetStats = ref<TodayStats>({
  todayExpense: 0,
  todayAvailable: 0,
  monthUsed: 0,
  monthBudget: 0,
  monthNet: 0
})
const monthOverview = ref<MonthOverview>({
  monthIncome: 0,
  monthExpense: 0,
  status: 'balanced',
  suggestion: ''
})
const categoryTop = ref<CategoryStat[]>([])

// 扇形图相关
const pieChartRef = ref<HTMLElement | null>(null)
let pieChartInstance: echarts.ECharts | null = null

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
  if (!budgetStats.value.monthBudget) return 0
  return (budgetStats.value.monthUsed / budgetStats.value.monthBudget) * 100
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
  if (!userSettingsStore.showAmount) return '***'
  return value.toFixed(2)
}

// 获取状态文本
function getStatusText(status: string): string {
  switch (status) {
    case 'surplus':
      return '✅ 收支盈余'
    case 'deficit':
      return '⚠️ 支出超标'
    case 'balanced':
      return '⚖️ 收支平衡'
    default:
      return '未知'
  }
}

function getCategoryDotColor(index: number): string {
  const colors = ['bg-indigo-500', 'bg-purple-500', 'bg-pink-500']
  return colors[index] || 'bg-gray-500'
}

// 扇形图数据
const pieChartData = computed(() => {
  if (!categoryTop.value || categoryTop.value.length === 0) {
    return []
  }

  const top3 = categoryTop.value.slice(0, 3)
  const top3PercentTotal = top3.reduce((sum, cat) => sum + cat.percent, 0)
  const totalExpense = budgetStats.value.monthUsed
  const top3AmountTotal = top3.reduce((sum, cat) => sum + cat.amount, 0)
  
  // 计算"其他"类别：优先使用百分比计算，如果百分比总和小于100%，则显示"其他"
  const otherPercent = Math.max(0, 100 - top3PercentTotal)
  const otherAmount = Math.max(0, totalExpense - top3AmountTotal)
  
  // 如果百分比计算出的"其他"存在，或者金额计算出的"其他"存在，都显示
  const shouldShowOther = otherPercent > 0.01 || otherAmount > 0.01

  const colors = ['#6366f1', '#a855f7', '#ec4899', '#94a3b8'] // indigo, purple, pink, gray

  const result = top3.map((cat, index) => ({
    name: cat.categoryKey,
    value: cat.amount,
    percent: cat.percent,
    color: colors[index],
    label: t(`category.${cat.categoryKey}`)
  }))

  // 如果有其他类别，添加到结果中
  if (shouldShowOther) {
    // 使用实际计算的金额，如果为0或负数，则使用百分比反推
    const finalOtherAmount = otherAmount > 0.01 
      ? otherAmount 
      : (otherPercent / 100) * totalExpense
    
    result.push({
      name: 'other',
      value: finalOtherAmount,
      percent: otherPercent,
      color: colors[3],
      label: t('category.other')
    })
  }

  return result
})

// 初始化扇形图
function initPieChart() {
  if (!pieChartRef.value || !categoryTop.value || categoryTop.value.length === 0) return

  // 如果已存在实例，先销毁
  if (pieChartInstance) {
    pieChartInstance.dispose()
  }

  // 创建新的图表实例
  pieChartInstance = echarts.init(pieChartRef.value)

  const chartData = pieChartData.value.map(item => ({
    value: item.value,
    name: item.label
  }))

  const colors = pieChartData.value.map(item => item.color)

  const option = {
    tooltip: {
      trigger: 'item',
      confine: true,
      appendToBody: false,
      position: (point: number[], params: any, dom: HTMLElement, rect: any, size: any) => {
        // 动态计算位置，确保 tooltip 不超出容器且不遮挡鼠标
        const [x, y] = point
        const viewWidth = size.viewSize[0]
        const viewHeight = size.viewSize[1]
        const boxWidth = size.contentSize[0]
        const boxHeight = size.contentSize[1]
        
        // 默认显示在鼠标右上方，避免遮挡鼠标指针
        let posX = x + 20  // 增加水平偏移，避免覆盖鼠标
        let posY = y - boxHeight - 10  // 显示在上方，避免覆盖鼠标
        
        // 如果右侧空间不足，显示在左侧
        if (posX + boxWidth > viewWidth) {
          posX = x - boxWidth - 20
        }
        
        // 如果上方空间不足，显示在下方
        if (posY < 0) {
          posY = y + 20  // 显示在鼠标下方
        }
        
        // 如果下方空间也不足，调整到中间位置
        if (posY + boxHeight > viewHeight) {
          posY = (viewHeight - boxHeight) / 2
        }
        
        // 确保不超出左边界
        if (posX < 0) {
          posX = 10
        }
        
        return [posX, posY]
      },
      formatter: (params: any) => {
        return `${params.name}<br/>¥${formatMoney(params.value)} (${params.percent}%)`
      }
    },
    legend: {
      show: false
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 4,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false
        },
        labelLine: {
          show: false
        },
        emphasis: {
          label: {
            show: false
          },
          labelLine: {
            show: false
          },
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        },
        data: chartData,
        color: colors
      }
    ]
  }

  pieChartInstance.setOption(option)
  
  // 添加窗口大小变化监听
  window.addEventListener('resize', handleResize)
}

// 监听数据变化，更新图表
watch(
  () => [categoryTop.value, pieChartData.value],
  () => {
    if (categoryTop.value && categoryTop.value.length > 0 && pieChartData.value.length > 0) {
      nextTick(() => {
        initPieChart()
      })
    }
  },
  { deep: true }
)

// 加载数据 - 使用独立 API
async function loadData() {
  loading.value = true
  error.value = null
  
  console.log('🔍 开始加载首页数据...')
  
  try {
    // 并行加载所有数据
    const [assetData, budgetData, monthData, categoryData] = await Promise.all([
      getAssetSummary(),
      getBudgetStats(),
      getMonthOverview(),
      getCategoryStats()
    ])
    
    console.log('✅ 数据加载成功')
    console.log('资产总览:', assetData)
    console.log('预算统计:', budgetData)
    console.log('本月概览:', monthData)
    console.log('分类统计:', categoryData)
    
    assetSummary.value = assetData
    budgetStats.value = budgetData
    monthOverview.value = monthData
    categoryTop.value = categoryData || []  // 如果为 null，设置为空数组
  } catch (e) {
    console.error('❌ 数据加载失败:', e)
    error.value = e instanceof Error ? e.message : t('common.error')
  } finally {
    loading.value = false
    console.log('📊 加载完成')
  }
}

// 初始化
onMounted(() => {
  // 初始化用户设置
  userSettingsStore.initFromLocalStorage()
  
  // 加载数据
  loadData()
})

// 窗口大小变化时调整图表大小
function handleResize() {
  if (pieChartInstance) {
    pieChartInstance.resize()
  }
}

// 组件卸载时销毁图表
onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (pieChartInstance) {
    pieChartInstance.dispose()
    pieChartInstance = null
  }
})
</script>

<style scoped>
.home-page {
  width: 100%;
  /* 确保没有额外的 margin 或 padding */
}

/* PC端欢迎头部 */
.welcome-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  /* margin-bottom: 2rem; */
  padding: 0 0 1.5rem 0;
}

.greeting-text {
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.5rem;
}

.welcome-title {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.welcome-title h1 {
  font-size: 2rem;
  font-weight: 700;
  color: #1f2937;
}

.badge {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
}

/* 移动端头部 */
.mobile-header {
  position: relative;
  padding: 1.5rem 1rem 1rem;
  color: white;
}

.mobile-header-content {
  padding-right: 3rem;
}

/* 内容网格 */
.content-grid {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1.5rem;
}

.spending-overview {
  width: 100%;
}

/* 卡片样式 */
.card {
  background: white;
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  transition: all 0.2s;
}

.card:hover {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  transform: translateY(-2px);
}

.card-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

/* 资产卡片 */
.asset-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.stat-item {
  text-align: center;
}

.stat-label {
  font-size: 0.75rem;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
}

/* 预算卡片 */
.budget-card .flex {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

/* 消费概览卡片 */
.spending-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: center;
}

.chart-container {
  display: flex;
  justify-content: center;
  align-items: center;
}

.pie-chart {
  width: 100%;
  max-width: 300px;
  aspect-ratio: 1;
}

.category-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.category-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: #f9fafb;
  border-radius: 10px;
  transition: all 0.2s;
}

.category-item:hover {
  background: #f3f4f6;
  transform: translateX(4px);
}

.category-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.category-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.category-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.category-stats {
  text-align: right;
}

.category-amount {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1f2937;
}

.category-percent {
  font-size: 0.75rem;
  color: #9ca3af;
}

/* ============================================
   响应式设计
   ============================================ */

/* 小手机 (< 640px) */
@media (max-width: 639px) {
  .stats-row {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .spending-content {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .chart-container {
    max-width: 280px;
    margin: 0 auto;
  }

  .asset-stats {
    grid-template-columns: 1fr;
  }
}

/* 大手机/小平板 (640px - 767px) */
@media (min-width: 640px) and (max-width: 767px) {
  .stats-row {
    grid-template-columns: 1fr;
    gap: 1.25rem;
  }
  
  .spending-content {
    grid-template-columns: 1fr;
  }

  .chart-container {
    max-width: 320px;
    margin: 0 auto 1.5rem;
  }
}

/* 平板 (768px - 1023px) */
@media (min-width: 768px) and (max-width: 1023px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }

  /* 让第三个卡片跨两列 */
  .stats-row > .card:nth-child(3) {
    grid-column: span 2;
  }
  
  .spending-content {
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }
}

/* 小笔记本 (1024px - 1279px) */
@media (min-width: 1024px) and (max-width: 1279px) {
  .stats-row {
    grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem;
  }
  
  .spending-content {
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }
}

/* 标准桌面 (1280px - 1919px) */
@media (min-width: 1280px) and (max-width: 1919px) {
  .stats-row {
    grid-template-columns: repeat(3, 1fr);
    gap: 2rem;
  }
  
  .spending-content {
    grid-template-columns: 1fr 1fr;
    gap: 2.5rem;
  }

  .pie-chart {
    max-width: 320px;
  }
}

/* 2K显示器 (1920px - 2559px) */
@media (min-width: 1920px) and (max-width: 2559px) {
  .welcome-title h1 {
    font-size: 2.5rem;
  }

  .card {
    padding: 2rem;
  }

  .card-title {
    font-size: 1.25rem;
  }

  .stat-value {
    font-size: 1.5rem;
  }

  .stats-row {
    grid-template-columns: repeat(3, 1fr);
    gap: 2.5rem;
  }
  
  .spending-content {
    grid-template-columns: 1fr 1.2fr;
    gap: 3rem;
  }

  .pie-chart {
    max-width: 360px;
  }

  .category-item {
    padding: 1rem;
  }

  .category-name {
    font-size: 1rem;
  }

  .category-amount {
    font-size: 1rem;
  }
}

/* 4K显示器 (≥ 2560px) */
@media (min-width: 2560px) {
  .welcome-title h1 {
    font-size: 3rem;
  }

  .badge {
    padding: 0.625rem 1.5rem;
    font-size: 0.875rem;
  }

  .card {
    padding: 2.5rem;
    border-radius: 1.5rem;
  }

  .card-title {
    font-size: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .stat-label {
    font-size: 1rem;
  }

  .stat-value {
    font-size: 2rem;
  }

  .stats-row {
    grid-template-columns: repeat(3, 1fr);
    gap: 3rem;
  }
  
  .spending-content {
    grid-template-columns: 1fr 1.2fr;
    gap: 4rem;
  }

  .pie-chart {
    max-width: 400px;
  }

  .category-item {
    padding: 1.25rem;
  }

  .category-name {
    font-size: 1.125rem;
  }

  .category-amount {
    font-size: 1.125rem;
  }

  .category-dot {
    width: 16px;
    height: 16px;
  }
}

/* 移动端 (< 768px) */
@media (max-width: 767px) {
  .desktop-only {
    display: none !important;
  }

  .mobile-only {
    display: block !important;
  }

  .content-grid {
    gap: 1rem;
  }

  .card {
    padding: 1rem;
  }

  .stat-value {
    font-size: 1rem;
  }
}

/* 桌面端 (≥ 768px) */
@media (min-width: 768px) {
  .mobile-only {
    display: none !important;
  }

  .desktop-only {
    display: block !important;
  }
}
</style>
