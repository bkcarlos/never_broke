<script setup lang="ts">
import type { CalcSalaryMonth } from '@/types'
import { formatMoney } from '@/utils/format'

defineProps<{
  months: CalcSalaryMonth[]
}>()
</script>

<template>
  <div class="overflow-x-auto">
    <table class="min-w-full bg-white rounded-lg shadow-sm border border-gray-100">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            月份
          </th>
          <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            税前
          </th>
          <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            社保(个人)
          </th>
          <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            公积金(个人)
          </th>
          <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            预缴个税
          </th>
          <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            税后到手
          </th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-100">
        <tr v-for="month in months" :key="month.month" class="hover:bg-gray-50">
          <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
            {{ month.month }}月
          </td>
          <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900">
            {{ formatMoney(month.gross) }}
          </td>
          <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-orange-600">
            -{{ formatMoney(month.ssPerson) }}
          </td>
          <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-blue-600">
            -{{ formatMoney(month.hfPerson) }}
          </td>
          <td class="px-4 py-3 whitespace-nowrap text-sm text-right text-red-600">
            -{{ formatMoney(month.tax) }}
          </td>
          <td class="px-4 py-3 whitespace-nowrap text-sm text-right font-semibold text-green-600">
            {{ formatMoney(month.net) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
