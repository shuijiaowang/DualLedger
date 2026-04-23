<template>
  <div class="record-page">
    <header class="top-nav">
      <router-link to="/home" class="back">← 返回工作台</router-link>
      <span class="title">记一笔</span>
      <span></span>
    </header>

    <main class="content">
      <el-card>
        <el-form :model="form" label-width="90px">
          <el-form-item label="类型">
            <el-radio-group v-model="form.type" @change="onTypeChange">
              <el-radio-button value="EXPENSE">支出</el-radio-button>
              <el-radio-button value="INCOME">收入</el-radio-button>
              <el-radio-button value="TRANSFER">转账</el-radio-button>
              <el-radio-button value="LOAN">借贷</el-radio-button>
              <el-radio-button value="DEPOSIT">押金</el-radio-button>
              <el-radio-button value="ADJUST">调整</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="金额">
            <el-input v-model="form.amount" placeholder="大于 0 的数字">
              <template #prefix>¥</template>
            </el-input>
          </el-form-item>

          <el-form-item label="方向" v-if="needDirectionPicker">
            <el-radio-group v-model="form.direction">
              <el-radio-button value="IN">IN（收入/还入）</el-radio-button>
              <el-radio-button value="OUT">OUT（支出/借出）</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="发生时间">
            <el-date-picker v-model="form.occur_at" type="datetime" placeholder="默认当前时间" />
          </el-form-item>

          <el-form-item label="账户">
            <el-select v-model="form.account_id" placeholder="选择账户">
              <el-option
                v-for="a in accounts"
                :key="accountId(a)"
                :value="accountId(a)"
                :label="`${a.name}（¥${a.balance}）`"
              />
            </el-select>
          </el-form-item>

          <el-form-item v-if="form.type === 'TRANSFER'" label="目标账户">
            <el-select v-model="form.to_account_id" placeholder="选择目标账户">
              <el-option
                v-for="a in accounts.filter((x) => accountId(x) !== form.account_id)"
                :key="accountId(a)"
                :value="accountId(a)"
                :label="a.name"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="分类">
            <el-select v-model="form.category_code" placeholder="选择分类" clearable filterable>
              <el-option-group v-for="p in filteredCategoryTree" :key="p.code" :label="p.name">
                <el-option :value="p.code" :label="p.name" />
                <el-option
                  v-for="child in p.children"
                  :key="child.code"
                  :value="child.code"
                  :label="'　' + child.name"
                />
              </el-option-group>
            </el-select>
          </el-form-item>

          <el-form-item v-if="['LOAN','DEPOSIT','REFUND'].includes(form.type)" label="对手方">
            <el-input v-model="form.counterparty" placeholder="如：小王 / 房东 / 女友" />
          </el-form-item>

          <el-form-item label="标题">
            <el-input v-model="form.title" placeholder="一句话描述，如：午饭 鸭腿饭" />
          </el-form-item>

          <el-form-item label="标签">
            <el-select v-model="form.tags" multiple filterable allow-create default-first-option placeholder="回车自定义，或从建议词中选">
              <el-option v-for="t in metaStore.tags" :key="t" :value="t" :label="t" />
            </el-select>
          </el-form-item>

          <el-form-item label="备注">
            <el-input v-model="form.note" type="textarea" :rows="2" />
          </el-form-item>

          <el-alert
            v-if="formErrors.length"
            type="error"
            :closable="false"
            :title="formErrors.join('；')"
            show-icon
            style="margin-bottom: 12px"
          />

          <div class="actions">
            <el-button @click="reset">重置</el-button>
            <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
          </div>
        </el-form>
      </el-card>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { listAccounts } from '@/api/account.js'
import { createTransaction } from '@/api/transaction.js'
import { useMetaStore } from '@/stores/meta.js'
import { defaultDirection, validateTxForm } from '@/utils/money.js'

const router = useRouter()
const metaStore = useMetaStore()

const accounts = ref([])
const submitting = ref(false)
const accountId = (a) => a?.id ?? a?.ID ?? null

const initialForm = () => ({
  type: 'EXPENSE',
  direction: 'OUT',
  amount: '',
  occur_at: new Date(),
  account_id: null,
  to_account_id: null,
  category_code: '',
  counterparty: '',
  title: '',
  tags: [],
  note: ''
})
const form = ref(initialForm())

const needDirectionPicker = computed(() =>
  ['LOAN', 'DEPOSIT', 'ADJUST'].includes(form.value.type)
)

const filteredCategoryTree = computed(() => {
  const kind =
    form.value.type === 'INCOME'
      ? 'INCOME'
      : form.value.type === 'EXPENSE'
      ? 'EXPENSE'
      : null
  if (!kind) return metaStore.categoryTree
  return metaStore.categoryTree.filter((p) => p.kind === kind)
})

const onTypeChange = () => {
  const d = defaultDirection(form.value.type)
  if (d) form.value.direction = d
  form.value.to_account_id = null
  form.value.category_code = ''
}

const formErrors = computed(() => validateTxForm(form.value))

const reset = () => {
  form.value = initialForm()
  if (accounts.value.length > 0) form.value.account_id = accountId(accounts.value[0])
}

const submit = async () => {
  const errs = validateTxForm(form.value)
  if (errs.length) {
    ElMessage.error(errs.join('；'))
    return
  }
  submitting.value = true
  try {
    const payload = { ...form.value }
    if (payload.occur_at instanceof Date) {
      payload.occur_at = payload.occur_at.toISOString()
    }
    if (payload.type !== 'TRANSFER') delete payload.to_account_id
    await createTransaction(payload)
    ElMessage.success('已记录')
    router.push('/ledger')
  } catch {
    /* request 拦截器已提示 */
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await metaStore.load()
  try {
    const res = await listAccounts()
    accounts.value = res?.data || []
    if (accounts.value.length > 0) form.value.account_id = accountId(accounts.value[0])
  } catch {
    /* noop */
  }
})
</script>

<style scoped>
.record-page { min-height: 100vh; background: #f5f7fb; }
.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
}
.back { color: #606266; text-decoration: none; }
.title { font-weight: 600; font-size: 16px; }
.content { max-width: 720px; margin: 24px auto; padding: 0 16px; }
.actions { text-align: right; }
</style>
