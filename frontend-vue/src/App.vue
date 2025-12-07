<template>
  <div class="app-container">
    <main class="main-content pb-0">
      <router-view />
    </main>
    
    <nav class="bottom-nav fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-lg">
      <div class="flex justify-around items-center h-16">
        <router-link 
          v-for="tab in tabs" 
          :key="tab.path"
          :to="tab.path"
          class="nav-item flex flex-col items-center justify-center flex-1 h-full transition-all duration-200"
          :class="{ 
            'text-blue-600 bg-blue-50': isActive(tab.path), 
            'text-gray-500 hover:text-gray-700 hover:bg-gray-50': !isActive(tab.path) 
          }"
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
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import Icon from './components/Icon.vue'

const route = useRoute()

const tabs = [
  { path: '/', icon: 'home', label: '首页' },
  { path: '/detail', icon: 'list', label: '明细' },
  { path: '/actions', icon: 'add', label: '操作' },
  { path: '/reports', icon: 'chart', label: '报表' },
  { path: '/profile', icon: 'user', label: '我的' }
]

const isActive = (path: string) => {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
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

