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
        <h1 class="text-2xl font-bold text-gray-900">新增账户</h1>
      </div>

      <!-- 表单卡片 -->
      <div class="bg-white rounded-lg shadow p-6 space-y-6">
        <!-- 账户类型 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            账户类型 <span class="text-red-500">*</span>
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
              {{ type.label }}
              <span v-if="type.value === 'cash' && hasCashAccount" class="text-xs block text-gray-400">
                (已存在)
              </span>
            </button>
          </div>
          <p v-if="errors.type" class="mt-1 text-sm text-red-500">{{ errors.type }}</p>
        </div>

        <!-- 机构选择（bank/credit/wallet 必填） -->
        <div v-if="form.type === 'bank' || form.type === 'credit' || form.type === 'wallet'">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            <span v-if="form.type === 'bank' || form.type === 'credit'">银行</span>
            <span v-else>平台</span>
            <span class="text-red-500">*</span>
          </label>
          <select
            v-model="form.institution"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.institution }"
          >
            <option value="">请选择</option>
            <option
              v-for="inst in availableInstitutions"
              :key="inst.value"
              :value="inst.value"
            >
              {{ inst.label }}
            </option>
          </select>
          <p v-if="errors.institution" class="mt-1 text-sm text-red-500">{{ errors.institution }}</p>
        </div>

        <!-- 卡片尾号（bank/credit 必填） -->
        <div v-if="form.type === 'bank' || form.type === 'credit'">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            卡片尾号 <span class="text-red-500">*</span>
          </label>
          <input
            v-model="form.cardLast4"
            type="text"
            maxlength="4"
            placeholder="2-4位数字"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.cardLast4 }"
            @input="form.cardLast4 = form.cardLast4.replace(/\D/g, '')"
          />
          <p v-if="errors.cardLast4" class="mt-1 text-sm text-red-500">{{ errors.cardLast4 }}</p>
          <p class="mt-1 text-xs text-gray-500">请输入2-4位数字</p>
        </div>

        <!-- 信用卡额度（仅 credit 类型） -->
        <template v-if="form.type === 'credit'">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              总额度
            </label>
            <input
              v-model.number="form.creditLimit"
              type="number"
              step="0.01"
              placeholder="例如 20000"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.creditLimit }"
            />
            <p v-if="errors.creditLimit" class="mt-1 text-sm text-red-500">{{ errors.creditLimit }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              临时额度
            </label>
            <input
              v-model.number="form.tempLimit"
              type="number"
              step="0.01"
              placeholder="没有可留空"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.tempLimit }"
            />
            <p v-if="errors.tempLimit" class="mt-1 text-sm text-red-500">{{ errors.tempLimit }}</p>
          </div>
        </template>

        <!-- 初始余额 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            <span v-if="form.type === 'credit'">已用额度</span>
            <span v-else>初始余额</span>
          </label>
          <input
            v-model.number="form.balance"
            type="number"
            step="0.01"
            placeholder="0.00"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.balance }"
          />
          <p v-if="errors.balance" class="mt-1 text-sm text-red-500">{{ errors.balance }}</p>
          <p class="mt-1 text-xs text-gray-500">
            <span v-if="form.type === 'credit'">信用卡已使用的额度</span>
            <span v-else>可以为 0 或正数，允许为负</span>
          </p>
        </div>

        <!-- 币种 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            币种
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
          <p class="text-sm text-gray-600 mb-2">预览：</p>
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
          {{ submitting ? '创建中...' : '创建账户' }}
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
  { value: 'cash' as AccountType, label: '现金' },
  { value: 'bank' as AccountType, label: '银行卡' },
  { value: 'wallet' as AccountType, label: '钱包' },
  { value: 'credit' as AccountType, label: '信用卡' }
]

const bankInstitutions: Array<{ value: AccountInstitution; label: string }> = [
  { value: 'icbc', label: '工商银行' },
  { value: 'ccb', label: '建设银行' },
  { value: 'abc', label: '农业银行' },
  { value: 'boc', label: '中国银行' },
  { value: 'bocom', label: '交通银行' },
  { value: 'cmb', label: '招商银行' },
  { value: 'spdb', label: '浦发银行' },
  { value: 'citic', label: '中信银行' },
  { value: 'ceb', label: '光大银行' },
  { value: 'hxb', label: '华夏银行' },
  { value: 'cmbc', label: '民生银行' },
  { value: 'cgb', label: '广发银行' },
  { value: 'pab', label: '平安银行' },
  { value: 'cib', label: '兴业银行' },
  { value: 'psbc', label: '邮储银行' },
  { value: 'other', label: '其他' }
]

const walletInstitutions: Array<{ value: AccountInstitution; label: string }> = [
  { value: 'wechat', label: '微信' },
  { value: 'alipay', label: '支付宝' },
  { value: 'other', label: '其他' }
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
    errors.type = '请选择账户类型'
    isValid = false
  }

  // 检查现金账户是否已存在
  if (form.type === 'cash' && hasCashAccount.value) {
    errors.type = '现金账户已存在，系统只能有一个现金账户'
    isValid = false
  }

  // 根据类型校验必填字段
  if (form.type === 'bank' || form.type === 'credit') {
    if (!form.institution) {
      errors.institution = '请选择银行'
      isValid = false
    }
    if (!form.cardLast4 || form.cardLast4.length < 2 || form.cardLast4.length > 4) {
      errors.cardLast4 = '请输入2-4位卡片尾号'
      isValid = false
    }
  } else if (form.type === 'wallet') {
    if (!form.institution) {
      errors.institution = '请选择平台'
      isValid = false
    }
  }

  // 校验币种
  if (!form.currency) {
    errors.currency = '请选择币种'
    isValid = false
  }

  // 校验余额格式
  if (isNaN(form.balance)) {
    errors.balance = '请输入有效的金额'
    isValid = false
  }

  // 校验信用卡额度
  if (form.type === 'credit') {
    if (form.creditLimit !== undefined && form.creditLimit < 0) {
      errors.creditLimit = '总额度不能为负数'
      isValid = false
    }
    if (form.tempLimit !== undefined && form.tempLimit < 0) {
      errors.tempLimit = '临时额度不能为负数'
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

    alert('创建成功')
    router.push('/accounts')
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : '创建账户失败'
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
