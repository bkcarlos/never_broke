<template>
  <section id="fixed-expenses" class="panel fixed-expenses">
    <div class="panel-head">
      <div><h2>固定支出</h2><p class="subtitle">房租、订阅、保险等，按扣款日期纳入预测。</p></div>
      <button class="button primary" :disabled="busy || loading" @click="create">＋ 新增固定支出</button>
    </div>
    <div class="panel-body">
      <p class="scope-note">已有订阅会一并显示。这里只维护预计支出，实际扣款仍需记账。信用卡账单与分期请在 <RouterLink to="/plans">账单与还款</RouterLink> 管理。</p>
      <div v-if="error" role="alert" class="error-box">{{error}} <button v-if="!draft" class="button" @click="load">重试</button></div>
      <p v-if="message" role="status" class="notice">{{message}}</p>
      <form v-if="draft" class="expense-form" @submit.prevent="save">
        <h3>{{draft.key ? '编辑固定支出' : '新增固定支出'}}</h3>
        <div class="form-grid">
          <label class="field">名称<input v-model.trim="draft.name" class="control" required maxlength="80" placeholder="如：房租" /></label>
          <label class="field">每次金额（{{currency}}）<input v-model.number="inputAmount" type="number" class="control" required min="0.01" max="99999999999.99" step="0.01" /></label>
          <label class="field">扣款周期<select v-model="draft.period" class="control"><option value="monthly">每月</option><option value="quarterly">每季度</option><option value="yearly">每年</option><option value="weekly">每周</option></select></label>
          <label class="field">首次 / 下次扣款日期<input v-model="draft.date" type="date" class="control" required min="1900-01-01" max="2200-12-31" /></label>
          <label class="field">结束日期（选填，含当天）<input v-model="draft.endDate" type="date" class="control" :min="draft.date" max="2200-12-31" /></label>
          <label class="field">扣款账户<select v-model="draft.accountId" class="control"><option value="">暂不指定（仅纳入全部账户预测）</option><option v-for="a in cashAccounts" :key="a.id" :value="a.id">{{formatAccountDisplayName(a)}}</option></select></label>
          <label class="field full-width">账单匹配关键词（选填）<input v-model.trim="draft.match" class="control" maxlength="80" placeholder="如：房租；留空时使用名称" /><span>匹配商户或摘要的历史消费会从日常消费估算中排除，请使用明确的名称。</span></label>
        </div>
        <p class="scope-note">月底扣款遇到短月时取当月最后一天，后续月份仍保留原扣款日。</p>
        <div class="actions"><button type="submit" class="button primary" :disabled="busy">{{busy ? '保存中…' : '保存并更新预测'}}</button><button type="button" class="button" :disabled="busy" @click="draft=null">取消</button></div>
      </form>
      <p v-if="loading" role="status" class="empty">正在读取固定支出…</p>
      <p v-else-if="!visible.length" class="empty">还没有固定支出。添加房租或订阅后，即可看到它们对每月余额的影响。</p>
      <div v-for="s in visible" :key="s.key" class="insight-row">
        <div><strong>{{s.name}}</strong> <span class="tag" :class="s.enabled?'green':''">{{s.enabled?'已启用':'已停用'}}</span> <span v-if="s.kind==='subscription'" class="tag">订阅</span>
          <p class="subtitle">{{money(s.cents)}} / {{periodLabel(s.period)}} · 起始 {{s.date}}<template v-if="s.endDate"> · 截止 {{s.endDate}}</template></p>
          <p class="subtitle">{{accountLabel(s.accountId)}} · 匹配：{{s.match || s.name}}</p>
        </div>
        <div class="actions"><button class="button" :disabled="busy" @click="edit(s)">编辑</button><button class="button ghost" :disabled="busy" @click="toggle(s)">{{s.enabled?'停用':'启用'}}</button></div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {getSettings, saveSetting, amount, localDate} from '@/api/workspace'
import {getAccounts} from '@/api/accounts'
import {formatAccountDisplayName} from '@/utils/accountDisplay'
import {useUserSettingsStore} from '@/stores/userSettings'
import type {Setting} from '@/api/workspace'
import type {Account} from '@/types'

const props=defineProps<{currency:string}>()
const emit=defineEmits<{(e:'saved'):void}>()
const user=useUserSettingsStore()
const settings=ref<Setting[]>([]), accounts=ref<Account[]>([]), draft=ref<Setting|null>(null)
const inputAmount=ref(0), busy=ref(false), loading=ref(false), error=ref(''), message=ref('')
const visible=computed(()=>settings.value.filter(s=>(s.kind==='fixed_expense'||s.kind==='subscription')&&s.currency===props.currency))
const cashAccounts=computed(()=>accounts.value.filter(a=>!a.archived&&a.type!=='credit'&&a.currency===props.currency))
const money=(v:number)=>amount(v,props.currency,user.showAmount)
const periodLabel=(period:string)=>({monthly:'月',quarterly:'季度',yearly:'年',weekly:'周'}[period]||period)
function accountLabel(id?:string){const a=accounts.value.find(a=>a.id===id);return a?formatAccountDisplayName(a):'未指定扣款账户'}
let loadId=0
async function load(){const id=++loadId;loading.value=true;error.value='';try{const [s,a]=await Promise.all([getSettings(),getAccounts()]);if(id===loadId){settings.value=s;accounts.value=a}}catch(e){if(id===loadId)error.value=e instanceof Error?e.message:'读取固定支出失败'}finally{if(id===loadId)loading.value=false}}
function create(){message.value='';draft.value={key:'',kind:'fixed_expense',name:'',currency:props.currency,cents:0,category:'',match:'',accountId:'',date:localDate(),endDate:'',period:'monthly',enabled:true,version:0};inputAmount.value=0}
function edit(s:Setting){message.value='';draft.value={...s,accountId:s.accountId||'',endDate:s.endDate||''};inputAmount.value=s.cents/100}
async function save(){
  if(!draft.value||busy.value)return
  if(!Number.isFinite(inputAmount.value)||inputAmount.value<=0){error.value='请输入大于零的金额';return}
  busy.value=true;error.value=''
  try{await saveSetting({...draft.value,cents:Math.round(inputAmount.value*100)});draft.value=null;message.value='固定支出已保存，预测已刷新。';await load();emit('saved')}
  catch(e){error.value=e instanceof Error?e.message:'保存失败，请重试'}finally{busy.value=false}
}
async function toggle(s:Setting){busy.value=true;error.value='';try{await saveSetting({...s,enabled:!s.enabled});message.value=s.enabled?'已停用，不再计入未来预测。':'已启用，已纳入未来预测。';await load();emit('saved')}catch(e){error.value=e instanceof Error?e.message:'更新失败，请重试'}finally{busy.value=false}}
watch(()=>props.currency,()=>{draft.value=null;message.value='';load()},{immediate:true})
</script>

<style scoped>
.fixed-expenses{scroll-margin-top:80px}.expense-form{background:#f8f9fe;border:1px solid var(--line);border-radius:7px;padding:18px;margin:14px 0 22px}.expense-form h3{font-weight:600;margin-bottom:16px}.full-width{grid-column:1/-1}.insight-row strong{font-weight:550}.field span{line-height:1.7}@media(max-width:600px){.form-grid{grid-template-columns:1fr}.insight-row{flex-direction:column}.panel-head{align-items:flex-start;flex-wrap:wrap}}
</style>
