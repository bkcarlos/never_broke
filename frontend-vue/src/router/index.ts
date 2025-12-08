import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/HomePage.vue')
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
    path: '/transfer',
    name: 'Transfer',
    component: () => import('@/views/TransferPage.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router

