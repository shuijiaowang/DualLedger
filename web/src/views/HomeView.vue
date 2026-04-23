<template>
  <div class="home-page">
    <header class="top-nav">
      <div class="nav-left">
        <span class="logo">DualLedger</span>
        <span class="tag">MVP</span>
      </div>
      <nav class="nav-center">
        <router-link to="/home">工作台</router-link>
        <router-link to="/record">记一笔</router-link>
        <router-link to="/ledger">流水</router-link>
      </nav>
      <div class="nav-right">
        <span class="nickname">{{ displayNickname }}</span>
        <el-button link type="primary" @click="logout">退出</el-button>
      </div>
    </header>

    <main class="content">
      <section class="card">
        <div class="card-head">
          <h3>我的账户</h3>
          <el-button type="primary" size="small" @click="showAccountDialog = true">
            新增账户
          </el-button>
        </div>
        <el-empty v-if="accounts.length === 0" description="尚无账户（注册时应自动建主账户）" />
        <el-table v-else :data="accounts" size="small" stripe>
          <el-table-column prop="name" label="名称" />
          <el-table-column label="余额" width="140">
            <template #default="{ row }">
              <span :class="balanceClass(row.balance)">{{ row.balance }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="currency" label="币种" width="90" />
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button size="small" link @click="handleRebuild(row)">重算</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="card">
        <div class="card-head">
          <h3>最近 10 条流水</h3>
          <router-link class="more" to="/ledger">查看全部 →</router-link>
        </div>
        <el-empty v-if="recentTxs.length === 0" description="还没有交易，去【记一笔】录入第一条" />
        <el-table v-else :data="recentTxs" size="small">
          <el-table-column label="时间" width="160">
            <template #default="{ row }">{{ formatDate(row.occur_at) }}</template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="90" />
          <el-table-column prop="category_code" label="分类" />
          <el-table-column prop="title" label="备注" />
          <el-table-column label="金额" width="120" align="right">
            <template #default="{ row }">
              <span :class="amountClass(row)">{{ signed(row) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </section>
    </main>

    <!-- 新增账户 -->
    <el-dialog v-model="showAccountDialog" title="新增账户" width="420px">
      <el-form :model="accountForm" label-width="70px">
        <el-form-item label="名称">
          <el-input v-model="accountForm.name" placeholder="招商银行 / 微信 / 育儿基金" />
        </el-form-item>
        <el-form-item label="初始余额">
          <el-input v-model="accountForm.balance" placeholder="0.00" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="accountForm.note" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAccountDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAccount">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listAccounts, createAccount, rebuildBalance } from '@/api/account.js'
import { listTransactions } from '@/api/transaction.js'
import { signedAmount } from '@/utils/money.js'

const userStore = useUserStore()
const displayNickname = computed(() => userStore.userInfo.nickname || '已登录用户')

const accounts = ref([])
const recentTxs = ref([])

const showAccountDialog = ref(false)
const accountForm = ref({ name: '', balance: '0.00', note: '' })

const load = async () => {
  try {
    const accRes = await listAccounts()
    accounts.value = accRes?.data || []
    const txRes = await listTransactions({ limit: 10 })
    recentTxs.value = txRes?.data?.rows || []
  } catch {
    // 错误已由 request 拦截器提示
  }
}

onMounted(load)

const formatDate = (s) => (s ? String(s).slice(0, 16).replace('T', ' ') : '')
const signed = (row) => signedAmount(row)
const amountClass = (row) => ({
  pos: row.direction === 'IN' && row.type !== 'TRANSFER',
  neg: row.direction === 'OUT' && row.type !== 'TRANSFER'
})
const balanceClass = (v) => (Number(v) < 0 ? 'neg' : '')

const saveAccount = async () => {
  if (!accountForm.value.name) {
    ElMessage.warning('请输入账户名')
    return
  }
  try {
    await createAccount({ ...accountForm.value })
    ElMessage.success('已创建')
    showAccountDialog.value = false
    accountForm.value = { name: '', balance: '0.00', note: '' }
    await load()
  } catch {
    /* noop */
  }
}

const handleRebuild = async (row) => {
  try {
    await ElMessageBox.confirm(
      `将根据历史交易重算【${row.name}】的余额，当前显示 ${row.balance}。继续？`,
      '余额重算',
      { type: 'warning' }
    )
    const res = await rebuildBalance(row.id)
    ElMessage.success(`重算完成：${res?.data?.balance ?? '-'}`)
    await load()
  } catch {
    /* cancel or error */
  }
}

const logout = async () => {
  await userStore.logout()
}
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #f5f7fb;
}
.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 0 24px;
  height: 56px;
  border-bottom: 1px solid #ebeef5;
}
.nav-left { display: flex; align-items: center; gap: 10px; }
.logo { font-weight: 700; font-size: 18px; color: #42b883; }
.tag { background: #eef7f1; color: #42b883; padding: 2px 8px; border-radius: 6px; font-size: 12px; }
.nav-center { display: flex; gap: 24px; }
.nav-center a {
  color: #606266;
  text-decoration: none;
  padding: 4px 2px;
  border-bottom: 2px solid transparent;
}
.nav-center a.router-link-active {
  color: #42b883;
  border-bottom-color: #42b883;
}
.nav-right { display: flex; align-items: center; gap: 12px; }
.nickname { color: #303133; font-weight: 500; }
.content {
  padding: 24px;
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.card {
  background: #fff;
  border-radius: 10px;
  padding: 18px 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.card-head h3 { margin: 0; }
.more { color: #409eff; text-decoration: none; font-size: 14px; }
.pos { color: #42b883; font-weight: 500; }
.neg { color: #f56c6c; font-weight: 500; }
</style>
