<template>
  <div class="rule-source">
    <p>官方基数 · {{rule.year}} 年度 · 适用 {{rule.effectiveFrom}} 至 {{rule.effectiveThrough}}</p>
    <p>核验日期：{{rule.verifiedAt}} · 查询日期：{{rule.asOf}}</p>
    <div><a v-for="source in rule.sources" :key="source.url" :href="source.url" target="_blank" rel="noopener noreferrer">{{source.name}} ↗</a></div>
    <p>来源用于核对缴费基数上下限；个人比例和实际缴费基数仍应与工资单一致。未收录或超过已核验期间时，不套用默认值。</p>
  </div>
</template>
<script setup lang="ts">
import type {PayrollRegionRule} from '@/types'
defineProps<{rule:PayrollRegionRule}>()
</script>
<style scoped>
.rule-source{font-size:11px;line-height:1.7;color:#64748b;margin-top:12px}.rule-source>div{display:flex;flex-direction:column;gap:4px;margin:6px 0}.rule-source a{color:#2563eb;text-decoration:underline;overflow-wrap:anywhere}
</style>
