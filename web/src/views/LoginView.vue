<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="card-header">
        <h2>欢迎登录</h2>
        <p>登录成功后进入工作台（模版页，可替换为业务路由）。</p>
      </div>
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>邮箱</label>
          <input type="email" v-model="loginForm.email" placeholder="请输入邮箱" required>
        </div>
        <div class="form-group">
          <label>密码</label>
          <input type="password" v-model="loginForm.password" placeholder="请输入密码" required>
        </div>
        <button type="submit" class="primary-btn">登录</button>
      </form>
      <div class="switch-text">
        还没有账号？
        <router-link to="/register">去注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useUserStore } from '@/stores/user.js'
import { ElMessage } from 'element-plus'

// 登录表单数据
const loginForm = ref({
  email: '',
  password: ''
})

// 获取用户存储
const userStore = useUserStore()

// 处理登录逻辑
const handleLogin = async () => {
  if (!loginForm.value.email || !loginForm.value.password) {
    ElMessage.warning('请输入邮箱和密码')
    return
  }

  const success = await userStore.loginIn({ ...loginForm.value })
  if (!success) {
    ElMessage.error('登录失败，请检查邮箱和密码')
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, #f8fffb 0%, #f2f6ff 100%);
  padding: 24px;
}

.auth-card {
  width: min(420px, 100%);
  padding: 28px;
  border-radius: 12px;
  background-color: #fff;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
}

.card-header {
  margin-bottom: 18px;
}

.card-header h2 {
  margin: 0 0 6px;
}

.card-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.form-group {
  margin-bottom: 14px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  color: #333;
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #333;
}

.form-group input:focus {
  outline: none;
  border-color: #42b883;
}

.primary-btn {
  width: 100%;
  padding: 10px 12px;
  background-color: #42b883;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.primary-btn:hover {
  background-color: #35996d;
}

.switch-text {
  margin-top: 14px;
  font-size: 14px;
  color: #666;
  text-align: right;
}
</style>