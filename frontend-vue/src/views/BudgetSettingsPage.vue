<template>
  <PageFrame :title="t('budget.title')" eyebrow="MONTHLY BUDGET" subtitle="为每个月设定支出边界，让日常消费更有计划。" back-to="/profile" compact>
    <div class="form-columns">
      <form class="form-surface" @submit.prevent="handleSubmit" :aria-busy="loading">
        <h2 class="budget-section-title">预算设置</h2>
        <label class="field budget-field" for="budget-month">{{ t('budget.selectMonth') }}</label>
        <input id="budget-month" v-model="selectedMonth" type="month" required class="control budget-month" :disabled="saving" />
        <label class="field budget-field" for="budget-amount">{{ t('budget.totalBudget') }}</label>
        <div class="budget-amount"><span>¥</span><input id="budget-amount" v-model.number="totalBudget" type="number" step="0.01" min="0" required placeholder="0.00" :disabled="loading || saving" :aria-invalid="!!errors.totalBudget" aria-describedby="budget-help budget-validation" /></div>
        <p id="budget-validation" class="expense text-xs mt-2">{{ errors.totalBudget }}</p>
        <p id="budget-help" class="subtitle">{{ t('budget.description') }}</p>
        <p v-if="loading" class="notice" role="status">正在读取当月预算…</p>
        <div v-if="submitError" class="error-box" role="alert">{{ submitError }}</div>
        <div class="form-actions"><RouterLink to="/profile" class="button">返回设置</RouterLink><button type="submit" class="button primary" :disabled="saving || loading || !selectedMonth">{{ saving ? t('budget.saving') : t('budget.save') }}</button></div>
      </form>
      <aside class="form-aside">
        <p class="eyebrow">DAILY REFERENCE</p><h2>{{ t('budget.dailySuggestion') }}</h2>
        <div class="daily-amount">{{ loading ? '—' : '¥' + dailySuggestion.toFixed(2) }}</div>
        <p>按所选月份的 {{ daysInMonth || '—' }} 天平均分配预算，作为每天消费的参考。</p>
        <p>每日参考 = 月度预算 ÷ 当月天数</p>
        <RouterLink to="/insights" class="button budget-link">查看订阅与预算 →</RouterLink>
      </aside>
    </div>
  </PageFrame>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { getMonthlyBudget, saveMonthlyBudget } from '@/api/budget'
import PageFrame from '@/components/PageFrame.vue'
import { t } from '@/i18n'

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
const loading = ref(false)
let loadSequence = 0
const submitError = ref<string | null>(null)
const errors = reactive<Record<string, string>>({})

// 计算当月天数
const daysInMonth = computed(() => {
  const [year, month] = selectedMonth.value.split('-').map(Number)
  return new Date(year, month, 0).getDate()
})

// 计算建议每日可花
const dailySuggestion = computed(() => {
  if (totalBudget.value <= 0 || !Number.isFinite(daysInMonth.value) || daysInMonth.value === 0) {
    return 0
  }
  return totalBudget.value / daysInMonth.value
})

// 加载预算数据
async function loadBudget() {
  const sequence = ++loadSequence
  if (!selectedMonth.value) {
    totalBudget.value = 0
    loading.value = false
    submitError.value = null
    return
  }
  loading.value = true
  submitError.value = null
  try {
    const budget = await getMonthlyBudget(selectedMonth.value)
    if (sequence === loadSequence) totalBudget.value = budget.totalBudget || 0
  } catch (error) {
    if (sequence === loadSequence) {
      submitError.value = error instanceof Error ? error.message : t('budget.error.load')
      totalBudget.value = 0
    }
  } finally {
    if (sequence === loadSequence) loading.value = false
  }
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

  if (!Number.isFinite(totalBudget.value) || totalBudget.value < 0) {
    errors.totalBudget = t('budget.error.negative')
    isValid = false
  }

  return isValid
}


async function handleSubmit() {
  if (loading.value || saving.value || !validateForm()) {
    return
  }

  saving.value = true
  submitError.value = null

  try {
    await saveMonthlyBudget({
      month: selectedMonth.value,
      totalBudget: totalBudget.value
    })

    alert(t('budget.success'))
    router.back()
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : t('budget.error.save')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadBudget()
})
</script>

<style scoped>
.budget-section-title { font-size: 14px; font-weight: 600; padding-bottom: 20px; border-bottom: 1px solid var(--line); }
.budget-field { margin-top: 24px; margin-bottom: 10px; }
.budget-month { width: 220px; min-height: 40px; }
.budget-amount { display: flex; align-items: center; gap: 12px; padding: 14px 0; border-bottom: 2px solid #dddff0; color: var(--ink); }
.budget-amount:focus-within { border-color: var(--brand); }
.budget-amount span { font-size: 24px; color: var(--muted); }
.budget-amount input { width: 100%; font-size: 32px; font-weight: 600; background: transparent; border: 0; outline: none; font-variant-numeric: tabular-nums; }
.daily-amount { margin: 20px 0; font-size: 32px; letter-spacing: -1px; font-weight: 600; color: var(--brand); overflow-wrap: anywhere; }
.budget-link { margin-top: 24px; }
</style>
