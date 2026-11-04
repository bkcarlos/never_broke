<template>
  <div class="work-page forecast-page">
    <header class="page-heading">
      <div><p class="eyebrow">CASH FLOW PLANNING</p><h1>每月收支预测</h1><p class="subtitle">提前看清每月结余，以及工资到账前的资金缺口。</p></div>
      <div class="actions"><RouterLink to="/me/salary" class="button">工资预估</RouterLink><RouterLink to="/settings/recurring-income" class="button">其他收入</RouterLink><a href="#fixed-expenses" class="button primary">管理固定支出</a></div>
    </header>
    <section class="panel">
      <div class="forecast-filters">
        <label class="field">预测范围<select v-model.number="monthCount" class="control" @change="loadForecast"><option :value="3">本月及未来 2 个月</option><option :value="6">本月及未来 5 个月</option><option :value="12">本月及未来 11 个月</option></select></label>
        <label class="field">币种<select v-model="currency" class="control" @change="changeCurrency"><option value="CNY">CNY · 人民币</option><option value="USD">USD · 美元</option></select></label>
        <label class="field">资金范围<select v-model="accountId" class="control" @change="loadForecast"><option value="">全部现金账户</option><option v-for="a in cashAccounts" :key="a.id" :value="a.id">{{formatAccountDisplayName(a)}}</option></select></label>
        <label class="field">安全垫（{{currency}}）<input v-model.number="safeAmount" type="number" min="0" step="100" class="control" @change="loadForecast" /></label>
        <button class="button" :disabled="loading" @click="loadForecast">{{loading?'计算中…':'刷新预测'}}</button>
      </div>
    </section>
    <div v-if="accountError" class="error-box" role="alert">{{accountError}} <button class="button" @click="loadAccounts">重试账户列表</button></div>
    <div v-if="error" class="error-box" role="alert">{{error}} <button class="button" @click="loadForecast">重试</button></div>
    <p v-if="loading" class="notice" role="status">正在更新收支与资金缺口预测…</p>
    <template v-if="forecast && !error">
      <div class="forecast-result" :aria-busy="loading">
        <section class="risk-banner" :class="forecast.status" aria-live="polite">
          <div><span class="risk-label">{{statusLabel(forecast.status)}}</span><h2>{{riskTitle}}</h2><p>{{riskDetail}}</p></div>
          <RouterLink v-if="!forecast.balanceKnown" to="/accounts" class="button">核验账户余额</RouterLink>
        </section>
        <div class="metrics">
          <div class="metric"><p class="metric-label">预测期收入 · 含本月已入账</p><p class="metric-value">{{money(totalIncome)}}</p><p class="metric-foot">税后工资 + 周期收入</p></div>
          <div class="metric"><p class="metric-label">预测期现金支出 · 含本月已发生</p><p class="metric-value">{{money(totalOutflow)}}</p><p class="metric-foot">固定支出 + 日常消费预留 + 还款</p></div>
          <div class="metric accent"><p class="metric-label">未来最低余额</p><p class="metric-value" :class="minimum!==null&&minimum<0?'expense':''">{{money(minimum)}}</p><p class="metric-foot">{{minDate || '余额待核验'}}{{!forecast.spendingKnown?' · 仅含已知支出':''}}</p></div>
          <div class="metric"><p class="metric-label">需提前补足的最大金额</p><p class="metric-value expense">{{money(gap)}}</p><p class="metric-foot">{{!forecast.spendingKnown ? '仅含已知支出，实际缺口可能更大' : firstGap ? '首次缺口：'+firstGap : '以预测期最低余额计算'}}</p></div>
        </div>
        <section class="panel monthly-panel">
          <div class="panel-head"><div><h2>每月收入与支出</h2><p class="subtitle">本月 = 已发生现金收支 + 剩余日期预测。点选月份查看构成与日期。</p></div><span class="tag">{{forecast.currency}}</span></div>
          <div class="month-grid">
            <button v-for="m in forecast.months" :key="m.month" class="month-card" :class="{selected:selectedMonth===m.month}" :aria-pressed="selectedMonth===m.month" @click="selectedMonth=m.month">
              <div class="month-heading"><strong>{{m.month}}</strong><span class="tag" :class="m.status==='danger'||m.net<0?'amber':m.status==='safe'?'green':''">{{m.net<0?'当月收不抵支':m.firstGapDate?'月内有资金缺口':statusLabel(m.status)}}</span></div>
              <dl><div><dt>预计收入</dt><dd>{{money(m.income)}}</dd></div><div><dt>预计现金支出</dt><dd>{{money(m.outflow)}}</dd></div><div class="month-net"><dt>当月结余</dt><dd :class="m.net<0?'expense':'refund'">{{money(m.net)}}</dd></div><div><dt>月末余额{{m.partial?'（区间末）':''}}</dt><dd>{{money(m.closingBalance)}}</dd></div></dl>
              <p v-if="m.firstGapDate" class="gap-note">{{m.firstGapDate}} 起有缺口 · 最多 {{money(m.gapAmount)}}</p>
              <p v-else class="subtitle">{{m.minBalanceDate?m.minBalanceDate+' 最低 '+money(m.minBalance):'余额待核验，暂不判断缺口'}}</p>
              <p v-if="m.spendingSource==='unknown'" class="subtitle">日常消费待补充</p>
            </button>
          </div>
        </section>
        <section v-if="month" class="panel month-detail">
          <div class="panel-head"><div><h2>{{month.month}} · 收支构成与资金压力</h2><p class="subtitle">未来日期：{{month.from}} 至 {{month.to}}</p></div><RouterLink to="/plans" class="button">核对还款安排</RouterLink></div>
          <div class="panel-body">
            <div class="breakdown"><div><span>已发生收入 / 支出</span><strong>{{money(month.actualIncome)}} / {{money(month.actualOutflow)}}</strong></div><div><span>未来预计收入</span><strong>{{money(month.forecastIncome)}}</strong></div><div><span>固定支出</span><strong>{{money(month.fixedExpense)}}</strong></div><div><span>日常消费预留</span><strong>{{month.spendingSource==='unknown'?'待估算':money(month.dailyExpense)}}</strong></div><div><span>账单与分期还款</span><strong>{{money(month.repayment)}}</strong></div></div>
            <p v-if="month.net>=0&&month.firstGapDate" class="notice">整月收支有结余，但 {{month.firstGapDate}} 起会出现短期缺口，需要在该日期前预留 {{money(month.gapAmount)}}。</p>
            <p class="scope-note">日常消费依据：{{sourceLabel(month.spendingSource)}}<template v-if="forecast.historyDays">（可用样本 {{forecast.historyDays}} 天）</template>。还款计划不再作为历史消费重复计入。</p>
            <div v-if="forecast.balanceKnown && user.showAmount" class="balance-chart">
              <svg viewBox="0 0 900 225" role="img" :aria-label="month.month+' 预计日终余额曲线'">
                <line x1="90" :y1="chartY(0)" x2="880" :y2="chartY(0)" stroke="#e5e7ef" />
                <line x1="90" :y1="chartY(forecast.safeAmount)" x2="880" :y2="chartY(forecast.safeAmount)" stroke="#d5a953" stroke-dasharray="5 5" />
                <text x="85" :y="chartY(forecast.safeAmount)-7" text-anchor="end" class="chart-label">安全垫</text>
                <text x="85" :y="chartY(chartMax)+5" text-anchor="end" class="chart-label">{{money(chartMax)}}</text>
                <text x="85" :y="chartY(chartMin)+5" text-anchor="end" class="chart-label">{{money(chartMin)}}</text>
                <polyline :points="chartLine" fill="none" stroke="#6363db" stroke-width="2.5" />
                <text x="90" y="220" class="chart-label">{{month.from}}</text><text x="880" y="220" text-anchor="end" class="chart-label">{{month.to}}</text>
              </svg>
            </div>
            <div v-else class="notice">{{!user.showAmount?'金额已隐藏。':'余额未核验或存在未知事项，暂不绘制余额曲线。'}}</div>
          </div>
          <div class="table-scroll"><table class="ledger-table"><caption class="sr-only">{{month.month}} 每日收支及缺口</caption><thead><tr><th>日期</th><th>预计事项</th><th class="money">流入</th><th class="money">流出</th><th class="money">日终余额</th><th class="money">资金缺口</th></tr></thead><tbody><tr v-for="p in monthPoints" :key="p.date" :class="{'pressure-row':p.closingBalance!==null&&p.closingBalance<forecast.safeAmount}"><td>{{p.date.slice(5)}}</td><td class="event-cell"><span v-for="(e,i) in p.events" :key="i" class="event-chip">{{e.name || kindLabel(e.kind)}} {{money(e.amount)}}</span><span v-if="!p.events.length" class="subtle">无已知计划</span></td><td class="money">{{money(p.inflow)}}</td><td class="money">{{money(p.outflow)}}</td><td class="money">{{money(p.closingBalance)}}</td><td class="money"><span :class="p.closingBalance!==null&&p.closingBalance<0?'expense':''">{{money(p.closingBalance===null?null:Math.max(0,-p.closingBalance))}}</span></td></tr></tbody></table></div>
        </section>
        <details class="panel assumptions"><summary>预测依据与待补充信息 <span class="subtle">{{forecast.warnings.length}} 项</span></summary><ul><li v-for="w in forecast.warnings" :key="w">{{w}}</li></ul><p class="scope-note">预测不会自动生成交易。当前余额已包含已记账流水；本月实际收支仅用于月度汇总，不再次从余额扣减。</p></details>
      </div>
    </template>
    <FixedExpenseManager :currency="currency" @saved="loadForecast" />
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue'
import {fetchCashflowForecast} from '@/api/cashflow'
import {getAccounts} from '@/api/accounts'
import {amount, localDate} from '@/api/workspace'
import {formatAccountDisplayName} from '@/utils/accountDisplay'
import {useUserSettingsStore} from '@/stores/userSettings'
import FixedExpenseManager from '@/components/FixedExpenseManager.vue'
import type {Account, CashflowForecast, ForecastStatus} from '@/types'

const user=useUserSettingsStore()
const monthCount=ref(3),currency=ref('CNY'),accountId=ref(''),safeAmount=ref(2000)
const forecast=ref<CashflowForecast|null>(null),accounts=ref<Account[]>([]),loading=ref(false),error=ref(''),accountError=ref(''),selectedMonth=ref('')
const cashAccounts=computed(()=>accounts.value.filter(a=>!a.archived&&a.type!=='credit'&&a.currency===currency.value))
const money=(value:number|null)=>amount(value===null?null:Math.round(value*100),forecast.value?.currency||currency.value,user.showAmount)
const statusLabel=(s:ForecastStatus)=>({safe:'高于安全垫',warning:'低于安全垫',danger:'存在资金缺口',unknown:'信息待补充'}[s])
const sourceLabel=(s:string)=>({history:'近三个月历史 + 本月消费',current:'本月日均消费',budget:'该月预算扣除已消费及固定支出后均摊',unknown:'缺少历史与预算，暂未计入日常消费'}[s]||s)
const kindLabel=(s:string)=>({income:'预计收入',fixed:'固定支出',daily:'日常消费',repayment:'还款'}[s]||s)
const month=computed(()=>forecast.value?.months.find(m=>m.month===selectedMonth.value))
const monthPoints=computed(()=>forecast.value?.points.filter(p=>p.date.startsWith(selectedMonth.value))||[])
const totalIncome=computed(()=>forecast.value?.months.reduce((n,m)=>n+m.income,0)||0)
const totalOutflow=computed(()=>forecast.value?.months.reduce((n,m)=>n+m.outflow,0)||0)
const minimum=computed(()=>{const values=forecast.value?.points.flatMap(p=>p.closingBalance===null?[]:[p.closingBalance])||[];return values.length?Math.min(...values):null})
const minDate=computed(()=>minimum.value===null?'':forecast.value?.points.find(p=>p.closingBalance===minimum.value)?.date||'')
const gap=computed(()=>minimum.value===null?null:Math.max(0,-minimum.value))
const firstGap=computed(()=>forecast.value?.dangerDates[0]||'')
const riskTitle=computed(()=>firstGap.value?`${firstGap.value} 起，预计出现资金缺口`:forecast.value?.status==='warning'?'余额可能低于你设置的安全垫':forecast.value?.status==='safe'?'按当前计划，预测期内余额高于安全垫':'收支可以预估，资金风险还需补充信息')
const riskDetail=computed(()=>firstGap.value?`预测期内最多需提前补足 ${money(gap.value)}；可在下方查看对应日期的支出。`:!forecast.value?.balanceKnown?'核验现金账户余额，并补齐未知还款金额或扣款账户后，即可判断缺口。':!forecast.value.spendingKnown?'部分月份缺少日常消费依据，当前余额仅按已知计划推算。':'留意收入到账日期和还款日期；新增消费或修改计划后可以刷新预测。')
const chartValues=computed(()=>monthPoints.value.map(p=>p.closingBalance??0))
const chartMin=computed(()=>Math.min(0,...chartValues.value))
const chartMax=computed(()=>Math.max(1,forecast.value?.safeAmount||0,...chartValues.value))
const chartY=(v:number)=>190-(v-chartMin.value)/(chartMax.value-chartMin.value)*165
const chartLine=computed(()=>chartValues.value.map((v,i)=>`${90+i/Math.max(1,chartValues.value.length-1)*790},${chartY(v)}`).join(' '))
let requestId=0
async function loadAccounts(){accountError.value='';try{accounts.value=await getAccounts()}catch(e){accountError.value=e instanceof Error?e.message:'账户列表加载失败'}}
async function loadForecast(){
  const id=++requestId
  if(!Number.isFinite(safeAmount.value)||safeAmount.value<0){error.value='安全垫应为不小于零的金额';loading.value=false;return}
  loading.value=true;error.value=''
  const now=new Date(),end=new Date(now.getFullYear(),now.getMonth()+monthCount.value,1)
  const startDate=localDate(now)
  const days=Math.round((Date.UTC(end.getFullYear(),end.getMonth(),end.getDate())-Date.UTC(now.getFullYear(),now.getMonth(),now.getDate()))/86400000)
  try{
    const result=await fetchCashflowForecast({startDate,days,currency:currency.value,base:accountId.value?'account':'total_assets',accountId:accountId.value||undefined,safeAmount:safeAmount.value})
    if(id!==requestId)return
    forecast.value=result
    if(!result.months.some(m=>m.month===selectedMonth.value))selectedMonth.value=result.months[0]?.month||''
  }catch(e){if(id===requestId)error.value=e instanceof Error?e.message:'预测加载失败，请重试'}finally{if(id===requestId)loading.value=false}
}
function changeCurrency(){accountId.value='';forecast.value=null;loadForecast()}
onMounted(()=>{loadAccounts();loadForecast()})
</script>

<style scoped>
.forecast-page{padding-bottom:24px}.forecast-filters{padding:18px 20px;display:flex;gap:16px;align-items:flex-end;flex-wrap:wrap}.forecast-filters .field{flex:1;min-width:140px}.forecast-filters input{width:140px}.risk-banner{margin-top:20px;border:1px solid #dcece5;border-radius:9px;background:#f0f8f4;padding:22px 24px;display:flex;align-items:center;justify-content:space-between;gap:18px}.risk-banner h2{font-size:18px;font-weight:600;margin:8px 0}.risk-banner p{font-size:12px;line-height:1.8;color:#6d7486}.risk-label{font-size:11px;font-weight:600;color:#34836d}.risk-banner.danger{background:#fff3f3;border-color:#efd5d9}.risk-banner.danger .risk-label{color:#be455a}.risk-banner.warning{background:#fff9ef;border-color:#ecdfc8}.risk-banner.unknown{background:#f0f1fa;border-color:#e0e2f2}.month-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;padding:20px}.month-card{text-align:left;border:1px solid var(--line);border-radius:8px;padding:18px;background:white;min-width:0}.month-card:hover{background:#fafaff}.month-card.selected{border-color:#8585dd;box-shadow:0 0 0 1px #8585dd;background:#fafaff}.month-heading{display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:18px;flex-wrap:wrap}.month-heading strong{font-size:17px;font-weight:600}.month-card dl>div{display:flex;justify-content:space-between;gap:12px;font-size:12px;margin:13px 0}.month-card dt{color:var(--muted)}.month-card dd{font-variant-numeric:tabular-nums;font-weight:550}.month-net{padding-top:13px;border-top:1px solid var(--line)}.gap-note{font-size:11px;color:var(--danger);margin-top:16px;line-height:1.7}.month-detail,.assumptions,.fixed-expenses{margin-top:20px}.breakdown{display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:16px}.breakdown div{display:flex;flex-direction:column;gap:10px;font-size:11px;color:var(--muted)}.breakdown strong{font-size:14px;font-weight:550;color:var(--ink)}.balance-chart{margin:20px 0 0}.balance-chart svg{width:100%;height:auto;min-height:140px}.chart-label{font-size:11px;fill:#777e92}.event-cell{white-space:normal;min-width:220px;max-width:450px}.event-chip{display:inline-block;font-size:10px;background:#f4f4fa;border-radius:4px;padding:4px 7px;margin:3px 6px 3px 0}.pressure-row{background:#fff8f5}.assumptions{padding:16px 20px;margin-bottom:20px;font-size:12px}.assumptions summary{cursor:pointer}.assumptions ul{list-style:disc;margin:14px 0 0 18px;color:var(--muted);line-height:1.9}.forecast-result[aria-busy=true]{opacity:.65;pointer-events:none}@media(max-width:1100px){.month-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.breakdown{grid-template-columns:repeat(3,minmax(0,1fr))}}@media(max-width:767px){.page-heading{flex-direction:column}.forecast-filters{gap:12px;padding:15px}.forecast-filters .field{min-width:calc(50% - 12px)}.forecast-filters input{width:100%}.risk-banner{padding:18px;align-items:flex-start;flex-direction:column}.risk-banner h2{font-size:16px}.month-grid{grid-template-columns:1fr;padding:15px}.breakdown{grid-template-columns:repeat(2,minmax(0,1fr))}.month-card{padding:16px}.metrics .metric-value{font-size:20px}}
</style>
