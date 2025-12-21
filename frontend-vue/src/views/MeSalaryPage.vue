<template>
  <div class="min-h-screen flex flex-col bg-slate-50">
    <!-- Header -->
    <header class="px-4 pt-4 pb-3 flex items-center gap-2 bg-white shadow-sm">
      <button class="text-2xl text-slate-600" @click="router.back()">‹</button>
      <h1 class="text-lg font-semibold">
        {{ t('me.salary.title', '薪资控制') }}
      </h1>
    </header>

    <main class="flex-1 overflow-y-auto px-4 pt-4 pb-32 space-y-4">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-8">
        <p class="text-gray-500">{{ t('common.loading', '加载中...') }}</p>
      </div>

      <template v-else>
        <!-- Section 1: Current Salary Plan Card -->
        <section v-if="currentPlan" class="bg-white rounded-2xl shadow-sm p-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-medium text-slate-800">
              {{ t('me.salary.currentPlan', '当前工资方案') }}
            </h2>
            <span
              class="px-2 py-1 rounded-full text-xs"
              :class="currentPlan.active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'"
            >
              {{ currentPlan.active ? t('me.salary.status.active', '启用') : t('me.salary.status.inactive', '停用') }}
            </span>
          </div>

          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-slate-600">{{ t('me.salary.preTaxMonthly', '税前月薪') }}</span>
              <span class="font-medium">¥{{ (currentPlan.salaryPreTaxMonthly || 0).toFixed(2) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-600">{{ t('me.salary.payDay', '发薪日') }}</span>
              <span class="font-medium">{{ currentPlan.salaryPayDay || currentPlan.dayOfMonth }}日</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-600">{{ t('me.salary.account', '到账账户') }}</span>
              <span class="font-medium">{{ getAccountName(currentPlan.accountId) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-600">{{ t('me.salary.city', '纳税城市') }}</span>
              <span class="font-medium">{{ getCityName(currentPlan.cityCode || '') }}</span>
            </div>
          </div>

          <div class="mt-4 flex gap-2">
            <button
              @click="handleEditPlan"
              class="flex-1 py-2 rounded-lg bg-indigo-500 text-white text-sm font-medium"
            >
              {{ t('me.salary.action.edit', '编辑') }}
            </button>
            <button
              @click="handleDeletePlan"
              class="px-4 py-2 rounded-lg bg-red-50 text-red-600 text-sm font-medium"
            >
              {{ t('me.salary.action.delete', '删除') }}
            </button>
          </div>
        </section>

        <div v-else class="bg-white rounded-2xl shadow-sm p-6 text-center">
          <p class="text-slate-500">{{ t('me.salary.noPlan', '暂未配置工资方案') }}</p>
        </div>

        <!-- Section 2: Salary Configuration Form -->
        <section class="bg-white rounded-2xl shadow-sm p-4 space-y-4">
          <h2 class="text-sm font-medium text-slate-800">
            {{ t('me.salary.config.title', '工资参数配置') }}
          </h2>

          <div class="space-y-3">
            <!-- Pre-tax Monthly Salary -->
            <div>
              <label class="block text-xs text-slate-600 mb-1">
                {{ t('me.salary.preTaxMonthly', '税前月薪') }} <span class="text-red-500">*</span>
              </label>
              <input
                v-model.number="form.preTaxMonthly"
                type="number"
                step="0.01"
                class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
                :placeholder="t('me.salary.preTaxMonthly', '税前月薪')"
              />
            </div>

            <!-- City -->
            <div>
              <label class="block text-xs text-slate-600 mb-1">
                {{ t('me.salary.city', '纳税城市') }} <span class="text-red-500">*</span>
              </label>
              <select
                v-model="form.cityCode"
                @change="handleCityChange"
                class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
              >
                <option value="beijing">{{ t('city.beijing', '北京') }}</option>
                <option value="shanghai">{{ t('city.shanghai', '上海') }}</option>
                <option value="guangzhou">{{ t('city.guangzhou', '广州') }}</option>
                <option value="shenzhen">{{ t('city.shenzhen', '深圳') }}</option>
                <option value="default">{{ t('city.default', '其他城市') }}</option>
              </select>
            </div>

            <!-- Social Rate with Detail -->
            <div>
              <label class="block text-xs text-slate-600 mb-1">
                {{ t('me.salary.socialRate', '社保+公积金比例') }} (%)
                <span class="text-red-500">*</span>
              </label>
              <input
                v-model.number="form.socialRate"
                type="number"
                step="0.01"
                min="0"
                max="100"
                class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
                :placeholder="t('me.salary.socialRateHint', '个人缴纳部分')"
              />
              <p class="text-xs text-slate-500 mt-1">
                {{ t('me.salary.socialRateDefault', '当前城市默认: {rate}%').replace('{rate}', (cityDefaultRate * 100).toFixed(1)) }}
              </p>

              <!-- Social Rate Detail -->
              <div v-if="cityRateDetail" class="mt-2 p-3 bg-slate-50 rounded-lg">
                <div class="text-xs font-medium text-slate-700 mb-2">
                  {{ t('me.salary.socialRateDetail', '社保比例明细') }}
                </div>
                <div class="space-y-1 text-xs">
                  <div
                    v-for="comp in cityRateDetail.components"
                    :key="comp.name"
                    class="flex justify-between text-slate-600"
                  >
                    <span>{{ comp.name }}</span>
                    <span>{{ (comp.rate * 100).toFixed(1) }}%</span>
                  </div>
                  <div class="flex justify-between font-medium text-slate-800 pt-1 border-t border-slate-200">
                    <span>{{ t('me.salary.total', '合计') }}</span>
                    <span>{{ (cityRateDetail.totalRate * 100).toFixed(1) }}%</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Extra Deductions -->
            <div>
              <label class="block text-xs text-slate-600 mb-1">
                {{ t('me.salary.extraDeductions', '专项附加扣除') }} (¥/月)
              </label>
              <input
                v-model.number="form.extraDeductions"
                type="number"
                step="0.01"
                min="0"
                class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
                :placeholder="t('me.salary.extraDeductionsHint', '子女教育、房贷利息等')"
              />
            </div>

            <!-- Start Date & Pay Day -->
            <div class="flex gap-3">
              <div class="flex-1">
                <label class="block text-xs text-slate-600 mb-1">
                  {{ t('me.salary.startDate', '开始日期') }} <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="form.startDate"
                  type="date"
                  class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
                />
              </div>
              <div class="flex-1">
                <label class="block text-xs text-slate-600 mb-1">
                  {{ t('me.salary.payDay', '发薪日') }} (1-31) <span class="text-red-500">*</span>
                </label>
                <input
                  v-model.number="form.payDay"
                  type="number"
                  min="1"
                  max="31"
                  class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
                  placeholder="10"
                />
              </div>
            </div>

            <!-- Account -->
            <div>
              <label class="block text-xs text-slate-600 mb-1">
                {{ t('me.salary.account', '到账账户') }} <span class="text-red-500">*</span>
              </label>
              <select
                v-model="form.accountId"
                class="w-full border border-slate-200 rounded-xl px-3 py-2 text-sm bg-slate-50"
              >
                <option value="" disabled>{{ t('me.salary.selectAccount', '请选择账户') }}</option>
                <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                  {{ formatAccountDisplayName(acc) }}
                </option>
              </select>
            </div>

            <!-- Action Buttons -->
            <div class="pt-2 space-y-2">
              <button
                @click="handleSaveAndPreview"
                :disabled="saving || !isFormValid"
                class="w-full py-2.5 rounded-full bg-indigo-500 text-white text-sm font-medium disabled:opacity-60"
              >
                {{ saving ? t('me.salary.saving', '保存中...') : t('me.salary.saveAndApply', '保存为工资方案') }}
              </button>
              <button
                @click="handlePreviewOnly"
                :disabled="previewing || !isFormValid"
                class="w-full py-2.5 rounded-full bg-slate-100 text-slate-700 text-sm font-medium disabled:opacity-60"
              >
                {{ previewing ? t('me.salary.generating', '生成中...') : t('me.salary.previewOnly', '仅预览') }}
              </button>
            </div>
          </div>
        </section>

        <!-- Section 3: Salary Preview -->
        <section v-if="salaryPreview" class="bg-white rounded-2xl shadow-sm p-4 space-y-4">
          <h2 class="text-sm font-medium text-slate-800">
            {{ t('me.salary.preview.title', '薪资预览') }}
          </h2>

          <!-- Annual Summary -->
          <div class="grid grid-cols-3 gap-3">
            <div class="bg-slate-50 rounded-lg p-3">
              <div class="text-xs text-slate-600 mb-1">{{ t('me.salary.preview.totalPreTax', '全年税前') }}</div>
              <div class="text-sm font-semibold text-slate-800">¥{{ salaryPreview.totalPreTax.toFixed(2) }}</div>
            </div>
            <div class="bg-slate-50 rounded-lg p-3">
              <div class="text-xs text-slate-600 mb-1">{{ t('me.salary.preview.totalTax', '全年个税') }}</div>
              <div class="text-sm font-semibold text-red-600">¥{{ salaryPreview.totalTax.toFixed(2) }}</div>
            </div>
            <div class="bg-emerald-50 rounded-lg p-3">
              <div class="text-xs text-slate-600 mb-1">{{ t('me.salary.preview.totalPostTax', '全年税后') }}</div>
              <div class="text-sm font-semibold text-emerald-600">¥{{ salaryPreview.totalPostTax.toFixed(2) }}</div>
            </div>
          </div>

          <!-- Monthly Detail Table -->
          <div class="overflow-x-auto">
            <table class="w-full text-xs">
              <thead class="bg-slate-50">
                <tr>
                  <th class="px-2 py-2 text-left">{{ t('me.salary.preview.month', '月份') }}</th>
                  <th class="px-2 py-2 text-right">{{ t('me.salary.preview.preTax', '税前') }}</th>
                  <th class="px-2 py-2 text-right">{{ t('me.salary.preview.social', '社保') }}</th>
                  <th class="px-2 py-2 text-right">{{ t('me.salary.preview.tax', '个税') }}</th>
                  <th class="px-2 py-2 text-right font-semibold text-emerald-600">
                    {{ t('me.salary.preview.postTax', '到手') }}
                  </th>
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
            </table>
          </div>

          <!-- Disclaimer -->
          <p class="text-xs text-slate-500 text-center">
            {{ t('me.salary.preview.disclaimer', '仅供参考,以实际发薪与当地政策为准') }}
          </p>
        </section>

        <!-- Error Message -->
        <div v-if="errorMessage" class="bg-red-50 text-red-600 rounded-lg p-4 text-sm">
          {{ errorMessage }}
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '@/i18n'
import type {
  RecurringIncomePlan,
  Account,
  SalaryPreviewResponse,
  CitySocialRateDetail,
  SalaryPlanConfig,
} from '@/types'
import { getSalaryPlan, upsertSalaryPlan, deleteSalaryPlan } from '@/api/meSalary'
import { getCitySocialRateDetail } from '@/api/citySocialRate'
import { getAccounts } from '@/api/accounts'
import { previewSalaryToYearEnd } from '@/api/recurringIncome'
import { formatAccountDisplayName } from '@/utils/accountDisplay'

const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const previewing = ref(false)
const errorMessage = ref<string | null>(null)

const currentPlan = ref<RecurringIncomePlan | null>(null)
const accounts = ref<Account[]>([])
const cityRateDetail = ref<CitySocialRateDetail | null>(null)
const salaryPreview = ref<SalaryPreviewResponse | null>(null)

const cityDefaultRate = ref(0.11)

const form = ref({
  preTaxMonthly: 0,
  cityCode: 'beijing',
  socialRate: 0,
  extraDeductions: 0,
  startDate: '',
  payDay: 10,
  accountId: '',
  currency: 'CNY',
})

const isFormValid = computed(() => {
  return (
    form.value.preTaxMonthly > 0 &&
    form.value.cityCode !== '' &&
    form.value.startDate !== '' &&
    form.value.payDay >= 1 &&
    form.value.payDay <= 31 &&
    form.value.accountId !== ''
  )
})

function getCityName(cityCode: string): string {
  return t(`city.${cityCode}`, cityCode)
}

function getAccountName(accountId: string): string {
  const account = accounts.value.find((acc) => acc.id === accountId)
  return account ? formatAccountDisplayName(account) : accountId
}

async function handleCityChange() {
  try {
    const detail = await getCitySocialRateDetail(form.value.cityCode)
    cityRateDetail.value = detail
    cityDefaultRate.value = detail.totalRate
    form.value.socialRate = detail.totalRate * 100 // Convert to percentage
  } catch (error) {
    console.error('Failed to load city rate detail:', error)
  }
}

async function handlePreviewOnly() {
  previewing.value = true
  errorMessage.value = null

  try {
    const preview = await previewSalaryToYearEnd({
      preTaxAmount: form.value.preTaxMonthly,
      cityCode: form.value.cityCode,
      startDate: form.value.startDate,
      payDay: form.value.payDay,
      frequency: 'monthly',
      currency: form.value.currency,
      socialRate: form.value.socialRate / 100, // Convert percentage to decimal
      extraDeductions: form.value.extraDeductions,
    })
    salaryPreview.value = preview
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '预览失败'
  } finally {
    previewing.value = false
  }
}

async function handleSaveAndPreview() {
  saving.value = true
  errorMessage.value = null

  try {
    const payload: SalaryPlanConfig = {
      preTaxMonthly: form.value.preTaxMonthly,
      cityCode: form.value.cityCode,
      startDate: form.value.startDate,
      payDay: form.value.payDay,
      socialRate: form.value.socialRate / 100, // Convert percentage to decimal
      extraDeductions: form.value.extraDeductions,
      accountId: form.value.accountId,
      currency: form.value.currency,
    }

    const result = await upsertSalaryPlan(payload)
    currentPlan.value = result.plan
    salaryPreview.value = result.preview

    alert(t('me.salary.success.saved', '工资方案保存成功'))
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '保存失败'
  } finally {
    saving.value = false
  }
}

function handleEditPlan() {
  if (!currentPlan.value) return

  form.value = {
    preTaxMonthly: currentPlan.value.salaryPreTaxMonthly || 0,
    cityCode: currentPlan.value.cityCode || 'beijing',
    socialRate: (currentPlan.value.salarySocialRate || 0) * 100,
    extraDeductions: currentPlan.value.salaryExtraDeductions || 0,
    startDate: currentPlan.value.salaryStartDate || '',
    payDay: currentPlan.value.salaryPayDay || currentPlan.value.dayOfMonth || 10,
    accountId: currentPlan.value.accountId,
    currency: currentPlan.value.currency,
  }

  handleCityChange()
}

async function handleDeletePlan() {
  if (!currentPlan.value) return

  if (!confirm('确定要删除当前工资方案吗？')) return

  try {
    await deleteSalaryPlan(currentPlan.value.id)
    currentPlan.value = null
    salaryPreview.value = null
    alert(t('me.salary.success.deleted', '工资方案已删除'))
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '删除失败'
  }
}

async function loadData() {
  loading.value = true
  errorMessage.value = null

  try {
    const [accountsData, salaryPlanData] = await Promise.all([
      getAccounts(),
      getSalaryPlan(),
    ])

    accounts.value = accountsData.filter((acc) => !acc.archived)
    if (salaryPlanData.plan) {
      currentPlan.value = salaryPlanData.plan
      handleEditPlan()
    } else {
      // Initialize form with default values
      const today = new Date()
      form.value.startDate = today.toISOString().split('T')[0]
      if (accounts.value.length > 0) {
        form.value.accountId = accounts.value[0].id
      }
      await handleCityChange()
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载数据失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
