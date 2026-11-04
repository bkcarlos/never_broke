<template><header class="workspace-header"><div><span class="workspace-label">我的空间</span><span class="divider">/</span><strong>{{title}}</strong></div><div class="header-actions"><button class="icon-button" @click="settings.toggleShowAmount" :aria-label="settings.showAmount?'隐藏金额':'显示金额'"><Icon :name="settings.showAmount?'eye':'eye-close'" size="18px"/></button><RouterLink to="/settings/import-bills" class="header-import">导入账单</RouterLink><span class="avatar">我</span></div></header></template>
<script setup lang="ts">
import {computed} from 'vue'
import {useRoute} from 'vue-router'
import {useUserSettingsStore} from '@/stores/userSettings'
import Icon from './Icon.vue'
const route=useRoute(),settings=useUserSettingsStore()
const titles:Record<string,string>={'/':'财务总览','/detail':'交易账本','/reports':'分析报表','/review':'核对工作台','/plans':'账单与还款','/accounts':'账户管理','/accounts/new':'新增账户','/insights':'订阅与预算','/settings/import-bills':'导入与同步','/record':'记一笔','/timeline':'账本日历','/profile':'设置与偏好','/cashflow':'每月收支预测'}
const title=computed(()=>titles[route.path]||'财务设置')
</script>
<style scoped>.workspace-header{height:48px;border-bottom:1px solid #e9eaf0;background:white;display:flex;align-items:center;justify-content:space-between;padding:0 16px;font-size:12px}.workspace-label{color:#9297a5}.divider{margin:0 14px;color:#ced1da}.workspace-header strong{font-weight:500}.header-actions{display:flex;align-items:center;gap:18px}.header-import{font-size:12px;color:#666d80}.avatar{width:26px;height:26px;display:grid;place-content:center;background:#edeeff;border:1px solid #e0e2fc;color:#6b6ace;border-radius:50%;font-size:11px}</style>
