<template>
  <PageFrame :title="t('recurring.title', '固定收入')" eyebrow="RECURRING INCOME" subtitle="统一安排租金、副业和其他周期收入。" back-to="/profile" compact>
    <template #actions><button class="button primary" @click="openCreateDialog">＋ 新增固定收入</button></template>
    <div class="space-y-4">
      <p v-if="actionError" class="error-box" role="alert">{{ actionError }}</p>
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-8">
        <p class="text-gray-500">{{ t('common.loading', '加载中...') }}</p>
      </div>

      <div v-else-if="loadError" class="error-box" role="alert">{{ loadError }} <button class="button" @click="loadData">重试</button></div>
      <template v-else>
        <section v-if="!activePlans.length && !dueToday.length" class="panel empty"><strong>还没有固定收入计划</strong><p class="subtitle">添加一项周期收入，在到账时确认入账。</p><button class="button primary mt-4" @click="openCreateDialog">＋ 新增固定收入</button></section>
        <!-- 今日待确认收入 -->
        <section v-if="dueToday.length" class="panel p-4 space-y-3">
          <div class="text-sm font-medium text-slate-800">
            {{ t('recurring.dueToday.title', '今日待确认收入') }}
          </div>
          <div
            v-for="plan in dueToday"
            :key="plan.id"
            class="border border-slate-100 rounded-xl p-3 flex items-center justify-between gap-3"
          >
            <div>
              <div class="text-sm font-semibold text-slate-900">
                {{ plan.label }}
              </div>
              <div class="text-xs text-slate-500 mt-1">
                {{ t('recurring.amount', '金额') }}：¥{{ plan.amount.toFixed(2) }}
                ·
                {{ t('recurring.account', '到账账户') }}：
                {{
                  formatAccountDisplayName(
                    accounts.find((a) => a.id === plan.accountId) || {
                      id: '',
                      type: 'cash',
                      currency: 'CNY',
                      balance: 0,
                      archived: false,
                    }
                  )
                }}
              </div>
            </div>
            <button
              class="px-3 py-1.5 rounded-full bg-emerald-500 text-white text-xs font-medium disabled:opacity-60"
              :disabled="realizingId !== null"
              @click="onRealize(plan)"
            >
              {{ t('recurring.confirm', '确认到账') }}
            </button>
          </div>
        </section>

        <!-- 已配置的周期收入预览区 -->
        <section v-if="activePlans.length" class="recurring-grid">
          <div
            v-for="plan in activePlans"
            :key="plan.id"
            class="panel p-4"
          >
            <div class="flex items-center justify-between mb-3">
              <div>
                <div class="text-sm font-semibold text-slate-900">
                  {{ plan.label }}
                </div>
                <div class="text-xs text-slate-500 mt-1">
                  {{ getTypeLabel(plan.type) }}
                  · ¥{{ plan.amount.toFixed(2) }}
                  · {{ getFrequencyLabel(plan.frequency) }}
                </div>
              </div>
              <button
                class="text-xs text-indigo-500"
                @click="editPlan(plan)"
              >
                {{ t('recurring.edit') }}
              </button>
            </div>
            <!-- 未来几次收入预览 -->
            <div class="space-y-2">
              <div class="text-xs text-slate-500 mb-2">{{ t('recurring.futurePreview') }}</div>
              <div
                v-for="(item, idx) in getFutureIncomes(plan, 3)"
                :key="idx"
                class="flex items-center justify-between text-xs py-1.5 border-b border-slate-50 last:border-0"
              >
                <span class="text-slate-600">{{ item.date }}</span>
                <span class="font-medium text-emerald-600">¥{{ item.amount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </section>
      </template>
    </div>

    <!-- 新增/编辑周期收入对话框 -->
    <div
      v-if="showCreateDialog"
      class="fixed inset-0 bg-black/50 z-[100] flex flex-col justify-end items-center"
      @click.self="!saving && closeCreateDialog()"
    >
      <div
        class="w-full max-w-lg bg-white rounded-t-3xl flex flex-col"
        style="max-height: calc(100vh - 60px);"
        @click.stop
      >
        <div class="flex-shrink-0 bg-white border-b border-slate-200 px-4 py-3 flex items-center justify-between rounded-t-3xl z-10">
          <h3 class="text-lg font-semibold">
            {{ editId ? t('recurring.dialog.title.edit') : t('recurring.dialog.title.new') }}
          </h3>
          <button class="text-2xl text-slate-400" aria-label="关闭固定收入表单" :disabled="saving" @click="closeCreateDialog">×</button>
        </div>

        <div class="flex-1 min-h-0 overflow-y-auto p-4 space-y-4">
          <!-- 类型切换 -->
          <div>
            <div class="text-xs text-slate-500 mb-2">{{ t('recurring.type') }}</div>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="tType in ['rent', 'side', 'other']"
                :key="tType"
                class="py-2 rounded-xl text-xs border"
                :class="
                  formType === tType
                    ? 'bg-indigo-500 border-indigo-500 text-white'
                    : 'bg-slate-50 border-slate-200 text-slate-600'
                "
                @click="formType = tType as RecurringIncomeType"
              >
                {{ getTypeLabel(tType as RecurringIncomeType) }}
              </button>
            </div>
          </div>

          <!-- 名称 -->
          <div>
            <div class="text-xs text-slate-500 mb-1">{{ t('recurring.name') }}</div>
            <input
              v-model="formLabel"
              type="text"
              class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              :placeholder="t('recurring.namePlaceholder')"
            />
          </div>

          <!-- 金额 & 账户 -->
          <div class="flex gap-3">
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">{{ t('recurring.amountLabel') }}</div>
              <input
                v-model.number="formAmount"
                type="number"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                :placeholder="t('recurring.amountPlaceholder')"
              />
            </div>
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">{{ t('recurring.accountLabel') }}</div>
              <select
                v-model="formAccountId"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              >
                <option value="" disabled>{{ t('recurring.accountPlaceholder') }}</option>
                <option
                  v-for="acc in accounts"
                  :key="acc.id"
                  :value="acc.id"
                >
                  {{ formatAccountDisplayName(acc) }}
                </option>
              </select>
            </div>
          </div>

          <!-- 频率 & 下次日期 -->
          <div class="flex gap-3">
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">{{ t('recurring.frequencyLabel') }}</div>
              <select
                v-model="formFrequency"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              >
                <option value="monthly">{{ t('recurring.frequency.monthly') }}</option>
                <option value="bimonthly">{{ t('recurring.frequency.bimonthly') }}</option>
                <option value="quarterly">{{ t('recurring.frequency.quarterly') }}</option>
                <option value="semiannual">{{ t('recurring.frequency.semiannual') }}</option>
                <option value="weekly">{{ t('recurring.frequency.weekly') }}</option>
                <option value="yearly">{{ t('recurring.frequency.yearly') }}</option>
              </select>
            </div>
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">{{ t('recurring.nextDueDate') }}</div>
              <input
                v-model="formNextDueDate"
                type="date"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              />
            </div>
          </div>

          <p v-if="formError" class="error-box" role="alert">{{ formError }}</p>
          <div class="pt-2 pb-4 space-y-2">
            <button
              class="w-full py-2.5 rounded-full bg-indigo-500 text-white text-sm font-medium disabled:opacity-60"
              :disabled="saving"
              @click="onSavePlan"
            >
              {{ saving ? t('recurring.saving') : t('recurring.save') }}
            </button>
            <button
              class="w-full py-2.5 rounded-full bg-slate-100 text-slate-600 text-sm font-medium"
              :disabled="saving"
              @click="closeCreateDialog"
            >
              {{ t('recurring.cancel') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </PageFrame>
</template>

<script setup lang="ts">
import PageFrame from '@/components/PageFrame.vue'
import { computed, onMounted, ref } from 'vue'
import type {
  RecurringIncomePlan,
  RecurringIncomeType,
  RecurringIncomeFrequency,
  UpsertRecurringIncomePayload,
  Account,
} from '@/types'
import {
  listRecurringIncomePlans,
  listDueRecurringIncomePlans,
  upsertRecurringIncomePlan,
  realizeRecurringIncome,
} from '@/api/recurringIncome'
import { getAccounts } from '@/api/accounts'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import { t } from '@/i18n'


const plans = ref<RecurringIncomePlan[]>([])
const dueToday = ref<RecurringIncomePlan[]>([])
const accounts = ref<Account[]>([])
const loading = ref<boolean>(false)
const loadError = ref(''), formError = ref(''), actionError = ref('')
const saving = ref<boolean>(false)
const realizingId = ref<string | null>(null)

// 对话框状态
const showCreateDialog = ref<boolean>(false)
const editId = ref<string | undefined>(undefined)

// 表单数据
const formType = ref<RecurringIncomeType>('rent')
const formLabel = ref<string>('')
const formAmount = ref<number>(0)
const formAccountId = ref<string>('')
const formFrequency = ref<RecurringIncomeFrequency>('monthly')
const formNextDueDate = ref<string>('')

const activePlans = computed(() =>
  plans.value.filter((p) => p.active && p.type !== 'salary'),
)

function getTypeLabel(type: RecurringIncomeType): string {
  const typeKey = `recurring.type.${type}` as const
  return t(typeKey, type)
}

function getFrequencyLabel(frequency: RecurringIncomeFrequency): string {
  return t(`recurring.frequency.${frequency}`)
}

// 计算未来几次收入日期和金额
function getFutureIncomes(plan: RecurringIncomePlan, count: number): Array<{ date: string; amount: number }> {
  const result: Array<{ date: string; amount: number }> = []
  let currentDate = new Date(plan.nextDueDate)

  for (let i = 0; i < count; i++) {
    result.push({
      date: currentDate.toISOString().split('T')[0],
      amount: plan.amount,
    })

    // 计算下一个日期
    switch (plan.frequency) {
      case 'monthly':
        currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, currentDate.getDate())
        break
      case 'bimonthly':
        currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 2, currentDate.getDate())
        break
      case 'quarterly':
        currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 3, currentDate.getDate())
        break
      case 'semiannual':
        currentDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 6, currentDate.getDate())
        break
      case 'weekly':
        currentDate = new Date(currentDate.getTime() + 7 * 24 * 60 * 60 * 1000)
        break
      case 'yearly':
        currentDate = new Date(currentDate.getFullYear() + 1, currentDate.getMonth(), currentDate.getDate())
        break
    }
  }

  return result
}

function initDateToday() {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  formNextDueDate.value = `${yyyy}-${mm}-${dd}`
}

async function loadAccounts() {
  const list = await getAccounts()
  accounts.value = list.filter((acc) => !acc.archived)
  if (!formAccountId.value && accounts.value.length > 0) {
    formAccountId.value = accounts.value[0].id
  }
}

async function loadPlans() {
  const list = await listRecurringIncomePlans()
  plans.value = list
}

async function loadDueToday() {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  const today = `${yyyy}-${mm}-${dd}`
  const list = await listDueRecurringIncomePlans(today)
  dueToday.value = list.filter((p) => p.type !== 'salary')
}

function openCreateDialog() {
  formError.value = ''
  editId.value = undefined
  formType.value = 'rent'
  formLabel.value = ''
  formAmount.value = 0
  formAccountId.value = accounts.value.length > 0 ? accounts.value[0].id : ''
  formFrequency.value = 'monthly'
  initDateToday()
  showCreateDialog.value = true
}

function closeCreateDialog() {
  showCreateDialog.value = false
  editId.value = undefined
}

function editPlan(plan: RecurringIncomePlan) {
  formError.value = ''
  editId.value = plan.id
  formType.value = plan.type === 'salary' ? 'other' : plan.type
  formLabel.value = plan.label
  formAmount.value = plan.amount
  formAccountId.value = plan.accountId
  formFrequency.value = plan.frequency
  formNextDueDate.value = plan.nextDueDate
  showCreateDialog.value = true
}

async function onSavePlan() {
  if (saving.value) return
  formError.value = ''
  if (!Number.isFinite(formAmount.value) || formAmount.value <= 0) {
    formError.value = t('recurring.error.amount')
    return
  }
  if (!formAccountId.value) {
    formError.value = t('recurring.error.account')
    return
  }
  if (!formNextDueDate.value) {
    formError.value = t('recurring.error.nextDueDate')
    return
  }

  const payload: UpsertRecurringIncomePayload = {
    id: editId.value,
    type: formType.value,
    label: formLabel.value || getTypeLabel(formType.value),
    accountId: formAccountId.value,
    amount: formAmount.value,
    currency: 'CNY',
    frequency: formFrequency.value,
    nextDueDate: formNextDueDate.value,
    active: true,
  }

  try {
    saving.value = true
    const saved = await upsertRecurringIncomePlan(payload)
    const idx = plans.value.findIndex((p) => p.id === saved.id)
    if (idx >= 0) {
      plans.value.splice(idx, 1, saved)
    } else {
      plans.value.push(saved)
    }
    closeCreateDialog()
    await loadDueToday().catch(error => { actionError.value = error instanceof Error ? error.message : '刷新到账提醒失败' })
  } catch (error) {
    formError.value = error instanceof Error ? error.message : '保存失败，请重试'
  } finally {
    saving.value = false
  }
}

async function onRealize(plan: RecurringIncomePlan) {
  if (realizingId.value) return
  actionError.value = ''
  realizingId.value = plan.id
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  const today = `${yyyy}-${mm}-${dd}`

  try {
    const result = await realizeRecurringIncome(plan.id, { date: today })
    const idx = plans.value.findIndex((p) => p.id === result.plan.id)
    if (idx >= 0) {
      plans.value.splice(idx, 1, result.plan)
    }
    dueToday.value = dueToday.value.filter((p) => p.id !== plan.id)
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : '确认到账失败，请重试'
  } finally {
    realizingId.value = null
  }
}

async function loadData() {
  loading.value = true
  loadError.value = ''
  try {
    await Promise.all([loadAccounts(), loadPlans(), loadDueToday()])
    initDateToday()
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : '加载固定收入失败'
  } finally {
    loading.value = false
  }
}
onMounted(loadData)
</script>

<style scoped>
.recurring-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }
.recurring-grid > .panel { padding: 22px; }
@media (max-width: 640px) { .recurring-grid { grid-template-columns: minmax(0, 1fr); } }
@media (min-width: 768px) { .fixed { justify-content: center; padding: 24px; } .fixed > div { border-radius: 12px; overflow: hidden; } }
</style>
