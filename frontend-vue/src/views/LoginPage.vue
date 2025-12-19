<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Logo和标题 -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-indigo-600 rounded-2xl mb-4">
          <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h1 class="text-3xl font-bold text-gray-900 mb-2">{{ t('auth.welcome') }}</h1>
        <p class="text-gray-600">{{ t('auth.subtitle') }}</p>
      </div>

      <!-- 登录表单卡片 -->
      <div class="bg-white rounded-2xl shadow-xl p-8">
        <!-- 标签切换 -->
        <div class="flex gap-2 mb-6 bg-gray-100 p-1 rounded-lg">
          <button
            @click="mode = 'login'"
            :class="[
              'flex-1 py-2 px-4 rounded-md font-medium transition-all',
              mode === 'login'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            ]"
          >
            {{ t('auth.login') }}
          </button>
          <button
            @click="mode = 'register'"
            :class="[
              'flex-1 py-2 px-4 rounded-md font-medium transition-all',
              mode === 'register'
                ? 'bg-white text-indigo-600 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            ]"
          >
            {{ t('auth.register') }}
          </button>
        </div>

        <!-- 错误提示 -->
        <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-start gap-2">
          <svg class="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="text-sm text-red-700">{{ errorMessage }}</p>
        </div>

        <!-- 表单 -->
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <!-- 昵称（仅注册） -->
          <div v-if="mode === 'register'">
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('auth.nickname') }}
            </label>
            <input
              v-model="form.nickname"
              type="text"
              required
              :placeholder="t('auth.nicknamePlaceholder')"
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
            />
          </div>

          <!-- 邮箱 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('auth.email') }}
            </label>
            <input
              v-model="form.email"
              type="email"
              required
              :placeholder="t('auth.emailPlaceholder')"
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
            />
          </div>

          <!-- 密码 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('auth.password') }}
            </label>
            <div class="relative">
              <input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                required
                :minlength="mode === 'register' ? 8 : undefined"
                :placeholder="t('auth.passwordPlaceholder')"
                class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all pr-12"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                <svg v-if="!showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              </button>
            </div>
            <p v-if="mode === 'register'" class="mt-1 text-xs text-gray-500">
              {{ t('auth.passwordHint') }}
            </p>
          </div>

          <!-- 记住我（仅登录） -->
          <div v-if="mode === 'login'" class="flex items-center justify-between">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="rememberMe"
                type="checkbox"
                class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
              />
              <span class="text-sm text-gray-600">{{ t('auth.rememberMe') }}</span>
            </label>
            <a href="#" class="text-sm text-indigo-600 hover:text-indigo-700">
              {{ t('auth.forgotPassword') }}
            </a>
          </div>

          <!-- 提交按钮 -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 px-4 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            <svg v-if="loading" class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span v-if="!loading">
              {{ mode === 'login' ? t('auth.loginButton') : t('auth.registerButton') }}
            </span>
            <span v-else>
              {{ mode === 'login' ? t('auth.loggingIn') : t('auth.registering') }}
            </span>
          </button>
        </form>

        <!-- OAuth登录 -->
        <div class="mt-6">
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-gray-300"></div>
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-2 bg-white text-gray-500">{{ t('auth.orContinueWith') }}</span>
            </div>
          </div>

          <div class="mt-4 grid grid-cols-2 gap-3">
            <button
              @click="handleOAuthLogin('wechat')"
              class="flex items-center justify-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <svg class="w-5 h-5 text-green-500" viewBox="0 0 24 24" fill="currentColor">
                <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-5.523 3.397-6.95 3.24-1.427 7.017-.817 9.376 1.517.11-3.885-3.785-7.199-8.584-7.199zm-2.46 5.703a.868.868 0 0 1-.869-.855c0-.471.388-.856.869-.856.482 0 .869.385.869.856a.868.868 0 0 1-.869.855zm5.53 0a.868.868 0 0 1-.868-.855c0-.471.388-.856.869-.856s.868.385.868.856a.868.868 0 0 1-.869.855zm8.305 4.39c0-2.939-2.942-5.32-6.568-5.32-3.626 0-6.569 2.381-6.569 5.32 0 2.939 2.943 5.32 6.569 5.32.645 0 1.29-.07 1.917-.21a.543.543 0 0 1 .45.061l1.197.7c.03.017.065.027.098.027.103 0 .184-.08.184-.184 0-.044-.019-.088-.03-.133l-.245-.93a.371.371 0 0 1 .133-.417c1.152-.855 1.864-2.11 1.864-3.534zm-8.42-1.294a.543.543 0 1 1-.001 1.086.543.543 0 0 1 .001-1.086zm3.26 0a.543.543 0 1 1-.001 1.086.543.543 0 0 1 .001-1.086z"/>
              </svg>
              <span class="text-sm font-medium text-gray-700">{{ t('auth.wechat') }}</span>
            </button>
            <button
              @click="handleOAuthLogin('google')"
              class="flex items-center justify-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <svg class="w-5 h-5" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              <span class="text-sm font-medium text-gray-700">{{ t('auth.google') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 底部链接 -->
      <p class="text-center text-sm text-gray-600 mt-6">
        {{ t('auth.termsPrefix') }}
        <a href="#" class="text-indigo-600 hover:text-indigo-700">{{ t('auth.terms') }}</a>
        {{ t('auth.and') }}
        <a href="#" class="text-indigo-600 hover:text-indigo-700">{{ t('auth.privacy') }}</a>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '@/i18n'
import { login, register } from '@/api/auth'

const router = useRouter()

// 表单模式
const mode = ref<'login' | 'register'>('login')

// 表单数据
const form = reactive({
  email: '',
  password: '',
  nickname: ''
})

// UI状态
const loading = ref(false)
const showPassword = ref(false)
const rememberMe = ref(false)
const errorMessage = ref('')

// 提交表单
async function handleSubmit() {
  errorMessage.value = ''
  loading.value = true

  try {
    if (mode.value === 'login') {
      await login({
        email: form.email,
        password: form.password
      })
      
      // 登录成功，跳转到首页
      router.push('/')
    } else {
      // 注册
      if (!form.nickname.trim()) {
        errorMessage.value = t('auth.nicknameRequired')
        return
      }
      
      await register({
        email: form.email,
        password: form.password,
        nickname: form.nickname
      })
      
      // 注册成功，跳转到首页
      router.push('/')
    }
  } catch (error: any) {
    console.error('Auth error:', error)
    
    // 显示错误信息
    if (error.code === 401) {
      errorMessage.value = t('auth.invalidCredentials')
    } else if (error.code === 409) {
      errorMessage.value = t('auth.emailExists')
    } else {
      errorMessage.value = error.message || t('auth.unknownError')
    }
  } finally {
    loading.value = false
  }
}

// OAuth登录
async function handleOAuthLogin(provider: 'wechat' | 'google') {
  errorMessage.value = ''
  
  try {
    // TODO: 实现OAuth登录
    console.log('OAuth login:', provider)
    errorMessage.value = t('auth.oauthNotImplemented')
  } catch (error: any) {
    console.error('OAuth error:', error)
    errorMessage.value = error.message || t('auth.unknownError')
  }
}
</script>
