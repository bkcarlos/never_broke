<template>
  <div class="app-container">
    <main class="main-content pb-0">
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
            'text-blue-600 bg-blue-50': isActive(tab.path), 
            'text-gray-500 hover:text-gray-700 hover:bg-gray-50': !isActive(tab.path) 
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
            'text-blue-600 bg-blue-50': isActive(tab.path), 
            'text-gray-500 hover:text-gray-700 hover:bg-gray-50': !isActive(tab.path) 
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
    label: '记支出',
    icon: 'minus-circle',
    route: '/actions'
  },
  {
    key: 'record_income',
    label: '记收入',
    icon: 'add-circle',
    route: '/actions'
  },
  {
    key: 'record_transfer',
    label: '转账',
    icon: 'transfer',
    route: '/actions'
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
  router.push(action.route)
}
</script>

<style scoped>
.app-container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.main-content {
  min-height: calc(100vh - 4rem);
}

.bottom-nav {
  z-index: 1000;
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

