<template>
  <PageFrame :title="t('salary.title')" eyebrow="INCOME PLANS" subtitle="设置收入金额、到账账户与发放日期。" back-to="/profile" compact>
    <p class="notice">此页面支持查看和新增收入计划，编辑与删除暂未开放。工资方案请前往 <RouterLink to="/me/salary">薪资管理 →</RouterLink></p>
    <div class="salary-settings-content">
      <!-- 现有收入计划列表 -->
      <div class="mb-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">{{ t('salary.existingPlans') }}</h2>
        <div v-if="loading" class="text-center py-8">
          <p class="text-gray-500">{{ t('salary.loading') }}</p>
        </div>
        <div v-else-if="incomePlans.length === 0" class="bg-gray-50 rounded-lg p-6 text-center">
          <p class="text-gray-500">{{ t('salary.noPlans') }}</p>
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="plan in incomePlans"
            :key="plan.id"
            class="panel p-4"
          >
            <div class="flex justify-between items-start mb-2">
              <div>
                <p class="font-semibold text-gray-900">{{ plan.name }}</p>
                <p class="text-sm text-gray-500 mt-1">
                  {{ t('salary.frequencyAmount', { frequency: getFrequencyLabel(plan.frequency), amount: plan.amount.toFixed(2) }) }}
                </p>
                <p class="text-xs text-gray-400 mt-1">
                  {{ t('salary.accountInfo', { account: getAccountName(plan.accountId) }) }}
                </p>
                <p class="text-xs text-gray-400 mt-1">
                  {{ t('salary.firstPayInfo', { date: plan.firstPayDate }) }}
                </p>
              </div>
              <div class="flex gap-2">
                <button
                  disabled title="暂未开放编辑"
                  class="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded"
                >
                  {{ t('salary.edit') }}
                </button>
                <button
                  disabled title="暂未开放删除"
                  class="px-3 py-1 text-sm text-red-600 hover:bg-red-50 rounded"
                >
                  {{ t('salary.delete') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 新增/编辑表单 -->
      <div class="panel p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">
          {{ editingPlan ? t('salary.editPlan') : t('salary.newPlan') }}
        </h2>

        <div class="space-y-4">
          <!-- 名称 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('salary.name') }} <span class="text-red-500">{{ t('common.required') }}</span>
            </label>
            <input
              v-model="form.name"
              type="text"
              :placeholder="t('salary.namePlaceholder')"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.name }"
            />
            <p v-if="errors.name" class="mt-1 text-sm text-red-500">{{ errors.name }}</p>
          </div>

          <!-- 金额 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('salary.amount') }} <span class="text-red-500">{{ t('common.required') }}</span>
            </label>
            <input
              v-model.number="form.amount"
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.amount }"
            />
            <p v-if="errors.amount" class="mt-1 text-sm text-red-500">{{ errors.amount }}</p>
          </div>

          <!-- 频率 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('salary.frequency') }} <span class="text-red-500">{{ t('common.required') }}</span>
            </label>
            <select
              v-model="form.frequency"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.frequency }"
            >
              <option value="monthly">{{ t('salary.frequency.monthly') }}</option>
              <option value="weekly">{{ t('salary.frequency.weekly') }}</option>
              <option value="biweekly">{{ t('salary.frequency.biweekly') }}</option>
              <option value="yearly">{{ t('salary.frequency.yearly') }}</option>
            </select>
            <p v-if="errors.frequency" class="mt-1 text-sm text-red-500">{{ errors.frequency }}</p>
          </div>

          <!-- 到账账户 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('salary.account') }} <span class="text-red-500">{{ t('common.required') }}</span>
            </label>
            <select
              v-model="form.accountId"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.accountId }"
            >
              <option value="">{{ t('salary.accountPlaceholder') }}</option>
              <option
                v-for="account in accounts"
                :key="account.id"
                :value="account.id"
              >
                {{ formatAccountDisplayName(account) }}
              </option>
            </select>
            <p v-if="errors.accountId" class="mt-1 text-sm text-red-500">{{ errors.accountId }}</p>
          </div>

          <!-- 第一次发薪日期 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('salary.firstPayDate') }} <span class="text-red-500">{{ t('common.required') }}</span>
            </label>
            <input
              v-model="form.firstPayDate"
              type="date"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :class="{ 'border-red-500': errors.firstPayDate }"
            />
            <p v-if="errors.firstPayDate" class="mt-1 text-sm text-red-500">{{ errors.firstPayDate }}</p>
          </div>

          <!-- 到期日期（可选） -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              {{ t('salary.endDate') }}
            </label>
            <input
              v-model="form.endDate"
              type="date"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('salary.endDateHint') }}</p>
          </div>
        </div>

        <!-- 提交按钮 -->
        <div class="mt-6">
          <button
            @click="handleSubmit"
            :disabled="submitting"
            class="button primary"
          >
            {{ submitting ? t('salary.submitting') : editingPlan ? t('salary.update') : t('salary.submit') }}
          </button>
        </div>

        <!-- 错误提示 -->
        <div v-if="submitError" class="mt-4 p-4 bg-red-50 text-red-600 rounded-lg">
          {{ submitError }}
        </div>
      </div>
    </div>
  </PageFrame>
</template>

<script setup lang="ts">
import PageFrame from '@/components/PageFrame.vue'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import { ref, reactive, onMounted } from 'vue'
import { listIncomePlans, createIncomePlan } from '@/api/incomePlans'
import { getAccounts } from '@/api/accounts'
import type { IncomePlan, IncomeFrequency, CreateIncomePlanPayload, Account } from '@/types'
import { t } from '@/i18n'


const loading = ref(false)
const accounts = ref<Account[]>([])
const incomePlans = ref<IncomePlan[]>([])
const submitting = ref(false)
const submitError = ref<string | null>(null)
const editingPlan = ref<IncomePlan | null>(null)

const form = reactive<CreateIncomePlanPayload>({
  name: '',
  amount: 0,
  frequency: 'monthly',
  accountId: '',
  firstPayDate: '',
  endDate: ''
})

const errors = reactive<Record<string, string>>({})

function getFrequencyLabel(frequency: IncomeFrequency): string {
  return t(`salary.frequency.${frequency}`)
}

function getAccountName(accountId: string): string {
  const account = accounts.value.find(acc => acc.id === accountId)
  return account ? formatAccountDisplayName(account) : accountId
}

function validateForm(): boolean {
  Object.keys(errors).forEach(key => {
    errors[key] = ''
  })

  let isValid = true

  if (!form.name.trim()) {
    errors.name = t('salary.error.name')
    isValid = false
  }

  if (form.amount <= 0) {
    errors.amount = t('salary.error.amount')
    isValid = false
  }

  if (!form.accountId) {
    errors.accountId = t('salary.error.account')
    isValid = false
  }

  if (!form.firstPayDate) {
    errors.firstPayDate = t('salary.error.firstPayDate')
    isValid = false
  }

  return isValid
}

function resetForm() {
  form.name = ''
  form.amount = 0
  form.frequency = 'monthly'
  form.accountId = ''
  form.firstPayDate = ''
  form.endDate = ''
  editingPlan.value = null
  Object.keys(errors).forEach(key => {
    errors[key] = ''
  })
}

async function handleSubmit() {
  if (!validateForm()) {
    return
  }

  submitting.value = true
  submitError.value = null

  try {
    const payload: CreateIncomePlanPayload = {
      name: form.name.trim(),
      amount: form.amount,
      frequency: form.frequency,
      accountId: form.accountId,
      firstPayDate: form.firstPayDate
    }

    if (form.endDate) {
      payload.endDate = form.endDate
    }

    const newPlan = await createIncomePlan(payload)
    incomePlans.value.push(newPlan)
    resetForm()
    alert(t('salary.success'))
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : t('salary.error.create')
  } finally {
    submitting.value = false
  }
}

async function loadData() {
  loading.value = true
  try {
    const [accountsData, plansData] = await Promise.all([
      getAccounts(),
      listIncomePlans()
    ])
    accounts.value = accountsData
    incomePlans.value = plansData
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : t('salary.error.load')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.salary-settings-content { display: grid; grid-template-columns: 300px minmax(0, 1fr); gap: 24px; align-items: start; }
@media (max-width: 900px) { .salary-settings-content { grid-template-columns: minmax(0, 1fr); } }
</style>




