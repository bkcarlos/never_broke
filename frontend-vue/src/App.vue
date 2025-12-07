<template>
  <div class="app-container">
    <main class="main-content pb-20">
      <router-view />
    </main>
    
    <nav class="bottom-nav fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-lg">
      <div class="flex justify-around items-center h-16">
        <router-link 
          v-for="tab in tabs" 
          :key="tab.path"
          :to="tab.path"
          class="flex flex-col items-center justify-center flex-1 h-full"
          :class="{ 'text-blue-600': isActive(tab.path), 'text-gray-500': !isActive(tab.path) }"
        >
          <span class="text-2xl mb-1">{{ tab.icon }}</span>
          <span class="text-xs">{{ tab.label }}</span>
        </router-link>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const tabs = [
  { path: '/', icon: '🏠', label: '首页' },
  { path: '/detail', icon: '📋', label: '明细' },
  { path: '/actions', icon: '➕', label: '操作' },
  { path: '/reports', icon: '📊', label: '报表' },
  { path: '/profile', icon: '👤', label: '我的' }
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
</style>

