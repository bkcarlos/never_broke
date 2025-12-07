<template>
  <div class="app-container">
    <main class="main-content">
      <router-view />
    </main>
    
    <nav class="bottom-nav fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-lg">
      <div class="flex justify-around items-center h-16">
        <!-- 左侧两个 tab -->
        <router-link 
          v-for="tab in leftTabs" 
          :key="tab.path"
          :to="tab.path"
          class="nav-item flex flex-col items-center justify-center flex-1 h-full transition-all duration-200"
          :class="{ 
            'text-blue-600': isActive(tab.path), 
            'text-gray-500 hover:text-gray-700': !isActive(tab.path) 
          }"
          @click="handleNavClick"
        >
          <Icon 
            :name="tab.icon" 
            size="24px"
            class="mb-1 transition-transform"
            :style="{ transform: isActive(tab.path) ? 'scale(1.1)' : 'scale(1)' }"
          />
          <span class="text-xs font-medium" :class="{ 'font-semibold': isActive(tab.path) }">
            {{ tab.label }}
          </span>
        </router-link>

        <!-- 中间的 + 按钮 -->
        <button
          class="w-14 h-14 rounded-full bg-indigo-500 text-white flex items-center justify-center text-3xl shadow-lg -mt-6 border-4 border-white transition-transform hover:scale-105"
          @click="toggleRadial"
        >
          <span :class="{ 'rotate-45 transition-transform duration-300': isRadialOpen, 'transition-transform duration-300': !isRadialOpen }">+</span>
        </button>

        <!-- 右侧两个 tab -->
        <router-link 
          v-for="tab in rightTabs" 
          :key="tab.path"
          :to="tab.path"
          class="nav-item flex flex-col items-center justify-center flex-1 h-full transition-all duration-200"
          :class="{ 
            'text-blue-600': isActive(tab.path), 
            'text-gray-500 hover:text-gray-700': !isActive(tab.path) 
          }"
          @click="handleNavClick"
        >
          <Icon 
            :name="tab.icon" 
            size="24px"
            class="mb-1 transition-transform"
            :style="{ transform: isActive(tab.path) ? 'scale(1.1)' : 'scale(1)' }"
          />
          <span class="text-xs font-medium" :class="{ 'font-semibold': isActive(tab.path) }">
            {{ tab.label }}
          </span>
        </router-link>
      </div>
    </nav>

    <!-- 圆弧菜单 -->
    <RadialActionMenu
      :visible="isRadialOpen"
      :actions="radialActions"
      @close="isRadialOpen = false"
      @select="handleSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from './components/Icon.vue'
import RadialActionMenu from './components/RadialActionMenu.vue'
import type { RadialAction } from './types/actions'

const route = useRoute()
const router = useRouter()

const isRadialOpen = ref<boolean>(false)

// 监听路由变化，关闭圆弧菜单
watch(() => route.path, () => {
  if (isRadialOpen.value) {
    isRadialOpen.value = false
  }
})

// 左侧两个 tab
const leftTabs = [
  { path: '/', icon: 'home', label: '首页' },
  { path: '/detail', icon: 'list', label: '明细' }
]

// 右侧两个 tab
const rightTabs = [
  { path: '/reports', icon: 'chart', label: '报表' },
  { path: '/profile', icon: 'user', label: '我的' }
]

const radialActions: RadialAction[] = [
  {
    key: 'record_expense',
    label: '支出',
    icon: 'minus-circle',
    route: '/record?type=expense'
  },
  {
    key: 'record_income',
    label: '收入',
    icon: 'add-circle',
    route: '/record?type=income'
  },
  {
    key: 'record_transfer',
    label: '转账',
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
  // 点击导航项时关闭圆弧菜单
  if (isRadialOpen.value) {
    isRadialOpen.value = false
  }
}

function handleSelect(action: RadialAction) {
  isRadialOpen.value = false
  // 解析路由路径和查询参数
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
.app-container {
  min-height: 800px;
  background-color: #f5f5f5;
  /* 限制最小宽度为手机合理尺寸，最大宽度并居中 */
  min-width: 320px;
  max-width: 100%;
  margin: 0 auto;
  position: relative;
}

/* 在大屏幕上居中显示，限制最大宽度 */
@media (min-width: 768px) {
  .app-container {
    max-width: 768px;
    box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
  }
}

.main-content {
  min-height: calc(100vh - 4rem);
  padding-bottom: 4rem; /* 为底部导航栏留出空间 */
  background-color: transparent; /* 让背景透明，由各个页面自己控制背景色 */
}

.bottom-nav {
  z-index: 1000;
  /* 确保导航栏始终在最上层 */
  position: fixed;
}

/* 全局样式：确保所有下拉框的z-index低于导航栏 */
select,
input[type="date"],
input[type="month"] {
  position: relative;
  z-index: 1;
}

/* 当下拉框获得焦点时，确保下拉选项的z-index仍然低于导航栏 */
select:focus,
input[type="date"]:focus,
input[type="month"]:focus {
  z-index: 999;
}

/* 在移动端，原生下拉选项可能会全屏显示，但导航栏应该始终可见 */
@media (max-width: 768px) {
  /* 确保导航栏在移动端也始终可见 */
  .bottom-nav {
    z-index: 1000 !important;
  }
}

.nav-item {
  position: relative;
}

.nav-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 3px;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6, #ec4899);
  border-radius: 0 0 3px 3px;
  transition: width 0.3s ease;
}

.nav-item.text-blue-600::before {
  width: 60%;
}
</style>

