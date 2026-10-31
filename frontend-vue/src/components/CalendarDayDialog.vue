<template>
  <Teleport to="body">
    <div class="day-dialog-backdrop" @click.self="$emit('close')" @keydown.esc.stop.prevent="$emit('close')" @keydown.tab="trapFocus">
      <section ref="dialog" class="day-dialog" role="dialog" aria-modal="true" aria-labelledby="calendar-day-title" tabindex="-1">
        <header class="panel-head">
          <div><h2 id="calendar-day-title">{{date}} · 当日明细</h2><p class="subtitle">{{currency}} · {{repayments?'消费、退款与还款安排':'消费与退款'}}<template v-if="current"> · 共 {{current.total}} 笔</template></p></div>
          <button class="icon-button" aria-label="关闭当日明细" @click="$emit('close')">×</button>
        </header>
        <div class="day-dialog-body" :aria-busy="loading">
          <div v-if="error" class="error-box" role="alert">{{error}} <button class="button" @click="$emit('retry')">重试</button></div>
          <div v-if="loading" class="dialog-loading" role="status">正在读取明细…</div>
          <template v-if="current && !loading">
            <LedgerTable v-if="current.items.length" :items="current.items" :currency="currency" :show="show" @open="$emit('open',$event)" />
            <div v-else class="empty"><strong>当天暂无{{repayments?'消费或待还款记录':'消费或退款记录'}}</strong>关闭弹窗即可继续查看其他日期。</div>
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
import type {CalendarResult} from '@/api/workspace'
import LedgerTable from './LedgerTable.vue'
const props=defineProps<{date:string;currency:string;repayments:boolean;data:CalendarResult|null;loading:boolean;error:string;show:boolean}>()
defineEmits<{(e:'close'):void;(e:'open',key:string):void;(e:'page',offset:number):void;(e:'retry'):void}>()
const current=computed(()=>props.data?.day===props.date?props.data:null)
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
.day-dialog-backdrop{position:fixed;inset:0;z-index:1100;background:#1d24444d;display:flex;align-items:center;justify-content:center;padding:24px}.day-dialog{background:white;width:1000px;max-width:100%;max-height:85dvh;border-radius:12px;box-shadow:0 20px 80px #1822482b;display:flex;flex-direction:column;overflow:hidden}.day-dialog>.panel-head,.day-dialog>.pagination{flex-shrink:0}.day-dialog-body{overflow:auto;overscroll-behavior:contain;min-height:100px}.day-dialog-body .error-box{margin:16px}.dialog-loading{padding:24px;text-align:center;font-size:12px;color:var(--muted)}
@media(max-width:767px){.day-dialog-backdrop{padding:16px}.day-dialog{max-height:85dvh;border-radius:10px}.day-dialog .panel-head{padding:15px}.day-dialog .panel-head h2{font-size:15px}}
</style>
