<template>
  <div class="page-container">
    <div class="page-content">
      <!-- 头部：模式切换按钮 -->
      <div class="mb-6">
        <div class="flex gap-2 mb-4">
          <button
            @click="switchMode('expense')"
            :class="[
              'flex-1 py-3 px-4 rounded-lg font-semibold transition-all',
              mode === 'expense'
                ? 'bg-red-500 text-white shadow-md'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
          >
            {{ t('record.expense') }}
          </button>
          <button
            @click="switchMode('income')"
            :class="[
              'flex-1 py-3 px-4 rounded-lg font-semibold transition-all',
              mode === 'income'
                ? 'bg-green-500 text-white shadow-md'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
          >
            {{ t('record.income') }}
          </button>
          <button
            @click="switchMode('transfer')"
            :class="[
              'flex-1 py-3 px-4 rounded-lg font-semibold transition-all',
              mode === 'transfer'
                ? 'bg-blue-500 text-white shadow-md'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
          >
            {{ t('record.transfer') }}
          </button>
        </div>
      </div>

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
          <select
            v-model="categoryKey"
            :placeholder="t('record.categoryPlaceholder')"
            required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          >
            <option value="">{{ t('record.categoryPlaceholder') }}</option>
            <option v-for="cat in categories" :key="cat.key" :value="cat.key">
              {{ t(`category.${cat.key}`) }}
            </option>
          </select>
        </div>

        <!-- 账户选择 -->
        <div v-if="mode === 'expense' || mode === 'income'">
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
              {{ acc.name }}
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
                {{ acc.name }}
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
                {{ acc.name }}
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

        <!-- 提交按钮 -->
        <button
          type="submit"
          :disabled="submitting"
          class="w-full py-3 bg-indigo-500 text-white rounded-lg font-semibold hover:bg-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {{ submitting ? t('record.submitting') : t('record.submit') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getAccounts } from '@/api/accounts'
import { createTransaction } from '@/api/ledger'
import type { Account, AccountOption, CreateTransactionPayload } from '@/types'
import { t } from '@/i18n'

const route = useRoute()
const router = useRouter()

// 模式：expense, income, transfer
const mode = ref<'expense' | 'income' | 'transfer'>('expense')

// 切换模式时清空相关字段
function switchMode(newMode: 'expense' | 'income' | 'transfer') {
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
const accounts = ref<AccountOption[]>([])

// 提交状态
const submitting = ref<boolean>(false)

// 分类列表
const categories = computed(() => {
  if (mode.value === 'expense') {
    return [
      { key: 'food' },
      { key: 'transport' },
      { key: 'shopping' },
      { key: 'entertainment' },
      { key: 'housing' },
      { key: 'healthcare' },
      { key: 'education' },
      { key: 'other' }
    ]
  } else if (mode.value === 'income') {
    return [
      { key: 'salary' },
      { key: 'bonus' },
      { key: 'investment' },
      { key: 'other' }
    ]
  }
  return []
})

// 从路由参数初始化模式（如果存在）
function initModeFromRoute() {
  const type = route.query.type as string
  if (type === 'expense' || type === 'income' || type === 'transfer') {
    mode.value = type
  }
}

// 初始化日期时间
function initDateTime() {
  const now = new Date()
  date.value = now.toISOString().split('T')[0]
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  time.value = `${hours}:${minutes}`
}

// 加载账户列表
async function loadAccounts() {
  try {
    const list: Account[] = await getAccounts()
    accounts.value = list
      .filter((acc) => !acc.archived)
      .map((acc) => ({ id: acc.id, name: acc.name }))
    
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

// 构建请求 payload
function buildPayload(): CreateTransactionPayload {
  const [datePart] = date.value.split('T')
  const occurredAt = `${datePart}T${time.value}:00`
  
  const payload: CreateTransactionPayload = {
    type: mode.value,
    amount: amount.value,
    occurredAt,
  }

  if (remark.value) {
    payload.remark = remark.value
  }

  if (mode.value === 'expense' || mode.value === 'income') {
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

// 提交表单
async function onSubmit() {
  if (submitting.value) return

  submitting.value = true
  try {
    const payload = buildPayload()
    await createTransaction(payload)
    alert(t('record.success'))
    router.back()
  } catch (e) {
    console.error('Failed to create transaction:', e)
    alert(t('record.error'))
  } finally {
    submitting.value = false
  }
}

// 监听路由参数变化
watch(() => route.query.type, (newType) => {
  if (newType === 'expense' || newType === 'income' || newType === 'transfer') {
    switchMode(newType)
  }
}, { immediate: true })

onMounted(() => {
  initModeFromRoute()
  initDateTime()
  loadAccounts()
})
</script>

<style scoped>
/* 页面样式已移至全局样式文件 */
</style>
