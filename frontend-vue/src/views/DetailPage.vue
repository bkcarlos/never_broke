<template>
  <div class="detail-page">
    <!-- 日期选择器 -->
    <div class="date-selector-card">
      <button 
        @click="changeDate(-1)"
        class="nav-btn"
      >
        <span>←</span>
        <span class="btn-text">{{ t('detail.prevDay') }}</span>
      </button>
      <input 
        v-model="selectedDate" 
        type="date" 
        @change="loadData"
        class="date-input"
      />
      <button 
        @click="changeDate(1)"
        class="nav-btn"
      >
        <span class="btn-text">{{ t('detail.nextDay') }}</span>
        <span>→</span>
      </button>
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
    <div v-else-if="data" class="detail-content">
      <!-- 当日统计 -->
      <div class="stats-card">
        <div class="stats-grid">
          <div class="stat-box income">
            <p class="stat-label">{{ t('detail.totalIncome') }}</p>
            <p class="stat-amount text-green-600">
              ¥{{ formatAmount(data.stats.totalIncome, showAmount) }}
            </p>
          </div>
          <div class="stat-box expense">
            <p class="stat-label">{{ t('detail.totalExpense') }}</p>
            <p class="stat-amount text-red-600">
              ¥{{ formatAmount(data.stats.totalExpense, showAmount) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 账单列表 -->
      <div v-if="data.list.length > 0" class="transactions-list">
        <div 
          v-for="item in data.list" 
          :key="item.id"
          class="transaction-card"
        >
          <div class="transaction-content">
            <div class="transaction-info">
              <div class="transaction-header">
                <span class="transaction-category">{{ t(`category.${item.categoryKey}`) }}</span>
                <span class="transaction-time">{{ item.time }}</span>
              </div>
              <p v-if="item.remark" class="transaction-remark">{{ item.remark }}</p>
            </div>
            <div class="transaction-amount">
              <p 
                class="amount-value"
                :class="{
                  'text-green-600': item.type === 'income',
                  'text-red-600': item.type === 'expense',
                  'text-blue-600': item.type === 'transfer'
                }"
              >
                {{ item.type === 'income' ? '+' : '-' }}¥{{ formatAmount(item.amount, showAmount) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 无数据 -->
      <div v-else class="empty-state">
        <svg class="empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
        </svg>
        <p class="empty-text">{{ t('detail.noData') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDailyTransactions } from '@/api/ledger'
import { t, formatAmount } from '@/i18n'
import type { DailyTransactionResponse } from '@/types'

const loading = ref(false)
const error = ref<string | null>(null)
const data = ref<DailyTransactionResponse | null>(null)
const showAmount = ref(true)

// 默认今天
const today = new Date().toISOString().split('T')[0]
const selectedDate = ref(today)

async function loadData() {
  loading.value = true
  error.value = null
  
  try {
    data.value = await getDailyTransactions(selectedDate.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '未知错误'
  } finally {
    loading.value = false
  }
}

function changeDate(days: number) {
  const date = new Date(selectedDate.value)
  date.setDate(date.getDate() + days)
  selectedDate.value = date.toISOString().split('T')[0]
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.detail-page {
  width: 100%;
}

/* 日期选择器 */
.date-selector-card {
  background: white;
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  background: #f3f4f6;
  color: #4f46e5;
  border: none;
  border-radius: 10px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.nav-btn:hover {
  background: #e5e7eb;
  transform: scale(1.02);
}

.date-input {
  padding: 0.75rem 1.25rem;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 500;
  color: #1f2937;
  transition: all 0.2s;
  cursor: pointer;
}

.date-input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

/* 详情内容 */
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* 统计卡片 */
.stats-card {
  background: white;
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.stat-box {
  padding: 1.25rem;
  background: #f9fafb;
  border-radius: 12px;
  text-align: center;
  transition: all 0.2s;
}

.stat-box:hover {
  background: #f3f4f6;
  transform: translateY(-2px);
}

.stat-box.income {
  border-left: 4px solid #10b981;
}

.stat-box.expense {
  border-left: 4px solid #ef4444;
}

.stat-label {
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.stat-amount {
  font-size: 1.75rem;
  font-weight: 700;
}

/* 交易列表 */
.transactions-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.transaction-card {
  background: white;
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  transition: all 0.2s;
}

.transaction-card:hover {
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.transaction-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.transaction-info {
  flex: 1;
  min-width: 0;
}

.transaction-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}

.transaction-category {
  font-weight: 600;
  color: #1f2937;
  font-size: 1rem;
}

.transaction-time {
  font-size: 0.75rem;
  color: #9ca3af;
  background: #f3f4f6;
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
}

.transaction-remark {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.transaction-amount {
  flex-shrink: 0;
}

.amount-value {
  font-size: 1.5rem;
  font-weight: 700;
  white-space: nowrap;
}

/* 空状态 */
.empty-state {
  background: white;
  border-radius: 16px;
  padding: 4rem 2rem;
  text-align: center;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.empty-icon {
  width: 80px;
  height: 80px;
  color: #d1d5db;
  margin: 0 auto 1rem;
}

.empty-text {
  font-size: 1rem;
  color: #9ca3af;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .date-selector-card {
    padding: 1rem;
    flex-wrap: wrap;
  }

  .btn-text {
    display: none;
  }

  .nav-btn {
    padding: 0.625rem 1rem;
  }

  .date-input {
    flex: 1;
    min-width: 150px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .stat-amount {
    font-size: 1.5rem;
  }

  .transaction-card {
    padding: 1rem;
  }

  .transaction-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .transaction-amount {
    align-self: flex-end;
  }

  .amount-value {
    font-size: 1.25rem;
  }
}

@media (min-width: 1280px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 2rem;
  }

  .stat-amount {
    font-size: 2rem;
  }

  .amount-value {
    font-size: 1.75rem;
  }
}

@media (min-width: 1920px) {
  .date-selector-card {
    padding: 2rem;
  }

  .stats-card {
    padding: 2rem;
  }

  .transaction-card {
    padding: 2rem;
  }

  .stat-amount {
    font-size: 2.25rem;
  }
}

/* 确保日期选择器不影响导航 */
input[type="date"] {
  z-index: 1;
  position: relative;
}

input[type="date"]:focus {
  z-index: 999;
}
</style>

