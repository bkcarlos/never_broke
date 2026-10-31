<template>
  <div class="work-page calendar-page">
    <header class="page-heading">
      <div>
        <p class="eyebrow">SPENDING & REPAYMENTS</p>
        <h1>消费与还款日历</h1>
        <p class="subtitle">每天花了多少，哪天需要还款，一眼看清。</p>
      </div>
      <div class="actions">
        <RouterLink to="/detail" class="button">交易账本</RouterLink>
        <RouterLink to="/plans" class="button">管理还款</RouterLink>
      </div>
    </header>

    <section class="panel calendar-controls" aria-label="日历筛选">
      <div class="calendar-toolbar">
        <div class="actions month-actions">
          <button class="button" aria-label="上个月" :disabled="month==='1900-01'" @click="changeMonth(-1)">‹</button>
          <input v-model="month" type="month" min="1900-01" max="2200-12" class="control month-input" aria-label="选择月份" @change="selectMonth" />
          <button class="button" aria-label="下个月" :disabled="month==='2200-12'" @click="changeMonth(1)">›</button>
          <button class="button ghost" @click="goToday">本月</button>
        </div>
        <div class="actions calendar-filters">
          <select v-model="account" class="control" aria-label="日历账户" @change="refresh">
            <option value="">全部账户</option>
            <option v-for="a in accounts.filter(a=>a.currency===currency)" :key="a.id" :value="a.id">{{accountName(a)}}</option>
          </select>
          <select v-model="currency" class="control workspace-currency" aria-label="日历币种" @change="account='';refresh()">
            <option value="CNY">CNY · 人民币</option><option value="USD">USD · 美元</option>
          </select>
        </div>
      </div>
      <div class="tabs" aria-label="日历内容">
        <button v-for="tab in views" :key="tab.key" :class="{active:view===tab.key}" :aria-pressed="view===tab.key" @click="view=tab.key;refresh()">{{tab.label}}</button>
      </div>
    </section>
    <div v-if="error" class="error-box" role="alert">{{error}} <button class="button" @click="load">重试</button></div>
    <div v-if="!data && loading" class="empty" role="status">正在读取日历…</div>

    <template v-if="data">
      <div class="metrics calendar-metrics" :class="{pair:view!=='all'}">
        <template v-if="view!=='repayment'">
          <div class="metric"><p class="metric-label"><i class="dot" />本月消费</p><div class="metric-value">{{money(data.totals.gross)}}</div><p class="metric-foot">{{data.totals.count}} 笔 · 不含还款</p></div>
          <div class="metric"><p class="metric-label"><i class="dot refund-dot" />本月退款</p><div class="metric-value refund">{{money(data.totals.refund)}}</div><p class="metric-foot">{{data.totals.refundCount}} 笔 · 净支出 {{money(data.totals.net)}}</p></div>
        </template>
        <template v-if="view!=='spending'">
          <div class="metric"><p class="metric-label"><i class="dot billed-dot" />本月已出账应还</p><div class="metric-value billed">{{money(data.totals.billed)}}</div><p class="metric-foot">{{data.totals.billedCount}} 项 · 未标记已还的正式计划</p></div>
          <div class="metric"><p class="metric-label"><i class="dot estimated-dot" />本月预计应还</p><div class="metric-value estimated">{{money(data.totals.estimated)}}</div><p class="metric-foot">{{data.totals.estimatedCount}} 项 · 以正式账单为准</p></div>
        </template>
      </div>
      <p v-if="data.totals.unknownRepayments" class="notice" role="status">有 {{data.totals.unknownRepayments}} 项还款金额待确认，上方仅合计已知金额。</p>

      <section class="panel month-calendar" :aria-busy="loading" aria-label="月历">
        <div class="calendar-week"><span v-for="day in weekdays" :key="day">{{day}}</span></div>
        <div class="calendar-grid">
          <template v-for="cell in cells" :key="cell.key">
            <button v-if="cell.day" class="calendar-day" :class="{selected:cell.day.date===selectedDate,today:cell.day.date===today}" :aria-label="dayLabel(cell.day)" :aria-pressed="cell.day.date===selectedDate" @click="selectDay(cell.day.date)">
              <span class="day-number">{{Number(cell.day.date.slice(-2))}}<span v-if="cell.day.date===today" class="today-label">今天</span></span>
              <span v-if="cell.day.count" class="day-amount spending"><span class="day-label">消费</span><b>{{compact(cell.day.gross)}}</b></span>
              <span v-if="cell.day.refundCount" class="day-amount refund"><span class="day-label">退款</span><b>{{compact(cell.day.refund)}}</b></span>
              <span v-if="cell.day.billedCount" class="day-amount billed"><span class="day-label">应还</span><b>{{compact(cell.day.billed)}}</b></span>
              <span v-if="cell.day.estimatedCount" class="day-amount estimated"><span class="day-label">预计</span><b>{{compact(cell.day.estimated)}}</b></span>
              <span v-if="cell.day.unknownRepayments" class="day-unknown">金额待确认</span>
            </button>
            <div v-else class="calendar-day outside" aria-hidden="true" />
          </template>
        </div>
        <div class="calendar-legend"><span class="spending">● 消费</span><span class="refund">● 退款</span><span class="billed">● 已出账应还</span><span class="estimated">● 预计应还</span><span>{{currency}} · 点日期看明细</span></div>
      </section>
      <p class="scope-note">还款按到期日展示，不重复计入消费。已还、已取消、已替代的计划不计入应还；无记录不代表没有消费。格内金额采用简写，完整金额见下方明细。</p>

      <section class="panel calendar-detail" :aria-busy="loading" aria-label="当日明细">
        <div class="panel-head"><div><h2>{{selectedDate}} · {{views.find(t=>t.key===view)?.label}}明细</h2><p class="subtitle">{{loading?'正在更新…':`共 ${data.total} 条 · 点击记录查看与核对`}}</p></div><span class="tag">{{currency}}</span></div>
        <template v-if="data.day===selectedDate">
          <LedgerTable v-if="data.items.length" :items="data.items" :currency="currency" :show="settings.showAmount" @open="recordKey=$event" />
          <div v-else class="empty"><strong>当天没有{{view==='repayment'?'待还款安排':view==='spending'?'消费或退款':'消费或待还款记录'}}</strong>点击日历中有金额的日期查看明细。</div>
          <footer v-if="data.total>50" class="pagination"><span>{{data.offset+1}}–{{data.offset+data.items.length}} / {{data.total}}</span><div class="actions"><button class="button" :disabled="!data.offset||loading" @click="offset-=50;load()">上一页</button><button class="button" :disabled="!data.hasMore||loading" @click="offset+=50;load()">下一页</button></div></footer>
        </template>
        <div v-else class="empty">正在读取当日明细…</div>
      </section>
      <RecordDrawer v-if="recordKey" :record-key="recordKey" :categories="data.categories" :show="settings.showAmount" @close="recordKey=''" @open="recordKey=$event" @saved="load" />
    </template>
  </div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRoute} from 'vue-router'
import {getCalendar, amount, accountName, localDate} from '@/api/workspace'
import type {Account, CalendarDay, CalendarResult} from '@/api/workspace'
import {useUserSettingsStore} from '@/stores/userSettings'
import LedgerTable from '@/components/LedgerTable.vue'
import RecordDrawer from '@/components/RecordDrawer.vue'

type View = 'all'|'spending'|'repayment'
const route=useRoute(), settings=useUserSettingsStore()
const today=localDate(), month=ref(today.slice(0,7)), selectedDate=ref(today)
const currency=ref('CNY'), account=ref(''), view=ref<View>('all')
const data=ref<CalendarResult|null>(null), accounts=ref<Account[]>([])
const loading=ref(false), error=ref(''), recordKey=ref(''), offset=ref(0)
const views:{key:View;label:string}[]=[{key:'all',label:'消费与还款'},{key:'spending',label:'消费日历'},{key:'repayment',label:'还款日历'}]
const weekdays=['一','二','三','四','五','六','日']
let sequence=0
const money=(cents:number)=>amount(cents,currency.value,settings.showAmount)
function compact(cents:number){
  if(!settings.showAmount)return '•••'
  const n=cents/100
  return n>=10000?`${(n/10000).toFixed(1)}万`:n.toLocaleString('zh-CN',{maximumFractionDigits:2})
}
const cells=computed(()=>{
  if(!data.value)return []
  const [year,m]=month.value.split('-').map(Number)
  const start=(new Date(year,m-1,1).getDay()+6)%7
  const size=Math.ceil((start+data.value.days.length)/7)*7
  return Array.from({length:size},(_,i)=>({key:i,day:data.value?.days[i-start]}))
})
function dayLabel(day:CalendarDay){
  return `${day.date}，消费 ${money(day.gross)}，退款 ${money(day.refund)}，已出账应还 ${money(day.billed)}，预计应还 ${money(day.estimated)}${day.unknownRepayments?'，有金额待确认事项':''}`
}
async function load(){
  const seq=++sequence
  loading.value=true;error.value=''
  try{
    const result=await getCalendar({month:month.value,day:selectedDate.value,currency:currency.value,account:account.value,view:view.value,offset:offset.value})
    if(seq===sequence){data.value=result;accounts.value=result.accounts||[]}
  }catch(e){if(seq===sequence)error.value=e instanceof Error?e.message:String(e)}
  finally{if(seq===sequence)loading.value=false}
}
function refresh(){offset.value=0;data.value=null;recordKey.value='';void load()}
function selectDay(date:string){selectedDate.value=date;offset.value=0;void load()}
function selectMonth(){
  if(!/^(19\d{2}|20\d{2}|21\d{2}|2200)-(0[1-9]|1[0-2])$/.test(month.value))month.value=today.slice(0,7)
  selectedDate.value=month.value===today.slice(0,7)?today:month.value+'-01'
  refresh()
}
function changeMonth(delta:number){
  const [year,m]=month.value.split('-').map(Number)
  month.value=localDate(new Date(year,m-1+delta,1)).slice(0,7)
  selectMonth()
}
function goToday(){month.value=today.slice(0,7);selectMonth()}
watch(()=>route.query.view,v=>{view.value=v==='spending'||v==='repayment'?v:'all';refresh()},{immediate:true})
</script>

<style scoped>
.calendar-toolbar{display:flex;justify-content:space-between;gap:16px;flex-wrap:wrap;padding:16px}.month-actions{flex-wrap:nowrap}.month-actions>.button{min-width:34px}.month-input{width:150px;font-weight:600}.calendar-filters{min-width:0}.calendar-filters select{max-width:230px}.calendar-controls .tabs{border-bottom:0;border-top:1px solid var(--line)}.calendar-metrics.pair{grid-template-columns:repeat(2,minmax(0,1fr))}.calendar-metrics .metric-value{font-size:clamp(18px,2.1vw,27px)}.spending{color:#5557ca}.billed{color:#3978aa}.estimated{color:#a57521}.refund-dot{background:var(--green)}.billed-dot{background:#3978aa}.estimated-dot{background:#b28b47}.calendar-week,.calendar-grid{display:grid;grid-template-columns:repeat(7,minmax(0,1fr))}.calendar-week{background:#fafbfe;border-bottom:1px solid var(--line)}.calendar-week span{padding:13px 0;text-align:center;font-size:11px;color:var(--muted)}.calendar-day{min-width:0;min-height:125px;padding:10px 12px;background:white;text-align:left;border:0;border-right:1px solid var(--line);border-bottom:1px solid var(--line);display:flex;flex-direction:column;gap:5px;position:relative}.calendar-day:nth-child(7n){border-right:0}.calendar-day:hover{background:#fafaff}.calendar-day.selected{background:#f5f4ff;box-shadow:inset 0 0 0 2px #7777d8;z-index:1}.calendar-day.outside{background:#fafbfc}.day-number{font-size:12px;font-weight:600;color:#535d70;display:flex;align-items:center;gap:7px;margin-bottom:5px}.today .day-number{color:var(--brand)}.today-label{font-size:9px;font-weight:400}.day-amount{display:flex;align-items:baseline;gap:5px;font-size:11px;font-variant-numeric:tabular-nums;line-height:1.4}.day-label{font-size:9px;flex-shrink:0}.day-amount b{font-weight:550;overflow-wrap:anywhere;min-width:0}.day-unknown{font-size:9px;color:#a57521}.calendar-legend{display:flex;flex-wrap:wrap;gap:14px;padding:13px 16px;font-size:10px}.calendar-legend>span:last-child{color:var(--muted);margin-left:auto}.calendar-detail{margin-top:15px}.error-box{margin-top:15px}
@media(max-width:767px){.calendar-toolbar{padding:12px;gap:12px}.calendar-filters{width:100%;flex-wrap:nowrap}.calendar-filters select{min-width:0;flex:1;width:50%;max-width:none}.calendar-filters .workspace-currency{min-width:136px;flex:0 0 136px}.calendar-metrics .metric-value{font-size:21px}.month-actions{width:100%;justify-content:space-between;gap:4px}.month-input{width:145px}.calendar-day{padding:7px 4px;min-height:105px;gap:4px}.day-number{font-size:11px;margin-bottom:3px;gap:3px}.today-label{font-size:8px}.day-amount{font-size:10px;display:block;line-height:1.25}.day-label{display:none}.day-amount b{font-weight:550;letter-spacing:-.2px}.day-amount:before{content:'';display:inline-block;width:3px;height:3px;border-radius:50%;background:currentColor;vertical-align:middle;margin-right:2px}.day-unknown{font-size:8px;line-height:1.3}.calendar-week span{padding:10px 0}.calendar-legend{padding:12px 10px;gap:8px 12px}.calendar-legend>span:last-child{margin-left:0;width:100%}.calendar-controls .tabs{padding:0 9px}.calendar-detail .panel-head h2{font-size:13px}}
</style>
