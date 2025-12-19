<template>
  <header class="topbar">
    <div class="topbar-content">
      <!-- 左侧：页面标题 -->
      <div class="topbar-left">
        <h2 class="page-title">{{ pageTitle }}</h2>
        <p v-if="pageSubtitle" class="page-subtitle">{{ pageSubtitle }}</p>
      </div>

      <!-- 右侧：操作按钮 -->
      <div class="topbar-right">
        <!-- 金额显隐切换 -->
        <button 
          @click="toggleShowAmount" 
          class="icon-btn"
          :title="showAmount ? t('profile.hideAmount') : t('profile.showAmount')"
        >
          <Icon :name="showAmount ? 'eye' : 'eye-close'" size="20px" />
        </button>

        <!-- 通知 -->
        <button class="icon-btn" title="通知">
          <Icon name="bell" size="20px" />
          <span class="notification-badge">3</span>
        </button>

        <!-- 设置 -->
        <button @click="$router.push('/profile')" class="icon-btn" title="设置">
          <Icon name="settings" size="20px" />
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from './Icon.vue'
import { t } from '@/i18n'

const route = useRoute()
const router = useRouter()

// 金额显隐
const showAmount = ref<boolean>(true)

// 页面标题映射
const pageTitles: Record<string, { title: string; subtitle?: string }> = {
  '/': { 
    title: t('nav.home'), 
    subtitle: t('home.subtitle')
  },
  '/detail': { 
    title: t('nav.detail'),
    subtitle: t('detail.subtitle')
  },
  '/reports': { 
    title: t('nav.reports'),
    subtitle: t('reports.subtitle')
  },
  '/timeline': { 
    title: t('nav.timeline'),
    subtitle: t('timeline.subtitle')
  },
  '/profile': { 
    title: t('nav.profile'),
    subtitle: t('profile.subtitle')
  },
  '/record': { 
    title: t('record.title'),
    subtitle: t('record.subtitle')
  }
}

const pageTitle = computed(() => {
  const path = route.path
  return pageTitles[path]?.title || t('app.title')
})

const pageSubtitle = computed(() => {
  const path = route.path
  return pageTitles[path]?.subtitle
})

function toggleShowAmount() {
  showAmount.value = !showAmount.value
  localStorage.setItem('showAmount', String(showAmount.value))
  // 触发全局事件通知其他组件
  window.dispatchEvent(new CustomEvent('toggle-amount-visibility', { 
    detail: { showAmount: showAmount.value } 
  }))
}
</script>

<style scoped>
.topbar {
  height: 70px;
  background: white;
  border-bottom: 1px solid #e5e7eb;
  position: fixed;
  top: 0;
  right: 0;
  left: 260px;
  z-index: 100;
  transition: left 0.3s;
}

.topbar-content {
  height: 100%;
  padding: 0 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.topbar-left {
  flex: 1;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}

.page-subtitle {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0.25rem 0 0;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.icon-btn {
  position: relative;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: none;
  background: #f3f4f6;
  color: #4b5563;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: #e5e7eb;
  transform: translateY(-2px);
}

.notification-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 16px;
  height: 16px;
  background: #ef4444;
  color: white;
  border-radius: 50%;
  font-size: 0.625rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .topbar {
    left: 70px;
  }
}

@media (max-width: 768px) {
  .topbar {
    left: 0;
    height: 60px;
  }

  .topbar-content {
    padding: 0 1rem;
  }

  .page-title {
    font-size: 1.25rem;
  }

  .page-subtitle {
    display: none;
  }

  .icon-btn {
    width: 36px;
    height: 36px;
  }
}
</style>
