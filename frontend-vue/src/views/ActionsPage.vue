<template>
  <div class="actions-page p-4">
    <div class="actions-grid grid grid-cols-2 gap-4">
      <!-- 记支出 -->
      <button 
        @click="handleAction('addExpense')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <i class="iconfont icon-minus-circle text-4xl mb-2 text-red-500"></i>
        <div class="font-semibold">{{ t('actions.addExpense') }}</div>
      </button>

      <!-- 记收入 -->
      <button 
        @click="handleAction('addIncome')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <i class="iconfont icon-add-circle text-4xl mb-2 text-green-500"></i>
        <div class="font-semibold">{{ t('actions.addIncome') }}</div>
      </button>

      <!-- 转账 -->
      <button 
        @click="handleAction('transfer')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <i class="iconfont icon-transfer text-4xl mb-2 text-blue-500"></i>
        <div class="font-semibold">{{ t('actions.transfer') }}</div>
      </button>

      <!-- 设置预算 -->
      <button 
        @click="handleAction('setBudget')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <i class="iconfont icon-target text-4xl mb-2 text-purple-500"></i>
        <div class="font-semibold">{{ t('actions.setBudget') }}</div>
      </button>

      <!-- 新增账户 -->
      <button 
        @click="handleAction('addAccount')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <i class="iconfont icon-bank text-4xl mb-2 text-indigo-500"></i>
        <div class="font-semibold">{{ t('actions.addAccount') }}</div>
      </button>

      <!-- 导入账单 -->
      <button 
        @click="handleAction('import')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <i class="iconfont icon-import text-4xl mb-2 text-orange-500"></i>
        <div class="font-semibold">{{ t('actions.import') }}</div>
      </button>
    </div>

    <!-- 提示信息 -->
    <div class="mt-6 p-4 bg-blue-50 rounded-lg flex items-center gap-2">
      <i class="iconfont icon-info-circle text-blue-500"></i>
      <p class="text-sm text-blue-700">
        {{ t('actions.comingSoon') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from '@/i18n'

function handleAction(action: string) {
  // TODO: 后续实现具体功能
  console.log('Action clicked:', action)
  const actionName = t(`actions.${action}`)
  alert(`TODO: ${t('actions.todoAlert').replace('{action}', actionName)}`)
}
</script>

<style scoped>
.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 120px;
}

.action-btn:active {
  transform: scale(0.98);
}
</style>

