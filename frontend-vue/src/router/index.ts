import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { isAuthenticated } from '@/api/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginPage.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/HomePage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/detail',
    name: 'Detail',
    component: () => import('@/views/DetailPage.vue')
  },
  {
    path: '/actions',
    name: 'Actions',
    component: () => import('@/views/ActionsPage.vue')
  },
  {
    path: '/record',
    name: 'Record',
    component: () => import('@/views/RecordFormPage.vue')
  },
  {
    path: '/reports',
    name: 'Reports',
    component: () => import('@/views/ReportsPage.vue')
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfilePage.vue')
  },
  {
    path: '/accounts',
    name: 'AccountManagement',
    component: () => import('@/views/AccountManagementPage.vue')
  },
  {
    path: '/accounts/new',
    name: 'NewAccount',
    component: () => import('@/views/NewAccountPage.vue')
  },
  {
    path: '/settings/budget',
    name: 'BudgetSettings',
    component: () => import('@/views/BudgetSettingsPage.vue')
  },
  {
    path: '/settings/import-bills',
    name: 'ImportBills',
    component: () => import('@/views/ImportBillsPage.vue')
  },
  {
    path: '/settings/salary',
    name: 'SalarySettings',
    component: () => import('@/views/SalarySettingsPage.vue')
  },
  {
    path: '/settings/recurring-income',
    name: 'RecurringIncomeSettings',
    component: () => import('@/views/RecurringIncomeSettingsPage.vue')
  },
  {
    path: '/me/salary',
    name: 'MeSalary',
    component: () => import('@/views/MeSalaryPage.vue')
  },
  {
    path: '/transfer',
    name: 'Transfer',
    component: () => import('@/views/TransferPage.vue')
  },
  {
    path: '/cashflow',
    name: 'CashflowForecast',
    component: () => import('@/views/CashflowForecastPage.vue')
  },
  {
    path: '/timeline',
    name: 'Timeline',
    component: () => import('@/views/TimelinePage.vue')
  },
  {
    path: '/tools/salary-calculator',
    name: 'SalaryCalculator',
    component: () => import('@/views/SalaryCalculatorPage.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫：检查认证状态
router.beforeEach((to, from, next) => {
  const requiresAuth = to.meta.requiresAuth !== false // 默认需要认证
  const authenticated = isAuthenticated()

  if (requiresAuth && !authenticated) {
    // 需要认证但未登录，跳转到登录页
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authenticated) {
    // 已登录用户访问登录页，跳转到首页
    next({ name: 'Home' })
  } else {
    next()
  }
})

export default router

