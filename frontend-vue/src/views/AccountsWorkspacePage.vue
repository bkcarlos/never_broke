<template>
  <PageFrame title="账户管理" eyebrow="ACCOUNTS" subtitle="核实当前余额与欠款，和历史消费分开管理。">
    <template #actions><RouterLink to="/transfer" class="button">账户转账</RouterLink><RouterLink to="/accounts/new" class="button primary">＋ 新增账户</RouterLink></template>
    <div class="notice">导入的账户未核验前显示“待核实”。信用卡填写当前欠款，其他账户填写可用余额；这不会生成消费记录。</div>
    <div v-if="error" class="error-box" role="alert">{{ error }} <button v-if="!draft" class="button" @click="load">重试</button></div>
    <p v-if="message" class="notice" role="status">{{ message }}</p>
    <p v-if="loading" class="empty" role="status">正在读取账户与核验记录…</p>
    <template v-else>
      <form v-if="draft" ref="verificationForm" class="settings-form verification-form" @submit.prevent="save">
        <div class="verification-heading"><div><h2>{{ draft.name }} · 核验余额</h2><p class="subtitle">以核验当天的实际余额或欠款为准。</p></div><span class="tag purple">{{ draft.currency }}</span></div>
        <div class="form-grid">
          <label class="field">金额（{{ draft.currency }}）<input ref="balanceInput" v-model.number="balance" class="control" type="number" step="0.01" required :disabled="busy" /></label>
          <label class="field">核验日期<input v-model="draft.date" class="control" type="date" required :disabled="busy" /></label>
        </div>
        <div class="actions verification-actions"><button type="submit" class="button primary" :disabled="busy">{{ busy ? '保存中…' : '保存已核验余额' }}</button><button type="button" class="button" :disabled="busy" @click="draft=null">取消</button></div>
      </form>
      <section v-if="!accounts.length && !error" class="panel empty"><strong>还没有账户</strong><p class="subtitle">添加现金、银行卡或钱包，开始管理余额。</p><RouterLink to="/accounts/new" class="button primary empty-action">＋ 新增账户</RouterLink></section>
      <div class="two-grid account-grid">
        <section v-for="a in accounts" :key="a.id" class="panel" :class="{ 'account-editing': draft?.match === a.id }">
          <div class="panel-head"><h2>{{ accountName(a) }}</h2><span class="tag">{{ a.currency }}</span></div>
          <div class="panel-body">
            <p class="metric-label">{{ a.type === 'credit' ? '当前欠款' : '当前余额' }}</p>
            <div class="metric-value">{{ verified(a.id) ? amount(Math.round(a.balance * 100), a.currency, show) : '待核实' }}</div>
            <p class="subtitle">{{ verified(a.id) ? '最近核验：' + verified(a.id)?.date : a.type === 'credit' ? '尚未确认实际欠款，不能视为零负债。' : '尚未确认实际余额，暂不用于余额预测。' }}</p>
            <p v-if="!supportsVerification(a)" class="subtitle">余额核验目前支持人民币和美元。</p>
            <div class="actions account-actions"><button class="button" :disabled="busy || !supportsVerification(a)" @click="start(a)">{{ !supportsVerification(a) ? '暂不支持核验' : draft?.match === a.id ? '正在核验' : '核验余额' }}</button><RouterLink :to="{path:'/detail',query:{account:a.id,currency:a.currency}}" class="button ghost">查看交易 →</RouterLink></div>
          </div>
        </section>
      </div>
    </template>
  </PageFrame>
</template>
<script setup lang="ts">
import {ref, computed, onMounted, nextTick} from 'vue'
import {getWorkspace, getSettings, saveSetting, accountName, amount, localDate} from '@/api/workspace'
import type {Account, Setting} from '@/api/workspace'
import {useUserSettingsStore} from '@/stores/userSettings'
import PageFrame from '@/components/PageFrame.vue'

const accounts=ref<Account[]>([]), settings=ref<Setting[]>([]), draft=ref<Setting|null>(null)
const balance=ref(0), busy=ref(false), loading=ref(false), error=ref(''), message=ref('')
const verificationForm=ref<HTMLElement|null>(null), balanceInput=ref<HTMLInputElement|null>(null)
const user=useUserSettingsStore(), show=computed(()=>user.showAmount)
function supportsVerification(a:Account){return ['CNY','USD'].includes(a.currency)}
function verified(id:string){return settings.value.find(s=>s.kind==='balance'&&s.match===id)}
async function load(){
  loading.value=true;error.value=''
  try{
    const [r,s]=await Promise.all([getWorkspace({currency:'CNY',kind:'ledger'}),getSettings()])
    accounts.value=r.accounts;settings.value=s
  }catch(e){error.value=e instanceof Error?e.message:'读取账户失败'}finally{loading.value=false}
}
async function start(a:Account){
  const s=verified(a.id)
  error.value='';message.value=''
  draft.value=s?{...s,date:localDate()}:{key:'balance:'+a.id,kind:'balance',name:accountName(a),currency:a.currency,cents:0,category:'',match:a.id,date:localDate(),period:'',enabled:true,version:0}
  draft.value.name=accountName(a)
  balance.value=s?a.balance:0
  await nextTick()
  verificationForm.value?.scrollIntoView({block:'nearest'})
  balanceInput.value?.focus({preventScroll:true})
}
async function save(){
  if(!draft.value||busy.value)return
  if(!Number.isFinite(balance.value)||!draft.value.date){error.value='请填写有效金额与核验日期';return}
  busy.value=true;error.value=''
  try{
    await saveSetting({...draft.value,cents:Math.round(balance.value*100)})
    draft.value=null;message.value='账户余额已核验。'
    await load()
  }catch(e){error.value=e instanceof Error?e.message:'保存核验失败'}finally{busy.value=false}
}
onMounted(load)
</script>
<style scoped>
.verification-form { margin: 16px 0; border-color: #b8b8ec; scroll-margin: 16px; }
.verification-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 18px; }
.verification-heading h2 { font-size: 14px; font-weight: 600; }
.verification-actions { margin-top: 18px; }
.account-actions { margin-top: 20px; }
.account-editing { border-color: var(--brand); }
.account-grid .panel-head h2 { min-width: 0; overflow-wrap: anywhere; }
.empty-action { margin-top: 18px; }
@media (max-width: 480px) { .verification-form .form-grid { grid-template-columns: minmax(0, 1fr); } }
</style>
