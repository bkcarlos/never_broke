<template>
  <PageFrame title="设置与偏好" eyebrow="SETTINGS & PREFERENCES" subtitle="管理你的账本、收入计划和日常使用偏好。" compact>
    <p v-if="loading" class="notice" role="status">{{ t('common.loading') }}</p>
    <div v-if="error" class="error-box" role="alert">{{ error }}</div>
    <div class="settings-layout">
      <div class="settings-groups">
        <section v-for="group in groups" :key="group.title" class="panel">
          <div class="panel-head"><div><h2>{{ group.title }}</h2><p class="subtitle">{{ group.description }}</p></div></div>
          <RouterLink v-for="link in group.links" :key="link.path" :to="link.path" class="settings-link">
            <span class="setting-icon"><Icon :name="link.icon" size="20px" /></span>
            <div><strong>{{ link.title }}</strong><p>{{ link.description }}</p></div>
            <Icon name="chevron-right" size="16px" class="setting-chevron" />
          </RouterLink>
        </section>
      </div>
      <aside class="settings-preferences">
        <section class="panel">
          <div class="panel-head"><h2>使用偏好</h2><Icon name="settings" size="17px" class="subtle" /></div>
          <div class="preference-row">
            <label id="locale-label">{{ t('profile.language') }}</label>
            <div class="locale-options" role="group" aria-labelledby="locale-label">
              <button v-for="option in locales" :key="option.value" :class="{ selected: userSettingsStore.locale === option.value }" :aria-pressed="userSettingsStore.locale === option.value" :disabled="loading" @click="handleLocaleSelect(option.value)">{{ option.label }}</button>
            </div>
          </div>
          <div class="preference-row amount-preference">
            <div><label id="amount-label">金额显示</label><p>控制工作台中的金额显隐</p></div>
            <button class="amount-switch" role="switch" :aria-checked="userSettingsStore.showAmount" aria-labelledby="amount-label" :class="{ enabled: userSettingsStore.showAmount }" :disabled="loading" @click="toggleShowAmount"><span /></button>
          </div>
        </section>
        <section class="panel session-panel">
          <h2>登录与安全</h2><p class="subtitle">在共享设备上使用后，记得退出登录。</p>
          <button class="button danger" @click="handleLogout"><Icon name="logout" size="16px" />{{ t('auth.logout') }}</button>
        </section>
      </aside>
    </div>
  </PageFrame>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { logout } from '@/api/auth'
import { t } from '@/i18n'
import type { Locale } from '@/types'
import Icon from '@/components/Icon.vue'
import PageFrame from '@/components/PageFrame.vue'
import { useUserSettingsStore } from '@/stores/userSettings'

const router = useRouter()
const groups = [
  { title: '账本与账户', description: '管理账户、预算与账单来源', links: [
    { path: '/accounts', icon: 'wallet', title: '账户管理', description: '查看账户、核验余额与欠款' },
    { path: '/settings/budget', icon: 'target', title: '月度预算', description: '设置每月支出额度与每日参考' },
    { path: '/settings/import-bills', icon: 'import', title: '导入与同步', description: '导入账单，整理历史交易' },
  ] },
  { title: '收入与规划', description: '把未来的收入和支出安排清楚', links: [
    { path: '/me/salary', icon: 'money', title: '薪资管理', description: '配置工资方案，预览税后收入' },
    { path: '/settings/recurring-income', icon: 'repeat', title: '固定收入', description: '管理租金、副业等周期收入' },
    { path: '/tools/salary-calculator', icon: 'chart', title: '薪资计算器', description: '试算社保、公积金与个税' },
    { path: '/cashflow', icon: 'calendar', title: '收支预测', description: '查看未来结余与资金缺口' },
  ] },
]
const userSettingsStore = useUserSettingsStore()
const locales: { value: Locale; label: string }[] = [{ value: 'zh-CN', label: '简体中文' }, { value: 'en-US', label: 'English' }]

const error = computed(() => userSettingsStore.error)
const loading = computed(() => userSettingsStore.loading)

async function handleLocaleSelect(locale: Locale) {
  await userSettingsStore.updateLocale(locale)
}

async function toggleShowAmount() {
  await userSettingsStore.toggleShowAmount()
}

// 退出登录
async function handleLogout() {
  if (confirm(t('auth.logoutConfirm'))) {
    try {
      await logout()
      router.push('/login')
    } catch (error) {
      console.error('Logout error:', error)
      // 即使失败也清除本地数据并跳转
      router.push('/login')
    }
  }
}

onMounted(() => {
  userSettingsStore.loadSettings()
})
</script>

<style scoped>
.settings-layout { display: grid; grid-template-columns: minmax(0, 1fr) 320px; gap: 20px; align-items: start; }
.settings-groups, .settings-preferences { display: grid; gap: 20px; min-width: 0; }
.settings-link { display: flex; align-items: center; gap: 14px; padding: 18px 20px; text-decoration: none; color: var(--ink); border-bottom: 1px solid #f0f1f6; }
.settings-link:last-child { border-bottom: 0; }
.settings-link:hover { background: #fafaff; }
.setting-icon { display: grid; place-items: center; width: 38px; height: 38px; flex-shrink: 0; border: 1px solid #e9e9f7; border-radius: 9px; background: #f4f4fc; color: var(--brand); }
.settings-link strong { font-size: 13px; font-weight: 550; }
.settings-link p, .preference-row p { font-size: 11px; line-height: 1.7; color: var(--muted); margin-top: 4px; }
.setting-chevron { margin-left: auto; flex-shrink: 0; color: #a3a8b6; }
.preference-row { padding: 20px; border-bottom: 1px solid var(--line); }
.preference-row:last-child { border-bottom: 0; }
.preference-row label, .session-panel h2 { font-size: 13px; font-weight: 550; }
.locale-options { display: flex; padding: 3px; background: #f3f4f8; border-radius: 7px; margin-top: 12px; gap: 3px; }
.locale-options button { flex: 1; min-height: 34px; border-radius: 5px; font-size: 12px; color: var(--muted); }
.locale-options button.selected { background: white; color: var(--brand); box-shadow: 0 1px 3px #242a3b12; }
.amount-preference { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.amount-switch { border-radius: 20px; width: 40px; height: 24px; flex-shrink: 0; padding: 3px; background: #c9cdd8; transition: background .15s; }
.amount-switch.enabled { background: var(--brand); }
.amount-switch span { display: block; width: 18px; height: 18px; border-radius: 50%; background: white; transition: transform .15s; }
.amount-switch.enabled span { transform: translateX(16px); }
.session-panel { padding: 20px; }
.session-panel .button { margin-top: 18px; width: 100%; }
@media (max-width: 900px) { .settings-layout { grid-template-columns: minmax(0, 1fr); } .settings-preferences { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 560px) { .settings-preferences { grid-template-columns: minmax(0, 1fr); } .settings-link { padding: 16px; } }
</style>
