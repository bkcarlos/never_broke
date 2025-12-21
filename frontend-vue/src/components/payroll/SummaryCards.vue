<script setup lang="ts">
import type { CalcSalarySummary } from '@/types'
import { formatMoney, formatPercent } from '@/utils/format'

defineProps<{
  summary: CalcSalarySummary
  ssPersonRate: number
  ssCompanyRate: number
  hfPersonRate: number
  hfCompanyRate: number
}>()
</script>

<template>
  <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
    <!-- 税前月薪 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">税前月薪</div>
      <div class="text-xl font-semibold text-gray-900">
        {{ formatMoney(summary.grossMonthly) }}
      </div>
    </div>

    <!-- 个人社保 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">
        个人社保 <span class="text-xs">({{ formatPercent(ssPersonRate) }})</span>
      </div>
      <div class="text-xl font-semibold text-orange-600">
        -{{ formatMoney(summary.ssPersonMonthly) }}
      </div>
    </div>

    <!-- 公司社保 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">
        公司社保 <span class="text-xs">({{ formatPercent(ssCompanyRate) }})</span>
      </div>
      <div class="text-xl font-semibold text-gray-600">
        {{ formatMoney(summary.ssCompanyMonthly) }}
      </div>
    </div>

    <!-- 个人公积金 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">
        个人公积金 <span class="text-xs">({{ formatPercent(hfPersonRate) }})</span>
      </div>
      <div class="text-xl font-semibold text-blue-600">
        -{{ formatMoney(summary.hfPersonMonthly) }}
      </div>
    </div>

    <!-- 公司公积金 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">
        公司公积金 <span class="text-xs">({{ formatPercent(hfCompanyRate) }})</span>
      </div>
      <div class="text-xl font-semibold text-gray-600">
        {{ formatMoney(summary.hfCompanyMonthly) }}
      </div>
    </div>

    <!-- 专项扣除 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">专项附加扣除</div>
      <div class="text-xl font-semibold text-purple-600">
        {{ formatMoney(summary.specialMonthly) }}
      </div>
    </div>

    <!-- 公司成本 -->
    <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
      <div class="text-sm text-gray-500 mb-1">公司用工成本</div>
      <div class="text-xl font-semibold text-gray-900">
        {{ formatMoney(summary.companyCostMonthly) }}
      </div>
    </div>
  </div>
</template>
