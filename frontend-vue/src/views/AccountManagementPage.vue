<template>
  <div class="min-h-screen flex flex-col bg-slate-50">
    <!-- 顶部栏 -->
    <header class="px-4 pt-4 pb-3 flex items-center gap-2 bg-white shadow-sm">
      <button class="text-2xl text-slate-600" @click="router.back()">‹</button>
      <h1 class="text-lg font-semibold">
        {{ t('account.manage.title', '账户管理') }}
      </h1>
      <div class="flex-1" />
      <button
        class="w-8 h-8 rounded-full bg-indigo-500 text-white flex items-center justify-center text-xl shadow-sm"
        @click="goNewAccount"
        aria-label="add account"
      >
        +
      </button>
    </header>

    <main class="flex-1 overflow-y-auto px-4 pt-4 pb-32 space-y-4">

      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-8">
        <p class="text-gray-500">{{ t('common.loading', '加载中...') }}</p>
      </div>

      <!-- 信用卡 -->
      <section v-if="!loading && creditAccounts.length" class="space-y-2">
        <h2 class="text-xs font-medium text-slate-500">
          {{ t('account.section.credit', '信用卡') }}
        </h2>
        <div class="space-y-3">
          <div
            v-for="acc in creditAccounts"
            :key="acc.id"
            class="bg-white rounded-2xl shadow-sm p-4 space-y-2"
          >
            <div class="flex items-center justify-between">
              <div>
                <div class="text-sm font-semibold text-slate-900">
                  {{ formatAccountDisplayName(acc) }}
                </div>
                <div class="text-xs text-slate-500 mt-1">
                  {{ t('account.balance', '当前欠款') }}：
                  ¥{{ acc.balance.toFixed(2) }}
                </div>
              </div>
              <button
                class="text-xs px-2 py-1 rounded-full border border-indigo-200 text-indigo-600"
                @click="startEditLimits(acc)"
              >
                {{ t('account.credit.edit', '管理额度') }}
              </button>
            </div>

            <div class="grid grid-cols-3 gap-2 text-xs text-slate-600 mt-2">
              <div>
                <div class="text-slate-400">
                  {{ t('account.credit.total', '总额度') }}
                </div>
                <div class="font-medium">
                  ¥{{ (acc.creditLimit ?? 0).toFixed(0) }}
                </div>
              </div>
              <div>
                <div class="text-slate-400">
                  {{ t('account.credit.temp', '临时额度') }}
                </div>
                <div class="font-medium">
                  ¥{{ (acc.tempLimit ?? 0).toFixed(0) }}
                </div>
              </div>
              <div>
                <div class="text-slate-400">
                  {{ t('account.credit.available', '剩余额度') }}
                </div>
                <div class="font-medium">
                  ¥{{ (getAvailableCredit(acc) ?? 0).toFixed(0) }}
                </div>
              </div>
            </div>

            <!-- 额度编辑区域 -->
            <div
              v-if="editingId === acc.id"
              class="mt-3 pt-3 border-t border-slate-100 space-y-2"
            >
              <div class="flex gap-2">
                <div class="flex-1">
                  <div class="text-[11px] text-slate-500 mb-1">
                    {{ t('account.credit.total', '总额度') }}
                  </div>
                  <input
                    v-model="editCreditLimit"
                    type="number"
                    class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-xs bg-slate-50"
                    :placeholder="t('account.credit.totalPlaceholder', '例如 20000')"
                  />
                </div>
                <div class="flex-1">
                  <div class="text-[11px] text-slate-500 mb-1">
                    {{ t('account.credit.temp', '临时额度') }}
                  </div>
                  <input
                    v-model="editTempLimit"
                    type="number"
                    class="w-full border border-slate-200 rounded-xl px-3 py-1.5 text-xs bg-slate-50"
                    :placeholder="t('account.credit.tempPlaceholder', '没有可留空')"
                  />
                </div>
              </div>
              <div class="flex justify-end gap-2 pt-1">
                <button
                  class="text-xs px-2 py-1 rounded-full text-slate-500"
                  @click="cancelEditLimits"
                >
                  {{ t('common.cancel', '取消') }}
                </button>
                <button
                  class="text-xs px-3 py-1 rounded-full bg-indigo-500 text-white disabled:opacity-60"
                  :disabled="savingLimits"
                  @click="saveLimits"
                >
                  {{ t('common.save', '保存') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 银行卡 -->
      <section v-if="!loading && bankAccounts.length" class="space-y-2">
        <h2 class="text-xs font-medium text-slate-500">
          {{ t('account.section.bank', '银行卡') }}
        </h2>
        <div class="space-y-3">
          <div
            v-for="acc in bankAccounts"
            :key="acc.id"
            class="bg-white rounded-2xl shadow-sm p-4"
          >
            <div class="flex items-center justify-between">
              <div>
                <div class="text-sm font-semibold text-slate-900">
                  {{ formatAccountDisplayName(acc) }}
                </div>
                <div class="text-xs text-slate-500 mt-1">
                  {{ t('account.balance', '余额') }}：
                  ¥{{ acc.balance.toFixed(2) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 钱包 -->
      <section v-if="!loading && walletAccounts.length" class="space-y-2">
        <h2 class="text-xs font-medium text-slate-500">
          {{ t('account.section.wallet', '钱包') }}
        </h2>
        <div class="space-y-3">
          <div
            v-for="acc in walletAccounts"
            :key="acc.id"
            class="bg-white rounded-2xl shadow-sm p-4"
          >
            <div class="flex items-center justify-between">
              <div>
                <div class="text-sm font-semibold text-slate-900">
                  {{ formatAccountDisplayName(acc) }}
                </div>
                <div class="text-xs text-slate-500 mt-1">
                  {{ t('account.balance', '余额') }}：
                  ¥{{ acc.balance.toFixed(2) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 现金 -->
      <section v-if="!loading && cashAccounts.length" class="space-y-2">
        <h2 class="text-xs font-medium text-slate-500">
          {{ t('account.section.cash', '现金') }}
        </h2>
        <div class="space-y-3">
          <div
            v-for="acc in cashAccounts"
            :key="acc.id"
            class="bg-white rounded-2xl shadow-sm p-4"
          >
            <div class="flex items-center justify-between">
              <div>
                <div class="text-sm font-semibold text-slate-900">
                  {{ formatAccountDisplayName(acc) }}
                </div>
                <div class="text-xs text-slate-500 mt-1">
                  {{ t('account.balance', '余额') }}：
                  ¥{{ acc.balance.toFixed(2) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 空状态 -->
      <div v-if="!loading && accounts.length === 0" class="text-center py-12">
        <p class="text-sm text-slate-500 mb-4">
          {{ t('account.empty', '暂无账户') }}
        </p>
        <button
          class="px-4 py-2 bg-indigo-500 text-white rounded-lg text-sm"
          @click="goNewAccount"
        >
          {{ t('account.manage.add', '新增账户') }}
        </button>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { Account } from '@/types'
import { getAccounts, updateCreditLimits } from '@/api/accounts'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import { t } from '@/i18n'

const router = useRouter()

const accounts = ref<Account[]>([])
const loading = ref<boolean>(false)
const error = ref<string | null>(null)

// 正在编辑额度的信用卡
const editingId = ref<string | null>(null)
const editCreditLimit = ref<string>('')
const editTempLimit = ref<string>('')
const savingLimits = ref<boolean>(false)

const creditAccounts = computed(() =>
  accounts.value.filter((a) => a.type === 'credit' && !a.archived)
)
const bankAccounts = computed(() =>
  accounts.value.filter((a) => a.type === 'bank' && !a.archived)
)
const walletAccounts = computed(() =>
  accounts.value.filter((a) => a.type === 'wallet' && !a.archived)
)
const cashAccounts = computed(() =>
  accounts.value.filter((a) => a.type === 'cash' && !a.archived)
)

// 可用额度计算：available = creditLimit + tempLimit - balance
function getAvailableCredit(acc: Account): number | null {
  if (acc.type !== 'credit') return null
  const total = (acc.creditLimit ?? 0) + (acc.tempLimit ?? 0)
  const used = acc.balance // 约定：余额为已用额度
  return total - used
}

async function loadAccounts() {
  try {
    loading.value = true
    error.value = null
    const list = await getAccounts()
    accounts.value = list
  } catch (e) {
    const errorMsg = t('account.error.load', '加载账户失败')
    error.value = errorMsg
    console.error('Failed to load accounts:', e)
  } finally {
    loading.value = false
  }
}

function startEditLimits(acc: Account) {
  editingId.value = acc.id
  editCreditLimit.value = acc.creditLimit !== undefined ? String(acc.creditLimit) : ''
  editTempLimit.value = acc.tempLimit !== undefined ? String(acc.tempLimit) : ''
}

function cancelEditLimits() {
  editingId.value = null
  editCreditLimit.value = ''
  editTempLimit.value = ''
}

async function saveLimits() {
  if (!editingId.value) return
  const credit = editCreditLimit.value.trim()
  const temp = editTempLimit.value.trim()

  const payload: { creditLimit?: number; tempLimit?: number } = {}
  if (credit !== '') {
    const n = Number(credit)
    if (Number.isNaN(n) || n < 0) {
      const errorMsg = t('account.error.invalidCreditLimit', '请输入有效的总额度')
      error.value = errorMsg
      console.error(errorMsg)
      return
    }
    payload.creditLimit = n
  }
  if (temp !== '') {
    const n = Number(temp)
    if (Number.isNaN(n) || n < 0) {
      const errorMsg = t('account.error.invalidTempLimit', '请输入有效的临时额度')
      error.value = errorMsg
      console.error(errorMsg)
      return
    }
    payload.tempLimit = n
  }

  try {
    savingLimits.value = true
    error.value = null
    const updated = await updateCreditLimits(editingId.value, payload)
    const idx = accounts.value.findIndex((a) => a.id === updated.id)
    if (idx >= 0) {
      accounts.value.splice(idx, 1, updated)
    }
    cancelEditLimits()
  } catch (e) {
    const errorMsg = t('common.error', '保存失败，请稍后重试')
    error.value = errorMsg
    console.error('Failed to save limits:', e)
  } finally {
    savingLimits.value = false
  }
}

function goNewAccount() {
  router.push('/accounts/new')
}

onMounted(() => {
  loadAccounts()
})
</script>

<style scoped>
/* 样式已通过 Tailwind 类实现 */
</style>
