import { get, post } from './http'
export interface Category {id:string;name:string;aliases:string[]}
export interface Allocation {key:string;cents:number}
export interface Edit {category?:string;tags?:string[];reviewed:boolean;relationReviewed:boolean;allocations:Allocation[];relatedKeys:string[];status?:string;replacedBy?:string;note?:string}
export interface Entry {key:string;transactionId?:string;kind:string;date:string;currency:string;cents:number|null;description:string;merchant:string;category:string;categoryName:string;accountId:string;accountName:string;bank:string;channel:string;confirmation:string;tags:string[];relatedKeys:string[];relationStatus:string;note:string;metadata:Record<string,unknown>|null;source:boolean;version:number;edit:Edit}
export interface Metric {gross:number;refund:number;net:number;income:number;count:number;refundCount:number;average:number;pending:number}
export interface Group extends Metric {key:string;name:string}
export interface Analysis extends Metric {categories:Group[];merchants:Group[];channels:Group[];trend:Group[];first:string;last:string;months:string[]}
export interface Account {id:string;type:string;institution?:string;cardLast4?:string;currency:string;balance:number}
export interface Result {quality:{receipts:number;matchedReceipts:number;refunds:number;confirmedRefunds:number;unclassified:number;platform:number;feeCents:number;feeRecords:number};items:Entry[];total:number;offset:number;hasMore:boolean;analysis:Analysis;accounts:Account[];categories:Category[];queues:Record<string,number>;years:Record<string,boolean>}
export interface Change {id:string;key:string;before:string;after:string;reason:string;createdAt:string}
export interface Attachment {id:string;name:string;mime:string;size:number}
export interface Detail {attachments:Attachment[];record:Entry;related:Entry[];candidates:Entry[];history:Change[]}
export interface Setting {key:string;kind:string;name:string;currency:string;cents:number;category:string;match:string;date:string;period:string;enabled:boolean;version:number;updatedAt?:string}
export interface Batch {id:string;status:string;new:number;updated:number;skipped:number;error?:string;createdAt:string}
export function queryString(filters:Record<string,string|number>) {const q=new URLSearchParams();Object.entries(filters).forEach(([k,v])=>{if(v!==''&&v!==undefined)q.set(k,String(v))});return q.toString()}
export const getWorkspace=(f:Record<string,string|number>)=>get<Result>('/workspace?'+queryString(f))
export const getDetail=(key:string)=>get<Detail>('/workspace/record?key='+encodeURIComponent(key))
export const saveEdits=(items:{key:string;version:number;edit:Edit}[],reason:string)=>post('/workspace/edit',{items,reason})
export const getSettings=()=>get<Setting[]>('/workspace/settings')
export const saveSetting=(s:Setting)=>post<Setting>('/workspace/settings',s)
export const getBatches=()=>get<Batch[]>('/workspace/batches')
export const kindNames:Record<string,string>={expense:'消费',refund:'退款',income:'收入',transfer:'转账',receipt:'凭证',statement:'月度账单',schedule:'还款计划',reference:'参考明细',daily_summary:'每日汇总'}
export const stateNames:Record<string,string>={formal:'正式账单',pending:'摘要待确认',manual:'手工记录',unknown:'未标注',estimated:'预计',billed:'已出账',paid:'已还',cancelled:'已取消',replaced:'已替代'}
export const bankName=(v?:string)=>({cmb:'招商银行',ceb:'光大银行'}[v||'']||v||'账户')
export const accountName=(a:Account)=>bankName(a.institution)+(a.cardLast4?' · '+a.cardLast4:'')
export function localDate(d=new Date()){return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`}
export function amount(cents:number|null,currency='CNY',show=true){if(cents===null)return '待确认';if(!show)return '••••';return (currency==='USD'?'US$':'¥')+(cents/100).toLocaleString('zh-CN',{minimumFractionDigits:2,maximumFractionDigits:2})}
export const blankEdit=():Edit=>({reviewed:false,relationReviewed:false,allocations:[],relatedKeys:[]})

export interface CalendarMetric extends Metric {
  billed:number; estimated:number; billedCount:number; estimatedCount:number; unknownRepayments:number
}
export interface CalendarDay extends CalendarMetric {date:string}
export interface CalendarResult {
  month:string; day:string; currency:string; nextRepayment:string; days:CalendarDay[]; totals:CalendarMetric
  items:Entry[]; total:number; offset:number; hasMore:boolean; accounts:Account[]; categories:Category[]
}
export const getCalendar=(f:Record<string,string|number>)=>get<CalendarResult>('/workspace/calendar?'+queryString(f))
