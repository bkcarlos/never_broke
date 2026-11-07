<template><div class="work-page"><header class="page-heading"><div><p class="eyebrow">RULES & PLANNING</p><h1>订阅、预算与分类规则</h1><p class="subtitle">把重复的核对变成可追踪的规则，让未来支出更清楚。</p></div><select v-model="currency" class="control" aria-label="币种" :disabled="busy" @change="changeCurrency"><option>CNY</option><option>USD</option></select></header><div v-if="error" class="error-box" role="alert">{{error}} <button v-if="!draft&&!busy" class="button" @click="load">重试</button></div><div v-if="message" class="notice" role="status">{{message}}</div><section class="panel"><div class="tabs"><button v-for="t in tabs" :key="t.key" :class="{active:kind===t.key}" :disabled="busy" :aria-pressed="kind===t.key" @click="changeKind(t.key)">{{t.name}}</button></div><div class="panel-head"><h2>{{tabs.find(t=>t.key===kind)?.name}}</h2><button class="button primary" :disabled="busy||loading" @click="create">＋ 新增</button></div><div class="panel-body"><form v-if="draft" ref="editor" class="settings-form rule-editor" @submit.prevent="save"><h2 style="margin-bottom:18px">{{draft.key?'编辑':'新增'}}{{tabs.find(t=>t.key===draft?.kind)?.name}}</h2><div class="form-grid"><label class="field">名称<input v-model.trim="draft.name" class="control" required/></label><label v-if="kind!=='subscription'" class="field">分类<select v-model="draft.category" class="control"><option v-for="c in categories" :key="c.id" :value="c.id">{{c.name}}</option></select></label><label v-if="kind==='rule'||kind==='subscription'" class="field">账单匹配关键词<input v-model.trim="draft.match" class="control" :required="kind==='rule'"/></label><label v-if="kind!=='rule'" class="field">金额（{{currency}}）<input v-model.number="inputAmount" type="number" required :min="kind==='subscription'?0.01:0" step="0.01" class="control"/></label><label v-if="kind==='subscription'" class="field">周期<select v-model="draft.period" class="control"><option value="monthly">每月</option><option value="yearly">每年</option><option value="quarterly">每季度</option><option value="weekly">每周</option></select></label><label v-if="kind==='subscription'" class="field">下次续费日期<input v-model="draft.date" type="date" required class="control"/></label></div><div class="actions" style="margin-top:16px"><button type="submit" class="button primary" :disabled="busy||!draft.name">{{busy?'保存中…':'保存'}}</button><button type="button" class="button" :disabled="busy" @click="draft=null">取消</button></div></form><p v-if="kind==='subscription'" class="scope-note"><RouterLink to="/cashflow#fixed-expenses">管理固定支出与查看收支预测 →</RouterLink> 已启用订阅会纳入预测，不会自动写入消费。实际扣费仍以账单为准。下方累计按关键词查找，可点开逐笔核实。</p><p v-if="kind==='rule'" class="scope-note">关键词匹配商户或摘要；只预览未人工确认的消费。每次应用最多 100 条，保留修改历史。</p><p v-if="loading" class="empty" role="status">正在读取设置…</p><div v-else-if="!visible.length&&!error" class="empty">尚未设置，点击“新增”开始。</div><div v-for="s in visible" :key="s.key" class="insight-row setting-row"><div><strong style="font-weight:500">{{s.name}}</strong><span class="tag" :class="s.enabled?'green':''" style="margin-left:10px">{{s.enabled?'已启用':'已停用'}}</span><p class="subtitle" v-if="kind==='subscription'">{{amount(s.cents,currency,show)}} / {{({monthly:'月',quarterly:'季度',weekly:'周',yearly:'年'}[s.period]||s.period)}} · 下次 {{s.date}} <span v-if="s.enabled&&s.date<=soon" class="tag amber">{{s.date<today?'待核对是否续费':'即将到期'}}</span></p><p v-if="kind==='subscription'" class="subtitle">账单关键词：{{s.match||s.name}}<template v-if="subscriptionStats[s.key]"> · {{subscriptionStats[s.key].count}} 笔消费，累计 {{amount(subscriptionStats[s.key].gross,currency,show)}}</template><RouterLink :to="{path:'/detail',query:{currency:s.currency,q:s.match||s.name}}"> 查看匹配账单 →</RouterLink></p><p v-else-if="kind==='budget'" class="subtitle">每月 {{amount(s.cents,currency,show)}} · 本月净支出 {{amount(used(s.category),currency,show)}} · <span :class="used(s.category)>s.cents?'expense':''">剩余 {{amount(s.cents-used(s.category),currency,show)}}</span></p><p v-else-if="kind==='rule'" class="subtitle">包含「{{s.match}}」→ {{categoryName(s.category)}}</p></div><div class="actions"><button v-if="kind==='rule'&&s.enabled" class="button" :disabled="busy" @click="previewRule(s)">预览匹配</button><button class="button" :disabled="busy||loading" @click="edit(s)">编辑</button><button class="button ghost" :disabled="busy" @click="toggle(s)">{{s.enabled?'停用':'启用'}}</button></div></div></div></section>

<section v-if="rule" class="panel" style="margin-top:20px"><div class="panel-head"><h2>规则预览 · {{rule.name}}</h2><button class="button primary" :disabled="!preview.length||busy" @click="applyRule">确认应用 {{preview.length}} 条</button></div><div class="panel-body"><p class="scope-note">只修改以下记录的分类，不改变金额；人工确认过的记录会保留。</p><div v-if="!preview.length" class="empty">暂无可应用记录</div><div class="rule-preview"><div v-for="r in preview" :key="r.key" class="insight-row"><span>{{r.date}} · {{r.description}}</span><span>{{r.categoryName}} → {{categoryName(rule.category)}}</span></div></div></div></section></div></template>
<script setup lang="ts">
import {ref,computed,onMounted,nextTick,onBeforeUnmount} from 'vue'
import {getWorkspace,getSettings,saveSetting,saveEdits,amount,localDate} from '@/api/workspace'
import type {Setting,Entry,Category,Group} from '@/api/workspace'
import {useUserSettingsStore} from '@/stores/userSettings'
const user=useUserSettingsStore(),show=computed(()=>user.showAmount)
const currency=ref('CNY'),kind=ref('subscription'),settings=ref<Setting[]>([]),categories=ref<Category[]>([]),spending=ref<Group[]>([]),draft=ref<Setting|null>(null),inputAmount=ref(0),rule=ref<Setting|null>(null),preview=ref<Entry[]>([]),busy=ref(false),error=ref(''),message=ref('')
const editor=ref<HTMLElement|null>(null),loading=ref(false)
let loadSequence=0
const subscriptionStats=ref<Record<string,{count:number;gross:number}>>({})
const today=localDate(),soon=localDate(new Date(Date.now()+7*86400000)),tabs=[{key:'subscription',name:'订阅与到期'},{key:'budget',name:'分类预算'},{key:'rule',name:'分类规则'}]
const visible=computed(()=>settings.value.filter(s=>s.kind===kind.value&&s.currency===currency.value))
const categoryName=(k:string)=>categories.value.find(c=>c.id===k)?.name||k
const used=(k:string)=>spending.value.find(g=>g.key===k)?.net||0
async function load(){
  const seq=++loadSequence,scope=currency.value
  loading.value=true;error.value='';subscriptionStats.value={}
  try{
    const [s,r]=await Promise.all([getSettings(),getWorkspace({currency:scope,from:today.slice(0,7)+'-01',to:today})])
    if(seq!==loadSequence)return
    settings.value=s;categories.value=r.categories;spending.value=r.analysis.categories
    const stats=await Promise.all(s.filter(x=>x.kind==='subscription'&&x.currency===scope).map(async x=>{
      const result=await getWorkspace({currency:scope,kind:'expense',q:x.match||x.name})
      return [x.key,{count:result.analysis.count,gross:result.analysis.gross}] as const
    }))
    if(seq===loadSequence)subscriptionStats.value=Object.fromEntries(stats)
  }catch(e){if(seq===loadSequence)error.value=e instanceof Error?e.message:String(e)}finally{if(seq===loadSequence)loading.value=false}
}
function clearEditor(){draft.value=null;preview.value=[];rule.value=null;message.value=''}
function changeCurrency(){clearEditor();settings.value=[];spending.value=[];load()}
function changeKind(value:string){kind.value=value;clearEditor()}
async function focusEditor(){await nextTick();editor.value?.scrollIntoView({block:'nearest'});editor.value?.querySelector('input')?.focus({preventScroll:true})}
function edit(s:Setting){draft.value={...s};inputAmount.value=s.cents/100;error.value='';focusEditor()}
function create(){draft.value={key:'',kind:kind.value,name:'',currency:currency.value,cents:0,category:'food',match:'',date:today,period:'monthly',enabled:true,version:0};inputAmount.value=0;error.value='';focusEditor()}
async function save(){if(!draft.value||busy.value)return;if(!draft.value.name.trim()||(kind.value==='rule'&&!draft.value.match.trim())||(kind.value!=='rule'&&(!Number.isFinite(inputAmount.value)||inputAmount.value<0))){error.value='请填写有效名称、关键词与金额';return}busy.value=true;error.value='';try{await saveSetting({...draft.value,cents:Math.round(inputAmount.value*100)});draft.value=null;await load()}catch(e){error.value=String(e)}finally{busy.value=false}}
async function toggle(s:Setting){busy.value=true;try{await saveSetting({...s,enabled:!s.enabled});await load()}catch(e){error.value=String(e)}finally{busy.value=false}}
async function previewRule(s:Setting){rule.value=s;busy.value=true;error.value='';preview.value=[];try{let offset=0,more=true;while(more&&preview.value.length<100){const r=await getWorkspace({currency:s.currency,kind:'expense',q:s.match,offset});preview.value.push(...r.items.filter(x=>!x.edit.reviewed&&x.category!==s.category&&(x.description+' '+x.merchant).toLowerCase().includes(s.match.toLowerCase())));more=r.hasMore;offset+=50}preview.value=preview.value.slice(0,100)}catch(e){preview.value=[];error.value=String(e)}finally{busy.value=false}}
async function applyRule(){if(!rule.value)return;busy.value=true;try{await saveEdits(preview.value.map(r=>({key:r.key,version:r.version,edit:{...r.edit,category:rule.value!.category,reviewed:false}})),'规则应用：'+rule.value.name);message.value=`已更新 ${preview.value.length} 条分类，可在记录处理历史中恢复。`;preview.value=[];rule.value=null;await load()}catch(e){error.value=String(e)}finally{busy.value=false}}
onMounted(load)
onBeforeUnmount(()=>{loadSequence++})
</script>

<style scoped>
.rule-editor { margin: 4px 0 20px; background: #fafaff; border-color: #d9daf0; scroll-margin: 16px; }
.setting-row > div:first-child { min-width: 0; overflow-wrap: anywhere; }
.setting-row > .actions { flex-shrink: 0; }
@media (max-width: 600px) {
  .setting-row { flex-direction: column; gap: 12px; padding-block: 16px; }
  .setting-row > .actions { justify-content: flex-end; }
  .rule-editor .form-grid { grid-template-columns: minmax(0, 1fr); }
}
</style>
