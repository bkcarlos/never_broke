<template><div ref="container" class="chart" role="img" :aria-label="label"></div></template>
<script setup lang="ts">
import {ref,onMounted,onBeforeUnmount,watch} from 'vue'
import * as echarts from 'echarts/core'
import {BarChart} from 'echarts/charts'
import {GridComponent,TooltipComponent,LegendComponent} from 'echarts/components'
import {CanvasRenderer} from 'echarts/renderers'
echarts.use([BarChart,GridComponent,TooltipComponent,LegendComponent,CanvasRenderer])
import type {Group} from '@/api/workspace'
const props=defineProps<{items:Group[];type:'trend'|'category';metric?:'gross'|'net';show:boolean;label:string}>()
const emit=defineEmits<{(e:'select',key:string):void}>()
const container=ref<HTMLElement|null>(null)
let chart:echarts.ECharts|undefined,observer:ResizeObserver|undefined
function draw(){if(!chart)return;const trend=props.type==='trend';const data=trend?props.items:props.items.slice(0,12);const metric=props.metric||'gross'
chart.setOption({animationDuration:250,color:['#7474d9','#9fcbbc'],textStyle:{fontFamily:'-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif',fontSize:10,color:'#9198a8'},tooltip:{trigger:'axis',valueFormatter:(v:unknown)=>props.show?Number(v).toLocaleString('zh-CN',{minimumFractionDigits:2,maximumFractionDigits:2}):'••••'},legend:trend?{bottom:0,itemWidth:8,itemHeight:8,textStyle:{fontSize:10,color:'#8a92a3'}}:undefined,grid:{left:trend?48:100,right:22,top:22,bottom:trend?52:25},xAxis:trend?{type:'category',data:data.map(x=>x.key),axisTick:{show:false},axisLine:{lineStyle:{color:'#eceef5'}},axisLabel:{formatter:(s:string)=>s.slice(5)}}:{type:'value',axisLabel:{formatter:(v:number)=>props.show?String(v):'••••'},splitLine:{lineStyle:{color:'#f0f1f6'}}},yAxis:trend?{type:'value',axisLabel:{formatter:(v:number)=>props.show?String(v):'••••'},splitLine:{lineStyle:{color:'#f0f1f6'}}}:{type:'category',inverse:true,data:data.map(x=>x.name),axisTick:{show:false},axisLine:{show:false},axisLabel:{width:88,overflow:'truncate'}},series:trend?[{name:'消费总额',type:'bar',barMaxWidth:18,itemStyle:{borderRadius:[3,3,0,0]},data:data.map(x=>x.gross/100)},{name:'退款',type:'bar',barMaxWidth:18,itemStyle:{borderRadius:[3,3,0,0]},data:data.map(x=>x.refund/100)}]:[{name:metric==='gross'?'消费总额':'净支出',type:'bar',barMaxWidth:12,itemStyle:{borderRadius:[0,3,3,0]},data:data.map(x=>x[metric]/100)}]},true)
}
onMounted(()=>{if(!container.value)return;chart=echarts.init(container.value);chart.on('click',p=>{const g=props.items[p.dataIndex];if(g)emit('select',g.key)});observer=new ResizeObserver(()=>chart?.resize());observer.observe(container.value);draw()})
watch(()=>[props.items,props.type,props.metric,props.show],draw,{deep:true})
onBeforeUnmount(()=>{observer?.disconnect();chart?.dispose()})
</script>
