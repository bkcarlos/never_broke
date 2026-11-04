<template>
  <div class="min-h-screen bg-gradient-to-b from-blue-50 to-gray-50">
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 sticky top-0 z-20 shadow-sm">
      <div class="max-w-4xl mx-auto px-4 py-3">
        <div class="flex items-center gap-3">
          <button 
            @click="router.back()"
            class="w-9 h-9 flex items-center justify-center rounded-full bg-gray-100 hover:bg-gray-200 transition-colors"
          >
            <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-lg font-semibold text-gray-900">{{ t('me.salary.title', '薪资管理') }}</h1>
            <p class="text-xs text-gray-500">五险一金 · 个税计算 · 工资方案</p>
          </div>
        </div>
      </div>
    </header>

    <main class="max-w-4xl mx-auto px-4 py-6 space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12">
        <div class="inline-flex items-center gap-2 text-gray-500">
          <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span>{{ t('common.loading', '加载中...') }}</span>
        </div>
      </div>

      <template v-else>
        <!-- Current Salary Plan Card -->
        <section v-if="currentPlan" class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
          <div class="bg-gradient-to-r from-emerald-600 to-emerald-500 px-5 py-3 flex items-center justify-between">
            <h2 class="text-white font-medium">当前工资方案</h2>
            <span class="px-2 py-1 rounded-full text-xs bg-white/20 text-white">
              {{ currentPlan.active ? '已启用' : '已停用' }}
            </span>
          </div>
          <div class="p-5">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="p-3 bg-gray-50 rounded-lg">
                <div class="text-xs text-gray-500 mb-1">税前月薪</div>
                <div class="text-lg font-bold text-gray-900">¥{{ (currentPlan.salaryPreTaxMonthly || 0).toLocaleString() }}</div>
              </div>
              <div class="p-3 bg-gray-50 rounded-lg">
                <div class="text-xs text-gray-500 mb-1">发薪日</div>
                <div class="text-lg font-bold text-gray-900">每月{{ currentPlan.salaryPayDay || currentPlan.dayOfMonth }}日</div>
              </div>
              <div class="p-3 bg-gray-50 rounded-lg">
                <div class="text-xs text-gray-500 mb-1">纳税城市</div>
                <div class="text-lg font-bold text-gray-900">{{ getCityName(currentPlan.cityCode || '') }}</div>
              </div>
              <div class="p-3 bg-gray-50 rounded-lg">
                <div class="text-xs text-gray-500 mb-1">到账账户</div>
                <div class="text-lg font-bold text-gray-900 truncate">{{ getAccountName(currentPlan.accountId) }}</div>
              </div>
            </div>
            <div class="mt-4 flex gap-3">
              <button
                @click="handleEditPlan"
                class="flex-1 py-2.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium transition-colors"
              >
                编辑方案
              </button>
              <button
                @click="handleDeletePlan"
                class="px-6 py-2.5 rounded-lg bg-red-50 hover:bg-red-100 text-red-600 text-sm font-medium transition-colors"
              >
                删除
              </button>
            </div>
          </div>
        </section>

        <div v-else class="bg-white rounded-xl shadow-sm border border-gray-100 p-8 text-center">
          <div class="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="text-gray-500">暂未配置工资方案</p>
          <p class="text-sm text-gray-400 mt-1">在下方填写信息后保存</p>
        </div>

        <!-- Salary Configuration Form -->
        <section class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
          <div class="bg-gradient-to-r from-blue-600 to-blue-500 px-5 py-3">
            <h2 class="text-white font-medium">工资参数配置</h2>
          </div>
          <div class="p-5 space-y-5">
            <!-- 基本信息 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- 税前月薪 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  税前月薪 <span class="text-red-500">*</span>
                </label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">¥</span>
                  <input
                    v-model.number="form.preTaxMonthly"
                    type="number"
                    step="100"
                    min="0"
                    class="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all text-lg font-medium"
                    placeholder="请输入税前月薪"
                  />
                </div>
              </div>

              <!-- 纳税城市 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  {{ t('me.salary.city') }} <span class="text-red-500">*</span>
                </label>
                <div class="relative">
                  <select
                    v-model="form.cityCode"
                    @change="handleCityChange"
                    class="w-full px-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                    :disabled="loadingRegionRule"
                  >
                    <!-- 优先使用 API 返回的地区列表 -->
                    <template v-if="regions.length > 0">
                      <option v-for="region in regions" :key="region.code" :value="mapRegionToCity(region.code)">
                        {{ region.name }}
                      </option>
                    </template>
                    <!-- 降级到本地选项 -->
                    <template v-else>
                      <option value="beijing">{{ t('city.beijing') }}</option>
                      <option value="shanghai">{{ t('city.shanghai') }}</option>
                      <option value="guangzhou">{{ t('city.guangzhou') }}</option>
                      <option value="shenzhen">{{ t('city.shenzhen') }}</option>
                      <option value="default">{{ t('city.default') }}</option>
                    </template>
                  </select>
                  <div v-if="loadingRegionRule" class="absolute right-10 top-1/2 -translate-y-1/2">
                    <svg class="animate-spin h-4 w-4 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                  </div>
                </div>
                <p v-if="regionRule" class="text-xs text-green-600 mt-1">
                  {{ t('me.salary.regionRuleLoaded', { region: regionRule.regionCode }) }}
                </p>
              </div>
            </div>

            <!-- 社保配置 -->
            <div class="border-t border-gray-100 pt-5">
              <div class="flex items-center justify-between mb-4">
                <div>
                  <h3 class="font-medium text-gray-900">社保缴纳比例</h3>
                  <p class="text-xs text-gray-500 mt-0.5">个人缴纳部分，可根据实际情况调整</p>
                </div>
                <div class="flex gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input 
                      type="radio" 
                      :checked="!customRatesEnabled" 
                      @change="customRatesEnabled = false"
                      class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                    />
                    <span class="text-sm text-gray-700">使用默认</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input 
                      type="radio" 
                      :checked="customRatesEnabled" 
                      @change="customRatesEnabled = true"
                      class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                    />
                    <span class="text-sm text-gray-700">自定义</span>
                  </label>
                </div>
              </div>

              <!-- 默认比例显示 -->
              <div v-if="!customRatesEnabled" class="bg-gray-50 rounded-lg p-4">
                <div class="grid grid-cols-2 md:grid-cols-5 gap-3 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-600">养老保险</span>
                    <span class="font-medium">8.0%</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">医疗保险</span>
                    <span class="font-medium">2.0%</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">失业保险</span>
                    <span class="font-medium">0.2%</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">住房公积金</span>
                    <span class="font-medium">{{ (form.hfRate * 100).toFixed(0) }}%</span>
                  </div>
                  <div class="flex justify-between font-medium">
                    <span class="text-gray-800">合计</span>
                    <span class="text-blue-600">{{ defaultTotalRate.toFixed(1) }}%</span>
                  </div>
                </div>
              </div>

              <!-- 自定义比例调整 -->
              <div v-else class="space-y-3">
                <!-- 养老保险 -->
                <div class="p-4 bg-orange-50 rounded-lg">
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-sm font-medium text-gray-700">🧓 养老保险</span>
                    <span class="text-xs text-gray-500">个人缴纳</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <input
                      v-model.number="ssRates.pensionPerson"
                      type="range"
                      min="0"
                      max="0.2"
                      step="0.005"
                      class="flex-1 h-2 bg-orange-200 rounded-lg appearance-none cursor-pointer"
                    />
                    <div class="w-20 text-right">
                      <span class="text-lg font-bold text-orange-600">{{ (ssRates.pensionPerson * 100).toFixed(1) }}%</span>
                    </div>
                  </div>
                </div>

                <!-- 医疗保险 -->
                <div class="p-4 bg-green-50 rounded-lg">
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-sm font-medium text-gray-700">🏥 医疗保险</span>
                    <span class="text-xs text-gray-500">个人缴纳</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <input
                      v-model.number="ssRates.medicalPerson"
                      type="range"
                      min="0"
                      max="0.1"
                      step="0.005"
                      class="flex-1 h-2 bg-green-200 rounded-lg appearance-none cursor-pointer"
                    />
                    <div class="w-20 text-right">
                      <span class="text-lg font-bold text-green-600">{{ (ssRates.medicalPerson * 100).toFixed(1) }}%</span>
                    </div>
                  </div>
                </div>

                <!-- 失业保险 -->
                <div class="p-4 bg-blue-50 rounded-lg">
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-sm font-medium text-gray-700">💼 失业保险</span>
                    <span class="text-xs text-gray-500">个人缴纳</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <input
                      v-model.number="ssRates.unemploymentPerson"
                      type="range"
                      min="0"
                      max="0.02"
                      step="0.001"
                      class="flex-1 h-2 bg-blue-200 rounded-lg appearance-none cursor-pointer"
                    />
                    <div class="w-20 text-right">
                      <span class="text-lg font-bold text-blue-600">{{ (ssRates.unemploymentPerson * 100).toFixed(1) }}%</span>
                    </div>
                  </div>
                </div>

                <!-- 住房公积金 -->
                <div class="p-4 bg-purple-50 rounded-lg">
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-sm font-medium text-gray-700">🏠 住房公积金</span>
                    <span class="text-xs text-gray-500">个人缴纳</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <input
                      v-model.number="form.hfRate"
                      type="range"
                      min="0.05"
                      max="0.12"
                      step="0.01"
                      class="flex-1 h-2 bg-purple-200 rounded-lg appearance-none cursor-pointer"
                    />
                    <div class="w-20 text-right">
                      <span class="text-lg font-bold text-purple-600">{{ (form.hfRate * 100).toFixed(0) }}%</span>
                    </div>
                  </div>
                </div>

                <!-- 比例汇总 -->
                <div class="p-4 bg-gray-100 rounded-lg flex justify-between items-center">
                  <span class="font-medium text-gray-700">个人缴纳合计</span>
                  <span class="text-xl font-bold text-blue-600">{{ customTotalRate.toFixed(1) }}%</span>
                </div>
              </div>

              <!-- 缴费基数信息 -->
              <div class="mt-4 p-4 bg-blue-50 border border-blue-100 rounded-lg">
                <div class="flex items-start gap-3">
                  <span class="text-xl">📊</span>
                  <div class="flex-1">
                    <h4 class="font-medium text-blue-800 mb-2">{{ currentCityLimit.name }}社保基数范围（2024年）</h4>
                    <div class="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <div class="text-blue-600 mb-1">社保基数</div>
                        <div class="flex justify-between">
                          <span class="text-gray-600">下限</span>
                          <span class="font-medium">¥{{ currentCityLimit.ssMin.toLocaleString() }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-gray-600">上限</span>
                          <span class="font-medium">¥{{ currentCityLimit.ssMax.toLocaleString() }}</span>
                        </div>
                      </div>
                      <div>
                        <div class="text-blue-600 mb-1">公积金基数</div>
                        <div class="flex justify-between">
                          <span class="text-gray-600">下限</span>
                          <span class="font-medium">¥{{ currentCityLimit.hfMin.toLocaleString() }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-gray-600">上限</span>
                          <span class="font-medium">¥{{ currentCityLimit.hfMax.toLocaleString() }}</span>
                        </div>
                      </div>
                    </div>
                    
                    <!-- 实际使用的基数 -->
                    <div v-if="form.preTaxMonthly > 0" class="mt-3 pt-3 border-t border-blue-200">
                      <div class="text-blue-700 font-medium mb-1">你的实际缴费基数</div>
                      <div class="grid grid-cols-2 gap-4 text-sm">
                        <div class="flex items-center justify-between">
                          <span class="text-gray-600">社保基数</span>
                          <div class="flex items-center gap-2">
                            <span class="font-bold text-blue-800">¥{{ actualSsBase.toLocaleString() }}</span>
                            <span v-if="isAtSsMaxBase" class="px-1.5 py-0.5 bg-amber-100 text-amber-700 text-xs rounded">封顶</span>
                            <span v-else-if="isAtSsMinBase" class="px-1.5 py-0.5 bg-amber-100 text-amber-700 text-xs rounded">保底</span>
                          </div>
                        </div>
                        <div class="flex items-center justify-between">
                          <span class="text-gray-600">公积金基数</span>
                          <div class="flex items-center gap-2">
                            <span class="font-bold text-blue-800">¥{{ actualHfBase.toLocaleString() }}</span>
                            <span v-if="isAtHfMaxBase" class="px-1.5 py-0.5 bg-amber-100 text-amber-700 text-xs rounded">封顶</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 专项附加扣除 -->
            <div class="border-t border-gray-100 pt-5">
              <div class="flex items-center justify-between mb-4">
                <div>
                  <h3 class="font-medium text-gray-900">专项附加扣除</h3>
                  <p class="text-xs text-gray-500 mt-0.5">可减少应纳税所得额，降低个税</p>
                </div>
                <div class="flex gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input 
                      type="radio" 
                      :checked="!specialDeductionEnabled" 
                      @change="specialDeductionEnabled = false"
                      class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                    />
                    <span class="text-sm text-gray-700">不扣除</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input 
                      type="radio" 
                      :checked="specialDeductionEnabled" 
                      @change="specialDeductionEnabled = true"
                      class="w-4 h-4 text-blue-600 focus:ring-blue-500"
                    />
                    <span class="text-sm text-gray-700">配置扣除</span>
                  </label>
                </div>
              </div>

              <!-- 不扣除时的提示 -->
              <div v-if="!specialDeductionEnabled" class="bg-gray-50 rounded-lg p-4 text-center text-sm text-gray-500">
                未配置专项附加扣除
              </div>

              <!-- 各项扣除配置 -->
              <div v-else class="space-y-3">
                <!-- 子女教育 -->
                <div class="p-4 bg-blue-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">👶</span>
                      <span class="text-sm font-medium text-gray-700">子女教育</span>
                    </div>
                    <span class="text-xs text-gray-500">每个子女2000元/月</span>
                  </div>
                  <select
                    v-model.number="specialDeductions.childEdu"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option v-for="opt in childEduOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>

                <!-- 继续教育 -->
                <div class="p-4 bg-green-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">📚</span>
                      <span class="text-sm font-medium text-gray-700">继续教育</span>
                    </div>
                    <span class="text-xs text-gray-500">学历教育400元/月</span>
                  </div>
                  <select
                    v-model.number="specialDeductions.contEdu"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option v-for="opt in contEduOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>

                <!-- 赡养老人 -->
                <div class="p-4 bg-orange-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">🧓</span>
                      <span class="text-sm font-medium text-gray-700">赡养老人</span>
                    </div>
                    <span class="text-xs text-gray-500">独生子女3000元/月</span>
                  </div>
                  <select
                    v-model.number="specialDeductions.elderlyCare"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option v-for="opt in elderlyCareOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>

                <!-- 住房贷款利息 -->
                <div class="p-4 bg-purple-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">🏠</span>
                      <span class="text-sm font-medium text-gray-700">住房贷款利息</span>
                    </div>
                    <span class="text-xs text-gray-500">首套房最长240个月</span>
                  </div>
                  <select
                    v-model.number="specialDeductions.housingLoan"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option v-for="opt in housingLoanOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>

                <!-- 住房租金 -->
                <div class="p-4 bg-teal-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">🏢</span>
                      <span class="text-sm font-medium text-gray-700">住房租金</span>
                    </div>
                    <span class="text-xs text-gray-500">与贷款利息二选一</span>
                  </div>
                  <select
                    v-model.number="specialDeductions.housingRent"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option v-for="opt in housingRentOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>

                <!-- 3岁以下婴幼儿照护 -->
                <div class="p-4 bg-pink-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">👼</span>
                      <span class="text-sm font-medium text-gray-700">3岁以下婴幼儿照护</span>
                    </div>
                    <span class="text-xs text-gray-500">每个婴幼儿2000元/月</span>
                  </div>
                  <select
                    v-model.number="specialDeductions.infantCare"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option v-for="opt in infantCareOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>

                <!-- 个人养老金 -->
                <div class="p-4 bg-indigo-50 rounded-lg">
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2">
                      <span class="text-lg">💰</span>
                      <span class="text-sm font-medium text-gray-700">个人养老金</span>
                    </div>
                    <span class="text-xs text-gray-500">年上限12000元</span>
                  </div>
                  <div class="relative">
                    <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 text-sm">¥</span>
                    <input
                      v-model.number="specialDeductions.pensionMonthly"
                      type="number"
                      min="0"
                      max="1000"
                      step="100"
                      class="w-full pl-8 pr-12 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500"
                      placeholder="0"
                    />
                    <span class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 text-sm">/月</span>
                  </div>
                  <p class="text-xs text-gray-500 mt-1">每月最多1000元</p>
                </div>

                <!-- 扣除汇总 -->
                <div class="p-4 bg-gray-100 rounded-lg flex justify-between items-center">
                  <span class="font-medium text-gray-700">每月扣除合计</span>
                  <span class="text-xl font-bold text-emerald-600">¥{{ totalSpecialDeduction.toLocaleString() }}</span>
                </div>

                <!-- 大病医疗提示 -->
                <div class="p-3 bg-amber-50 border border-amber-200 rounded-lg">
                  <p class="text-xs text-amber-700">
                    💡 大病医疗扣除需次年汇算清缴时自行申报，本工具暂不支持按月计算
                  </p>
                </div>
              </div>
            </div>

            <!-- 发薪设置 -->
            <div class="border-t border-gray-100 pt-5">
              <h3 class="font-medium text-gray-900 mb-4">发薪设置</h3>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    开始日期 <span class="text-red-500">*</span>
                  </label>
                  <input
                    v-model="form.startDate"
                    type="date"
                    class="w-full px-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    发薪日 <span class="text-red-500">*</span>
                  </label>
                  <div class="relative">
                    <input
                      v-model.number="form.payDay"
                      type="number"
                      min="1"
                      max="31"
                      class="w-full px-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                      placeholder="10"
                    />
                    <span class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 text-sm">日</span>
                  </div>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    到账账户 <span class="text-red-500">*</span>
                  </label>
                  <select
                    v-model="form.accountId"
                    class="w-full px-4 py-2.5 border border-gray-200 rounded-lg bg-gray-50 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                  >
                    <option value="" disabled>请选择账户</option>
                    <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
                      {{ formatAccountDisplayName(acc) }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="pt-4 space-y-3">
              <button
                @click="handleSaveAndPreview"
                :disabled="saving || !isFormValid"
                class="w-full py-3.5 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-700 hover:to-blue-600 disabled:from-gray-400 disabled:to-gray-400 text-white font-medium rounded-xl shadow-lg shadow-blue-500/25 transition-all flex items-center justify-center gap-2"
              >
                <svg v-if="saving" class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                {{ saving ? '保存中...' : '保存为工资方案' }}
              </button>
              <button
                @click="handlePreviewOnly"
                :disabled="previewing || !isFormValid"
                class="w-full py-3 bg-gray-100 hover:bg-gray-200 disabled:opacity-60 text-gray-700 font-medium rounded-xl transition-all flex items-center justify-center gap-2"
              >
                <svg v-if="previewing" class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                {{ previewing ? '生成中...' : '仅预览' }}
              </button>
            </div>
          </div>
        </section>

        <!-- Salary Preview Results -->
        <section v-if="salaryPreview" class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
          <div class="bg-gradient-to-r from-green-600 to-emerald-500 px-5 py-3">
            <h2 class="text-white font-medium">薪资预览</h2>
          </div>
          <div class="p-5 space-y-5">
            <!-- 年度汇总 - 第一行 -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="p-4 bg-gray-50 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">全年税前</div>
                <div class="text-xl font-bold text-gray-900">¥{{ salaryPreview.totalPreTax.toLocaleString() }}</div>
              </div>
              <div class="p-4 bg-emerald-50 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">全年税后到手</div>
                <div class="text-xl font-bold text-emerald-600">¥{{ salaryPreview.totalPostTax.toLocaleString() }}</div>
              </div>
              <div class="p-4 bg-red-50 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">全年个税</div>
                <div class="text-xl font-bold text-red-600">¥{{ salaryPreview.totalTax.toLocaleString() }}</div>
              </div>
              <div class="p-4 bg-orange-50 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">全年社保+公积金</div>
                <div class="text-xl font-bold text-orange-600">¥{{ totalSocialSecurityYear.toLocaleString() }}</div>
              </div>
            </div>

            <!-- 年度汇总 - 第二行：公积金详情 -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="p-4 bg-purple-50 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">个人公积金/月</div>
                <div class="text-lg font-bold text-purple-600">¥{{ monthlyHousingFundPerson.toLocaleString() }}</div>
              </div>
              <div class="p-4 bg-purple-50 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">公司公积金/月</div>
                <div class="text-lg font-bold text-purple-600">¥{{ monthlyHousingFundCompany.toLocaleString() }}</div>
              </div>
              <div class="p-4 bg-purple-100 rounded-lg text-center">
                <div class="text-sm text-gray-500 mb-1">公积金总计/月</div>
                <div class="text-lg font-bold text-purple-700">¥{{ monthlyHousingFundTotal.toLocaleString() }}</div>
              </div>
              <div class="p-4 bg-gradient-to-r from-purple-500 to-purple-600 rounded-lg text-center">
                <div class="text-sm text-purple-100 mb-1">全年公积金总计</div>
                <div class="text-xl font-bold text-white">¥{{ yearlyHousingFundTotal.toLocaleString() }}</div>
              </div>
            </div>

            <!-- 月度明细表格 -->
            <div class="overflow-x-auto">
              <table class="min-w-full text-sm">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-3 py-3 text-left text-xs font-medium text-gray-500 uppercase">月份</th>
                    <th class="px-3 py-3 text-right text-xs font-medium text-gray-500 uppercase">税前</th>
                    <th class="px-3 py-3 text-right text-xs font-medium text-gray-500 uppercase">社保</th>
                    <th class="px-3 py-3 text-right text-xs font-medium text-gray-500 uppercase">公积金(个人)</th>
                    <th class="px-3 py-3 text-right text-xs font-medium text-gray-500 uppercase">个税</th>
                    <th class="px-3 py-3 text-right text-xs font-medium text-gray-500 uppercase">税后到手</th>
                    <th class="px-3 py-3 text-right text-xs font-medium text-purple-600 uppercase bg-purple-50">公积金累计</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr v-for="(item, index) in salaryPreview.items" :key="item.date" class="hover:bg-gray-50">
                    <td class="px-3 py-3 whitespace-nowrap font-medium text-gray-900">
                      {{ item.date }}
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap text-right text-gray-700">
                      ¥{{ item.preTax.toLocaleString() }}
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap text-right text-orange-600">
                      -¥{{ monthlySocialSecurityOnly.toLocaleString() }}
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap text-right text-purple-600">
                      -¥{{ monthlyHousingFundPerson.toLocaleString() }}
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap text-right text-red-600">
                      -¥{{ item.taxAmount.toLocaleString() }}
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap text-right font-bold text-emerald-600">
                      ¥{{ item.postTax.toLocaleString() }}
                    </td>
                    <td class="px-3 py-3 whitespace-nowrap text-right font-bold text-purple-700 bg-purple-50">
                      ¥{{ ((index + 1) * monthlyHousingFundTotal).toLocaleString() }}
                    </td>
                  </tr>
                </tbody>
                <tfoot class="bg-gradient-to-r from-gray-100 to-purple-100">
                  <tr class="font-bold">
                    <td class="px-3 py-4 text-gray-900">年度合计</td>
                    <td class="px-3 py-4 text-right text-gray-900">¥{{ salaryPreview.totalPreTax.toLocaleString() }}</td>
                    <td class="px-3 py-4 text-right text-orange-600">-¥{{ (monthlySocialSecurityOnly * salaryPreview.items.length).toLocaleString() }}</td>
                    <td class="px-3 py-4 text-right text-purple-600">-¥{{ (monthlyHousingFundPerson * salaryPreview.items.length).toLocaleString() }}</td>
                    <td class="px-3 py-4 text-right text-red-600">-¥{{ salaryPreview.totalTax.toLocaleString() }}</td>
                    <td class="px-3 py-4 text-right text-emerald-600">¥{{ salaryPreview.totalPostTax.toLocaleString() }}</td>
                    <td class="px-3 py-4 text-right text-purple-700 bg-purple-100">¥{{ yearlyHousingFundTotal.toLocaleString() }}</td>
                  </tr>
                </tfoot>
              </table>
            </div>

            <!-- 公积金说明 -->
            <div class="p-4 bg-purple-50 border border-purple-100 rounded-lg">
              <div class="flex items-start gap-3">
                <span class="text-2xl">🏠</span>
                <div>
                  <h4 class="font-medium text-purple-800 mb-1">公积金账户说明</h4>
                  <p class="text-sm text-purple-700">
                    公积金由个人和公司各缴纳 <span class="font-bold">{{ (form.hfRate * 100).toFixed(0) }}%</span>，
                    每月入账 <span class="font-bold">¥{{ monthlyHousingFundTotal.toLocaleString() }}</span>，
                    全年累计 <span class="font-bold">¥{{ yearlyHousingFundTotal.toLocaleString() }}</span>。
                    可用于购房、租房提取或公积金贷款。
                  </p>
                </div>
              </div>
            </div>

            <!-- 免责声明 -->
            <p class="text-xs text-gray-400 text-center">
              * 计算结果仅供参考，实际收入以公司发放与当地政策为准
            </p>
          </div>
        </section>

        <!-- Error Message -->
        <div v-if="errorMessage" class="p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
          {{ errorMessage }}
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '@/i18n'
import type {
  RecurringIncomePlan,
  Account,
  SalaryPreviewResponse,
  SalaryPlanConfig,
  PayrollRegion,
  PayrollRegionRule
} from '@/types'
import { getSalaryPlan, upsertSalaryPlan, deleteSalaryPlan } from '@/api/meSalary'
import { getAccounts } from '@/api/accounts'
import { previewSalaryToYearEnd } from '@/api/recurringIncome'
import { formatAccountDisplayName } from '@/utils/accountDisplay'
import { getPayrollRegions, getPayrollRegionRule } from '@/api/payroll'

const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const previewing = ref(false)
const loadingRegionRule = ref(false)
const errorMessage = ref<string | null>(null)

const currentPlan = ref<RecurringIncomePlan | null>(null)
const accounts = ref<Account[]>([])
const salaryPreview = ref<SalaryPreviewResponse | null>(null)

// 地区列表和规则
const regions = ref<PayrollRegion[]>([])
const regionRule = ref<PayrollRegionRule | null>(null)

// 是否使用自定义比例
const customRatesEnabled = ref(false)

// 社保各项比例（默认值）
const ssRates = reactive({
  pensionPerson: 0.08,      // 养老保险个人
  medicalPerson: 0.02,      // 医疗保险个人
  unemploymentPerson: 0.002 // 失业保险个人
})

// 专项附加扣除开关
const specialDeductionEnabled = ref(false)

// 专项附加扣除各项
const specialDeductions = reactive({
  childEdu: 0,        // 子女教育
  contEdu: 0,         // 继续教育
  elderlyCare: 0,     // 赡养老人
  housingLoan: 0,     // 住房贷款利息
  housingRent: 0,     // 住房租金
  infantCare: 0,      // 3岁以下婴幼儿照护
  pensionMonthly: 0   // 个人养老金
})

// 专项扣除选项（使用 computed 以支持多语言）
const childEduOptions = computed(() => [
  { value: 0, label: t('deduction.notApplicable') },
  { value: 1000, label: t('deduction.childEdu.one1000') },
  { value: 2000, label: t('deduction.childEdu.one2000') },
  { value: 3000, label: t('deduction.childEdu.two2000') },
  { value: 4000, label: t('deduction.childEdu.two4000') }
])

const contEduOptions = computed(() => [
  { value: 0, label: t('deduction.notApplicable') },
  { value: 400, label: t('deduction.contEdu.degree') }
])

const elderlyCareOptions = computed(() => [
  { value: 0, label: t('deduction.notApplicable') },
  { value: 500, label: t('deduction.elderlyCare.six') },
  { value: 750, label: t('deduction.elderlyCare.four') },
  { value: 1000, label: t('deduction.elderlyCare.three') },
  { value: 1500, label: t('deduction.elderlyCare.two') },
  { value: 3000, label: t('deduction.elderlyCare.only') }
])

const housingLoanOptions = computed(() => [
  { value: 0, label: t('deduction.notApplicable') },
  { value: 500, label: t('deduction.housingLoan.both') },
  { value: 1000, label: t('deduction.housingLoan.one') }
])

const housingRentOptions = computed(() => [
  { value: 0, label: t('deduction.notApplicable') },
  { value: 800, label: t('deduction.housingRent.small') },
  { value: 1100, label: t('deduction.housingRent.medium') },
  { value: 1500, label: t('deduction.housingRent.large') }
])

const infantCareOptions = computed(() => [
  { value: 0, label: t('deduction.notApplicable') },
  { value: 1000, label: t('deduction.infantCare.one1000') },
  { value: 2000, label: t('deduction.infantCare.one2000') },
  { value: 3000, label: t('deduction.infantCare.two3000') },
  { value: 4000, label: t('deduction.infantCare.two4000') }
])

// 专项扣除合计
const totalSpecialDeduction = computed(() => {
  if (!specialDeductionEnabled.value) return 0
  return specialDeductions.childEdu +
         specialDeductions.contEdu +
         specialDeductions.elderlyCare +
         specialDeductions.housingLoan +
         specialDeductions.housingRent +
         specialDeductions.infantCare +
         specialDeductions.pensionMonthly
})

// 默认基数限制（当 API 未返回时使用）
const defaultCityLimit = {
  name: '默认',
  ssMin: 4000,
  ssMax: 25000,
  hfMin: 2000,
  hfMax: 30000
}

// 获取当前城市基数限制（从 API 获取）
const currentCityLimit = computed(() => {
  if (regionRule.value) {
    return {
      name: getCityName(form.value.cityCode),
      ssMin: regionRule.value.ssBaseMin,
      ssMax: regionRule.value.ssBaseMax,
      hfMin: regionRule.value.hfBaseMin,
      hfMax: regionRule.value.hfBaseMax
    }
  }
  return { ...defaultCityLimit, name: getCityName(form.value.cityCode) }
})

// 计算实际社保缴费基数（考虑上下限）
const actualSsBase = computed(() => {
  const salary = form.value.preTaxMonthly
  const limit = currentCityLimit.value
  if (salary < limit.ssMin) return limit.ssMin
  if (salary > limit.ssMax) return limit.ssMax
  return salary
})

// 计算实际公积金缴费基数（考虑上下限）
const actualHfBase = computed(() => {
  const salary = form.value.preTaxMonthly
  const limit = currentCityLimit.value
  if (salary < limit.hfMin) return limit.hfMin
  if (salary > limit.hfMax) return limit.hfMax
  return salary
})

// 是否触及基数上限
const isAtSsMaxBase = computed(() => form.value.preTaxMonthly > currentCityLimit.value.ssMax)
const isAtSsMinBase = computed(() => form.value.preTaxMonthly < currentCityLimit.value.ssMin)
const isAtHfMaxBase = computed(() => form.value.preTaxMonthly > currentCityLimit.value.hfMax)

const form = ref({
  preTaxMonthly: 0,
  cityCode: 'beijing',
  startDate: '',
  payDay: 10,
  accountId: '',
  currency: 'CNY',
  hfRate: 0.12 // 公积金比例
})

// 默认总比例
const defaultTotalRate = computed(() => {
  return 8 + 2 + 0.2 + form.value.hfRate * 100
})

// 自定义总比例
const customTotalRate = computed(() => {
  return (ssRates.pensionPerson + ssRates.medicalPerson + ssRates.unemploymentPerson + form.value.hfRate) * 100
})

// 当前使用的总比例
const currentSocialRate = computed(() => {
  if (customRatesEnabled.value) {
    return ssRates.pensionPerson + ssRates.medicalPerson + ssRates.unemploymentPerson + form.value.hfRate
  }
  return 0.08 + 0.02 + 0.002 + form.value.hfRate
})

const contributionConfig = computed(() => ({
  socialBase: Number(actualSsBase.value),
  housingBase: Number(actualHfBase.value),
  pensionRate: customRatesEnabled.value ? ssRates.pensionPerson : 0.08,
  medicalRate: customRatesEnabled.value ? ssRates.medicalPerson : 0.02,
  unemploymentRate: customRatesEnabled.value ? ssRates.unemploymentPerson : 0.002,
  housingRate: form.value.hfRate,
}))
const roundMoney = (value: number) => Math.round(value * 100) / 100

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

// 公积金计算（基于限制后的基数）
const monthlyHousingFundPerson = computed(() => {
  return roundMoney(actualHfBase.value * form.value.hfRate)
})

const monthlyHousingFundCompany = computed(() => {
  return roundMoney(actualHfBase.value * form.value.hfRate)
})

const monthlyHousingFundTotal = computed(() => {
  return monthlyHousingFundPerson.value + monthlyHousingFundCompany.value
})

const yearlyHousingFundTotal = computed(() => {
  const months = salaryPreview.value?.items?.length || 12
  return monthlyHousingFundTotal.value * months
})

// 社保计算（基于限制后的基数，不含公积金）
const monthlySocialSecurityOnly = computed(() => {
  const c = contributionConfig.value
  return roundMoney(c.socialBase*c.pensionRate) + roundMoney(c.socialBase*c.medicalRate) + roundMoney(c.socialBase*c.unemploymentRate)
})

// 全年社保+公积金总额（个人部分）
const totalSocialSecurityYear = computed(() => {
  const months = salaryPreview.value?.items?.length || 12
  return (monthlySocialSecurityOnly.value + monthlyHousingFundPerson.value) * months
})

function getCityName(cityCode: string): string {
  // 优先从 API 获取的地区列表中查找
  const region = regions.value.find(r => r.code === cityCode)
  if (region) return region.name
  
  // 降级到 i18n 翻译
  const cityKey = `city.${cityCode}` as const
  return t(cityKey) || cityCode
}

function getAccountName(accountId: string): string {
  const account = accounts.value.find((acc) => acc.id === accountId)
  return account ? formatAccountDisplayName(account) : accountId
}

// 加载地区规则
async function loadRegionRule(cityCode: string) {
  if (!cityCode) return
  
  loadingRegionRule.value = true
  try {
    // 将本地城市代码映射到 API 的地区代码
    const regionCode = mapCityToRegionCode(cityCode)
    const rule = await getPayrollRegionRule(regionCode, 2025)
    regionRule.value = rule
    
    // 更新公积金比例范围
    if (rule) {
      form.value.hfRate = rule.defaultHfRatePerson
    }
  } catch (error) {
    console.error('Failed to load region rule:', error)
    regionRule.value = null
  } finally {
    loadingRegionRule.value = false
  }
}

// 将本地城市代码映射到 API 地区代码
function mapCityToRegionCode(cityCode: string): string {
  const mapping: Record<string, string> = {
    'beijing': 'CN-BJ',
    'shanghai': 'CN-SH', 
    'guangzhou': 'CN-GZ',
    'shenzhen': 'CN-SZ',
    'default': 'CN-DEFAULT'
  }
  return mapping[cityCode] || 'CN-DEFAULT'
}

// 将 API 地区代码映射回本地城市代码
function mapRegionToCity(regionCode: string): string {
  const mapping: Record<string, string> = {
    'CN-BJ': 'beijing',
    'CN-SH': 'shanghai',
    'CN-GZ': 'guangzhou',
    'CN-SZ': 'shenzhen',
    'CN-DEFAULT': 'default'
  }
  return mapping[regionCode] || 'default'
}

// 城市变化时重新加载规则
async function handleCityChange() {
  await loadRegionRule(form.value.cityCode)
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
      socialRate: currentSocialRate.value,
      contributions: contributionConfig.value,
      extraDeductions: totalSpecialDeduction.value,
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
      socialRate: currentSocialRate.value,
      contributions: contributionConfig.value,
      extraDeductions: totalSpecialDeduction.value,
      accountId: form.value.accountId,
      currency: form.value.currency,
    }

    const result = await upsertSalaryPlan(payload)
    currentPlan.value = result.plan
    salaryPreview.value = result.preview

    alert('工资方案保存成功')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function handleEditPlan() {
  if (!currentPlan.value) return

  form.value = {
    preTaxMonthly: currentPlan.value.salaryPreTaxMonthly || 0,
    cityCode: currentPlan.value.cityCode || 'beijing',
    startDate: currentPlan.value.salaryStartDate || '',
    payDay: currentPlan.value.salaryPayDay || currentPlan.value.dayOfMonth || 10,
    accountId: currentPlan.value.accountId,
    currency: currentPlan.value.currency,
    hfRate: 0.12
  }

  // 加载对应城市的规则
  await loadRegionRule(form.value.cityCode)

  // 恢复专项附加扣除
  const savedDeductions = currentPlan.value.salaryExtraDeductions || 0
  if (savedDeductions > 0) {
    specialDeductionEnabled.value = true
    // 简化处理：将总额放入赡养老人项（实际应该分别保存各项）
    specialDeductions.elderlyCare = savedDeductions
  } else {
    specialDeductionEnabled.value = false
  }

  const contributions = currentPlan.value.salaryContributions
  customRatesEnabled.value = !!contributions
  if (contributions) {
    ssRates.pensionPerson = contributions.pensionRate
    ssRates.medicalPerson = contributions.medicalRate
    ssRates.unemploymentPerson = contributions.unemploymentRate
    form.value.hfRate = contributions.housingRate
  }
}

async function handleDeletePlan() {
  if (!currentPlan.value) return

  if (!confirm('确定要删除当前工资方案吗？')) return

  try {
    await deleteSalaryPlan(currentPlan.value.id)
    currentPlan.value = null
    salaryPreview.value = null
    alert('工资方案已删除')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '删除失败'
  }
}

async function loadData() {
  loading.value = true
  errorMessage.value = null

  try {
    // 并行加载账户、工资方案和地区列表
    const [accountsData, salaryPlanData, regionsData] = await Promise.all([
      getAccounts(),
      getSalaryPlan(),
      getPayrollRegions().catch(() => [])  // 如果失败则使用空数组
    ])

    accounts.value = accountsData.filter((acc) => !acc.archived)
    regions.value = regionsData

    if (salaryPlanData.plan) {
      currentPlan.value = salaryPlanData.plan
      await handleEditPlan()
    } else {
      // Initialize form with default values
      const today = new Date()
      form.value.startDate = today.toISOString().split('T')[0]
      if (accounts.value.length > 0) {
        form.value.accountId = accounts.value[0].id
      }
      // 加载默认城市的规则
      await loadRegionRule(form.value.cityCode)
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

<style scoped>
/* 自定义滑块样式 */
input[type="range"] {
  -webkit-appearance: none;
  appearance: none;
}

input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  border: 2px solid currentColor;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

input[type="range"]::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  border: 2px solid currentColor;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style>
