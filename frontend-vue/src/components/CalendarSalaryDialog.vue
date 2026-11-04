<template>
  <Teleport to="body">
    <div class="salary-backdrop" @click.self="$emit('close')" @keydown.esc.stop.prevent="$emit('close')" @keydown.tab="trapFocus">
      <section ref="dialog" class="salary-dialog" role="dialog" aria-modal="true" aria-labelledby="salary-preview-title" tabindex="-1">
        <header class="panel-head">
          <div><h2 id="salary-preview-title">{{salary.title}} · 薪资预估</h2><p class="subtitle">{{salary.date}} · {{accountLabel}} · {{salary.currency}}</p></div>
          <button class="icon-button" aria-label="关闭薪资预估" @click="$emit('close')">×</button>
        </header>
        <div class="salary-body">
          <div class="salary-net"><span>预估到手</span><strong>{{money(salary.postTax)}}</strong><small>税前薪资 − 五险一金个人缴纳 − 个人所得税</small></div>
          <dl class="salary-breakdown">
            <div class="major"><dt>税前薪资</dt><dd>{{money(salary.preTax)}}</dd></div>
            <div class="major"><dt>五险一金 · 个人合计</dt><dd>−{{money(salary.socialSecurity)}}</dd></div>
            <template v-if="salary.contributions">
              <div v-for="row in contributionRows" :key="row.label" class="detail"><dt>{{row.label}}<small v-if="row.note">{{row.note}}</small></dt><dd>{{money(row.value)}}</dd></div>
            </template>
            <div class="major"><dt>个人所得税<small>按本年度累计预扣估算</small></dt><dd>−{{money(salary.taxAmount)}}</dd></div>
            <div class="major total"><dt>预估到手</dt><dd>{{money(salary.postTax)}}</dd></div>
          </dl>
          <p v-if="!salary.contributions" class="salary-note legacy-note">此方案仅保存了五险一金合计比例，暂无法拆分各项。前往薪资管理重新保存配置后，即可查看险种和公积金明细。</p>
          <details class="tax-details">
            <summary>查看计税依据</summary>
            <dl class="salary-breakdown">
              <div><dt>基本减除费用</dt><dd>{{money(salary.basicDeduction)}}</dd></div>
              <div><dt>专项附加等税前扣除</dt><dd>{{money(salary.extraDeductions)}}</dd></div>
              <div><dt>本次应纳税所得额</dt><dd>{{money(salary.taxableIncome)}}</dd></div>
              <div><dt>本年累计税前薪资</dt><dd>{{money(salary.cumulativePreTax)}}</dd></div>
              <div><dt>本年累计个税</dt><dd>{{money(salary.cumulativeTax)}}</dd></div>
              <template v-if="salary.configuration">
                <div><dt>社保缴费基数</dt><dd>{{money(salary.configuration.socialBase)}}</dd></div>
                <div><dt>公积金缴费基数</dt><dd>{{money(salary.configuration.housingBase)}}</dd></div>
              </template>
            </dl>
            <p class="salary-note">基本减除费用和专项附加等扣除用于计算个税，不会再次从到手薪资中扣除。</p>
          </details>
          <p class="salary-note">按当前有效薪资方案预估，自方案起始日累计，同一年内连续计算，跨年重新累计；不含其他任职收入、奖金或实际已缴税调整。此预估不代表实际到账，以工资单为准。</p>
        </div>
        <footer class="salary-footer"><RouterLink to="/me/salary" class="button" @click="$emit('close')">管理薪资配置</RouterLink><button class="button primary" @click="$emit('close')">知道了</button></footer>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import {computed,nextTick,onBeforeUnmount,onMounted,ref} from 'vue'
import {amount} from '@/api/workspace'
import type {CalendarSalary} from '@/api/workspace'
const props=defineProps<{salary:CalendarSalary;accountLabel:string;show:boolean}>()
defineEmits<{(e:'close'):void}>()
const money=(value:number)=>amount(Math.round(value*100),props.salary.currency,props.show)
const contributionRows=computed(()=>{
  const c=props.salary.contributions, config=props.salary.configuration
  if(!c)return []
  const rate=(value?:number)=>value===undefined?'':`${Number((value*100).toFixed(4))}%`
  return [
    {label:'养老保险',value:c.pension,note:rate(config?.pensionRate)},
    {label:'医疗保险',value:c.medical,note:rate(config?.medicalRate)},
    {label:'失业保险',value:c.unemployment,note:rate(config?.unemploymentRate)},
    {label:'工伤保险',value:c.injury,note:'此方案无个人扣款'},
    {label:'生育保险',value:c.maternity,note:'此方案无个人扣款'},
    {label:'住房公积金',value:c.housing,note:rate(config?.housingRate)},
  ]
})
const dialog=ref<HTMLElement|null>(null), previousFocus=document.activeElement as HTMLElement|null
onMounted(async()=>{await nextTick();dialog.value?.focus()})
onBeforeUnmount(()=>{if(previousFocus?.isConnected)previousFocus.focus({preventScroll:true})})
function trapFocus(event:KeyboardEvent){
  const elements=dialog.value?.querySelectorAll<HTMLElement>('button:not(:disabled),a[href],summary,[tabindex="0"]')
  const visible=Array.from(elements||[]).filter(el=>el.getClientRects().length)
  const first=visible[0],last=visible[visible.length-1]
  if(!first){event.preventDefault();dialog.value?.focus();return}
  if(event.shiftKey&&(document.activeElement===first||document.activeElement===dialog.value)){event.preventDefault();last.focus()}
  else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first.focus()}
}
</script>

<style scoped>
.salary-backdrop{position:fixed;inset:0;z-index:1100;background:#1d24444d;display:flex;align-items:center;justify-content:center;padding:24px}.salary-dialog{width:580px;max-width:100%;max-height:90dvh;background:white;border-radius:14px;box-shadow:0 20px 80px #1822482b;display:flex;flex-direction:column;overflow:hidden}.panel-head,.salary-footer{flex-shrink:0}.salary-body{overflow:auto;padding:20px;overscroll-behavior:contain}.salary-net{padding:20px;background:#edf8f3;border:1px solid #d8eade;border-radius:10px;display:flex;flex-direction:column;gap:8px;color:#21745a}.salary-net>span{font-size:12px}.salary-net strong{font-size:30px;font-variant-numeric:tabular-nums}.salary-net small{font-size:11px;line-height:1.6}.salary-breakdown{margin:14px 0}.salary-breakdown>div{display:flex;justify-content:space-between;gap:16px;padding:10px 0;font-size:12px}.salary-breakdown dt small{display:block;font-size:10px;color:var(--muted);margin-top:4px}.salary-breakdown dd{margin:0;white-space:nowrap;font-variant-numeric:tabular-nums}.major{border-bottom:1px solid var(--line);font-weight:600}.salary-breakdown .detail{padding:7px 0 7px 14px;color:var(--muted)}.detail dt small{display:inline;margin-left:8px}.total{color:#21745a}.salary-note{font-size:11px;color:var(--muted);line-height:1.8;margin:14px 0 0}.legacy-note{padding:10px;background:#fff8ec;border-radius:6px;color:#94691f}.tax-details{border:1px solid var(--line);border-radius:8px;padding:12px;margin-top:16px}.tax-details summary{cursor:pointer;font-size:12px;font-weight:600}.salary-footer{border-top:1px solid var(--line);padding:14px 20px;display:flex;justify-content:space-between;gap:12px}
@media(max-width:767px){.salary-backdrop{padding:12px}.salary-body{padding:14px}.salary-dialog .panel-head{padding:15px}.salary-dialog h2{font-size:16px}.salary-net strong{font-size:26px}}
</style>
