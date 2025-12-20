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
            <button 
              @click="toggleShowAmount" 
              class="mobile-icon-btn touch-target no-touch-zoom"
              :title="showAmount ? t('profile.hideAmount') : t('profile.showAmount')"
            >
              <Icon :name="showAmount ? 'eye' : 'eye-close'" size="20px" />
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
          <button class="add-btn touch-target no-touch-zoom" @click="toggleRadial">
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
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from './components/Icon.vue'
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import RadialActionMenu from './components/RadialActionMenu.vue'
import { t } from './i18n'
import type { RadialAction } from './types/actions'
import { useBreakpoint } from './composables/useBreakpoint'

const route = useRoute()
const router = useRouter()

// 响应式断点
const { isMobile, isTablet, currentBreakpoint } = useBreakpoint()

const isRadialOpen = ref<boolean>(false)
const showAmount = ref<boolean>(true)

// 判断是否是登录页面
const isLoginPage = computed(() => route.path === '/login')

// 布局类名
const layoutClass = computed(() => ({
  'layout-mobile': isMobile.value,
  'layout-tablet': isTablet.value,
  'layout-desktop': !isMobile.value && !isTablet.value,
}))

// 页面标题（移动端顶部栏显示）
const pageTitle = computed(() => {
  const pageTitles: Record<string, string> = {
    '/': t('nav.home'),
    '/detail': t('nav.detail'),
    '/reports': t('nav.reports'),
    '/timeline': t('nav.timeline'),
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

// 初始化金额显隐设置
const stored = localStorage.getItem('showAmount')
if (stored !== null) {
  showAmount.value = stored === 'true'
}

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

function toggleShowAmount() {
  showAmount.value = !showAmount.value
  localStorage.setItem('showAmount', String(showAmount.value))
  // 触发全局事件通知其他组件
  window.dispatchEvent(new CustomEvent('toggle-amount-visibility', { 
    detail: { showAmount: showAmount.value } 
  }))
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
/* ============================================
   主应用布局 - 使用 CSS Grid
   ============================================ */
.app-layout {
  min-height: 100vh;
  height: 100%;
  background-color: #f9fafb;
  display: grid;
}

/* 桌面端布局：侧边栏 + 顶部栏 + 内容区 */
.app-layout.layout-desktop {
  grid-template-areas: 
    "sidebar header"
    "sidebar main";
  grid-template-columns: 260px 1fr;
  grid-template-rows: 70px 1fr;
}

/* 平板布局：精简侧边栏 + 顶部栏 + 内容区 */
.app-layout.layout-tablet {
  grid-template-areas: 
    "sidebar header"
    "sidebar main";
  grid-template-columns: 80px 1fr;
  grid-template-rows: 70px 1fr;
}

/* 移动端布局：顶部栏 + 内容区 + 底部导航 */
.app-layout.layout-mobile {
  grid-template-areas: 
    "header"
    "main"
    "footer";
  grid-template-columns: 1fr;
  grid-template-rows: auto 1fr auto;
}

.app-sidebar {
  grid-area: sidebar;
}

.app-topbar {
  grid-area: header;
}

.app-main {
  grid-area: main;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 2rem;
  background-color: #f9fafb;
}

.bottom-nav {
  grid-area: footer;
}

.login-container {
  width: 100%;
  min-height: 100vh;
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* ============================================
   移动端顶部栏 - 刘海屏适配
   ============================================ */
.mobile-topbar {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.mobile-topbar-safe-area {
  /* 关键：使用 env() 添加顶部安全区域 */
  padding-top: env(safe-area-inset-top, 0);
  padding-top: constant(safe-area-inset-top, 0); /* iOS 11.0-11.2 */
}

.mobile-topbar-content {
  height: 56px;
  padding: 0 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mobile-topbar-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  color: white;
}

.mobile-topbar-actions {
  display: flex;
  gap: 0.5rem;
}

.mobile-icon-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  border: none;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.mobile-icon-btn:active {
  transform: scale(0.95);
  background: rgba(255, 255, 255, 0.3);
}

/* ============================================
   移动端底部导航 - 安全区域适配
   ============================================ */
.bottom-nav {
  position: sticky;
  bottom: 0;
  border-top: 1px solid #e5e7eb;
  background: white;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  z-index: 1000;
}

.bottom-nav-safe-area {
  /* 底部安全区域适配（Home指示条） */
  padding-bottom: env(safe-area-inset-bottom, 0);
  padding-bottom: constant(safe-area-inset-bottom, 0);
  
  /* 左右安全区域 */
  padding-left: env(safe-area-inset-left, 0);
  padding-right: env(safe-area-inset-right, 0);
}

.bottom-nav-content {
  display: flex;
  justify-content: space-around;
  align-items: center;
  height: 64px;
  padding: 0 1rem;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 64px;
  color: #6b7280;
  text-decoration: none;
  transition: all 0.2s;
  position: relative;
  -webkit-tap-highlight-color: transparent;
}

.nav-item.active {
  color: #6366f1;
  font-weight: 600;
}

.nav-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 3px;
  background: #6366f1;
  border-radius: 0 0 3px 3px;
  transition: width 0.3s ease;
}

.nav-item.active::before {
  width: 60%;
}

.nav-icon {
  margin-bottom: 0.25rem;
  transition: transform 0.2s;
}

.nav-item.active .nav-icon {
  transform: scale(1.1);
}

.nav-label {
  font-size: 0.75rem;
  font-weight: 500;
}

.nav-item.active .nav-label {
  font-weight: 600;
}

.add-btn {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  color: white;
  border: none;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  font-size: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: -24px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  -webkit-tap-highlight-color: transparent;
}

.add-btn:active {
  transform: scale(0.95);
}

.add-btn span {
  transition: transform 0.3s;
}

.add-btn span.rotate-45 {
  transform: rotate(45deg);
}

/* ============================================
   响应式设计
   ============================================ */

/* 移动端 (< 768px) */
@media (max-width: 767px) {
  .app-main {
    padding: 1rem;
    /* 内容区不需要额外的 padding-top 和 padding-bottom，
       因为 Grid 布局会自动处理空间分配 */
  }

  /* 横屏优化 */
  @media (orientation: landscape) {
    .mobile-topbar-content {
      height: 48px;
      padding-left: max(1rem, env(safe-area-inset-left, 0));
      padding-right: max(1rem, env(safe-area-inset-right, 0));
    }
    
    .bottom-nav-content {
      height: 56px;
    }
  }
}

/* 平板 (768px - 1023px) */
@media (min-width: 768px) and (max-width: 1023px) {
  .app-main {
    padding: 1.5rem;
  }
}

/* 桌面端 (1024px - 1919px) */
@media (min-width: 1024px) and (max-width: 1919px) {
  .app-main {
    padding: 2rem;
    /* 移除 max-width 限制，让内容占满整个可用空间 */
  }
}

/* 2K显示器 (1920px - 2559px) */
@media (min-width: 1920px) and (max-width: 2559px) {
  .app-layout.layout-desktop {
    grid-template-columns: 280px 1fr;
    grid-template-rows: 80px 1fr;
  }
  
  .app-main {
    padding: 3rem;
    /* 移除 max-width 限制 */
  }
}

/* 4K显示器 (≥ 2560px) */
@media (min-width: 2560px) {
  .app-layout.layout-desktop {
    grid-template-columns: 320px 1fr;
    grid-template-rows: 90px 1fr;
  }
  
  .app-main {
    padding: 4rem;
    /* 移除 max-width 限制 */
  }
}

/* ============================================
   工具样式
   ============================================ */

/* 确保下拉框不影响导航 */
select,
input[type="date"],
input[type="month"] {
  position: relative;
  z-index: 1;
}

select:focus,
input[type="date"]:focus,
input[type="month"]:focus {
  z-index: 999;
}
</style>

