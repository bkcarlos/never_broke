<template>
  <div class="min-h-screen flex flex-col bg-slate-50">
    <header class="px-4 pt-4 pb-3 flex items-center gap-2 bg-white shadow-sm">
      <button class="text-2xl text-slate-600" @click="router.back()">‹</button>
      <h1 class="text-lg font-semibold">
        {{ t('recurring.title', '周期收入配置') }}
      </h1>
      <div class="flex-1" />
      <button
        class="w-8 h-8 rounded-full bg-indigo-500 text-white flex items-center justify-center text-xl shadow-sm"
        @click="openCreateDialog"
        aria-label="add recurring income"
      >
        +
      </button>
    </header>

    <main class="flex-1 overflow-y-auto px-4 pt-4 pb-32 space-y-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-8">
        <p class="text-gray-500">{{ t('common.loading', '加载中...') }}</p>
      </div>

      <template v-else>
        <!-- 今日待确认收入 -->
        <section v-if="dueToday.length" class="bg-white rounded-2xl shadow-sm p-4 space-y-3">
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
              :disabled="realizingId === plan.id"
              @click="onRealize(plan)"
            >
              {{ t('recurring.confirm', '确认到账') }}
            </button>
          </div>
        </section>

        <!-- 已配置的周期收入预览区 -->
        <section v-if="activePlans.length" class="space-y-3">
          <h2 class="text-sm font-medium text-slate-800">
            {{ t('recurring.section.active', '已配置的周期收入') }}
          </h2>
          <div
            v-for="plan in activePlans"
            :key="plan.id"
            class="bg-white rounded-2xl shadow-sm p-4"
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

        <!-- 工资税前配置 & 预览到年底 -->
        <section class="bg-white rounded-2xl shadow-sm p-4 space-y-4">
          <h2 class="text-sm font-medium text-slate-800">
            {{ t('recurring.salary.preview', '工资税前配置 & 预览到年底') }}
          </h2>

          <!-- 配置表单 -->
          <div class="space-y-3">
            <div>
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.salary.preTax', '税前收入（每期）') }}
              </div>
              <input
                v-model.number="salaryPreviewForm.preTaxAmount"
                type="number"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                :placeholder="t('recurring.salary.preTaxPlaceholder')"
              />
            </div>

            <div>
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.salary.city', '纳税城市') }}
              </div>
              <select
                v-model="salaryPreviewForm.cityCode"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              >
                <option value="beijing">{{ t('recurring.city.beijing') }}</option>
                <option value="shanghai">{{ t('recurring.city.shanghai') }}</option>
                <option value="guangzhou">{{ t('recurring.city.guangzhou') }}</option>
                <option value="shenzhen">{{ t('recurring.city.shenzhen') }}</option>
              </select>
            </div>

            <div class="flex gap-3">
              <div class="flex-1">
                <div class="text-xs text-slate-500 mb-1">
                  {{ t('recurring.salary.startDate', '开始日期') }}
                </div>
                <input
                  v-model="salaryPreviewForm.startDate"
                  type="date"
                  class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                />
              </div>
              <div class="flex-1">
                <div class="text-xs text-slate-500 mb-1">
                  {{ t('recurring.salary.payDay', '发薪日（1-31）') }}
                </div>
                <input
                  v-model.number="salaryPreviewForm.payDay"
                  type="number"
                  min="1"
                  max="31"
                  class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                  placeholder="10"
                />
              </div>
            </div>

            <div>
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.frequency', '频率') }}
              </div>
              <select
                v-model="salaryPreviewForm.frequency"
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

            <button
              class="w-full py-2.5 rounded-full bg-indigo-500 text-white text-sm font-medium disabled:opacity-60"
              :disabled="previewingSalary || !canPreviewSalary"
              @click="onPreviewSalary"
            >
              {{ previewingSalary ? t('recurring.salary.generating') : t('recurring.salary.generate') }}
            </button>
          </div>

          <!-- 预览表格 -->
          <div v-if="salaryPreview" class="mt-4">
            <div class="text-xs text-slate-500 mb-2">
              {{ t('recurring.salary.yearPreview', { year: salaryPreview.year, count: salaryPreview.items.length }) }}
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-xs text-slate-600">
                <thead class="bg-slate-50">
                  <tr>
                    <th class="px-2 py-2 text-left">{{ t('recurring.salary.table.date') }}</th>
                    <th class="px-2 py-2 text-right">{{ t('recurring.salary.table.preTax') }}</th>
                    <th class="px-2 py-2 text-right">{{ t('recurring.salary.table.social') }}</th>
                    <th class="px-2 py-2 text-right">{{ t('recurring.salary.table.tax') }}</th>
                    <th class="px-2 py-2 text-right font-semibold text-emerald-600">{{ t('recurring.salary.table.postTax') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in salaryPreview.items"
                    :key="item.date"
                    class="border-b border-slate-50"
                  >
                    <td class="px-2 py-2">{{ item.date }}</td>
                    <td class="px-2 py-2 text-right">¥{{ item.preTax.toFixed(2) }}</td>
                    <td class="px-2 py-2 text-right">¥{{ item.socialSecurity.toFixed(2) }}</td>
                    <td class="px-2 py-2 text-right">¥{{ item.taxAmount.toFixed(2) }}</td>
                    <td class="px-2 py-2 text-right font-semibold text-emerald-600">
                      ¥{{ item.postTax.toFixed(2) }}
                    </td>
                  </tr>
                </tbody>
                <tfoot class="bg-slate-50 font-semibold">
                  <tr>
                    <td class="px-2 py-2">{{ t('recurring.salary.table.total') }}</td>
                    <td class="px-2 py-2 text-right">¥{{ salaryPreview.totalPreTax.toFixed(2) }}</td>
                    <td class="px-2 py-2 text-right">-</td>
                    <td class="px-2 py-2 text-right">¥{{ salaryPreview.totalTax.toFixed(2) }}</td>
                    <td class="px-2 py-2 text-right text-emerald-600">
                      ¥{{ salaryPreview.totalPostTax.toFixed(2) }}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
        </section>
      </template>
    </main>

    <!-- 新增/编辑周期收入对话框 -->
    <div
      v-if="showCreateDialog"
      class="fixed inset-0 bg-black/50 flex items-end z-50"
      @click.self="closeCreateDialog"
    >
      <div
        class="w-full bg-white rounded-t-3xl max-h-[90vh] overflow-y-auto"
        @click.stop
      >
        <div class="sticky top-0 bg-white border-b border-slate-200 px-4 py-3 flex items-center justify-between">
          <h3 class="text-lg font-semibold">
            {{ editId ? t('recurring.dialog.title.edit') : t('recurring.dialog.title.new') }}
          </h3>
          <button class="text-2xl text-slate-400" @click="closeCreateDialog">×</button>
        </div>

        <div class="p-4 space-y-4">
          <!-- 类型切换 -->
          <div>
            <div class="text-xs text-slate-500 mb-2">{{ t('recurring.type') }}</div>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="tType in ['salary', 'rent', 'side', 'other']"
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

          <!-- 工资专用：城市代码 -->
          <div v-if="formType === 'salary'">
            <div class="text-xs text-slate-500 mb-1">{{ t('recurring.cityLabel') }}</div>
            <select
              v-model="formCityCode"
              class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
            >
              <option value="beijing">{{ t('recurring.city.beijing') }}</option>
              <option value="shanghai">{{ t('recurring.city.shanghai') }}</option>
              <option value="guangzhou">{{ t('recurring.city.guangzhou') }}</option>
              <option value="shenzhen">{{ t('recurring.city.shenzhen') }}</option>
            </select>
          </div>

          <div class="pt-2 space-y-2">
            <button
              class="w-full py-2.5 rounded-full bg-indigo-500 text-white text-sm font-medium disabled:opacity-60"
              :disabled="saving"
              @click="onSavePlan"
            >
              {{ saving ? t('recurring.saving') : t('recurring.save') }}
            </button>
            <button
              class="w-full py-2.5 rounded-full bg-slate-100 text-slate-600 text-sm font-medium"
              @click="closeCreateDialog"
            >
              {{ t('recurring.cancel') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type {
  RecurringIncomePlan,
  RecurringIncomeType,
  RecurringIncomeFrequency,
  UpsertRecurringIncomePayload,
  Account,
  SalaryPreviewResponse,
} from '@/types'
import {
  listRecurringIncomePlans,
  listDueRecurringIncomePlans,
  upsertRecurringIncomePlan,
  realizeRecurringIncome,
  previewSalaryToYearEnd,
} from '@/api/recurringIncome'
import { getAccounts } from '@/api/accounts'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import { t } from '@/i18n'

const router = useRouter()

const plans = ref<RecurringIncomePlan[]>([])
const dueToday = ref<RecurringIncomePlan[]>([])
const accounts = ref<Account[]>([])
const loading = ref<boolean>(false)
const saving = ref<boolean>(false)
const realizingId = ref<string | null>(null)
const previewingSalary = ref<boolean>(false)
const salaryPreview = ref<SalaryPreviewResponse | null>(null)

// 对话框状态
const showCreateDialog = ref<boolean>(false)
const editId = ref<string | undefined>(undefined)

// 表单数据
const formType = ref<RecurringIncomeType>('salary')
const formLabel = ref<string>('')
const formAmount = ref<number>(0)
const formAccountId = ref<string>('')
const formFrequency = ref<RecurringIncomeFrequency>('monthly')
const formNextDueDate = ref<string>('')
const formCityCode = ref<string>('beijing')

// 工资预览表单
const salaryPreviewForm = ref({
  preTaxAmount: 15000,
  cityCode: 'beijing',
  startDate: '',
  payDay: 10,
  frequency: 'monthly' as RecurringIncomeFrequency,
  currency: 'CNY',
})

const activePlans = computed(() =>
  plans.value.filter((p) => p.active),
)

const canPreviewSalary = computed(() => {
  return (
    salaryPreviewForm.value.preTaxAmount > 0 &&
    salaryPreviewForm.value.cityCode !== '' &&
    salaryPreviewForm.value.startDate !== '' &&
    salaryPreviewForm.value.payDay >= 1 &&
    salaryPreviewForm.value.payDay <= 31
  )
})

function getTypeLabel(type: RecurringIncomeType): string {
  const typeKey = `recurring.type.${type}` as const
  return t(typeKey, type)
}

function getFrequencyLabel(frequency: RecurringIncomeFrequency): string {
  return t(`recurring.frequency.${frequency}`)

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
  salaryPreviewForm.value.startDate = `${yyyy}-${mm}-${dd}`
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
  dueToday.value = list
}

function openCreateDialog() {
  editId.value = undefined
  formType.value = 'salary'
  formLabel.value = ''
  formAmount.value = 0
  formAccountId.value = accounts.value.length > 0 ? accounts.value[0].id : ''
  formFrequency.value = 'monthly'
  initDateToday()
  formCityCode.value = 'beijing'
  showCreateDialog.value = true
}

function closeCreateDialog() {
  showCreateDialog.value = false
  editId.value = undefined
}

function editPlan(plan: RecurringIncomePlan) {
  editId.value = plan.id
  formType.value = plan.type
  formLabel.value = plan.label
  formAmount.value = plan.amount
  formAccountId.value = plan.accountId
  formFrequency.value = plan.frequency
  formNextDueDate.value = plan.nextDueDate
  formCityCode.value = plan.cityCode || 'beijing'
  showCreateDialog.value = true
}

async function onSavePlan() {
  if (!formAmount.value || formAmount.value <= 0) {
    alert(t('recurring.error.amount'))
    return
  }
  if (!formAccountId.value) {
    alert(t('recurring.error.account'))
    return
  }
  if (!formNextDueDate.value) {
    alert(t('recurring.error.nextDueDate'))
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
    cityCode: formType.value === 'salary' ? formCityCode.value : undefined,
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
    await loadDueToday()
  } finally {
    saving.value = false
  }
}

async function onPreviewSalary() {
  if (!canPreviewSalary.value) {
    alert(t('recurring.error.salaryConfig'))
    return
  }

  try {
    previewingSalary.value = true
    const result = await previewSalaryToYearEnd({
      preTaxAmount: salaryPreviewForm.value.preTaxAmount,
      cityCode: salaryPreviewForm.value.cityCode,
      startDate: salaryPreviewForm.value.startDate,
      payDay: salaryPreviewForm.value.payDay,
      frequency: salaryPreviewForm.value.frequency,
      currency: salaryPreviewForm.value.currency,
    })
    salaryPreview.value = result
  } catch (error) {
    alert(error instanceof Error ? error.message : t('recurring.error.preview'))
  } finally {
    previewingSalary.value = false
  }
}

async function onRealize(plan: RecurringIncomePlan) {
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
  } finally {
    realizingId.value = null
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadAccounts(), loadPlans(), loadDueToday()])
    initDateToday()
  } finally {
    loading.value = false
  }
})
</script>
