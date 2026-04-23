<template>
  <div class="home-page">
    <header class="top-nav">
      <span class="nickname">{{ displayNickname }}</span>
      <button type="button" class="text-btn" @click="logout">退出</button>
    </header>

    <main class="content">
      <h1>工作台</h1>
      <p class="subtitle">登录注册已由模版接好；此处可替换为业务页面。</p>

      <section class="card">
        <h3>鉴权示例</h3>
        <p class="hint">调用需登录的后端 <code>POST /api/example/test</code>，用于确认 JWT 与代理正常。</p>
        <button type="button" class="primary-btn" @click="pingExample">调用示例接口</button>
        <p v-if="exampleMsg" class="result">{{ exampleMsg }}</p>
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { pingExampleApi } from '@/api/example'

const userStore = useUserStore()
const displayNickname = computed(() => userStore.userInfo.nickname || '已登录用户')
const exampleMsg = ref('')

const logout = async () => {
  await userStore.logout()
}

const pingExample = async () => {
  try {
    const res = await pingExampleApi()
    exampleMsg.value = res?.data?.message ?? JSON.stringify(res?.data ?? res)
    ElMessage.success('请求成功')
  } catch {
    exampleMsg.value = ''
  }
}
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: var(--color-background);
}

.top-nav {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  height: 56px;
  padding: 0 24px;
  border-bottom: 1px solid var(--color-border);
}

.nickname {
  color: var(--color-heading);
  font-weight: 600;
}

.text-btn {
  border: none;
  background: transparent;
  color: hsla(160, 100%, 37%, 1);
  cursor: pointer;
  font-size: 14px;
}

.content {
  padding: 32px 24px;
  max-width: 720px;
  margin: 0 auto;
}

.subtitle {
  color: #666;
  margin-bottom: 20px;
}

.card {
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 18px;
}

.hint {
  color: #666;
  font-size: 14px;
  margin: 0 0 12px;
}

.result {
  margin-top: 12px;
  color: #333;
  word-break: break-all;
}

.primary-btn {
  padding: 10px 16px;
  background-color: #42b883;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.primary-btn:hover {
  background-color: #35996d;
}
</style>
