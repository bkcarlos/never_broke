<template>
  <div class="work-page page-frame" :class="{ 'page-frame-compact': compact }">
    <RouterLink v-if="backTo" :to="backTo" class="frame-back">
      <Icon name="arrow-left" size="14px" />{{ backLabel }}
    </RouterLink>
    <header class="page-heading">
      <div>
        <p class="eyebrow">{{ eyebrow }}</p>
        <h1>{{ title }}</h1>
        <p class="subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="$slots.actions" class="actions"><slot name="actions" /></div>
    </header>
    <slot />
  </div>
</template>

<script setup lang="ts">
import Icon from './Icon.vue'

withDefaults(defineProps<{
  title: string
  subtitle: string
  eyebrow: string
  backTo?: string
  backLabel?: string
  compact?: boolean
}>(), { backLabel: '返回设置', compact: false })
</script>

<style scoped>
.page-frame { padding-bottom: 24px; }
.page-frame-compact { max-width: 1080px; }
.frame-back { display: inline-flex; align-items: center; gap: 6px; color: var(--muted); font-size: 12px; text-decoration: none; margin-bottom: 18px; min-height: 28px; }
.frame-back:hover { color: var(--brand); }
.page-heading > div { min-width: 0; }
.page-frame :deep(input:not([type='checkbox']):not([type='radio']):not([type='range'])),
.page-frame :deep(select), .page-frame :deep(textarea) { min-width: 0; max-width: 100%; }
.page-frame :deep(.form-surface) { background: #fff; border: 1px solid var(--line); border-radius: 9px; padding: 24px; }
.page-frame :deep(.form-actions) { display: flex; align-items: center; justify-content: flex-end; flex-wrap: wrap; gap: 10px; padding-top: 20px; margin-top: 24px; border-top: 1px solid var(--line); }
.page-frame :deep(.form-actions .button) { min-height: 40px; padding-inline: 22px; }
.page-frame :deep(.form-columns) { display: grid; grid-template-columns: minmax(0, 1fr) 280px; gap: 20px; align-items: start; }
.page-frame :deep(.form-aside) { padding: 24px; border: 1px solid #e1e1f4; background: #f0f0fa; border-radius: 9px; }
.page-frame :deep(.form-aside h2) { font-size: 14px; font-weight: 600; }
.page-frame :deep(.form-aside p) { font-size: 12px; line-height: 1.8; color: var(--muted); margin-top: 10px; }
.page-frame :deep(.section-heading) { padding: 16px 20px; border-bottom: 1px solid var(--line); background: #fafbfe; color: var(--ink); }
.page-frame :deep(.section-heading h2) { color: var(--ink); font-size: 14px; font-weight: 600; }
@media (max-width: 900px) {
  .page-frame :deep(.form-columns) { grid-template-columns: minmax(0, 1fr); }
}
@media (max-width: 767px) {
  .page-frame :deep(.form-surface), .page-frame :deep(.form-aside) { padding: 18px; }
  .page-heading { flex-wrap: wrap; gap: 14px; }
  .page-heading h1 { display: none; }
  .page-heading .subtitle { margin-top: 0; }
  .frame-back { margin-bottom: 12px; }
}
</style>
