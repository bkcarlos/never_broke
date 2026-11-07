<template>
  <PageFrame title="快捷记账" eyebrow="QUICK ACTIONS" subtitle="选择记录类型，开始整理这一笔收支。" compact>
    <div class="action-grid">
      <RouterLink v-for="action in actions" :key="action.type" :to="{ path: '/record', query: { type: action.type } }" class="panel action-card">
        <span class="action-icon"><Icon :name="action.icon" size="24px" /></span>
        <h2>{{ action.title }}</h2><p>{{ action.description }}</p><span class="action-link">开始记录 →</span>
      </RouterLink>
    </div>
  </PageFrame>
</template>
<script setup lang="ts">
import PageFrame from '@/components/PageFrame.vue'
import Icon from '@/components/Icon.vue'
const actions = [
  { type: 'expense', icon: 'minus-circle', title: '记支出', description: '记录消费金额、分类和付款账户。' },
  { type: 'income', icon: 'add-circle', title: '记收入', description: '记录实际到账的工资和其他收入。' },
  { type: 'transfer', icon: 'transfer', title: '账户转账', description: '记录账户之间的资金流动。' },
  { type: 'refund', icon: 'refresh', title: '记录退款', description: '记录退回款项，冲减原分类支出。' },
]
</script>
<style scoped>
.action-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 18px; }
.action-card { display: block; padding: 28px; color: var(--ink); text-decoration: none; transition: border-color .15s; }
.action-card:hover { border-color: var(--brand); }
.action-icon { display: grid; place-items: center; width: 44px; height: 44px; background: var(--soft); color: var(--brand); border-radius: 10px; margin-bottom: 24px; }
.action-card h2 { font-size: 17px; font-weight: 600; }
.action-card p { color: var(--muted); font-size: 12px; margin: 10px 0 24px; line-height: 1.8; }
.action-link { color: var(--brand); font-size: 12px; }
@media (max-width: 560px) { .action-grid { grid-template-columns: minmax(0, 1fr); gap: 12px; } .action-card { padding: 22px; } }
</style>
