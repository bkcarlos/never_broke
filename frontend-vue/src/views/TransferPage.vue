<template>
  <div class="min-h-screen flex flex-col bg-slate-50">
    <!-- 顶部栏 -->
    <header class="px-4 pt-4 pb-3 flex items-center gap-2 bg-white shadow-sm">
      <button class="text-2xl text-slate-600" @click="router.back()">‹</button>
      <h1 class="text-lg font-semibold">
        {{ t('transfer.title', '账户间转账') }}
      </h1>
    </header>

    <main class="flex-1 overflow-y-auto px-4 pt-4 pb-32 space-y-4">
      <!-- 金额卡片 -->
      <section class="bg-white rounded-2xl shadow-sm p-4">
        <div class="text-xs text-slate-500 mb-2">
          {{ t('transfer.amountLabel', '转账金额') }}
        </div>
        <div class="flex items-baseline gap-2">
          <span class="text-2xl text-slate-400">¥</span>
          <input
            v-model="amount"
            type="number"
            inputmode="decimal"
            class="flex-1 bg-transparent border-none outline-none text-3xl font-semibold text-slate-900 placeholder:text-slate-300"
            :placeholder="t('transfer.amountPlaceholder', '0.00')"
          />
        </div>
        <div class="mt-2 text-xs text-slate-500">
          <span v-if="fromAccount && toAccount">
            {{ t('transfer.summary', '从') }}
            <span class="font-medium">{{ fromAccount.name }}</span>
            {{ t('transfer.summaryTo', '转到') }}
            <span class="font-medium">{{ toAccount.name }}</span>
          </span>
        </div>
      </section>

      <!-- 账户选择 -->
      <section class="bg-white rounded-2xl shadow-sm p-4 space-y-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-slate-700">
            {{ t('transfer.fromAccount', '转出账户') }}
          </span>
          <span v-if="fromAccount" class="text-xs text-slate-500">
            {{ t('transfer.balance', '余额') }}：¥{{ fromAccount.balance.toFixed(2) }}
          </span>
        </div>
        <select
          v-model="fromAccountId"
          class="w-full mt-1 border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
          @change="handleFromAccountChange"
        >
          <option value="" disabled>
            {{ t('transfer.selectFrom', '请选择转出账户') }}
          </option>
          <option
            v-for="acc in accounts"
            :key="acc.id"
            :value="acc.id"
          >
            {{ acc.name }}
          </option>
        </select>

        <div class="flex justify-center py-1">
          <span class="text-slate-400 text-xl">↓</span>
        </div>

        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-slate-700">
            {{ t('transfer.toAccount', '转入账户') }}
          </span>
          <span v-if="toAccount" class="text-xs text-slate-500">
            {{ t('transfer.balance', '余额') }}：¥{{ toAccount.balance.toFixed(2) }}
          </span>
        </div>
        <select
          v-model="toAccountId"
          class="w-full mt-1 border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
          @change="handleToAccountChange"
        >
          <option value="" disabled>
            {{ t('transfer.selectTo', '请选择转入账户') }}
          </option>
          <option
            v-for="acc in accounts"
            :key="acc.id"
            :value="acc.id"
          >
            {{ acc.name }}
          </option>
        </select>
      </section>

      <!-- 日期时间 -->
      <section class="bg-white rounded-2xl shadow-sm p-4 space-y-3">
        <div class="text-sm font-medium text-slate-700">
          {{ t('transfer.time', '时间') }}
        </div>
        <div class="flex gap-3">
          <input
            v-model="date"
            type="date"
            class="flex-1 border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
          />
          <input
            v-model="time"
            type="time"
            class="w-28 border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
          />
        </div>
      </section>

      <!-- 备注 -->
      <section class="bg-white rounded-2xl shadow-sm p-4">
        <div class="text-sm font-medium text-slate-700 mb-2">
          {{ t('transfer.remark', '备注（可选）') }}
        </div>
        <input
          v-model="remark"
          type="text"
          class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
          :placeholder="t('transfer.remarkPlaceholder', '例如：还信用卡、存定期')"
        />
      </section>

      <!-- 错误提示 -->
      <p v-if="error" class="text-xs text-rose-500 px-1">
        {{ error }}
      </p>
    </main>

    <!-- 底部提交按钮 -->
    <footer
      class="fixed bottom-0 left-0 right-0 bg-white px-4 py-3 shadow-[0_-2px_6px_rgba(15,23,42,0.08)]"
    >
      <button
        class="w-full py-3 rounded-full text-white text-base font-medium bg-gradient-to-r from-indigo-500 via-purple-500 to-sky-500 disabled:opacity-60"
        :disabled="!canSubmit"
        @click="onSubmit"
      >
        {{ loading ? t('common.loading', '加载中...') : t('transfer.submit', '确认转账') }}
      </button>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '@/i18n'
import { getAccounts } from '@/api/accounts'
import { createTransaction } from '@/api/ledger'
import type { Account } from '@/types'

const router = useRouter()

const accounts = ref<Account[]>([])
const fromAccountId = ref<string>('')
const toAccountId = ref<string>('')
const amount = ref<string>('')
const remark = ref<string>('')
const date = ref<string>('') // YYYY-MM-DD
const time = ref<string>('') // HH:mm

const loading = ref<boolean>(false)
const error = ref<string | null>(null)

const fromAccount = computed(() =>
  accounts.value.find((a) => a.id === fromAccountId.value) || null,
)
const toAccount = computed(() =>
  accounts.value.find((a) => a.id === toAccountId.value) || null,
)

function initDateTime() {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  const hh = String(now.getHours()).padStart(2, '0')
  const mi = String(now.getMinutes()).padStart(2, '0')
  date.value = `${yyyy}-${mm}-${dd}`
  time.value = `${hh}:${mi}`
}

async function loadAccounts() {
  try {
    const list = await getAccounts()
    const active = list.filter((acc) => !acc.archived)
    accounts.value = active

    if (active.length > 0) {
      fromAccountId.value = active[0].id
    }
    if (active.length > 1) {
      toAccountId.value = active[1].id
    }
  } catch (e) {
    error.value = t('transfer.error.loadAccounts', '加载账户失败')
  }
}

function handleFromAccountChange() {
  // 如果转出账户和转入账户相同，自动调整转入账户
  if (fromAccountId.value === toAccountId.value) {
    const otherAccounts = accounts.value.filter((acc) => acc.id !== fromAccountId.value)
    if (otherAccounts.length > 0) {
      toAccountId.value = otherAccounts[0].id
    } else {
      toAccountId.value = ''
    }
  }
}

function handleToAccountChange() {
  // 如果转入账户和转出账户相同，自动调整转出账户
  if (toAccountId.value === fromAccountId.value) {
    const otherAccounts = accounts.value.filter((acc) => acc.id !== toAccountId.value)
    if (otherAccounts.length > 0) {
      fromAccountId.value = otherAccounts[0].id
    } else {
      fromAccountId.value = ''
    }
  }
}

onMounted(() => {
  initDateTime()
  loadAccounts()
})

const canSubmit = computed(() => {
  const a = Number(amount.value)
  return (
    !loading.value &&
    !!fromAccountId.value &&
    !!toAccountId.value &&
    fromAccountId.value !== toAccountId.value &&
    !Number.isNaN(a) &&
    a > 0
  )
})

async function onSubmit() {
  error.value = null
  const a = Number(amount.value)
  if (Number.isNaN(a) || a <= 0) {
    error.value = t('transfer.error.invalidAmount', '请输入正确的金额')
    return
  }
  if (!fromAccountId.value || !toAccountId.value || fromAccountId.value === toAccountId.value) {
    error.value = t('transfer.error.invalidAccounts', '请选择不同的转出和转入账户')
    return
  }
  if (!date.value || !time.value) {
    error.value = t('transfer.error.invalidDate', '请选择日期和时间')
    return
  }

  const occurredAt = `${date.value}T${time.value}`

  try {
    loading.value = true
    await createTransaction({
      type: 'transfer',
      amount: a,
      categoryKey: 'TRANSFER',
      remark: remark.value || undefined,
      occurredAt,
      fromAccountId: fromAccountId.value,
      toAccountId: toAccountId.value,
    })
    // 简单提示 + 返回首页
    alert(t('transfer.success', '转账成功'))
    router.push('/')
  } catch (e) {
    error.value = t('common.error', '保存失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
</style>
