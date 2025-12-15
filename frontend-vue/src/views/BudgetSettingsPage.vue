<template>
  <div class="page-container">
    <div class="page-content">
      <!-- 顶部栏 -->
      <div class="flex items-center gap-4 mb-6">
        <button
          @click="handleBack"
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <Icon name="arrow-left" size="24px" class="text-gray-700" />
        </button>
        <h1 class="text-2xl font-bold text-gray-900">预算设置</h1>
      </div>

      <!-- 当前月份选择 -->
      <div class="bg-white rounded-lg shadow p-6 mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          选择月份
        </label>
        <input
          v-model="selectedMonth"
          type="month"
          class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          @change="handleMonthChange"
        />
      </div>

      <!-- 预算输入卡片 -->
      <div class="bg-white rounded-lg shadow p-6 mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
          当月总预算
        </label>
        <div class="flex items-center gap-2">
          <span class="text-2xl font-bold text-gray-700">¥</span>
          <input
            v-model.number="totalBudget"
            type="number"
            step="0.01"
            min="0"
            placeholder="0.00"
            class="flex-1 text-2xl font-bold border-0 border-b-2 border-gray-300 focus:border-blue-500 focus:outline-none pb-2"
            :class="{ 'border-red-500': errors.totalBudget }"
          />
        </div>
        <p v-if="errors.totalBudget" class="mt-1 text-sm text-red-500">{{ errors.totalBudget }}</p>
        <p class="mt-2 text-sm text-gray-500">
          根据你的预算，我们会在首页给出每日建议可花金额
        </p>
      </div>

      <!-- 建议每日可花 -->
      <div class="bg-blue-50 rounded-lg shadow p-6 mb-4">
        <p class="text-sm text-gray-600 mb-2">建议每日可花</p>
        <p class="text-3xl font-bold text-blue-600">
          ¥{{ dailySuggestion.toFixed(2) }}
        </p>
      </div>

      <!-- 保存按钮 -->
      <div class="mt-6">
        <button
          @click="handleSubmit"
          :disabled="saving"
          class="w-full py-3 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
        >
          {{ saving ? '保存中...' : '保存本月预算' }}
        </button>
      </div>

      <!-- 错误提示 -->
      <div v-if="submitError" class="mt-4 p-4 bg-red-50 text-red-600 rounded-lg">
        {{ submitError }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { getMonthlyBudget, saveMonthlyBudget } from '@/api/budget'
import Icon from '@/components/Icon.vue'

const router = useRouter()

// 获取当前月份 YYYY-MM
function getCurrentMonth(): string {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  return `${year}-${month}`
}

const selectedMonth = ref<string>(getCurrentMonth())
const totalBudget = ref<number>(0)
const saving = ref(false)
const submitError = ref<string | null>(null)
const errors = reactive<Record<string, string>>({})

// 计算当月天数
const daysInMonth = computed(() => {
  const [year, month] = selectedMonth.value.split('-').map(Number)
  return new Date(year, month, 0).getDate()
})

// 计算建议每日可花
const dailySuggestion = computed(() => {
  if (totalBudget.value <= 0 || daysInMonth.value === 0) {
    return 0
  }
  return totalBudget.value / daysInMonth.value
})

// 加载预算数据
async function loadBudget() {
  try {
    const budget = await getMonthlyBudget(selectedMonth.value)
    totalBudget.value = budget.totalBudget || 0
    submitError.value = null
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : '加载预算失败'
    totalBudget.value = 0
  }
}

// 月份变化时重新加载
function handleMonthChange() {
  loadBudget()
}

// 监听月份变化
watch(selectedMonth, () => {
  loadBudget()
})

function validateForm(): boolean {
  Object.keys(errors).forEach(key => {
    errors[key] = ''
  })

  let isValid = true

  if (totalBudget.value < 0) {
    errors.totalBudget = '预算金额不能为负数'
    isValid = false
  }

  return isValid
}

function handleBack() {
  router.back()
}

async function handleSubmit() {
  if (!validateForm()) {
    return
  }

  saving.value = true
  submitError.value = null

  try {
    await saveMonthlyBudget({
      month: selectedMonth.value,
      totalBudget: totalBudget.value
    })

    alert('保存成功')
    router.back()
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : '保存预算失败'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadBudget()
})
</script>

<style scoped>
/* 页面样式已移至全局样式文件 */
</style>



