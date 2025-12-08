<template>
  <div class="min-h-screen flex flex-col bg-slate-50">
    <header class="px-4 pt-4 pb-3 flex items-center gap-2 bg-white shadow-sm">
      <button class="text-2xl text-slate-600" @click="router.back()">‹</button>
      <h1 class="text-lg font-semibold">
        {{ t('recurring.title', '周期收入配置') }}
      </h1>
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

        <!-- 已配置的计划列表：工资 -->
        <section v-if="salaryPlans.length" class="space-y-2">
          <h2 class="text-xs font-medium text-slate-500">
            {{ t('recurring.section.salary', '工资类收入') }}
          </h2>
          <div class="space-y-2">
            <div
              v-for="plan in salaryPlans"
              :key="plan.id"
              class="bg-white rounded-2xl shadow-sm p-3"
            >
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-sm font-semibold text-slate-900">
                    {{ plan.label }}
                  </div>
                  <div class="text-xs text-slate-500 mt-1">
                    ¥{{ plan.amount.toFixed(2) }}
                    · {{ t('recurring.next', '下次') }} {{ plan.nextDueDate }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 已配置的计划列表：房租 -->
        <section v-if="rentPlans.length" class="space-y-2">
          <h2 class="text-xs font-medium text-slate-500">
            {{ t('recurring.section.rent', '房租类收入') }}
          </h2>
          <div class="space-y-2">
            <div
              v-for="plan in rentPlans"
              :key="plan.id"
              class="bg-white rounded-2xl shadow-sm p-3"
            >
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-sm font-semibold text-slate-900">
                    {{ plan.label }}
                  </div>
                  <div class="text-xs text-slate-500 mt-1">
                    ¥{{ plan.amount.toFixed(2) }}
                    · {{ t('recurring.next', '下次') }} {{ plan.nextDueDate }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 已配置的计划列表：其他 -->
        <section v-if="otherPlans.length" class="space-y-2">
          <h2 class="text-xs font-medium text-slate-500">
            {{ t('recurring.section.other', '其他固定收入') }}
          </h2>
          <div class="space-y-2">
            <div
              v-for="plan in otherPlans"
              :key="plan.id"
              class="bg-white rounded-2xl shadow-sm p-3"
            >
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-sm font-semibold text-slate-900">
                    {{ plan.label }}
                  </div>
                  <div class="text-xs text-slate-500 mt-1">
                    ¥{{ plan.amount.toFixed(2) }}
                    · {{ t('recurring.next', '下次') }} {{ plan.nextDueDate }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 表单：新增/编辑 -->
        <section class="bg-white rounded-2xl shadow-sm p-4 space-y-3">
          <div class="flex items-center gap-2 text-sm font-medium text-slate-800">
            {{ t('recurring.form.title', '新增/编辑周期收入') }}
          </div>

          <!-- 类型切换 -->
          <div class="grid grid-cols-4 gap-2 mt-2">
            <button
              v-for="tType in ['salary', 'rent', 'side', 'other']"
              :key="tType"
              class="py-1.5 rounded-xl text-xs border"
              :class="
                type === tType
                  ? 'bg-indigo-500 border-indigo-500 text-white'
                  : 'bg-slate-50 border-slate-200 text-slate-600'
              "
              @click="
                () => {
                  type = tType as RecurringIncomeType
                  label = ''
                  initFormForType(type)
                }
              "
            >
              {{ t('recurring.type.' + tType, tType) }}
            </button>
          </div>

          <!-- 名称 -->
          <div>
            <div class="text-xs text-slate-500 mb-1">
              {{ t('recurring.label', '名称') }}
            </div>
            <input
              v-model="label"
              type="text"
              class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              :placeholder="t('recurring.label.placeholder', '例如：工资、房租收入')"
            />
          </div>

          <!-- 金额 & 账户 -->
          <div class="flex gap-3">
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.amount', '金额') }}
              </div>
              <input
                v-model="amount"
                type="number"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                :placeholder="t('recurring.amount.placeholder', '例如：12000')"
              />
            </div>
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.account', '到账账户') }}
              </div>
              <select
                v-model="accountId"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              >
                <option value="" disabled>
                  {{ t('recurring.account.placeholder', '请选择账户') }}
                </option>
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
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.frequency', '频率') }}
              </div>
              <select
                v-model="frequency"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              >
                <option value="monthly">
                  {{ t('recurring.frequency.monthly', '每月') }}
                </option>
                <option value="weekly">
                  {{ t('recurring.frequency.weekly', '每周') }}
                </option>
                <option value="biweekly">
                  {{ t('recurring.frequency.biweekly', '每两周') }}
                </option>
                <option value="yearly">
                  {{ t('recurring.frequency.yearly', '每年') }}
                </option>
              </select>
            </div>
            <div class="flex-1">
              <div class="text-xs text-slate-500 mb-1">
                {{ t('recurring.next', '下次收入日期') }}
              </div>
              <input
                v-model="nextDueDate"
                type="date"
                class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
              />
            </div>
          </div>

          <div class="pt-2">
            <button
              class="w-full py-2.5 rounded-full bg-indigo-500 text-white text-sm font-medium disabled:opacity-60"
              :disabled="saving"
              @click="onSave"
            >
              {{ t('recurring.save', '保存') }}
            </button>
          </div>
        </section>
      </template>
    </main>
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

const router = useRouter()

const plans = ref<RecurringIncomePlan[]>([])
const dueToday = ref<RecurringIncomePlan[]>([])
const accounts = ref<Account[]>([])
const loading = ref<boolean>(false)
const saving = ref<boolean>(false)
const realizingId = ref<string | null>(null)

// 表单
const editId = ref<string | undefined>(undefined)
const type = ref<RecurringIncomeType>('salary')
const label = ref<string>('')
const amount = ref<string>('')
const accountId = ref<string>('')
const frequency = ref<RecurringIncomeFrequency>('monthly')
const nextDueDate = ref<string>('') // YYYY-MM-DD

function initFormForType(tType: RecurringIncomeType) {
  type.value = tType
  if (!label.value) {
    if (tType === 'salary') label.value = t('recurring.salary.labelDefault', '工资')
    if (tType === 'rent') label.value = t('recurring.rent.labelDefault', '房租收入')
    if (tType === 'side') label.value = t('recurring.side.labelDefault', '副业收入')
    if (tType === 'other') label.value = t('recurring.other.labelDefault', '其他固定收入')
  }
}

function initDateToday() {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  nextDueDate.value = `${yyyy}-${mm}-${dd}`
}

async function loadAccounts() {
  const list = await getAccounts()
  accounts.value = list.filter((acc) => !acc.archived)
  if (!accountId.value && accounts.value.length > 0) {
    accountId.value = accounts.value[0].id
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

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([loadAccounts(), loadPlans(), loadDueToday()])
    initFormForType('salary')
    initDateToday()
  } finally {
    loading.value = false
  }
})

const salaryPlans = computed(() =>
  plans.value.filter((p) => p.type === 'salary' && p.active),
)
const rentPlans = computed(() =>
  plans.value.filter((p) => p.type === 'rent' && p.active),
)
const otherPlans = computed(() =>
  plans.value.filter((p) => p.type !== 'salary' && p.type !== 'rent' && p.active),
)

async function onSave() {
  const amt = Number(amount.value)
  if (!amt || amt <= 0) {
    alert(t('recurring.error.amount', '请输入正确的金额'))
    return
  }
  if (!accountId.value) {
    alert(t('recurring.error.account', '请选择收入到账账户'))
    return
  }
  if (!nextDueDate.value) {
    alert(t('recurring.error.date', '请选择下次收入日期'))
    return
  }

  const payload: UpsertRecurringIncomePayload = {
    id: editId.value,
    type: type.value,
    label: label.value || t('recurring.label.default', '固定收入'),
    accountId: accountId.value,
    amount: amt,
    currency: 'CNY',
    frequency: frequency.value,
    nextDueDate: nextDueDate.value,
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
    // 重置表单为"新增工资"
    editId.value = undefined
    amount.value = ''
    initFormForType('salary')
    initDateToday()
    await loadDueToday()
  } finally {
    saving.value = false
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
    // 从 dueToday 里移除
    dueToday.value = dueToday.value.filter((p) => p.id !== plan.id)
  } finally {
    realizingId.value = null
  }
}
</script>
