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
        <!-- 账户名称 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            账户名称 <span class="text-red-500">*</span>
          </label>
          <input
            v-model="form.name"
            type="text"
            placeholder="请输入账户名称"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            :class="{ 'border-red-500': errors.name }"
          />
          <p v-if="errors.name" class="mt-1 text-sm text-red-500">{{ errors.name }}</p>
        </div>

        <!-- 账户类型 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            账户类型 <span class="text-red-500">*</span>
          </label>
          <div class="grid grid-cols-2 gap-3">
            <button
              v-for="type in accountTypes"
              :key="type.value"
              @click="form.type = type.value"
              class="px-4 py-3 border-2 rounded-lg font-medium transition-all"
              :class="form.type === type.value
                ? 'border-blue-500 bg-blue-50 text-blue-700'
                : 'border-gray-300 text-gray-700 hover:border-gray-400'"
            >
              {{ type.label }}
            </button>
          </div>
          <p v-if="errors.type" class="mt-1 text-sm text-red-500">{{ errors.type }}</p>
        </div>

        <!-- 币种 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            币种 <span class="text-red-500">*</span>
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

        <!-- 初始余额 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            初始余额
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
          <p class="mt-1 text-xs text-gray-500">可以为 0 或正数，允许为负（如信用卡透支）</p>
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
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { createAccount } from '@/api/accounts'
import type { AccountType, CreateAccountPayload } from '@/types'
import Icon from '@/components/Icon.vue'

const router = useRouter()

const form = reactive<CreateAccountPayload>({
  name: '',
  type: 'cash' as AccountType,
  currency: 'CNY',
  balance: 0
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

function validateForm(): boolean {
  // 清空之前的错误
  Object.keys(errors).forEach(key => {
    errors[key] = ''
  })

  let isValid = true

  // 校验名称
  if (!form.name.trim()) {
    errors.name = '账户名称不能为空'
    isValid = false
  }

  // 校验类型（已在选择器中限制，这里可以跳过）

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

  return isValid
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
    await createAccount({
      name: form.name.trim(),
      type: form.type,
      currency: form.currency,
      balance: form.balance
    })

    alert('创建成功')
    router.push('/profile')
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : '创建账户失败'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
/* 页面样式已移至全局样式文件 */
</style>
