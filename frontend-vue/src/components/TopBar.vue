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
          @click="userSettingsStore.toggleShowAmount" 
          class="icon-btn"
          :title="userSettingsStore.showAmount ? t('profile.hideAmount') : t('profile.showAmount')"
        >
          <Icon :name="userSettingsStore.showAmount ? 'eye' : 'eye-close'" size="20px" />
        </button>

        <!-- 通知 - 暂未实现，已隐藏 -->
        <!-- <button class="icon-btn" title="通知">
          <Icon name="bell" size="20px" />
          <span class="notification-badge">3</span>
        </button> -->

        <!-- 设置 -->
        <button @click="$router.push('/profile')" class="icon-btn" title="设置">
          <Icon name="settings" size="20px" />
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from './Icon.vue'
import { t } from '@/i18n'
import { useUserSettingsStore } from '@/stores/userSettings'

const route = useRoute()
const router = useRouter()
const userSettingsStore = useUserSettingsStore()

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
</script>

<style scoped>
.topbar {
  height: 70px;
  border-bottom: 1px solid #e5e7eb;
  background: white;
  position: sticky;
  top: 0;
  z-index: 90;
  transition: all 0.3s ease;
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
  border: 1px solid #e5e7eb;
  background: transparent;
  color: #4b5563;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  border-color: #9ca3af;
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

/* ============================================
   响应式设计
   ============================================ */

/* 平板 (768px - 1023px) */
@media (min-width: 768px) and (max-width: 1023px) {
  .topbar {
    height: 64px;
  }

  .topbar-content {
    padding: 0 1.5rem;
  }

  .page-title {
    font-size: 1.375rem;
  }

  .page-subtitle {
    font-size: 0.8125rem;
  }

  .icon-btn {
    width: 38px;
    height: 38px;
  }
}

/* 桌面端 (1024px - 1919px) */
@media (min-width: 1024px) and (max-width: 1919px) {
  .topbar {
    height: 70px;
  }
}

/* 2K显示器 (1920px - 2559px) */
@media (min-width: 1920px) and (max-width: 2559px) {
  .topbar {
    height: 80px;
  }

  .topbar-content {
    padding: 0 3rem;
  }

  .page-title {
    font-size: 1.75rem;
  }

  .page-subtitle {
    font-size: 1rem;
  }

  .icon-btn {
    width: 44px;
    height: 44px;
  }
}

/* 4K显示器 (≥ 2560px) */
@media (min-width: 2560px) {
  .topbar {
    height: 90px;
  }

  .topbar-content {
    padding: 0 4rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .page-subtitle {
    font-size: 1.125rem;
  }

  .icon-btn {
    width: 48px;
    height: 48px;
  }

  .notification-badge {
    width: 18px;
    height: 18px;
    font-size: 0.6875rem;
  }
}

/* 移动端 (< 768px) - TopBar 在移动端由 App.vue 中的 mobile-topbar 替代 */
@media (max-width: 767px) {
  .topbar {
    display: none;
  }
}
</style>
