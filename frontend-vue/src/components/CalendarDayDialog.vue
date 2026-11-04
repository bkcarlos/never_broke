<template>
  <Teleport to="body">
    <div class="day-dialog-backdrop" @click.self="$emit('close')" @keydown.esc.stop.prevent="$emit('close')" @keydown.tab="trapFocus">
      <section ref="dialog" class="day-dialog" role="dialog" aria-modal="true" aria-labelledby="calendar-day-title" tabindex="-1">
        <header class="panel-head">
          <div><h2 id="calendar-day-title">{{date}} · 当日明细</h2><p class="subtitle">{{currency}} · {{repayments?'消费、退款与还款安排':'消费与退款'}}<template v-if="current"> · 共 {{current.total}} 笔<template v-if="salaries.length"> · {{salaries.length}} 次预估发薪</template></template></p></div>
          <button class="icon-button" aria-label="关闭当日明细" @click="$emit('close')">×</button>
        </header>
        <div class="day-dialog-body" :aria-busy="loading">
          <div v-if="error" class="error-box" role="alert">{{error}} <button class="button" @click="$emit('retry')">重试</button></div>
          <div v-if="loading" class="dialog-loading" role="status">正在读取明细…</div>
          <template v-if="current && !loading">
            <div v-if="salaries.length" class="day-records">
              <button v-for="salary in salaries" :key="salary.key" class="day-record salary" aria-haspopup="dialog" @click="$emit('salary',salary)"><span class="record-kind">预估到手</span><span class="record-title">{{salary.title}} · 查看扣除明细</span><strong>+{{amount(Math.round(salary.postTax*100),currency,show)}}</strong></button>
            </div>
            <div v-if="current.items.length" class="day-records">
              <button v-for="item in current.items" :key="item.key" class="day-record" :class="item.kind==='schedule'?item.edit.status:item.kind" @click="$emit('open',item.key)">
                <span class="record-kind">{{calendarKind(item.kind,item.edit.status)}}</span>
                <span class="record-title" :title="calendarTitle(item.merchant||item.description,item.kind)">{{calendarTitle(item.merchant||item.description||'未填写摘要',item.kind)}}</span>
                <strong>{{item.kind==='refund'?'+':''}}{{amount(item.cents,currency,show)}}</strong>
              </button>
            </div>
            <div v-else-if="!salaries.length" class="empty"><strong>当天暂无{{repayments?'消费或待还款记录':'消费或退款记录'}}</strong>关闭弹窗即可继续查看其他日期。</div>
          </template>
        </div>
        <footer v-if="current && current.total>50" class="pagination">
          <span>{{current.offset+1}}–{{current.offset+current.items.length}} / {{current.total}}</span>
          <div class="actions"><button class="button" :disabled="!current.offset||loading" @click="$emit('page',current.offset-50)">上一页</button><button class="button" :disabled="!current.hasMore||loading" @click="$emit('page',current.offset+50)">下一页</button></div>
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import {computed,nextTick,onBeforeUnmount,onMounted,ref} from 'vue'
import type {CalendarResult,CalendarSalary} from '@/api/workspace'
import {amount} from '@/api/workspace'
import {calendarTitle,calendarKind} from '@/utils/calendarDisplay'
const props=defineProps<{date:string;currency:string;repayments:boolean;data:CalendarResult|null;loading:boolean;error:string;show:boolean}>()
defineEmits<{(e:'close'):void;(e:'open',key:string):void;(e:'page',offset:number):void;(e:'retry'):void;(e:'salary',salary:CalendarSalary):void}>()
const current=computed(()=>props.data?.day===props.date?props.data:null)
const salaries=computed(()=>current.value?.salaries?.filter(s=>s.date===props.date)||[])
const dialog=ref<HTMLElement|null>(null), previousFocus=document.activeElement as HTMLElement|null
onMounted(async()=>{await nextTick();dialog.value?.focus()})
onBeforeUnmount(()=>{if(previousFocus?.isConnected)previousFocus.focus({preventScroll:true})})
function trapFocus(event:KeyboardEvent){
  const elements=dialog.value?.querySelectorAll<HTMLElement>('button:not(:disabled),a[href],input:not(:disabled),select:not(:disabled),textarea:not(:disabled),[tabindex="0"]')
  const visible=Array.from(elements||[]).filter(el=>el.getClientRects().length)
  const first=visible[0],last=visible[visible.length-1]
  if(!first){event.preventDefault();dialog.value?.focus();return}
  if(event.shiftKey&&(document.activeElement===first||document.activeElement===dialog.value)){event.preventDefault();last.focus()}
  else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first.focus()}
}
</script>

<style scoped>
.day-record.salary{background:#edf8f3}.day-record.salary .record-kind,.day-record.salary strong{color:#21745a}
.day-dialog-backdrop{position:fixed;inset:0;z-index:1100;background:#1d24444d;display:flex;align-items:center;justify-content:center;padding:24px}.day-dialog{background:white;width:640px;max-width:100%;max-height:85dvh;border-radius:12px;box-shadow:0 20px 80px #1822482b;display:flex;flex-direction:column;overflow:hidden}.day-dialog>.panel-head,.day-dialog>.pagination{flex-shrink:0}.day-dialog-body{overflow:auto;overscroll-behavior:contain;min-height:100px}.day-dialog-body .error-box{margin:16px}.dialog-loading{padding:24px;text-align:center;font-size:12px;color:var(--muted)}
.day-record{display:grid;grid-template-columns:64px minmax(0,1fr) auto;align-items:center;gap:12px;width:100%;min-height:48px;padding:12px 20px;background:white;text-align:left;border-bottom:1px solid var(--line);font-size:12px;color:var(--ink)}.day-record:last-child{border-bottom:0}.day-record:hover{background:#f8f8fd}.record-kind{color:#5557ca;font-size:11px}.record-title{white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.day-record strong{font-variant-numeric:tabular-nums;font-weight:550;white-space:nowrap}.day-record.refund .record-kind,.day-record.refund strong{color:var(--green)}.day-record.billed .record-kind{color:#3978aa}.day-record.estimated .record-kind{color:#a57521}
@media(max-width:767px){.day-record{grid-template-columns:48px minmax(0,1fr) auto;gap:8px;padding:12px;font-size:11px}.record-kind{font-size:10px}}
@media(max-width:767px){.day-dialog-backdrop{padding:16px}.day-dialog{max-height:85dvh;border-radius:10px}.day-dialog .panel-head{padding:15px}.day-dialog .panel-head h2{font-size:15px}}
</style>
