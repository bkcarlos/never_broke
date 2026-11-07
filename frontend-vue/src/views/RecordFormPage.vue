<template>
  <PageFrame title="记一笔" eyebrow="NEW TRANSACTION" subtitle="记录一笔真实收支，让账本保持完整。" back-to="/detail" back-label="返回交易账本" compact>
    <div class="form-columns">
    <div class="form-surface">
      <div class="record-modes" role="group" aria-label="交易类型">
        <button v-for="item in recordModes" :key="item.value" :class="{ selected: mode === item.value }" :aria-pressed="mode === item.value" @click="switchMode(item.value)">{{ item.label }}</button>
      </div>
      <p v-if="mode==='refund'" class="notice">输入退回金额与原消费分类。退款计入退款发生日期，冲减支出，不作为收入。</p>
      <!-- 表单 -->
      <form @submit.prevent="onSubmit" class="space-y-6">
        <!-- 金额 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('record.amount') }}
          </label>
          <input
            v-model.number="amount"
            type="number"
            step="0.01"
            min="0.01"
            :placeholder="t('record.amountPlaceholder')"
            required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </div>

        <!-- 分类（支出/收入时显示） -->
        <div v-if="mode !== 'transfer'">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('record.category') }}
          </label>
          <!-- 支出模式：快速选择按钮网格 -->
          <div v-if="(mode === 'expense' || mode === 'refund')" class="grid grid-cols-4 gap-2">
            <button
              v-for="cat in categories"
              :key="cat.key"
              type="button"
              @click="categoryKey = cat.key"
              :class="[
                'py-3 px-2 rounded-lg font-medium text-sm transition-all',
                'border-2',
                categoryKey === cat.key
                  ? 'bg-red-500 text-white border-red-500 shadow-md'
                  : 'bg-white text-gray-700 border-gray-200 hover:border-red-300 hover:bg-red-50'
              ]"
            >
              {{cat.name}}
            </button>
          </div>
          <!-- 收入模式：快速选择按钮网格 -->
          <div v-else-if="mode === 'income'" class="grid grid-cols-4 gap-2">
            <button
              v-for="cat in categories"
              :key="cat.key"
              type="button"
              @click="categoryKey = cat.key"
              :class="[
                'py-3 px-2 rounded-lg font-medium text-sm transition-all',
                'border-2',
                categoryKey === cat.key
                  ? 'bg-green-500 text-white border-green-500 shadow-md'
                  : 'bg-white text-gray-700 border-gray-200 hover:border-green-300 hover:bg-green-50'
              ]"
            >
              {{cat.name}}
            </button>
          </div>
        </div>

        <!-- 账户选择 -->
        <div v-if="mode === 'expense' || mode === 'income' || mode === 'refund'">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('record.account') }}
          </label>
          <select
            v-model="accountId"
            :placeholder="t('record.accountPlaceholder')"
            required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          >
            <option value="">{{ t('record.accountPlaceholder') }}</option>
            <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
              {{ formatAccountDisplayName(acc) }}
            </option>
          </select>
        </div>

        <!-- 转账账户选择 -->
        <div v-if="mode === 'transfer'" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('record.fromAccount') }}
            </label>
            <select
              v-model="fromAccountId"
              required
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            >
              <option value="">{{ t('record.accountPlaceholder') }}</option>
              <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                {{ formatAccountDisplayName(acc) }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('record.toAccount') }}
            </label>
            <select
              v-model="toAccountId"
              required
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            >
              <option value="">{{ t('record.accountPlaceholder') }}</option>
              <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                {{ formatAccountDisplayName(acc) }}
              </option>
            </select>
          </div>
        </div>

        <!-- 备注 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('record.remark') }}
          </label>
          <input
            v-model="remark"
            type="text"
            :placeholder="t('record.remarkPlaceholder')"
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </div>

        <!-- 日期时间 -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('record.date') }}
            </label>
            <input
              v-model="date"
              type="date"
              required
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('record.time') }}
            </label>
            <input
              v-model="time"
              type="time"
              required
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
        </div>

        <!-- 分期付款（仅支出模式） -->
        <section v-if="mode === 'expense'" class="panel p-4 space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm font-medium text-slate-800">
                {{ t('record.installment.title', '分期付款') }}
              </div>
              <div class="text-xs text-slate-500 mt-1">
                {{ t('record.installment.subtitle', '自动生成每期还款明细') }}
              </div>
            </div>
            <label class="inline-flex items-center cursor-pointer">
              <input
                v-model="isInstallment"
                type="checkbox"
                class="sr-only"
              />
              <span
                class="w-10 h-5 flex items-center bg-slate-200 rounded-full p-1 duration-200"
                :class="isInstallment ? 'bg-indigo-500' : ''"
              >
                <span
                  class="bg-white w-4 h-4 rounded-full shadow transform duration-200"
                  :class="isInstallment ? 'translate-x-5' : ''"
                />
              </span>
            </label>
          </div>

          <div v-if="isInstallment" class="space-y-3 pt-2 border-t border-slate-100">
            <div class="flex gap-3">
              <div class="flex-1">
                <div class="text-xs text-slate-500 mb-1">
                  {{ t('record.installment.term', '分期期数（月）') }}
                </div>
                <input
                  v-model="termMonths"
                  type="number"
                  min="2"
                  class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                  :placeholder="t('record.installment.termPlaceholder', '例如 12')"
                />
              </div>
              <div class="flex-1">
                <div class="text-xs text-slate-500 mb-1">
                  {{ t('record.installment.monthlyRate', '月利率（%）') }}
                </div>
                <input
                  v-model="monthlyRatePercent"
                  type="number"
                  min="0"
                  step="0.01"
                  class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                  :placeholder="t('record.installment.monthlyRatePlaceholder', '例如 0.5')"
                />
              </div>
            </div>

            <div class="flex gap-3">
              <div class="flex-1">
                <div class="text-xs text-slate-500 mb-1">
                  {{ t('record.installment.repayMethod', '还款方式') }}
                </div>
                <select
                  v-model="repayMethod"
                  class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                >
                  <option value="equal_principal">
                    {{ t('record.installment.repayMethod.equalPrincipal', '等额本金') }}
                  </option>
                  <option value="equal_payment">
                    {{ t('record.installment.repayMethod.equalPayment', '等额本息') }}
                  </option>
                </select>
              </div>
              <div class="flex-1">
                <div class="text-xs text-slate-500 mb-1">
                  {{ t('record.installment.firstDueDate', '首期扣款日') }}
                </div>
                <input
                  v-model="firstDueDate"
                  type="date"
                  class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-sm bg-slate-50"
                />
              </div>
            </div>

            <div class="text-[11px] text-slate-400">
              {{ t('record.installment.hint', '保存后会按期数自动生成每期还款明细记录，用于后续统计和现金流预测。') }}
            </div>
          </div>
        </section>

        <!-- 错误提示 -->
        <div v-if="error" class="p-4 bg-red-50 text-red-600 rounded-lg text-sm">
          {{ error }}
        </div>

        <!-- 提交按钮 -->
        <div class="form-actions"><RouterLink to="/detail" class="button">取消</RouterLink><button
          type="submit"
          :disabled="submitting"
          class="button primary"
        >
          {{ submitting ? t('record.submitting') : t('record.submit') }}
        </button></div>
      </form>
    </div>
    <aside class="form-aside"><p class="eyebrow">RECORD GUIDE</p><h2>{{ mode === 'refund' ? '退款如何记录' : mode === 'transfer' ? '账户之间的资金流动' : mode === 'income' ? '记录实际到账收入' : '记录实际发生的支出' }}</h2><p>{{ mode === 'refund' ? '退款按退回日期冲减支出。选择原消费的分类，便于后续核对。' : mode === 'transfer' ? '选择转出和转入账户，内部转账不会增加收入或支出。' : mode === 'income' ? '选择到账账户和收入分类，计划中的收入可在固定收入中管理。' : '选择消费分类和付款账户，填写日期与备注，方便之后查找。' }}</p><RouterLink to="/settings/import-bills" class="button record-import">批量导入账单 →</RouterLink></aside>
    </div>
  </PageFrame>
</template>

<script setup lang="ts">
import PageFrame from '@/components/PageFrame.vue'
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {get} from '@/api/http'
import {localDate} from '@/api/workspace'
import type {Category} from '@/api/workspace'
import { getAccounts } from '@/api/accounts'
import { createTransaction } from '@/api/ledger'
import { createExpenseInstallment } from '@/api/installments'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import type {
  Account,
  CreateTransactionPayload,
  CreateExpenseInstallmentPayload,
  InstallmentRepayMethod
} from '@/types'
import { t } from '@/i18n'

const route = useRoute()
const router = useRouter()

// 模式：expense, income, transfer
const mode = ref<'expense' | 'income' | 'transfer' | 'refund'>('expense')
const recordModes = computed(() => [
  { value: 'expense' as const, label: t('record.expense') },
  { value: 'income' as const, label: t('record.income') },
  { value: 'transfer' as const, label: t('record.transfer') },
  { value: 'refund' as const, label: '退款' },
])

// 切换模式时清空相关字段
function switchMode(newMode: 'expense' | 'income' | 'transfer' | 'refund') {
  mode.value = newMode
  // 切换模式时清空分类选择
  categoryKey.value = ''
}

// 表单数据
const amount = ref<number>(0)
const categoryKey = ref<string>('')
const accountId = ref<string>('')
const fromAccountId = ref<string>('')
const toAccountId = ref<string>('')
const remark = ref<string>('')
const date = ref<string>('')
const time = ref<string>('')

// 账户列表
const accounts = ref<Account[]>([])

// 提交状态
const submitting = ref<boolean>(false)
const error = ref<string | null>(null)

// 分期相关状态（仅支出模式）
const isInstallment = ref<boolean>(false)
const termMonths = ref<string>('3')
const monthlyRatePercent = ref<string>('0')
const repayMethod = ref<InstallmentRepayMethod>('equal_principal')
const firstDueDate = ref<string>('')

const expenseCategories = ref<{key:string;name:string}[]>([])
const categories = computed(() => mode.value === 'income'
 ? ['salary','bonus','investment','parttime','finance','rental','redpacket','other'].map(key=>({key,name:t(`category.${key}`)}))
 : expenseCategories.value)

// 从路由参数初始化模式（如果存在）
function initModeFromRoute() {
  const type = route.query.type as string
  if (type === 'expense' || type === 'income' || type === 'transfer' || type === 'refund') {
    mode.value = type
  }
}

// 初始化日期时间
function initDateTime() {
  const now = new Date()
  date.value = localDate(now)
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  time.value = `${hours}:${minutes}`
}

// 加载账户列表
async function loadAccounts() {
  try {
    const [list, dictionary] = await Promise.all([getAccounts(),get<Category[]>('/workspace/categories')])
    expenseCategories.value=dictionary.map(c=>({key:c.id,name:c.name}))
    accounts.value = list.filter((acc) => !acc.archived)
    
    // 设置默认账户
    if (accounts.value.length > 0) {
      accountId.value = accounts.value[0].id
      fromAccountId.value = accounts.value[0].id
      if (accounts.value.length > 1) {
        toAccountId.value = accounts.value[1].id
      } else {
        toAccountId.value = accounts.value[0].id
      }
    }
  } catch (e) {
    console.error('Failed to load accounts:', e)
  }
}

// 初始化首期扣款日
function initFirstDueDate() {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  firstDueDate.value = `${yyyy}-${mm}-${dd}`
}

// 构建请求 payload
function buildPayload(): CreateTransactionPayload {
  const [datePart] = date.value.split('T')
  const occurredAt = new Date(`${datePart}T${time.value}:00`).toISOString()
  
  const payload: CreateTransactionPayload = {
    type: mode.value,
    amount: amount.value,
    occurredAt,
  }

  if (remark.value) {
    payload.remark = remark.value
  }

  if (mode.value === 'expense' || mode.value === 'income' || mode.value === 'refund') {
    payload.accountId = accountId.value
    if (categoryKey.value) {
      payload.categoryKey = categoryKey.value
    }
  } else if (mode.value === 'transfer') {
    payload.fromAccountId = fromAccountId.value
    payload.toAccountId = toAccountId.value
  }

  return payload
}

// 提交普通交易
async function submitNormalTransaction() {
  error.value = null
  const payload = buildPayload()
  await createTransaction(payload)
  router.push('/detail')
}

// 提交分期支出
async function submitInstallmentExpense() {
  error.value = null

  const numericAmount = Number(amount.value)
  if (!numericAmount || numericAmount <= 0) {
    error.value = t('record.error.invalidAmount', '请输入有效金额')
    return
  }

  const term = Number(termMonths.value)
  if (!term || term <= 1) {
    error.value = t('record.error.invalidTerm', '分期期数至少为2期')
    return
  }

  const rate = Number(monthlyRatePercent.value || '0')
  if (Number.isNaN(rate) || rate < 0) {
    error.value = t('record.error.invalidRate', '请输入正确的月利率')
    return
  }

  if (!accountId.value) {
    error.value = t('record.error.invalidAccount', '请选择账户')
    return
  }

  if (!firstDueDate.value) {
    error.value = t('record.error.invalidFirstDueDate', '请选择首期扣款日')
    return
  }

  if (!date.value) {
    error.value = t('record.error.invalidDate', '请选择消费日期')
    return
  }

  const payload: CreateExpenseInstallmentPayload = {
    accountId: accountId.value,
    principal: numericAmount,
    termMonths: term,
    monthlyRatePercent: rate,
    repayMethod: repayMethod.value,
    firstDueDate: firstDueDate.value,
    currency: 'CNY',
    categoryKey: categoryKey.value || undefined,
    remark: remark.value || undefined,
    purchaseDate: date.value,
  }

  try {
    await createExpenseInstallment(payload)
    alert(t('record.success'))
    router.push('/')
  } catch (e) {
    error.value = t('common.error', '保存分期失败，请稍后重试')
  }
}

// 提交表单
async function onSubmit() {
  if (submitting.value) return

  submitting.value = true
  try {
    if (mode.value === 'expense' && isInstallment.value) {
      await submitInstallmentExpense()
    } else {
      await submitNormalTransaction()
    }
  } catch (e) {
    console.error('Failed to submit:', e)
    if (!error.value) {
      alert(t('record.error'))
    }
  } finally {
    submitting.value = false
  }
}

// 监听路由参数变化
watch(() => route.query.type, (newType) => {
  if (newType === 'expense' || newType === 'income' || newType === 'transfer' || newType === 'refund') {
    switchMode(newType)
  }
}, { immediate: true })

onMounted(() => {
  initModeFromRoute()
  initDateTime()
  initFirstDueDate()
  loadAccounts()
})
</script>

<style scoped>
.record-modes { display: flex; gap: 4px; padding: 4px; margin-bottom: 24px; background: #f3f4f8; border-radius: 8px; }
.record-modes button { flex: 1; padding: 10px 4px; font-size: 13px; color: var(--muted); border-radius: 5px; }
.record-modes button.selected { background: white; color: var(--brand); font-weight: 600; box-shadow: 0 1px 3px #242a3b12; }
.record-import { margin-top: 24px; }
@media (max-width: 420px) { form > .grid-cols-2 { grid-template-columns: minmax(0, 1fr); } }
</style>
