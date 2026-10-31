<template>
  <div class="work-page calendar-page" :style="{'--calendar-weeks':cells.length/7||6}">
    <header class="page-heading">
      <div>
        <p class="eyebrow">SPENDING & REPAYMENTS</p>
        <h1>账本日历</h1>
        <p class="subtitle">每天花了多少，哪天需要还款，一眼看清。</p>
      </div>
      <div class="actions">
        <button class="button repayment-toggle" :class="{enabled:showRepayments}" role="switch" :aria-checked="showRepayments" @click="toggleRepayments"><span class="toggle-track" aria-hidden="true"><span /></span>显示还款项</button>
        <button class="button mobile-filter-toggle" :aria-expanded="filtersOpen" aria-controls="calendar-filters" @click="filtersOpen=!filtersOpen">筛选 · {{currency}}{{account?' · 1':''}}</button>
        <RouterLink to="/detail" class="button calendar-secondary">交易账本</RouterLink>
        <RouterLink to="/plans" class="button calendar-secondary">管理还款</RouterLink>
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
        <div id="calendar-filters" class="actions calendar-filters" :class="{expanded:filtersOpen}">
          <select v-model="account" class="control" aria-label="日历账户" @change="refresh">
            <option value="">全部账户</option>
            <option v-for="a in accounts.filter(a=>a.currency===currency)" :key="a.id" :value="a.id">{{accountName(a)}}</option>
          </select>
          <select v-model="currency" class="control workspace-currency" aria-label="日历币种" @change="account='';refresh()">
            <option value="CNY">CNY · 人民币</option><option value="USD">USD · 美元</option>
          </select>
        </div>
      </div>
    </section>
    <div v-if="error" class="error-box" role="alert">{{error}} <button class="button" @click="load">重试</button></div>
    <div v-if="!data && loading" class="empty" role="status">正在读取日历…</div>

    <template v-if="data">
      <div class="metrics calendar-metrics" :class="{pair:!showRepayments}">
          <div class="metric"><p class="metric-label"><i class="dot" />本月消费</p><div class="metric-value">{{money(data.totals.gross)}}</div><p class="metric-foot">{{data.totals.count}} 笔 · 不含还款</p></div>
          <div class="metric"><p class="metric-label"><i class="dot refund-dot" />本月退款</p><div class="metric-value refund">{{money(data.totals.refund)}}</div><p class="metric-foot">{{data.totals.refundCount}} 笔 · 净支出 {{money(data.totals.net)}}</p></div>
        <template v-if="showRepayments">
          <div class="metric"><p class="metric-label"><i class="dot billed-dot" />本月已出账应还</p><div class="metric-value billed">{{money(data.totals.billed)}}</div><p class="metric-foot">{{data.totals.billedCount}} 项 · 未标记已还的正式计划</p></div>
          <div class="metric"><p class="metric-label"><i class="dot estimated-dot" />本月预计应还</p><div class="metric-value estimated">{{money(data.totals.estimated)}}</div><p class="metric-foot">{{data.totals.estimatedCount}} 项 · 以正式账单为准</p></div>
        </template>
      </div>
      <div class="mobile-calendar-summary" :class="{pair:!showRepayments}">
        <div v-for="metric in mobileMetrics" :key="metric.label" :title="money(metric.cents)" :aria-label="metric.label+' '+money(metric.cents)"><span>{{metric.label}}</span><strong :class="metric.tone">{{compactMoney(metric.cents)}}</strong></div>
      </div>
      <p v-if="data.totals.unknownRepayments" class="notice" role="status">有 {{data.totals.unknownRepayments}} 项还款金额待确认，上方仅合计已知金额。</p>
      <p v-if="showRepayments && !data.totals.billedCount && !data.totals.estimatedCount" class="notice">
        本月暂无待还款。
        <RouterLink v-if="data.nextRepayment" :to="{path:'/timeline',query:{repayments:'1',month:data.nextRepayment.slice(0,7),day:data.nextRepayment,currency,account}}">下一笔 · {{data.nextRepayment.slice(5)}} →</RouterLink>
      </p>

      <section class="panel month-calendar" :aria-busy="loading" aria-label="月历">
        <p class="calendar-scroll-hint">左右滑动查看整周，点击记录查看详情</p>
        <div class="calendar-scroll" tabindex="0" aria-label="月历，点击日期查看明细">
          <div class="calendar-week"><span v-for="day in weekdays" :key="day">{{day}}</span></div>
          <div class="calendar-grid">
            <template v-for="cell in cells" :key="cell.key">
              <section v-if="cell.day" class="calendar-day" :class="{selected:cell.day.date===selectedDate,today:cell.day.date===today}" :aria-label="cell.day.date">
                <button class="mobile-day-button" :aria-label="dayLabel(cell.day)" aria-haspopup="dialog" @click="selectDay(cell.day.date)">
                  <span class="mobile-day-date">{{Number(cell.day.date.slice(-2))}}</span>
                  <span class="mobile-day-amount" :class="{refund:!cell.day.count&&cell.day.refundCount}">{{cell.day.count?compactAmount(cell.day.gross):cell.day.refundCount?'+'+compactAmount(cell.day.refund):cell.day.entryCount?'还款':''}}</span>
                  <span v-if="cell.day.entryCount" class="mobile-day-indicators"><span>{{cell.day.entryCount}}笔</span><i v-if="cell.day.refundCount" class="dot refund-dot"/><i v-if="cell.day.billedCount" class="dot billed-dot"/><i v-if="cell.day.estimatedCount" class="dot estimated-dot"/></span>
                </button>
                <button class="day-number" :aria-label="dayLabel(cell.day)" aria-haspopup="dialog" @click="selectDay(cell.day.date)">{{Number(cell.day.date.slice(-2))}}<span v-if="cell.day.date===today" class="today-label">今天</span><span class="day-count">{{cell.day.entryCount?cell.day.entryCount+' 笔':''}}</span></button>
                <button v-for="event in cell.day.events" :key="event.key" class="calendar-event" :class="eventTone(event)" :aria-label="eventLabel(event)" :title="eventLabel(event)" aria-haspopup="dialog" @click="openRecord(event.key)">
                  <span class="event-kind">{{event.kind==='schedule'&&event.status!=='billed'?'预计':eventKind(event)}}</span><span class="event-title">{{calendarTitle(event.title,event.kind)}}</span><strong class="event-amount">{{event.kind==='refund'?'+':''}}{{amount(event.cents,currency,settings.showAmount)}}</strong>
                </button>
                <button v-if="cell.day.entryCount>cell.day.events.length" class="day-more" aria-haspopup="dialog" @click="selectDay(cell.day.date)">还有 {{cell.day.entryCount-cell.day.events.length}} 笔 · 展开</button>
                <span v-if="!cell.day.entryCount" class="day-empty">—</span>
              </section>
              <div v-else class="calendar-day outside" aria-hidden="true" />
            </template>
          </div>
        </div>
        <div class="calendar-legend"><span class="spending">● 消费</span><span class="refund">● 退款</span><span v-if="showRepayments" class="billed">● 已出账应还</span><span v-if="showRepayments" class="estimated">● 预计应还</span><span>{{currency}} · 点日期看明细</span></div>
      </section>
      <p class="scope-note">还款按到期日展示，不重复计入消费。已还、已取消、已替代的计划不计入应还；无记录不代表没有消费。日期格直接展示商户、事项与金额；点日期或“更多”弹窗查看全部记录。</p>

      <RecordDrawer v-if="recordKey" :record-key="recordKey" :categories="data.categories" :show="settings.showAmount" @close="recordKey=''" @open="recordKey=$event" @saved="load" />
    </template>
    <CalendarDayDialog v-if="dayOpen" :date="selectedDate" :currency="currency" :repayments="showRepayments" :data="data" :loading="loading" :error="error" :show="settings.showAmount" @close="dayOpen=false" @open="openRecord" @page="offset=$event;load()" @retry="load" />
  </div>
</template>

<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import type {LocationQueryRaw} from 'vue-router'
import {getCalendar, amount, accountName, localDate} from '@/api/workspace'
import type {Account, CalendarDay, CalendarEvent, CalendarResult} from '@/api/workspace'
import {useUserSettingsStore} from '@/stores/userSettings'
import CalendarDayDialog from '@/components/CalendarDayDialog.vue'
import RecordDrawer from '@/components/RecordDrawer.vue'
import {calendarTitle,calendarKind} from '@/utils/calendarDisplay'

const route=useRoute(), router=useRouter(), settings=useUserSettingsStore()
const today=localDate(), month=ref(today.slice(0,7)), selectedDate=ref(today)
const currency=ref('CNY'), account=ref(''), showRepayments=ref(false)
const data=ref<CalendarResult|null>(null), accounts=ref<Account[]>([])
const loading=ref(false), error=ref(''), recordKey=ref(''), offset=ref(0), dayOpen=ref(false)
const filtersOpen=ref(false)
const weekdays=['一','二','三','四','五','六','日']
let sequence=0
const money=(cents:number)=>amount(cents,currency.value,settings.showAmount)
const eventKind=(event:CalendarEvent)=>calendarKind(event.kind,event.status)
const eventTone=(event:CalendarEvent)=>event.kind==='schedule'?(event.status==='billed'?'billed':'estimated'):event.kind==='refund'?'refund':'spending'
const eventLabel=(event:CalendarEvent)=>`${calendarTitle(event.title,event.kind)}，${eventKind(event)} ${amount(event.cents,currency.value,settings.showAmount)}`
function compactAmount(cents:number){
  if(!settings.showAmount)return '***'
  const value=cents/100
  return value>=10000?(value/10000).toFixed(1)+'万':value>=1000?(value/1000).toFixed(1)+'千':String(Number(value.toFixed(2)))
}
function compactMoney(cents:number){
  if(!settings.showAmount)return '***'
  return cents>=1000000?(currency.value==='USD'?'$':'¥')+(cents/1000000).toFixed(2)+'万':money(cents)
}
const mobileMetrics=computed(()=>data.value?[
  {label:'本月消费',cents:data.value.totals.gross,tone:'spending'},
  {label:'本月退款',cents:data.value.totals.refund,tone:'refund'},
  ...(showRepayments.value?[
    {label:'已出账应还',cents:data.value.totals.billed,tone:'billed'},
    {label:'预计应还',cents:data.value.totals.estimated,tone:'estimated'}
  ]:[])
]:[])
async function openRecord(key:string){dayOpen.value=false;await nextTick();recordKey.value=key}
const cells=computed(()=>{
  if(!data.value)return []
  const [year,m]=month.value.split('-').map(Number)
  const start=(new Date(year,m-1,1).getDay()+6)%7
  const size=Math.ceil((start+data.value.days.length)/7)*7
  return Array.from({length:size},(_,i)=>({key:i,day:data.value?.days[i-start]}))
})
function dayLabel(day:CalendarDay){
  return `${day.date}，消费 ${money(day.gross)}，退款 ${money(day.refund)}`+(showRepayments.value?`，已出账应还 ${money(day.billed)}，预计应还 ${money(day.estimated)}${day.unknownRepayments?'，有金额待确认事项':''}`:'')
}
async function load(){
  const seq=++sequence
  loading.value=true;error.value=''
  try{
    const result=await getCalendar({month:month.value,day:selectedDate.value,currency:currency.value,account:account.value,view:showRepayments.value?'all':'spending',offset:offset.value})
    if(seq===sequence){data.value=result;accounts.value=result.accounts||[]}
  }catch(e){if(seq===sequence)error.value=e instanceof Error?e.message:String(e)}
  finally{if(seq===sequence)loading.value=false}
}
function toggleRepayments(){
  const query:LocationQueryRaw={...route.query,month:month.value,day:selectedDate.value,currency:currency.value,account:account.value,repayments:showRepayments.value?'0':'1'}
  delete query.view
  void router.replace({path:'/timeline',query})
}
function refresh(){offset.value=0;data.value=null;recordKey.value='';dayOpen.value=false;void load()}
function selectDay(date:string){selectedDate.value=date;offset.value=0;dayOpen.value=true;void load()}
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
watch(()=>route.fullPath,()=>{
  const q=route.query, m=String(q.month||''), d=String(q.day||'')
  showRepayments.value=q.repayments==='1'||(q.repayments===undefined&&(q.view==='all'||q.view==='repayment'))
  month.value=/^(19\d{2}|20\d{2}|21\d{2}|2200)-(0[1-9]|1[0-2])$/.test(m)?m:today.slice(0,7)
  const [y,n]=month.value.split('-').map(Number), last=new Date(y,n,0).getDate()
  selectedDate.value=d.startsWith(month.value+'-')&&/^\d{4}-\d{2}-\d{2}$/.test(d)&&Number(d.slice(-2))>=1&&Number(d.slice(-2))<=last?d:month.value===today.slice(0,7)?today:month.value+'-01'
  currency.value=q.currency==='USD'?'USD':'CNY';account.value=String(q.account||'')
  refresh()
},{immediate:true})
</script>

<style scoped>
.repayment-toggle{font-weight:550}.repayment-toggle.enabled{background:var(--soft);border-color:#bcbcf0;color:var(--brand)}.toggle-track{width:28px;height:16px;background:#c7ccd8;border-radius:9px;padding:2px;display:inline-flex;align-items:center;transition:background .15s}.toggle-track>span{width:12px;height:12px;background:white;border-radius:50%;transition:transform .15s}.enabled .toggle-track{background:var(--brand)}.enabled .toggle-track>span{transform:translateX(12px)}
.calendar-toolbar{display:flex;justify-content:space-between;gap:16px;flex-wrap:wrap;padding:16px}.month-actions{flex-wrap:nowrap}.month-actions>.button{min-width:34px}.month-input{width:150px;font-weight:600}.calendar-filters{min-width:0}.calendar-filters select{max-width:230px}.calendar-metrics.pair{grid-template-columns:repeat(2,minmax(0,1fr))}.calendar-metrics .metric-value{font-size:clamp(18px,2.1vw,27px)}.spending{color:#5557ca}.billed{color:#3978aa}.estimated{color:#a57521}.refund-dot{background:var(--green)}.billed-dot{background:#3978aa}.estimated-dot{background:#b28b47}
.calendar-scroll{overflow-x:auto;overscroll-behavior-x:contain}.calendar-week,.calendar-grid{display:grid;grid-template-columns:repeat(7,minmax(0,1fr));min-width:980px}.calendar-week{background:#fafbfe;border-bottom:1px solid var(--line)}.calendar-week span{padding:13px 0;text-align:center;font-size:11px;color:var(--muted)}.calendar-day{min-width:0;min-height:110px;padding:6px;background:white;border-right:1px solid var(--line);border-bottom:1px solid var(--line);display:flex;flex-direction:column;gap:6px;position:relative}.calendar-day:nth-child(7n){border-right:0}.calendar-day.selected{background:#faf9ff;box-shadow:inset 0 0 0 2px #aaa9e5}.calendar-day.outside{background:#fafbfc}.day-number{font-size:12px;font-weight:600;color:#535d70;display:flex;align-items:center;gap:7px;background:transparent;min-height:28px;text-align:left;padding:0 3px;border-radius:4px}.day-number:hover{background:#ededfa}.today .day-number{color:var(--brand)}.today-label,.day-count{font-size:9px;font-weight:400}.day-count{margin-left:auto;color:var(--muted)}
.calendar-event{border:1px solid #e6e5fa;background:#f3f2fc;border-left:3px solid #7774d6;border-radius:5px;padding:6px;text-align:left;width:100%;min-width:0;display:flex;flex-direction:column;gap:5px;min-height:48px}.calendar-event:hover{filter:brightness(.97)}.calendar-event.refund{background:#edf8f3;border-color:#d8eade;border-left-color:#54a889}.calendar-event.billed{background:#eef5fc;border-color:#dbe8f4;border-left-color:#548cb5}.calendar-event.estimated{background:#fff8ec;border-color:#f0e4c9;border-left-color:#bf9856}.event-title{font-size:11px;font-weight:500;line-height:1.5;display:-webkit-box;-webkit-box-orient:vertical;-webkit-line-clamp:2;overflow:hidden;overflow-wrap:anywhere}.day-more{font-size:10px;color:var(--brand);background:transparent;text-align:left;padding:5px 3px;min-height:28px}.day-more:hover{background:#f0effb}.day-empty{font-size:12px;color:#c1c5d1;padding:4px}
.calendar-legend{display:flex;flex-wrap:wrap;gap:14px;padding:13px 16px;font-size:10px}.calendar-legend>span:last-child{color:var(--muted);margin-left:auto}.calendar-scroll-hint{display:none}.error-box{margin-top:15px}
@media(max-width:1100px){.calendar-scroll-hint{display:block;padding:10px 12px 0;color:var(--muted);font-size:10px}}
/* Keep phone dates as full-width, tappable seven-column tiles. */
.mobile-filter-toggle,.mobile-day-button,.mobile-calendar-summary{display:none}
.calendar-event{flex-direction:row;align-items:center;gap:3px;min-height:30px;padding:5px 4px}
.event-kind{flex-shrink:0;font-size:9px;white-space:nowrap}
.event-title{display:block;flex:1;min-width:0;white-space:nowrap;text-overflow:ellipsis;font-size:10px}
.event-amount{flex-shrink:0;white-space:nowrap;font-size:10px;font-weight:600;font-variant-numeric:tabular-nums}
@media(max-width:767px){
  .calendar-page{height:100%;display:flex;flex-direction:column;gap:8px;min-height:0}
  .calendar-page>*{flex-shrink:0}
  .page-heading{margin:0}.page-heading>div:first-child,.calendar-secondary,.calendar-metrics,.scope-note,.calendar-scroll-hint{display:none}
  .page-heading>.actions{margin:0;justify-content:space-between;flex-wrap:nowrap}
  .mobile-filter-toggle{display:inline-flex}.page-heading .button{font-size:11px;padding:6px 9px}
  .calendar-toolbar{padding:5px 7px;gap:7px}.calendar-filters{display:none;width:100%;flex-wrap:nowrap}.calendar-filters.expanded{display:flex}
  .calendar-filters select{min-width:0;flex:1;width:50%;max-width:none}.calendar-filters .workspace-currency{min-width:125px;flex:0 0 125px}
  .month-actions{width:100%;justify-content:space-between;gap:4px}.month-input{width:145px}
  .mobile-calendar-summary{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));border:1px solid var(--line);border-radius:8px;background:white;padding:9px 0}
  .mobile-calendar-summary.pair{grid-template-columns:repeat(2,minmax(0,1fr))}
  .mobile-calendar-summary>div{min-width:0;padding:0 6px;border-right:1px solid var(--line);display:flex;flex-direction:column;gap:6px;text-align:center}
  .mobile-calendar-summary>div:last-child{border:0}.mobile-calendar-summary span{font-size:9px;color:var(--muted);white-space:nowrap}
  .mobile-calendar-summary strong{font-size:clamp(10px,3vw,13px);font-variant-numeric:tabular-nums;white-space:nowrap}
  .notice{margin:0;padding:5px 7px;font-size:10px;line-height:1.5}
  .month-calendar{flex:1;display:flex;flex-direction:column;min-height:calc(var(--calendar-weeks)*49px + 62px)}
  .calendar-scroll{display:flex;flex-direction:column;flex:1;min-height:0;overflow:visible}
  .calendar-week,.calendar-grid{min-width:0;width:100%}.calendar-week{flex-shrink:0}.calendar-week span{padding:7px 0;font-size:10px}
  .calendar-grid{flex:1;min-height:0;grid-template-rows:repeat(var(--calendar-weeks),minmax(0,1fr))}
  .calendar-day{padding:0;gap:0;min-height:0}.calendar-day>.day-number,.calendar-event,.day-more,.day-empty{display:none}
  .mobile-day-button{display:flex;flex:1;min-width:0;width:100%;min-height:48px;padding:4px 2px;background:transparent;flex-direction:column;align-items:center;justify-content:center;gap:3px}
  .mobile-day-date{font-size:12px;font-weight:600;line-height:16px;color:#535d70}
  .today .mobile-day-date{background:var(--brand);color:white;border-radius:5px;min-width:21px}
  .mobile-day-amount{font-size:10px;line-height:12px;min-height:12px;font-variant-numeric:tabular-nums;color:#5557ca}.mobile-day-amount.refund{color:var(--green)}
  .mobile-day-indicators{display:flex;gap:3px;align-items:center;font-size:8px;line-height:10px;color:var(--muted)}.mobile-day-indicators .dot{width:4px;height:4px}
  .calendar-legend{flex-shrink:0;padding:6px 7px;gap:4px 9px;font-size:9px;line-height:12px}
  .calendar-legend>span:last-child{width:100%;margin-left:0}
}
</style>
