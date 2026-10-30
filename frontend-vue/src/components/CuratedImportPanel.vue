<template>
 <section class="bg-white rounded-lg shadow p-4 mb-6 space-y-4">
  <h2 class="font-semibold text-lg">整理账单导入与查阅</h2>
  <p class="text-sm text-gray-600">消费和退款计入账本；收据仅补充说明，预计还款单独展示。相同记录重复导入会跳过，历史导入不改变当前余额。新账户余额需另行核实设置。</p>
  <input aria-label="选择整理账单 JSON 文件" type="file" accept=".json,application/json" @change="selectFile" :disabled="busy" />
  <div v-if="summary" class="bg-blue-50 p-3 rounded text-sm">
   共 {{ summary.records }} 条，新增 {{ summary.new }} 条，已有 {{ summary.skipped }} 条。
   <div v-for="(cents, currency) in summary.netExpenseCents" :key="currency">{{ currency }} 净消费：{{ (cents / 100).toFixed(2) }}</div>
   <button v-if="!done" class="mt-2 px-4 py-2 bg-blue-600 text-white rounded" :disabled="busy" @click="commit">{{ busy ? '正在导入…' : '确认导入' }}</button>
   <p v-else class="text-green-700">导入完成，日历和报表已可查看。</p>
  </div>
  <p v-if="error" role="alert" class="text-red-600">{{ error }}</p>
  <div class="flex gap-3 flex-wrap items-center">
   <label>类型 <select v-model="kind" @change="offset=0; load()" class="border rounded p-1"><option value="">全部记录</option><option v-for="(name, key) in names" :key="key" :value="key">{{ name }}</option></select></label>
   <label>币种 <select v-model="currency" @change="offset=0; load()" class="border rounded p-1"><option>CNY</option><option>USD</option></select></label>
   <span class="text-sm text-gray-500">{{ total }} 条</span>
  </div>
  <div v-if="detail" class="border border-blue-200 rounded p-3 bg-blue-50 whitespace-pre-wrap">
   <button @click="detail=null" class="float-right text-blue-600">关闭详情</button>
   <p class="font-semibold">{{ detail.description }}</p><p>{{ detail.date }} · {{ detail.currency }} {{ amount(detail) }}</p>
   <p>{{ detail.category }} · {{ detail.tags?.join('、') }}</p><p>{{ detail.note }}</p><p v-if="detail.relationStatus">关联状态：{{ detail.relationStatus }}</p>
   <button v-for="(key,i) in detail.relatedKeys" :key="key" @click="openRelated(key)" class="text-blue-600 mr-3 underline">查看关联记录 {{ i+1 }}</button>
  </div>
  <div class="overflow-x-auto max-h-[440px]">
   <table class="w-full text-sm text-left"><thead class="sticky top-0 bg-gray-50"><tr><th class="p-2">日期</th><th>类型</th><th>摘要 / 分类</th><th class="text-right">金额</th></tr></thead>
    <tbody><tr v-for="r in records" :key="r.key" class="border-t cursor-pointer hover:bg-blue-50" @click="detail=r"><td class="p-2 whitespace-nowrap">{{ r.date }}</td><td>{{ names[r.kind] }}</td><td class="p-2">{{ r.description }}<div class="text-xs text-gray-500">{{ r.category }} {{ r.relationStatus }}</div></td><td class="text-right whitespace-nowrap">{{ r.currency }} {{ amount(r) }}</td></tr></tbody>
   </table>
  </div>
  <div class="flex gap-4"><button :disabled="offset===0" @click="offset-=100; load()">上一页</button><button :disabled="!hasMore" @click="offset+=100; load()">下一页</button></div>
 </section>
</template>
<script setup lang="ts">
import {ref,onMounted} from 'vue'
import {useRoute} from 'vue-router'
import {request} from '@/api/http'
interface Entry {key:string;kind:string;date:string;currency:string;description:string;amountCents?:number;category?:string;tags?:string[];note?:string;relatedKeys?:string[];relationStatus?:string}
interface Summary {records:number;new:number;skipped:number;netExpenseCents:Record<string,number>}
const names:Record<string,string>={expense:'消费',refund:'退款',receipt:'收据',statement:'月度账单',schedule:'待核实 / 预计还款',reference:'其他明细',daily_summary:'每日汇总'}
const records=ref<Entry[]>([]),detail=ref<Entry|null>(null),summary=ref<Summary|null>(null),bundle=ref<unknown>(null)
const kind=ref('expense'),currency=ref('CNY'),offset=ref(0),total=ref(0),hasMore=ref(false),busy=ref(false),done=ref(false),error=ref('')
const amount=(r:Entry)=>r.amountCents==null?'待确认':(r.amountCents/100).toFixed(2)
async function load(){try {const q=new URLSearchParams({kind:kind.value,currency:currency.value,offset:String(offset.value)});const d=await request<{items:Entry[];total:number;hasMore:boolean}>(`/import/curated?${q}`);records.value=d.items;total.value=d.total;hasMore.value=d.hasMore}catch(e){error.value=String(e)}}
async function selectFile(e:Event){bundle.value=null;summary.value=null;done.value=false;error.value='';const f=(e.target as HTMLInputElement).files?.[0];if(!f)return;if(f.size>16*1024*1024){error.value='文件不能超过 16 MB';return}busy.value=true;try{const b=JSON.parse(await f.text());summary.value=await request<Summary>('/import/curated/preview',{method:'POST',body:JSON.stringify(b)});bundle.value=b}catch(e){error.value=String(e)}finally{busy.value=false}}
async function commit(){if(!bundle.value||busy.value)return;busy.value=true;error.value='';try{summary.value=await request<Summary>('/import/curated/commit',{method:'POST',body:JSON.stringify(bundle.value)});done.value=true;await load()}catch(e){error.value=String(e)}finally{busy.value=false}}
async function openRelated(key:string){try{const d=await request<{items:Entry[]}>(`/import/curated?key=${encodeURIComponent(key)}`);detail.value=d.items[0]??null}catch(e){error.value=String(e)}}
const route=useRoute()
onMounted(async()=>{await load();if(typeof route.query.transactionId==='string'){try{const d=await request<{items:Entry[]}>(`/import/curated?transactionId=${encodeURIComponent(route.query.transactionId)}`);detail.value=d.items[0]??null}catch(e){error.value=String(e)}}})
</script>
