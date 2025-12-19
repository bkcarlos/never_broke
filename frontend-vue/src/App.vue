<template>
  <!-- PC端布局 -->
  <div v-if="!isLoginPage" class="app-container-desktop">
    <!-- 侧边栏 -->
    <Sidebar class="desktop-only" />
    
    <!-- 主内容区域 -->
    <div class="main-wrapper">
      <!-- 顶部栏 -->
      <TopBar class="desktop-only" />
      
      <!-- 内容区域 -->
      <main class="content-area">
        <router-view />
      </main>
    </div>

    <!-- 移动端底部导航 -->
    <nav class="bottom-nav mobile-only">
      <div class="bottom-nav-content">
        <!-- 左侧两个 tab -->
        <router-link 
          v-for="tab in leftTabs" 
          :key="tab.path"
          :to="tab.path"
          class="nav-item"
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
        <button class="add-btn" @click="toggleRadial">
          <span :class="{ 'rotate-45': isRadialOpen }">+</span>
        </button>

        <!-- 右侧两个 tab -->
        <router-link 
          v-for="tab in rightTabs" 
          :key="tab.path"
          :to="tab.path"
          class="nav-item"
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
    </nav>

    <!-- 圆弧菜单（移动端） -->
    <RadialActionMenu
      class="mobile-only"
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

const route = useRoute()
const router = useRouter()

const isRadialOpen = ref<boolean>(false)

// 判断是否是登录页面
const isLoginPage = computed(() => route.path === '/login')

// 监听路由变化，关闭圆弧菜单
watch(() => route.path, () => {
  if (isRadialOpen.value) {
    isRadialOpen.value = false
  }
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
/* PC端布局 */
.app-container-desktop {
  min-height: 100vh;
  height: 100%;
  background-color: #f9fafb;
  display: flex;
}

.main-wrapper {
  flex: 1;
  margin-left: 260px;
  display: flex;
  flex-direction: column;
  transition: margin-left 0.3s;
}

.content-area {
  flex: 1;
  margin-top: 70px;
  padding: 2rem;
  overflow-y: auto;
  background-color: #f9fafb;
}

.login-container {
  width: 100%;
  min-height: 100vh;
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 移动端底部导航 */
.bottom-nav {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  border-top: 1px solid #e5e7eb;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  z-index: 1000;
  height: 64px;
}

.bottom-nav-content {
  display: flex;
  justify-content: space-around;
  align-items: center;
  height: 100%;
  padding: 0 1rem;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  height: 100%;
  color: #6b7280;
  text-decoration: none;
  transition: all 0.2s;
  position: relative;
}

.nav-item.active {
  color: #1f2937;
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
  background: #6b7280;
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
  color: #6b7280;
  border: 2px solid #e5e7eb;
  background: transparent;
  font-size: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: -24px;
  cursor: pointer;
  transition: all 0.2s;
}

.add-btn:hover {
  transform: scale(1.05);
  border-color: #9ca3af;
}

.add-btn span {
  transition: transform 0.3s;
}

.add-btn span.rotate-45 {
  transform: rotate(45deg);
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .main-wrapper {
    margin-left: 70px;
  }
}

@media (max-width: 768px) {
  .desktop-only {
    display: none !important;
  }

  .mobile-only {
    display: block !important;
  }

  .main-wrapper {
    margin-left: 0;
  }

  .content-area {
    margin-top: 0;
    padding: 1rem;
    padding-bottom: 80px;
  }

  .bottom-nav {
    display: block;
  }
}

@media (min-width: 769px) {
  .mobile-only {
    display: none !important;
  }

  .desktop-only {
    display: block !important;
  }
}

/* 全局样式：确保下拉框不影响导航 */
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

