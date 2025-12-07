<template>
  <div class="page-container">
    <div class="page-content">
    <!-- 日期选择 -->
    <div class="date-selector bg-white rounded-lg shadow p-4 mb-4">
      <div class="flex items-center justify-between">
        <button 
          @click="changeDate(-1)"
          class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
        >
          ← {{ t('detail.prevDay') }}
        </button>
        <input 
          v-model="selectedDate" 
          type="date" 
          @change="loadData"
          class="px-3 py-1 border rounded"
        />
        <button 
          @click="changeDate(1)"
          class="px-3 py-1 text-blue-500 hover:bg-blue-50 rounded"
        >
          {{ t('detail.nextDay') }} →
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
    <div v-else-if="data">
      <!-- 当日统计 -->
      <div class="stats bg-white rounded-lg shadow p-4 mb-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-500">{{ t('detail.totalIncome') }}</p>
            <p class="text-xl font-bold text-green-500">
              ¥{{ formatAmount(data.stats.totalIncome, showAmount) }}
            </p>
          </div>
          <div>
            <p class="text-sm text-gray-500">{{ t('detail.totalExpense') }}</p>
            <p class="text-xl font-bold text-red-500">
              ¥{{ formatAmount(data.stats.totalExpense, showAmount) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 账单列表 -->
      <div v-if="data.list.length > 0" class="transactions space-y-2">
        <div 
          v-for="item in data.list" 
          :key="item.id"
          class="transaction-item bg-white rounded-lg shadow p-4"
        >
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-semibold">{{ t(`category.${item.categoryKey}`) }}</span>
                <span class="text-xs text-gray-500">{{ item.time }}</span>
              </div>
              <p v-if="item.remark" class="text-sm text-gray-600">{{ item.remark }}</p>
            </div>
            <div class="text-right">
              <p 
                class="text-lg font-bold"
                :class="{
                  'text-green-500': item.type === 'income',
                  'text-red-500': item.type === 'expense',
                  'text-blue-500': item.type === 'transfer'
                }"
              >
                {{ item.type === 'income' ? '+' : '-' }}¥{{ formatAmount(item.amount, showAmount) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 无数据 -->
      <div v-else class="text-center py-8 text-gray-500">
        {{ t('detail.noData') }}
      </div>
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
/* 页面样式已移至全局样式文件 (style.css) 的 .page-container 和 .page-content */

/* 确保日期选择器的下拉选项不影响导航栏 */
input[type="date"] {
  z-index: 1;
  position: relative;
}

/* 在移动端，date input的下拉选项可能超出容器，需要特殊处理 */
@media (max-width: 768px) {
  input[type="date"]:focus {
    z-index: 999;
  }
}
</style>

