<template>
  <!-- 主应用布局 -->
  <div v-if="!isLoginPage" class="app-layout" :class="layoutClass">
    <!-- 侧边栏 - 桌面端和平板 -->
    <Sidebar v-if="!isMobile" class="app-sidebar" />
    
    <!-- 移动端顶部栏 -->
    <header v-if="isMobile" class="mobile-topbar">
      <div class="mobile-topbar-safe-area">
        <div class="mobile-topbar-content">
          <h1 class="mobile-topbar-title">{{ pageTitle }}</h1>
          <div class="mobile-topbar-actions">
            <RouterLink to="/timeline" class="mobile-icon-btn touch-target" aria-label="账本日历"><Icon name="calendar" size="20px" /></RouterLink>
            <button 
              @click="userSettingsStore.toggleShowAmount" 
              class="mobile-icon-btn touch-target no-touch-zoom"
              :title="userSettingsStore.showAmount ? t('profile.hideAmount') : t('profile.showAmount')"
            >
              <Icon :name="userSettingsStore.showAmount ? 'eye' : 'eye-close'" size="20px" />
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- PC端顶部栏 -->
    <TopBar v-if="!isMobile" class="app-topbar" />
    
    <!-- 主内容区域 -->
    <main class="app-main">
      <router-view />
    </main>

    <!-- 移动端底部导航 -->
    <nav v-if="isMobile" class="bottom-nav">
      <div class="bottom-nav-safe-area">
        <div class="bottom-nav-content">
          <!-- 左侧两个 tab -->
          <router-link 
            v-for="tab in leftTabs" 
            :key="tab.path"
            :to="tab.path"
            class="nav-item touch-target no-touch-zoom"
            :class="{ 'active': isActive(tab.path) }"
            @click="handleNavClick"
          >
            <Icon 
              :name="tab.icon" 
              size="24px"
              class="nav-icon"
            />
            <span class="nav-label">{{ tab.label }}</span>
          </router-link>

          <!-- 中间的 + 按钮 -->
          <button class="add-btn touch-target no-touch-zoom" aria-label="记一笔" :aria-expanded="isRadialOpen" @click="toggleRadial">
            <span :class="{ 'rotate-45': isRadialOpen }">+</span>
          </button>

          <!-- 右侧两个 tab -->
          <router-link 
            v-for="tab in rightTabs" 
            :key="tab.path"
            :to="tab.path"
            class="nav-item touch-target no-touch-zoom"
            :class="{ 'active': isActive(tab.path) }"
            @click="handleNavClick"
          >
            <Icon 
              :name="tab.icon" 
              size="24px"
              class="nav-icon"
            />
            <span class="nav-label">{{ tab.label }}</span>
          </router-link>
        </div>
      </div>
    </nav>

    <!-- 圆弧菜单（移动端） -->
    <RadialActionMenu
      :visible="isRadialOpen"
      :actions="radialActions"
      @close="isRadialOpen = false"
      @select="handleSelect"
    />
  </div>

  <!-- 登录页面独立布局 -->
  <div v-else class="login-container">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from './components/Icon.vue'
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import RadialActionMenu from './components/RadialActionMenu.vue'
import { t } from './i18n'
import type { RadialAction } from './types/actions'
import { useBreakpoint } from './composables/useBreakpoint'
import { useUserSettingsStore } from './stores/userSettings'

const route = useRoute()
const router = useRouter()
const userSettingsStore = useUserSettingsStore()

// 响应式断点
const { isMobile, isTablet } = useBreakpoint()

const isRadialOpen = ref<boolean>(false)

// 判断是否是登录页面
const isLoginPage = computed(() => route.path === '/login')

// 布局类名
const layoutClass = computed(() => ({
  'layout-mobile': isMobile.value,
  'calendar-layout': route.path === '/timeline',
  'layout-tablet': isTablet.value,
  'layout-desktop': !isMobile.value && !isTablet.value,
}))

// 页面标题（移动端顶部栏显示）
const pageTitle = computed(() => {
  const pageTitles: Record<string, string> = {
    '/': t('nav.home'),
    '/detail': '交易账本',
    '/review':'核对工作台', '/plans':'账单与还款', '/insights':'订阅与预算', '/accounts':'账户管理', '/settings/import-bills':'导入与同步',
    '/reports': t('nav.reports'),
    '/timeline': '账本日历',
    '/profile': t('nav.profile'),
    '/record': t('actions.record'),
  }
  return pageTitles[route.path] || t('app.title')
})

// 监听路由变化，关闭圆弧菜单
watch(() => route.path, () => {
  if (isRadialOpen.value) {
    isRadialOpen.value = false
  }
})

// 初始化用户设置
onMounted(() => {
  userSettingsStore.initFromLocalStorage()
})

// 左侧两个 tab (移动端)
const leftTabs = [
  { path: '/', icon: 'home', label: t('nav.home') },
  { path: '/detail', icon: 'list', label: t('nav.detail') }
]

// 右侧两个 tab (移动端)
const rightTabs = [
  { path: '/reports', icon: 'chart', label: t('nav.reports') },
  { path: '/profile', icon: 'user', label: t('nav.profile') }
]

const radialActions: RadialAction[] = [
  {
    key: 'record_expense',
    label: t('actions.expense'),
    icon: 'minus-circle',
    route: '/record?type=expense'
  },
  {
    key: 'record_income',
    label: t('actions.income'),
    icon: 'add-circle',
    route: '/record?type=income'
  },
  {
    key: 'record_transfer',
    label: t('actions.transfer'),
    icon: 'transfer',
    route: '/record?type=transfer'
  }
]

const isActive = (path: string) => {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}

function toggleRadial() {
  isRadialOpen.value = !isRadialOpen.value
}

function handleNavClick() {
  if (isRadialOpen.value) {
    isRadialOpen.value = false
  }
}

function handleSelect(action: RadialAction) {
  isRadialOpen.value = false
  const [path, queryString] = action.route.split('?')
  const query: Record<string, string> = {}
  if (queryString) {
    queryString.split('&').forEach(param => {
      const [key, value] = param.split('=')
      if (key && value) {
        query[key] = decodeURIComponent(value)
      }
    })
  }
  router.push({ path, query })
}
</script>

<style scoped>
.mobile-topbar-actions{display:flex;align-items:center;gap:2px}
.app-layout{display:grid;height:100dvh;min-height:0;background:#f6f7fa;grid-template-areas:"sidebar header" "sidebar main";grid-template-columns:192px minmax(0,1fr);grid-template-rows:48px minmax(0,1fr)}
.app-sidebar{grid-area:sidebar}.app-topbar{grid-area:header}.app-main{grid-area:main;min-width:0;overflow:auto;min-height:0;padding:16px}.layout-tablet{grid-template-columns:76px minmax(0,1fr)}.login-container{min-height:100dvh}
.mobile-topbar{grid-area:header;background:#fff;border-bottom:1px solid #e6e8ee}.mobile-topbar-content{height:52px;display:flex;align-items:center;justify-content:space-between;padding:0 18px}.mobile-topbar-title{font-size:15px;font-weight:600}.mobile-icon-btn{display:grid;place-items:center;background:transparent;border:0;color:#667085}.bottom-nav{grid-area:footer;background:white;border-top:1px solid #e6e8ee;padding-bottom:env(safe-area-inset-bottom);z-index:30}.bottom-nav-content{display:flex;align-items:center;justify-content:space-around;height:59px}.nav-item{display:flex;flex-direction:column;align-items:center;gap:3px;font-size:10px;color:#8a90a0;text-decoration:none;min-width:56px}.nav-item.active{color:#5658d7}.add-btn{width:38px;height:38px;background:#5859dc;color:white;border-radius:11px;font-size:25px}.rotate-45{display:block;transform:rotate(45deg)}
.layout-mobile{grid-template-areas:"header" "main" "footer";grid-template-columns:minmax(0,1fr);grid-template-rows:52px minmax(0,1fr) auto}.layout-mobile .app-main{padding:12px}.layout-mobile.calendar-layout .app-main{padding:10px}.mobile-topbar-safe-area{padding-top:env(safe-area-inset-top)}
</style>
