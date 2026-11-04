<template>
  <main class="login-page">
    <div class="login-wrapper">
      <div class="brand"><span>N·</span> NeverBroke</div>
      <section class="login-card">
        <p class="eyebrow">YOUR FINANCIAL WORKSPACE</p>
        <h1>{{ mode === 'login' ? '欢迎回来' : '创建你的工作空间' }}</h1>
        <p class="subtitle">整理账单，核对每一笔，了解钱的去向。</p>
        <div class="mode-tabs"><button :class="{active:mode==='login'}" @click="mode='login'">登录</button><button :class="{active:mode==='register'}" @click="mode='register'">注册</button></div>
        <p v-if="errorMessage" class="error-box" role="alert">{{ errorMessage }}</p>
        <form @submit.prevent="handleSubmit">
          <label v-if="mode==='register'" class="field">昵称<input v-model="form.nickname" autocomplete="nickname" required class="control" placeholder="如何称呼你" /></label>
          <label class="field">邮箱<input v-model="form.email" type="email" autocomplete="username" required class="control" placeholder="name@example.com" /></label>
          <div class="field"><label for="password">密码</label><div class="password-field"><input id="password" v-model="form.password" :type="showPassword?'text':'password'" :autocomplete="mode==='login'?'current-password':'new-password'" :minlength="mode==='register'?8:undefined" required class="control" placeholder="输入密码" /><button type="button" :aria-label="showPassword?'隐藏密码':'显示密码'" @click="showPassword=!showPassword">{{showPassword?'隐藏':'显示'}}</button></div><span v-if="mode==='register'" class="subtle">至少 8 个字符</span></div>
          <button class="button primary submit" :disabled="loading">{{loading?'正在处理…':mode==='login'?'登录工作台':'创建账户'}}</button>
        </form>
      </section>
      <footer class="login-footer">
        <p>个人财务工作台 · 让每一笔都有迹可循</p>
        <a v-if="icpFiling" class="filing-link" href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer">{{ icpFiling }}</a>
      </footer>
    </div>
  </main>
</template>
<script setup lang="ts">
import {ref,reactive} from 'vue'
import {useRoute,useRouter} from 'vue-router'
import {login,register,HttpError} from '@/api/auth'
const icpFiling = import.meta.env.VITE_ICP_FILING?.trim()
const router=useRouter(),route=useRoute(),mode=ref<'login'|'register'>('login')
const form=reactive({email:'',password:'',nickname:''}),showPassword=ref(false),loading=ref(false),errorMessage=ref('')
async function handleSubmit(){
  loading.value=true;errorMessage.value=''
  try{
    if(mode.value==='login')await login({email:form.email,password:form.password})
    else await register({...form,nickname:form.nickname.trim()})
    const redirect=String(route.query.redirect||'/')
    await router.push(redirect.startsWith('/')&&!redirect.startsWith('//')?redirect:'/')
  }catch(e){
    errorMessage.value=e instanceof HttpError&&e.status===401?'邮箱或密码不正确':e instanceof HttpError&&e.status===409?'该邮箱已注册，请登录':e instanceof Error?e.message:'暂时无法连接，请稍后重试'
  }finally{loading.value=false}
}
</script>
<style scoped>
.filing-link{display:inline-block;margin-top:8px;color:#737887;text-decoration:none;line-height:24px}.filing-link:hover{text-decoration:underline;color:#554bc8}.filing-link:focus-visible{outline:2px solid #554bc8;outline-offset:3px;border-radius:2px}
.login-page{min-height:100dvh;background:#f6f7fb;display:grid;place-items:center;padding:32px 20px;color:#222632}.login-wrapper{width:100%;max-width:420px}.brand{display:flex;align-items:center;gap:10px;font-size:20px;font-weight:650;margin:0 0 28px 3px}.brand span{display:grid;place-items:center;width:38px;height:38px;background:#554bc8;color:white;border-radius:10px;font-size:23px}.login-card{background:white;border:1px solid #e3e5ec;border-radius:12px;padding:32px;box-shadow:0 8px 32px #26304b06}.login-card h1{font-size:25px;letter-spacing:-.6px;margin:6px 0 10px}.mode-tabs{display:flex;border-bottom:1px solid #e6e7ee;margin:26px 0 24px}.mode-tabs button{flex:1;padding:12px;border:0;background:none;color:#7c8090;border-bottom:2px solid transparent}.mode-tabs .active{color:#554bc8;border-bottom-color:#554bc8}.field{display:flex;flex-direction:column;gap:8px;font-size:13px;margin-bottom:18px}.field .control{width:100%;height:43px;box-sizing:border-box}.password-field{position:relative}.password-field input{padding-right:55px}.password-field button{position:absolute;right:10px;top:13px;border:0;background:transparent;color:#6c7183;font-size:12px}.submit{width:100%;height:43px;margin-top:6px}.login-footer{font-size:12px;color:#9599a7;text-align:center;margin-top:24px}@media(max-width:480px){.login-card{padding:26px 22px}.login-page{padding:24px 16px}}
</style>
