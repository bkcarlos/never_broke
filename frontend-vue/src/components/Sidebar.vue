<template>
  <aside class="sidebar">
    <!-- Logo 区域 -->
    <div class="logo-section">
      <div class="logo-icon">
        <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <h1 class="logo-text">NeverBroke</h1>
    </div>

    <!-- 导航菜单 -->
    <nav class="nav-menu">
      <router-link 
        v-for="item in navItems" 
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ 'active': isActive(item.path) }"
      >
        <Icon :name="item.icon" size="20px" />
        <span>{{ item.label }}</span>
      </router-link>
    </nav>

    <!-- 快捷操作区域 -->
    <div class="quick-actions">
      <div class="section-title">{{ t('sidebar.quickActions') }}</div>
      <button 
        v-for="action in quickActions"
        :key="action.key"
        @click="handleActionClick(action)"
        class="action-btn"
      >
        <Icon :name="action.icon" size="18px" />
        <span>{{ action.label }}</span>
      </button>
    </div>

    <!-- 底部用户区域 -->
    <div class="user-section">
      <router-link to="/profile" class="user-profile">
        <div class="avatar">
          <Icon name="user" size="20px" />
        </div>
        <div class="user-info">
          <div class="user-name">{{ t('sidebar.myAccount') }}</div>
          <div class="user-email">{{ t('sidebar.settings') }}</div>
        </div>
      </router-link>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import Icon from './Icon.vue'
import { t } from '@/i18n'
import type { RadialAction } from '@/types/actions'

const route = useRoute()
const router = useRouter()

const navItems = [
  { path: '/', icon: 'home', label: t('nav.home') },
  { path: '/detail', icon: 'list', label: t('nav.detail') },
  { path: '/reports', icon: 'chart', label: t('nav.reports') },
  { path: '/timeline', icon: 'calendar', label: t('nav.timeline') }
]

const quickActions: RadialAction[] = [
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

function handleActionClick(action: RadialAction) {
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
.sidebar {
  width: 260px;
  height: 100vh;
  border-right: 1px solid #e5e7eb;
  color: #374151;
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 1000;
}

.logo-section {
  padding: 2rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
}

.logo-icon {
  width: 40px;
  height: 40px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.logo-text {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.025em;
}

.nav-menu {
  flex: 1;
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 10px;
  color: #6b7280;
  text-decoration: none;
  transition: all 0.2s;
  font-weight: 500;
}

.nav-item:hover {
  background: #f3f4f6;
  color: #374151;
  transform: translateX(4px);
}

.nav-item.active {
  background: #f3f4f6;
  color: #1f2937;
  font-weight: 600;
}

.quick-actions {
  padding: 1rem;
  border-top: 1px solid #e5e7eb;
}

.section-title {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #9ca3af;
  margin-bottom: 0.75rem;
  padding: 0 0.5rem;
}

.action-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: transparent;
  color: #4b5563;
  font-weight: 500;
  font-size: 0.875rem;
  transition: all 0.2s;
  cursor: pointer;
  margin-bottom: 0.5rem;
}

.action-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
  transform: translateY(-2px);
}

.user-section {
  padding: 1rem;
  border-top: 1px solid #e5e7eb;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 10px;
  text-decoration: none;
  color: #374151;
  transition: all 0.2s;
}

.user-profile:hover {
  background: #f3f4f6;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: #6b7280;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-weight: 600;
  font-size: 0.875rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-email {
  font-size: 0.75rem;
  color: #9ca3af;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .sidebar {
    width: 70px;
  }

  .logo-text,
  .nav-item span,
  .section-title,
  .action-btn span,
  .user-info {
    display: none;
  }

  .logo-section {
    justify-content: center;
    padding: 1.5rem 0.5rem;
  }

  .nav-item,
  .action-btn,
  .user-profile {
    justify-content: center;
    padding: 0.75rem;
  }

  .nav-item:hover {
    transform: translateX(0);
  }
}

@media (max-width: 768px) {
  .sidebar {
    display: none;
  }
}
</style>
