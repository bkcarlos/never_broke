<template>
  <div class="page-container">
    <div class="page-content px-4">
    <div class="actions-grid grid grid-cols-2 gap-4">
      <!-- 记支出 -->
      <button 
        @click="handleAction('addExpense')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <Icon name="minus-circle" size="40px" class="mb-2 text-red-500" />
        <div class="font-semibold">{{ t('actions.addExpense') }}</div>
      </button>

      <!-- 记收入 -->
      <button 
        @click="handleAction('addIncome')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <Icon name="add-circle" size="40px" class="mb-2 text-green-500" />
        <div class="font-semibold">{{ t('actions.addIncome') }}</div>
      </button>

      <!-- 转账 -->
      <button 
        @click="handleAction('transfer')"
        class="action-btn bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow"
      >
        <Icon name="transfer" size="40px" class="mb-2 text-blue-500" />
        <div class="font-semibold">{{ t('actions.transfer') }}</div>
      </button>
    </div>

    <!-- 提示信息 -->
    <div class="mt-6 p-4 bg-blue-50 rounded-lg flex items-center gap-2">
      <Icon name="info-circle" size="20px" class="text-blue-500 flex-shrink-0" />
      <p class="text-sm text-blue-700">
        {{ t('actions.comingSoon') }}
      </p>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { t } from '@/i18n'
import Icon from '@/components/Icon.vue'

const router = useRouter()

function handleAction(action: string) {
  if (action === 'transfer') {
    router.push('/transfer')
    return
  }
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

