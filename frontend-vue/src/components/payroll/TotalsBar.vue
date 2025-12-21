<script setup lang="ts">
import type { CalcSalaryTotals } from '@/types'
import { formatMoney } from '@/utils/format'

defineProps<{
  totals: CalcSalaryTotals
}>()
</script>

<template>
  <div class="flex flex-col sm:flex-row gap-4 p-4 bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-100">
    <div class="flex-1 text-center sm:text-left">
      <div class="text-sm text-gray-600 mb-1">年税后合计</div>
      <div class="text-2xl font-bold text-green-600">
        {{ formatMoney(totals.netYear) }}
      </div>
    </div>
    <div class="hidden sm:block w-px bg-gray-200"></div>
    <div class="flex-1 text-center sm:text-left">
      <div class="text-sm text-gray-600 mb-1">年预缴个税合计</div>
      <div class="text-2xl font-bold text-red-600">
        {{ formatMoney(totals.taxYear) }}
      </div>
    </div>
  </div>
</template>
