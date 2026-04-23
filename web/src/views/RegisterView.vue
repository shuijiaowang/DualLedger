<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="card-header">
        <h2>创建账号</h2>
        <p>注册成功后可使用邮箱和密码登录</p>
      </div>
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label>昵称</label>
          <input v-model="registerForm.nickname" placeholder="请输入昵称" required>
        </div>
        <div class="form-group">
          <label>邮箱</label>
          <input type="email" v-model="registerForm.email" placeholder="请输入邮箱" required>
        </div>
        <div class="form-group">
          <label>密码</label>
          <input type="password" v-model="registerForm.password" placeholder="请输入密码" required>
        </div>
        <button type="submit" class="primary-btn">注册</button>
      </form>
      <div class="switch-text">
        已有账号？
        <router-link to="/login">去登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register } from '@/api/user'

const router = useRouter()

const registerForm = ref({
  nickname: '',
  email: '',
  password: ''
})

const handleRegister = async () => {
  const { nickname, email, password } = registerForm.value
  if (!nickname || !email || !password) {
    ElMessage.warning('请完整填写注册信息')
    return
  }

  try {
    await register({ nickname, email, password })
    ElMessage.success('注册成功，请登录')
    await router.replace({ name: 'login' })
  } catch (error) {
    console.error('注册失败:', error)
    ElMessage.error('注册失败，请稍后重试')
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
