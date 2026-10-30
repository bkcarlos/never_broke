<template>
  <div class="page-container">
    <div class="page-content px-4">
      <!-- 顶部栏 -->
      <div class="flex items-center gap-4 mb-6">
        <button
          @click="handleBack"
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <Icon name="arrow-left" size="24px" class="text-gray-700" />
        </button>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('newAccount.title') }}</h1>
      </div>

      <!-- 表单卡片 -->
      <div class="bg-white rounded-lg shadow p-6 space-y-6">
        <!-- 账户类型 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('newAccount.type') }} <span class="text-red-500">{{ t('common.required') }}</span>
          </label>
          <div class="grid grid-cols-2 gap-3">
            <button
              v-for="type in accountTypes"
              :key="type.value"
              @click="handleTypeChange(type.value)"
              :disabled="type.value === 'cash' && hasCashAccount"
              class="px-4 py-3 border-2 rounded-lg font-medium transition-all"
              :class="[
                form.type === type.value
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-300 text-gray-700 hover:border-gray-400',
                type.value === 'cash' && hasCashAccount
                  ? 'opacity-50 cursor-not-allowed'
                  : ''
              ]"
            >
              {{ t(`account.type.${type.value}`) }}
              <span v-if="type.value === 'cash' && hasCashAccount" class="text-xs block text-gray-400">
                {{ t('newAccount.cashExists') }}
              </span>
            </button>
          </div>
          <p v-if="errors.type" class="mt-1 text-sm text-red-500">{{ errors.type }}</p>
        </div>

        <!-- 机构选择（bank/credit/wallet 必填） -->
        <div v-if="form.type === 'bank' || form.type === 'credit' || form.type === 'wallet'">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            <span v-if="form.type === 'bank' || form.type === 'credit'">{{ t('newAccount.institution') }}</span>
            <span v-else>{{ t('newAccount.platform') }}</span>
            <span class="text-red-500">{{ t('common.required') }}</span>
          </label>
          <select
            v-model="form.institution"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.institution }"
          >
            <option value="">{{ form.type === 'wallet' ? t('newAccount.selectPlatform') : t('newAccount.selectInstitution') }}</option>
            <option
              v-for="inst in availableInstitutions"
              :key="inst.value"
              :value="inst.value"
            >
              {{ t(`account.institution.${inst.value}`) }}
            </option>
          </select>
          <p v-if="errors.institution" class="mt-1 text-sm text-red-500">{{ errors.institution }}</p>
        </div>

        <!-- 卡片尾号（bank/credit 必填） -->
        <div v-if="form.type === 'bank' || form.type === 'credit'">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('newAccount.cardLast4') }} <span class="text-red-500">{{ t('common.required') }}</span>
          </label>
          <input
            v-model="form.cardLast4"
            type="text"
            maxlength="4"
            :placeholder="t('newAccount.cardLast4Placeholder')"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.cardLast4 }"
            @input="form.cardLast4 = (form.cardLast4 || '').replace(/\D/g, '')"
          />
          <p v-if="errors.cardLast4" class="mt-1 text-sm text-red-500">{{ errors.cardLast4 }}</p>
          <p class="mt-1 text-xs text-gray-500">{{ t('newAccount.cardLast4Hint') }}</p>
        </div>

        <!-- 信用卡额度（仅 credit 类型） -->
        <template v-if="form.type === 'credit'">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('newAccount.creditLimit') }}
            </label>
            <input
              v-model.number="form.creditLimit"
              type="number"
              step="0.01"
              :placeholder="t('newAccount.creditLimitPlaceholder')"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.creditLimit }"
            />
            <p v-if="errors.creditLimit" class="mt-1 text-sm text-red-500">{{ errors.creditLimit }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('newAccount.tempLimit') }}
            </label>
            <input
              v-model.number="form.tempLimit"
              type="number"
              step="0.01"
              :placeholder="t('newAccount.tempLimitHint')"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.tempLimit }"
            />
            <p v-if="errors.tempLimit" class="mt-1 text-sm text-red-500">{{ errors.tempLimit }}</p>
          </div>
        </template>

        <!-- 初始余额 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            <span v-if="form.type === 'credit'">{{ t('newAccount.usedAmount') }}</span>
            <span v-else>{{ t('newAccount.balance') }}</span>
          </label>
          <input
            v-model.number="form.balance"
            type="number"
            step="0.01"
            :placeholder="t('newAccount.balancePlaceholder')"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.balance }"
          />
          <p v-if="errors.balance" class="mt-1 text-sm text-red-500">{{ errors.balance }}</p>
          <p class="mt-1 text-xs text-gray-500">
            <span v-if="form.type === 'credit'">{{ t('newAccount.creditBalanceHint') }}</span>
            <span v-else>{{ t('newAccount.balanceHint') }}</span>
          </p>
        </div>

        <!-- 币种 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            {{ t('newAccount.currency') }}
          </label>
          <select
            v-model="form.currency"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.currency }"
          >
            <option value="CNY">CNY (人民币)</option>
            <option value="USD">USD (美元)</option>
            <option value="EUR">EUR (欧元)</option>
          </select>
          <p v-if="errors.currency" class="mt-1 text-sm text-red-500">{{ errors.currency }}</p>
        </div>

        <!-- 预览 -->
        <div class="pt-4 border-t border-gray-200">
          <p class="text-sm text-gray-600 mb-2">{{ t('newAccount.preview') }}</p>
          <p class="text-lg font-semibold text-gray-900">{{ previewName }}</p>
        </div>
      </div>

      <!-- 提交按钮 -->
      <div class="mt-6">
        <button
          @click="handleSubmit"
          :disabled="submitting"
          class="w-full py-3 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
        >
          {{ submitting ? t('newAccount.creating') : t('newAccount.create') }}
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { createAccount, getAccounts } from '@/api/accounts'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import type { AccountType, AccountInstitution, CreateAccountPayload, Account } from '@/types'
import Icon from '@/components/Icon.vue'
import { t } from '@/i18n'

const router = useRouter()

const existingAccounts = ref<Account[]>([])
const hasCashAccount = computed(() => 
  existingAccounts.value.some(acc => acc.type === 'cash' && !acc.archived)
)

const form = reactive<CreateAccountPayload & { creditLimit?: number; tempLimit?: number }>({
  type: 'cash' as AccountType,
  institution: undefined,
  cardLast4: '',
  currency: 'CNY',
  balance: 0,
  creditLimit: undefined,
  tempLimit: undefined
})

const errors = reactive<Record<string, string>>({})
const submitting = ref(false)
const submitError = ref<string | null>(null)

const accountTypes = [
  { value: 'cash' as AccountType },
  { value: 'bank' as AccountType },
  { value: 'wallet' as AccountType },
  { value: 'credit' as AccountType }
]

const bankInstitutions: Array<{ value: AccountInstitution }> = [
  { value: 'icbc' },
  { value: 'ccb' },
  { value: 'abc' },
  { value: 'boc' },
  { value: 'bocom' },
  { value: 'cmb' },
  { value: 'spdb' },
  { value: 'citic' },
  { value: 'ceb' },
  { value: 'hxb' },
  { value: 'cmbc' },
  { value: 'cgb' },
  { value: 'pab' },
  { value: 'cib' },
  { value: 'psbc' },
  { value: 'other' }
]

const walletInstitutions: Array<{ value: AccountInstitution }> = [
  { value: 'wechat' },
  { value: 'alipay' },
  { value: 'other' }
]

const availableInstitutions = computed(() => {
  if (form.type === 'bank' || form.type === 'credit') {
    return bankInstitutions
  } else if (form.type === 'wallet') {
    return walletInstitutions
  }
  return []
})

const previewAccount = computed<Account>(() => ({
  id: '',
  type: form.type,
  institution: form.institution,
  // 只有 bank 和 credit 类型才显示尾号
  cardLast4: (form.type === 'bank' || form.type === 'credit') && form.cardLast4 ? form.cardLast4 : undefined,
  currency: form.currency,
  balance: Number(form.balance) || 0,
  creditLimit: form.type === 'credit' ? form.creditLimit : undefined,
  tempLimit: form.type === 'credit' ? form.tempLimit : undefined,
  archived: false
}))

const previewName = computed(() => formatAccountDisplayName(previewAccount.value))

function validateForm(): boolean {
  // 清空之前的错误
  Object.keys(errors).forEach(key => {
    errors[key] = ''
  })

  let isValid = true

  // 校验类型
  if (!form.type) {
    errors.type = t('newAccount.error.type')
    isValid = false
  }

  // 检查现金账户是否已存在
  if (form.type === 'cash' && hasCashAccount.value) {
    errors.type = t('account.error.cashExists')
    isValid = false
  }

  // 根据类型校验必填字段
  if (form.type === 'bank' || form.type === 'credit') {
    if (!form.institution) {
      errors.institution = t('newAccount.error.institution')
      isValid = false
    }
    if (!form.cardLast4 || (form.cardLast4 || '').length < 2 || (form.cardLast4 || '').length > 4) {
      errors.cardLast4 = t('newAccount.error.cardLast4')
      isValid = false
    }
  } else if (form.type === 'wallet') {
    if (!form.institution) {
      errors.institution = t('newAccount.error.platform')
      isValid = false
    }
  }

  // 校验币种
  if (!form.currency) {
    errors.currency = t('newAccount.error.currency')
    isValid = false
  }

  // 校验余额格式
  if (isNaN(form.balance)) {
    errors.balance = t('newAccount.error.balance')
    isValid = false
  }

  // 校验信用卡额度
  if (form.type === 'credit') {
    if (form.creditLimit !== undefined && form.creditLimit < 0) {
      errors.creditLimit = t('newAccount.error.creditLimit')
      isValid = false
    }
    if (form.tempLimit !== undefined && form.tempLimit < 0) {
      errors.tempLimit = t('newAccount.error.tempLimit')
      isValid = false
    }
  }

  return isValid
}

function handleTypeChange(newType: AccountType) {
  form.type = newType
  
  // 清理不相关的字段
  if (newType === 'cash') {
    // 现金类型：不需要机构和尾号
    form.institution = undefined
    form.cardLast4 = ''
    form.creditLimit = undefined
    form.tempLimit = undefined
  } else if (newType === 'wallet') {
    // 钱包类型：需要机构（平台），不需要尾号和额度
    form.cardLast4 = ''
    form.creditLimit = undefined
    form.tempLimit = undefined
    // 如果之前选择的是银行机构，清空
    if (form.institution && !walletInstitutions.find(i => i.value === form.institution)) {
      form.institution = undefined
    }
  } else if (newType === 'bank') {
    // 银行卡：需要机构和尾号，不需要额度
    form.creditLimit = undefined
    form.tempLimit = undefined
    // 如果之前选择的是钱包平台，清空
    if (form.institution && !bankInstitutions.find(i => i.value === form.institution)) {
      form.institution = undefined
    }
  } else if (newType === 'credit') {
    // 信用卡：需要机构和尾号，可以有额度
    // 如果之前选择的是钱包平台，清空
    if (form.institution && !bankInstitutions.find(i => i.value === form.institution)) {
      form.institution = undefined
    }
  }
  
  // 清空相关错误
  errors.institution = ''
  errors.cardLast4 = ''
  errors.creditLimit = ''
  errors.tempLimit = ''
}

function handleBack() {
  router.back()
}

async function handleSubmit() {
  if (!validateForm()) {
    return
  }

  submitting.value = true
  submitError.value = null

  try {
    const payload: CreateAccountPayload = {
      type: form.type,
      institution: form.institution,
      cardLast4: form.cardLast4 || undefined,
      currency: form.currency,
      balance: form.balance,
      creditLimit: form.type === 'credit' && form.creditLimit !== undefined ? form.creditLimit : undefined,
      tempLimit: form.type === 'credit' && form.tempLimit !== undefined ? form.tempLimit : undefined
    }

    await createAccount(payload)

    alert(t('newAccount.success'))
    router.push('/accounts')
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : t('newAccount.error.create')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
    existingAccounts.value = await getAccounts()
  } catch (error) {
    console.error('Failed to load existing accounts:', error)
  }
})
</script>

<style scoped>
/* 页面样式已移至全局样式文件 */
</style>
