<template>
  <div class="login-page">
    <div class="login-wrapper">
      <!-- Logo和标题 -->
      <div class="logo-section">
        <div class="logo-icon">
          <svg class="logo-svg" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h1 class="welcome-title">{{ t('auth.welcome') }}</h1>
        <p class="welcome-subtitle">{{ t('auth.subtitle') }}</p>
      </div>

      <!-- 登录表单卡片 -->
      <div class="form-card">
        <!-- 标签切换 -->
        <div class="mode-tabs">
          <button
            @click="mode = 'login'"
            :class="['mode-tab', { 'active': mode === 'login' }]"
          >
            {{ t('auth.login') }}
          </button>
          <button
            @click="mode = 'register'"
            :class="['mode-tab', { 'active': mode === 'register' }]"
          >
            {{ t('auth.register') }}
          </button>
        </div>

        <!-- 错误提示 -->
        <div v-if="errorMessage" class="error-banner">
          <svg class="error-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="error-text">{{ errorMessage }}</p>
        </div>

        <!-- 表单 -->
        <form @submit.prevent="handleSubmit" class="auth-form">
          <!-- 昵称（仅注册） -->
          <div v-if="mode === 'register'" class="form-group">
            <label class="form-label">{{ t('auth.nickname') }}</label>
            <input
              v-model="form.nickname"
              type="text"
              required
              :placeholder="t('auth.nicknamePlaceholder')"
              class="form-input"
            />
          </div>

          <!-- 邮箱 -->
          <div class="form-group">
            <label class="form-label">{{ t('auth.email') }}</label>
            <input
              v-model="form.email"
              type="email"
              required
              :placeholder="t('auth.emailPlaceholder')"
              class="form-input"
            />
          </div>

          <!-- 密码 -->
          <div class="form-group">
            <label class="form-label">{{ t('auth.password') }}</label>
            <div class="password-wrapper">
              <input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                required
                :minlength="mode === 'register' ? 8 : undefined"
                :placeholder="t('auth.passwordPlaceholder')"
                class="form-input password-input"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="password-toggle"
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
            <p v-if="mode === 'register'" class="form-hint">
              {{ t('auth.passwordHint') }}
            </p>
          </div>

          <!-- 记住我（仅登录） -->
          <div v-if="mode === 'login'" class="form-options">
            <label class="remember-me">
              <input
                v-model="rememberMe"
                type="checkbox"
                class="checkbox-input"
              />
              <span class="checkbox-label">{{ t('auth.rememberMe') }}</span>
            </label>
            <a href="#" class="forgot-link">{{ t('auth.forgotPassword') }}</a>
          </div>

          <!-- 提交按钮 -->
          <button
            type="submit"
            :disabled="loading"
            class="submit-btn"
            :class="{ 'loading': loading }"
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
        <div class="oauth-section">
          <div class="divider">
            <div class="divider-line"></div>
            <span class="divider-text">{{ t('auth.orContinueWith') }}</span>
            <div class="divider-line"></div>
          </div>

          <div class="oauth-buttons">
            <button
              @click="handleOAuthLogin('wechat')"
              class="oauth-btn"
            >
              <svg class="oauth-icon text-green-500" viewBox="0 0 24 24" fill="currentColor">
                <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-5.523 3.397-6.95 3.24-1.427 7.017-.817 9.376 1.517.11-3.885-3.785-7.199-8.584-7.199zm-2.46 5.703a.868.868 0 0 1-.869-.855c0-.471.388-.856.869-.856.482 0 .869.385.869.856a.868.868 0 0 1-.869.855zm5.53 0a.868.868 0 0 1-.868-.855c0-.471.388-.856.869-.856s.868.385.868.856a.868.868 0 0 1-.869.855zm8.305 4.39c0-2.939-2.942-5.32-6.568-5.32-3.626 0-6.569 2.381-6.569 5.32 0 2.939 2.943 5.32 6.569 5.32.645 0 1.29-.07 1.917-.21a.543.543 0 0 1 .45.061l1.197.7c.03.017.065.027.098.027.103 0 .184-.08.184-.184 0-.044-.019-.088-.03-.133l-.245-.93a.371.371 0 0 1 .133-.417c1.152-.855 1.864-2.11 1.864-3.534zm-8.42-1.294a.543.543 0 1 1-.001 1.086.543.543 0 0 1 .001-1.086zm3.26 0a.543.543 0 1 1-.001 1.086.543.543 0 0 1 .001-1.086z"/>
              </svg>
              <span class="oauth-label">{{ t('auth.wechat') }}</span>
            </button>
            <button
              @click="handleOAuthLogin('google')"
              class="oauth-btn"
            >
              <svg class="w-5 h-5" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              <span class="oauth-label">{{ t('auth.google') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 底部链接 -->
      <p class="footer-text">
        {{ t('auth.termsPrefix') }}
        <a href="#" class="footer-link">{{ t('auth.terms') }}</a>
        {{ t('auth.and') }}
        <a href="#" class="footer-link">{{ t('auth.privacy') }}</a>
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.login-wrapper {
  width: 100%;
  max-width: 480px;
  animation: fadeIn 0.5s ease-out;
}

/* Logo区域 */
.logo-section {
  text-align: center;
  margin-bottom: 2rem;
}

.logo-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  border-radius: 20px;
  margin-bottom: 1.5rem;
  box-shadow: 0 10px 25px rgba(79, 70, 229, 0.3);
}

.logo-svg {
  width: 48px;
  height: 48px;
  color: white;
}

.welcome-title {
  font-size: 2rem;
  font-weight: 700;
  color: white;
  margin-bottom: 0.5rem;
}

.welcome-subtitle {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1rem;
}

/* 表单卡片 */
.form-card {
  background: white;
  border-radius: 24px;
  padding: 2.5rem;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.2);
}

/* 模式切换标签 */
.mode-tabs {
  display: flex;
  gap: 0.5rem;
  background: #f3f4f6;
  padding: 0.375rem;
  border-radius: 12px;
  margin-bottom: 1.5rem;
}

.mode-tab {
  flex: 1;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.9375rem;
  color: #6b7280;
  background: transparent;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-tab.active {
  background: white;
  color: #4f46e5;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

/* 错误提示 */
.error-banner {
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.error-icon {
  width: 20px;
  height: 20px;
  color: #ef4444;
  flex-shrink: 0;
  margin-top: 2px;
}

.error-text {
  font-size: 0.875rem;
  color: #b91c1c;
  margin: 0;
}

/* 表单 */
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.form-input {
  width: 100%;
  padding: 0.875rem 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 1rem;
  color: #1f2937;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.1);
}

.password-wrapper {
  position: relative;
}

.password-input {
  padding-right: 3rem;
}

.password-toggle {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: transparent;
  color: #9ca3af;
  cursor: pointer;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-center;
  transition: color 0.2s;
}

.password-toggle:hover {
  color: #6b7280;
}

.form-hint {
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: #6b7280;
}

/* 表单选项 */
.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.remember-me {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.checkbox-input {
  width: 1rem;
  height: 1rem;
  border-radius: 0.25rem;
  border: 2px solid #d1d5db;
  cursor: pointer;
}

.checkbox-input:checked {
  background-color: #4f46e5;
  border-color: #4f46e5;
}

.checkbox-label {
  font-size: 0.875rem;
  color: #6b7280;
}

.forgot-link {
  font-size: 0.875rem;
  color: #4f46e5;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.forgot-link:hover {
  color: #4338ca;
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  padding: 1rem;
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(79, 70, 229, 0.4);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* OAuth区域 */
.oauth-section {
  margin-top: 1.5rem;
}

.divider {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: #e5e7eb;
}

.divider-text {
  font-size: 0.875rem;
  color: #9ca3af;
}

.oauth-buttons {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.oauth-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.oauth-btn:hover {
  background: #f9fafb;
  border-color: #d1d5db;
  transform: translateY(-2px);
}

.oauth-icon {
  width: 20px;
  height: 20px;
}

.oauth-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

/* 底部链接 */
.footer-text {
  text-align: center;
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.9);
  margin-top: 1.5rem;
}

.footer-link {
  color: white;
  text-decoration: none;
  font-weight: 600;
  transition: opacity 0.2s;
}

.footer-link:hover {
  opacity: 0.8;
}

/* 响应式设计 */
@media (max-width: 640px) {
  .login-page {
    padding: 1rem;
  }

  .form-card {
    padding: 1.5rem;
  }

  .logo-icon {
    width: 64px;
    height: 64px;
  }

  .logo-svg {
    width: 40px;
    height: 40px;
  }

  .welcome-title {
    font-size: 1.75rem;
  }

  .oauth-buttons {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 1920px) {
  .login-wrapper {
    max-width: 560px;
  }

  .logo-icon {
    width: 96px;
    height: 96px;
  }

  .logo-svg {
    width: 56px;
    height: 56px;
  }

  .welcome-title {
    font-size: 2.5rem;
  }

  .form-card {
    padding: 3rem;
  }
}

/* 动画 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

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
