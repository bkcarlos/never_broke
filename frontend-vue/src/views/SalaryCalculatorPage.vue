<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { t } from '@/i18n'
import { getPayrollRegions, getPayrollRegionRule, calcSalary } from '@/api/payroll'
import type {
  PayrollRegion,
  PayrollRegionRule,
  CalcSalaryRequest,
  CalcSalaryResponse,
  BaseMode,
  SpecialDeduction
} from '@/types'

// 地区列表
const regions = ref<PayrollRegion[]>([])
const selectedRegionCode = ref('CN-DEFAULT')
const selectedYear = ref(2025)

// 地区规则
const regionRule = ref<PayrollRegionRule | null>(null)

// 社保各项比例（默认值）
const ssRates = reactive({
  // 养老保险
  pensionPerson: 0.08,
  pensionCompany: 0.16,
  // 医疗保险
  medicalPerson: 0.02,
  medicalCompany: 0.095,
  // 失业保险
  unemploymentPerson: 0.005,
  unemploymentCompany: 0.005,
  // 工伤保险（个人不缴纳）
  injuryCompany: 0.004,
  // 生育保险（个人不缴纳，已合并入医疗）
  maternityCompany: 0.008
})

// 表单数据
const form = reactive({
  grossMonthly: 20000,
  
  // 社保配置
  ssEnabled: true,
  ssBaseMode: 'BY_WAGE' as BaseMode,
  ssBaseCustom: null as number | null,
  ssRateCustomEnabled: false,
  
  // 公积金配置
  hfEnabled: true,
  hfBaseMode: 'BY_WAGE' as BaseMode,
  hfBaseCustom: null as number | null,
  hfRatePerson: 0.12,
  hfRateCompany: 0.12,
  
  // 专项附加扣除
  specialEnabled: false,
  special: {
    childEdu: 0,
    contEdu: 0,
    elderlyCare: 0,
    housingLoan: 0,
    housingRent: 0,
    infantCare: 0,
    pensionMonthly: 0
  } as SpecialDeduction
})

// 计算社保个人/公司比例总和
const ssRatePersonSum = computed(() => {
  return ssRates.pensionPerson + ssRates.medicalPerson + ssRates.unemploymentPerson
})

const ssRateCompanySum = computed(() => {
  return ssRates.pensionCompany + ssRates.medicalCompany + 
         ssRates.unemploymentCompany + ssRates.injuryCompany + ssRates.maternityCompany
})

// 状态
const loading = ref(false)
const calculating = ref(false)
const errorMessage = ref('')
const result = ref<CalcSalaryResponse | null>(null)

// 专项扣除选项 - 更详细的描述
const childEduOptions = [
  { value: 0, label: '不符合，扣除0元' },
  { value: 1000, label: '一个孩子，夫妻各扣除1000元' },
  { value: 2000, label: '一个孩子，仅有一方扣除2000元' },
  { value: 3000, label: '两个孩子，夫妻各扣除2000元（自己分摊3000元）' },
  { value: 4000, label: '两个孩子，仅有一方扣除4000元' }
]

const contEduOptions = [
  { value: 0, label: '不符合，扣除0元' },
  { value: 400, label: '接受学历教育中，扣除400元' }
]

const elderlyCareOptions = [
  { value: 0, label: '不符合，扣除0元' },
  { value: 500, label: '六兄弟姐妹分摊，扣除500元' },
  { value: 1000, label: '三兄弟姐妹分摊，扣除1000元' },
  { value: 1500, label: '两兄弟姐妹分摊，扣除1500元' },
  { value: 3000, label: '独生子女，扣除3000元' }
]

const housingLoanOptions = [
  { value: 0, label: '不符合，扣除0元' },
  { value: 500, label: '夫妻各扣除500元' },
  { value: 1000, label: '仅由一方扣除1000元' }
]

const housingRentOptions = [
  { value: 0, label: '不符合，扣除0元' },
  { value: 800, label: '所在城市户籍人口少于100万，扣除800元' },
  { value: 1100, label: '所在城市户籍人口超过100万，扣除1100元' },
  { value: 1500, label: '直辖市、省会城市、计划单列市等，扣除1500元' }
]

const infantCareOptions = [
  { value: 0, label: '不符合，扣除0元' },
  { value: 1000, label: '夫妻各扣除1000元' },
  { value: 2000, label: '一方扣除2000元' }
]

// 计算属性
const baseModeOptions = computed(() => [
  { value: 'BY_WAGE', label: t('payroll.baseMode.byWage') },
  { value: 'BY_MIN', label: t('payroll.baseMode.byMin') },
  { value: 'CUSTOM', label: t('payroll.baseMode.custom') }
])

const hfRateRange = computed(() => {
  if (!regionRule.value) return { min: 0.05, max: 0.12 }
  return {
    min: regionRule.value.hfRateMin,
    max: regionRule.value.hfRateMax
  }
})

const hfRateOptions = computed(() => {
  const options = []
  const { min, max } = hfRateRange.value
  for (let rate = min; rate <= max; rate += 0.01) {
    options.push({
      value: Math.round(rate * 100) / 100,
      label: `${(rate * 100).toFixed(0)}%`
    })
  }
  return options
})

// 专项扣除月总额
const specialMonthlyTotal = computed(() => {
  if (!form.specialEnabled) return 0
  return form.special.childEdu + 
         form.special.contEdu + 
         form.special.elderlyCare + 
         form.special.housingLoan + 
         form.special.housingRent + 
         form.special.infantCare + 
         form.special.pensionMonthly
})

// 加载地区列表
async function loadRegions() {
  try {
    regions.value = await getPayrollRegions()
    if (regions.value.length > 0 && !regions.value.find(r => r.code === selectedRegionCode.value)) {
      selectedRegionCode.value = regions.value[0].code
    }
  } catch (error) {
    console.error('Failed to load regions:', error)
    errorMessage.value = t('payroll.error.loadRegions')
  }
}

// 加载地区规则
async function loadRegionRule() {
  try {
    loading.value = true
    regionRule.value = await getPayrollRegionRule(selectedRegionCode.value, selectedYear.value)
    
    if (regionRule.value) {
      form.hfRatePerson = regionRule.value.defaultHfRatePerson
      form.hfRateCompany = regionRule.value.defaultHfRateCompany
    }
  } catch (error) {
    console.error('Failed to load region rule:', error)
    errorMessage.value = t('payroll.error.loadRule')
  } finally {
    loading.value = false
  }
}

// 执行计算
async function handleCalculate() {
  if (!regionRule.value) {
    errorMessage.value = t('payroll.error.noRule')
    return
  }

  if (form.grossMonthly <= 0) {
    errorMessage.value = t('payroll.error.invalidGross')
    return
  }

  if (form.ssEnabled && form.ssBaseMode === 'CUSTOM' && (!form.ssBaseCustom || form.ssBaseCustom <= 0)) {
    errorMessage.value = t('payroll.error.invalidSSBase')
    return
  }

  if (form.hfEnabled && form.hfBaseMode === 'CUSTOM' && (!form.hfBaseCustom || form.hfBaseCustom <= 0)) {
    errorMessage.value = t('payroll.error.invalidHFBase')
    return
  }

  errorMessage.value = ''
  calculating.value = true

  try {
    const request: CalcSalaryRequest = {
      regionCode: selectedRegionCode.value,
      year: selectedYear.value,
      grossMonthly: form.grossMonthly,
      ssEnabled: form.ssEnabled,
      ssBaseMode: form.ssBaseMode,
      ssBaseCustom: form.ssBaseMode === 'CUSTOM' ? form.ssBaseCustom ?? undefined : undefined,
      ssRatePersonSum: form.ssRateCustomEnabled ? ssRatePersonSum.value : undefined,
      ssRateCompanySum: form.ssRateCustomEnabled ? ssRateCompanySum.value : undefined,
      hfEnabled: form.hfEnabled,
      hfBaseMode: form.hfBaseMode,
      hfBaseCustom: form.hfBaseMode === 'CUSTOM' ? form.hfBaseCustom ?? undefined : undefined,
      hfRatePerson: form.hfRatePerson,
      hfRateCompany: form.hfRateCompany,
      specialEnabled: form.specialEnabled,
      special: form.special
    }

    result.value = await calcSalary(request)
  } catch (error: any) {
    console.error('Failed to calculate salary:', error)
    errorMessage.value = error.message || t('payroll.error.calcFailed')
  } finally {
    calculating.value = false
  }
}

// 监听变化
watch(selectedRegionCode, () => {
  loadRegionRule()
  result.value = null
})

watch(selectedYear, () => {
  loadRegionRule()
  result.value = null
})

// 格式化金额
function formatMoney(value: number): string {
  return value.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// 格式化百分比
function formatPercent(value: number): string {
  return (value * 100).toFixed(1) + '%'
}

// 初始化
onMounted(async () => {
  await loadRegions()
  await loadRegionRule()
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-b from-blue-50 to-gray-50">
    <!-- 顶部导航 -->
    <div class="bg-white border-b border-gray-200 sticky top-0 z-20 shadow-sm">
      <div class="max-w-6xl mx-auto px-4 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <router-link 
              to="/" 
              class="w-9 h-9 flex items-center justify-center rounded-full bg-gray-100 hover:bg-gray-200 transition-colors"
            >
              <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
            </router-link>
            <div>
              <h1 class="text-lg font-semibold text-gray-900">{{ t('payroll.title') }}</h1>
              <p class="text-xs text-gray-500">五险一金 · 个税计算 · 专项扣除</p>
            </div>
          </div>
          <div class="text-sm text-gray-500">
            {{ selectedYear }}年
          </div>
        </div>
      </div>
    </div>

    <div class="max-w-6xl mx-auto px-4 py-6">
      <div class="lg:grid lg:grid-cols-12 lg:gap-6">
        <!-- 左侧表单区域 -->
        <div class="lg:col-span-5 space-y-4">
          <!-- 基本信息卡片 -->
          <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div class="bg-gradient-to-r from-blue-600 to-blue-500 px-5 py-3">
              <h2 class="text-white font-medium">基本信息</h2>
            </div>
            <div class="p-5 space-y-4">
              <!-- 地区选择 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">选择地区</label>
                <select
                  v-model="selectedRegionCode"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                >
                  <option v-for="region in regions" :key="region.code" :value="region.code">
                    {{ region.name }}
                  </option>
                </select>
              </div>

              <!-- 税前基本工资 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">税前基本工资</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 font-medium">¥</span>
                  <input
                    v-model.number="form.grossMonthly"
                    type="number"
                    min="0"
                    step="100"
                    class="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all text-lg font-medium"
                    placeholder="请输入金额"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- 社保配置卡片 -->
          <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 flex items-center justify-between">
              <h2 class="font-medium text-gray-900">是否缴纳社保</h2>
              <div class="flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input 
                    type="radio" 
                    :checked="form.ssEnabled" 
                    @change="form.ssEnabled = true"
                    class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-sm text-gray-700">缴纳</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input 
                    type="radio" 
                    :checked="!form.ssEnabled" 
                    @change="form.ssEnabled = false"
                    class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-sm text-gray-700">不缴纳</span>
                </label>
              </div>
            </div>

            <div v-if="form.ssEnabled" class="p-5 space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">社保缴纳基数</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="opt in baseModeOptions"
                    :key="opt.value"
                    @click="form.ssBaseMode = opt.value as BaseMode"
                    :class="[
                      'px-4 py-2 rounded-lg text-sm font-medium transition-all border',
                      form.ssBaseMode === opt.value
                        ? 'bg-blue-50 border-blue-500 text-blue-700'
                        : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100'
                    ]"
                  >
                    {{ opt.label }}
                  </button>
                </div>
              </div>

              <div v-if="form.ssBaseMode === 'CUSTOM'">
                <label class="block text-sm font-medium text-gray-700 mb-2">社保基数</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">¥</span>
                  <input
                    v-model.number="form.ssBaseCustom"
                    type="number"
                    min="0"
                    class="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <p v-if="regionRule" class="text-xs text-gray-500 mt-1.5">
                  📍 {{ regionRule.regionName }} 基数范围：¥{{ regionRule.ssBaseMin.toLocaleString() }} - ¥{{ regionRule.ssBaseMax.toLocaleString() }}
                </p>
              </div>

              <!-- 自定义比例开关 -->
              <div class="flex items-center justify-between pt-2 border-t border-gray-100">
                <span class="text-sm text-gray-700">自定义各项比例</span>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input type="checkbox" v-model="form.ssRateCustomEnabled" class="sr-only peer">
                  <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <!-- 各项社保比例调整 -->
              <div v-if="form.ssRateCustomEnabled" class="space-y-3 pt-2">
                <!-- 养老保险 -->
                <div class="p-3 bg-orange-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-sm font-medium text-gray-700">🧓 养老保险</span>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">个人</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.pensionPerson"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.pensionPerson) }}
                        </span>
                      </div>
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">公司</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.pensionCompany"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.pensionCompany) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 医疗保险 -->
                <div class="p-3 bg-green-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-sm font-medium text-gray-700">🏥 医疗保险</span>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">个人</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.medicalPerson"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.medicalPerson) }}
                        </span>
                      </div>
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">公司</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.medicalCompany"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.medicalCompany) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 失业保险 -->
                <div class="p-3 bg-blue-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-sm font-medium text-gray-700">💼 失业保险</span>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">个人</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.unemploymentPerson"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.unemploymentPerson) }}
                        </span>
                      </div>
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">公司</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.unemploymentCompany"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.unemploymentCompany) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 工伤保险 -->
                <div class="p-3 bg-red-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-sm font-medium text-gray-700">🚑 工伤保险</span>
                    <span class="text-xs text-gray-500">个人不缴纳</span>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">个人</label>
                      <input
                        type="text"
                        value="0%"
                        disabled
                        class="w-full px-3 py-1.5 text-sm bg-gray-100 border border-gray-200 rounded-lg text-gray-400"
                      />
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">公司</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.injuryCompany"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.injuryCompany) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 生育保险 -->
                <div class="p-3 bg-pink-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-sm font-medium text-gray-700">👶 生育保险</span>
                    <span class="text-xs text-gray-500">个人不缴纳</span>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">个人</label>
                      <input
                        type="text"
                        value="0%"
                        disabled
                        class="w-full px-3 py-1.5 text-sm bg-gray-100 border border-gray-200 rounded-lg text-gray-400"
                      />
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">公司</label>
                      <div class="relative">
                        <input
                          v-model.number="ssRates.maternityCompany"
                          type="number"
                          min="0"
                          max="1"
                          step="0.001"
                          class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
                        />
                        <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                          {{ formatPercent(ssRates.maternityCompany) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 比例汇总 -->
                <div class="p-3 bg-gray-100 rounded-lg">
                  <div class="flex justify-between text-sm">
                    <span class="text-gray-600">个人合计</span>
                    <span class="font-medium text-gray-900">{{ formatPercent(ssRatePersonSum) }}</span>
                  </div>
                  <div class="flex justify-between text-sm mt-1">
                    <span class="text-gray-600">公司合计</span>
                    <span class="font-medium text-gray-900">{{ formatPercent(ssRateCompanySum) }}</span>
                  </div>
                </div>
              </div>

              <!-- 默认比例提示 -->
              <div v-if="!form.ssRateCustomEnabled" class="text-sm text-gray-500 bg-gray-50 p-3 rounded-lg">
                <div class="flex justify-between">
                  <span>个人社保比例</span>
                  <span class="font-medium">{{ regionRule ? formatPercent(regionRule.ssPersonRateSum) : '10.5%' }}</span>
                </div>
                <div class="flex justify-between mt-1">
                  <span>公司社保比例</span>
                  <span class="font-medium">{{ regionRule ? formatPercent(regionRule.ssCompanyRateSum) : '29.5%' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 公积金配置卡片 -->
          <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 flex items-center justify-between">
              <h2 class="font-medium text-gray-900">是否缴纳公积金</h2>
              <div class="flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input 
                    type="radio" 
                    :checked="form.hfEnabled" 
                    @change="form.hfEnabled = true"
                    class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-sm text-gray-700">缴纳</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input 
                    type="radio" 
                    :checked="!form.hfEnabled" 
                    @change="form.hfEnabled = false"
                    class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-sm text-gray-700">不缴纳</span>
                </label>
              </div>
            </div>

            <div v-if="form.hfEnabled" class="p-5 space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">公积金缴纳基数</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="opt in baseModeOptions"
                    :key="opt.value"
                    @click="form.hfBaseMode = opt.value as BaseMode"
                    :class="[
                      'px-4 py-2 rounded-lg text-sm font-medium transition-all border',
                      form.hfBaseMode === opt.value
                        ? 'bg-blue-50 border-blue-500 text-blue-700'
                        : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100'
                    ]"
                  >
                    {{ opt.label }}
                  </button>
                </div>
              </div>

              <div v-if="form.hfBaseMode === 'CUSTOM'">
                <label class="block text-sm font-medium text-gray-700 mb-2">公积金基数</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">¥</span>
                  <input
                    v-model.number="form.hfBaseCustom"
                    type="number"
                    min="0"
                    class="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">公积金缴纳比例</label>
                <select
                  v-model="form.hfRatePerson"
                  @change="form.hfRateCompany = form.hfRatePerson"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in hfRateOptions" :key="opt.value" :value="opt.value">
                    个人 {{ opt.label }} + 公司 {{ opt.label }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <!-- 专项附加扣除卡片 -->
          <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div class="px-5 py-3 border-b border-gray-100 flex items-center justify-between">
              <div>
                <h2 class="font-medium text-gray-900">是否包含专项扣除</h2>
                <p class="text-xs text-gray-500 mt-0.5">子女教育、继续教育、赡养老人等</p>
              </div>
              <div class="flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input 
                    type="radio" 
                    :checked="form.specialEnabled" 
                    @change="form.specialEnabled = true"
                    class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-sm text-gray-700">包含</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input 
                    type="radio" 
                    :checked="!form.specialEnabled" 
                    @change="form.specialEnabled = false"
                    class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-sm text-gray-700">不包含</span>
                </label>
              </div>
            </div>

            <div v-if="form.specialEnabled" class="p-5 space-y-4">
              <!-- 子女教育扣除 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">子女教育扣除</label>
                <select
                  v-model.number="form.special.childEdu"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in childEduOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- 继续教育扣除 -->
              <div>
                <label class="flex items-center gap-1 text-sm font-medium text-gray-700 mb-2">
                  继续教育扣除
                  <span class="relative group">
                    <svg class="w-4 h-4 text-gray-400 cursor-help" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-48 p-2 bg-gray-800 text-white text-xs rounded-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-10">
                      接受学历（学位）继续教育时可扣除
                    </div>
                  </span>
                </label>
                <select
                  v-model.number="form.special.contEdu"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in contEduOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- 赡养老人扣除 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">赡养老人扣除</label>
                <select
                  v-model.number="form.special.elderlyCare"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in elderlyCareOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- 住房贷款利息扣除 -->
              <div>
                <label class="flex items-center gap-1 text-sm font-medium text-gray-700 mb-2">
                  住房贷款利息扣除
                  <span class="relative group">
                    <svg class="w-4 h-4 text-gray-400 cursor-help" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-56 p-2 bg-gray-800 text-white text-xs rounded-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-10">
                      首套住房贷款利息可扣除，最长240个月
                    </div>
                  </span>
                </label>
                <select
                  v-model.number="form.special.housingLoan"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in housingLoanOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- 住房租金扣除 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">住房租金扣除</label>
                <select
                  v-model.number="form.special.housingRent"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in housingRentOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- 3岁以下婴幼儿照护 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">3岁以下婴幼儿照护</label>
                <select
                  v-model.number="form.special.infantCare"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="opt in infantCareOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- 个人养老金 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">个人养老金</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">¥</span>
                  <input
                    v-model.number="form.special.pensionMonthly"
                    type="number"
                    min="0"
                    max="1000"
                    class="w-full pl-10 pr-16 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="0"
                  />
                  <span class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 text-sm">/月</span>
                </div>
                <p class="text-xs text-gray-500 mt-1.5">每年上限12000元，折合每月1000元</p>
              </div>

              <!-- 大病医疗提示 -->
              <div class="p-3 bg-amber-50 border border-amber-200 rounded-lg">
                <p class="text-xs text-amber-700">
                  💡 大病医疗扣除需次年纳税人自行扣除，系统暂不支持按月扣除计算
                </p>
              </div>
            </div>
          </div>

          <!-- 计算按钮 -->
          <button
            @click="handleCalculate"
            :disabled="calculating || loading"
            class="w-full py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-700 hover:to-blue-600 disabled:from-gray-400 disabled:to-gray-400 text-white font-medium rounded-xl shadow-lg shadow-blue-500/25 transition-all flex items-center justify-center gap-2"
          >
            <svg v-if="calculating" class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
            {{ calculating ? '计算中...' : '开始计算' }}
          </button>

          <!-- 错误提示 -->
          <div v-if="errorMessage" class="p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
            {{ errorMessage }}
          </div>
        </div>

        <!-- 右侧结果区域 -->
        <div class="lg:col-span-7 mt-6 lg:mt-0">
          <!-- 无结果提示 -->
          <div v-if="!result" class="bg-white rounded-xl shadow-sm border border-gray-100 p-12 text-center">
            <div class="w-20 h-20 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
              <svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 class="text-lg font-medium text-gray-900 mb-2">{{ t('payroll.noResult') }}</h3>
            <p class="text-gray-500 text-sm">{{ t('payroll.noResultTip') }}</p>
          </div>

          <!-- 计算结果 -->
          <template v-else>
            <!-- 结果概览卡片 -->
            <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden mb-6">
              <div class="bg-gradient-to-r from-green-600 to-emerald-500 px-5 py-3">
                <h2 class="text-white font-medium">计算结果（由于各地工伤保险略有差异，计算结果仅供查询）</h2>
              </div>
              <div class="p-5">
                <div class="grid grid-cols-2 gap-4">
                  <!-- 个人税前收入 -->
                  <div class="p-4 bg-gray-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">个人税前收入</div>
                    <div class="text-xl font-bold text-gray-900">¥{{ formatMoney(result.summary.grossMonthly) }}</div>
                  </div>
                  
                  <!-- 个人缴纳公积金 -->
                  <div class="p-4 bg-blue-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">个人缴纳公积金</div>
                    <div class="text-xl font-bold text-blue-600">¥{{ formatMoney(result.summary.hfPersonMonthly) }}</div>
                  </div>
                  
                  <!-- 公司缴纳公积金 -->
                  <div class="p-4 bg-gray-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">公司缴纳公积金</div>
                    <div class="text-xl font-bold text-gray-700">¥{{ formatMoney(result.summary.hfCompanyMonthly) }}</div>
                  </div>
                  
                  <!-- 个人缴纳社保 -->
                  <div class="p-4 bg-orange-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">个人缴纳社保</div>
                    <div class="text-xl font-bold text-orange-600">¥{{ formatMoney(result.summary.ssPersonMonthly) }}</div>
                  </div>
                  
                  <!-- 公司缴纳社保 -->
                  <div class="p-4 bg-gray-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">公司缴纳社保</div>
                    <div class="text-xl font-bold text-gray-700">¥{{ formatMoney(result.summary.ssCompanyMonthly) }}</div>
                  </div>
                  
                  <!-- 公司雇佣成本 -->
                  <div class="p-4 bg-purple-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">公司花费雇佣成本</div>
                    <div class="text-xl font-bold text-purple-600">¥{{ formatMoney(result.summary.companyCostMonthly) }}</div>
                  </div>
                  
                  <!-- 专项扣除免税总额 -->
                  <div class="col-span-2 p-4 bg-emerald-50 rounded-lg">
                    <div class="text-sm text-gray-500 mb-1">专项扣除免税总额</div>
                    <div class="text-xl font-bold text-emerald-600">¥{{ formatMoney(specialMonthlyTotal) }}/月</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 12个月明细表格 -->
            <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
              <div class="px-5 py-3 border-b border-gray-100">
                <h2 class="font-medium text-gray-900">每月收入明细</h2>
              </div>
              <div class="overflow-x-auto">
                <table class="min-w-full">
                  <thead class="bg-gray-50">
                    <tr>
                      <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">月份</th>
                      <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">税后收入</th>
                      <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">预缴个税</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="month in result.months" :key="month.month" class="hover:bg-gray-50">
                      <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                        第{{ month.month }}个月
                      </td>
                      <td class="px-4 py-3 whitespace-nowrap text-sm text-right font-semibold text-green-600">
                        ¥{{ formatMoney(month.net) }}
                      </td>
                      <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-red-600">
                        ¥{{ formatMoney(month.tax) }}
                      </td>
                    </tr>
                  </tbody>
                  <tfoot class="bg-gradient-to-r from-green-50 to-blue-50">
                    <tr>
                      <td class="px-4 py-4 whitespace-nowrap text-sm font-bold text-gray-900">
                        年度合计
                      </td>
                      <td class="px-4 py-4 whitespace-nowrap text-right">
                        <div class="text-xs text-gray-500 mb-1">个人税后总和</div>
                        <div class="text-lg font-bold text-green-600">¥{{ formatMoney(result.totals.netYear) }}</div>
                      </td>
                      <td class="px-4 py-4 whitespace-nowrap text-right">
                        <div class="text-xs text-gray-500 mb-1">应预缴额总和</div>
                        <div class="text-lg font-bold text-red-600">¥{{ formatMoney(result.totals.taxYear) }}</div>
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- 底部说明 -->
    <div class="max-w-6xl mx-auto px-4 py-6">
      <div class="text-center text-xs text-gray-400">
        <p>* 计算结果仅供参考，实际收入以公司发放为准</p>
        <p class="mt-1">* 各地社保、公积金政策可能有所不同，请以当地最新政策为准</p>
      </div>
    </div>
  </div>
</template>
