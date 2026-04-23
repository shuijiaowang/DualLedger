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

          <el-divider />
          <el-form-item label="权责设置" v-if="canConfigureAccrual">
            <el-switch v-model="form.enable_accrual" active-text="启用规则（如房租按期间分摊）" />
          </el-form-item>

          <template v-if="form.enable_accrual && canConfigureAccrual">
            <el-form-item label="资源名称">
              <el-input v-model="form.resource_name" placeholder="如：房租（2026-05）" />
            </el-form-item>

            <el-form-item label="规则类型">
              <el-select v-model="form.amortize_type">
                <el-option label="固定周期（推荐：房租）" value="FIXED_PERIOD" />
                <el-option label="按次数（如 10 次课）" value="BY_COUNT" />
                <el-option label="动态按天（预计天数）" value="DYNAMIC_BY_DAY" />
              </el-select>
            </el-form-item>

            <el-form-item v-if="isFixedPeriod" label="时间范围">
              <el-date-picker
                v-model="form.period_range"
                type="daterange"
                value-format="YYYY-MM-DD"
                range-separator="→"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
              />
            </el-form-item>

            <el-form-item v-if="isFixedPeriod && !(form.period_range && form.period_range.length === 2)" label="周期天数">
              <el-input v-model="form.amortize_days" placeholder="不选范围时手动填写天数，如 30" />
            </el-form-item>

            <el-form-item v-if="isByCount" label="总量">
              <el-input v-model="form.total_qty" placeholder="如：10（次/个）" />
            </el-form-item>

            <el-form-item v-if="isDynamicByDay" label="预计天数">
              <el-input v-model="form.expected_days" placeholder="可空；为空默认按今天" />
            </el-form-item>
          </template>

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
import { createResource } from '@/api/resource.js'
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
  note: '',
  enable_accrual: false,
  resource_name: '',
  amortize_type: 'FIXED_PERIOD',
  period_range: [],
  amortize_days: null,
  start_use_at: null,
  total_qty: null,
  expected_days: null,
  include_start_gap: false
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
  if (!canConfigureAccrual.value) form.value.enable_accrual = false
}

const canConfigureAccrual = computed(() => ['EXPENSE', 'INCOME'].includes(form.value.type))
const isFixedPeriod = computed(() => form.value.amortize_type === 'FIXED_PERIOD')
const isByCount = computed(() => form.value.amortize_type === 'BY_COUNT')
const isDynamicByDay = computed(() => form.value.amortize_type === 'DYNAMIC_BY_DAY')

const buildAccrualErrors = () => {
  const errs = []
  if (!form.value.enable_accrual) return errs
  if (!canConfigureAccrual.value) {
    errs.push('仅收入/支出支持权责规则')
    return errs
  }
  if (!form.value.resource_name?.trim()) errs.push('请填写资源名称')
  if (isFixedPeriod.value) {
    if (Array.isArray(form.value.period_range) && form.value.period_range.length === 2) {
      const [start, end] = form.value.period_range
      const s = new Date(start)
      const e = new Date(end)
      if (!(s instanceof Date) || Number.isNaN(s.getTime()) || Number.isNaN(e.getTime())) {
        errs.push('固定周期的起止日期格式不正确')
      } else if (e < s) {
        errs.push('固定周期结束日期不能早于开始日期')
      }
    } else {
      const days = Number(form.value.amortize_days)
      if (!Number.isFinite(days) || days <= 0) errs.push('固定周期请填写 > 0 的天数')
    }
  }
  if (isByCount.value) {
    const qty = Number(form.value.total_qty)
    if (!Number.isFinite(qty) || qty <= 0) errs.push('按次数请填写 > 0 的总量')
  }
  if (isDynamicByDay.value) {
    const raw = form.value.expected_days
    if (raw !== null && raw !== undefined && raw !== '') {
      const days = Number(raw)
      if (!Number.isFinite(days) || days <= 0) errs.push('动态按天预计天数若填写，必须 > 0')
    }
  }
  return errs
}
const formErrors = computed(() => [...validateTxForm(form.value), ...buildAccrualErrors()])

const getFixedPeriodDays = () => {
  if (!Array.isArray(form.value.period_range) || form.value.period_range.length !== 2) return null
  const [start, end] = form.value.period_range
  const s = new Date(start)
  const e = new Date(end)
  if (Number.isNaN(s.getTime()) || Number.isNaN(e.getTime())) return null
  const ms = e.getTime() - s.getTime()
  if (ms < 0) return null
  return Math.floor(ms / (24 * 60 * 60 * 1000)) + 1
}

const toISODateTime = (v) => {
  if (!v) return undefined
  if (v instanceof Date) return v.toISOString()
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return undefined
  return d.toISOString()
}

const buildAmortizeRule = () => {
  const rule = { type: form.value.amortize_type }
  if (isFixedPeriod.value) {
    const calcDays = getFixedPeriodDays()
    const days = calcDays ?? Number(form.value.amortize_days)
    rule.days = Math.max(1, Number(days))
    if (Array.isArray(form.value.period_range) && form.value.period_range.length === 2) {
      rule.start = form.value.period_range[0]
    }
    if (form.value.include_start_gap) rule.include_start_gap = true
  } else if (isByCount.value) {
    rule.total_qty = Number(form.value.total_qty)
  } else if (isDynamicByDay.value) {
    const raw = form.value.expected_days
    if (raw !== null && raw !== undefined && raw !== '') {
      const days = Number(raw)
      if (Number.isFinite(days) && days > 0) rule.expected_days = days
    }
  }
  return rule
}

const reset = () => {
  form.value = initialForm()
  if (accounts.value.length > 0) form.value.account_id = accountId(accounts.value[0])
}

const submit = async () => {
  const errs = formErrors.value
  if (errs.length) {
    ElMessage.error(errs.join('；'))
    return
  }
  submitting.value = true
  try {
    if (form.value.enable_accrual && canConfigureAccrual.value) {
      const purchaseAt =
        form.value.occur_at instanceof Date ? form.value.occur_at.toISOString() : form.value.occur_at
      const payload = {
        name: form.value.resource_name.trim(),
        category_code: form.value.category_code,
        unit: isByCount.value ? '次' : '天',
        total_qty: isByCount.value ? Number(form.value.total_qty) : undefined,
        total_cost: form.value.amount,
        amortize_rule: buildAmortizeRule(),
        purchase_at: purchaseAt,
        start_use_at:
          Array.isArray(form.value.period_range) && form.value.period_range.length === 2
            ? toISODateTime(form.value.period_range[0])
            : undefined,
        note: form.value.note,
        ext: form.value.tags?.length ? { tags: form.value.tags } : undefined,
        account_id: form.value.account_id,
        tx_type: form.value.type,
        tx_title: form.value.title || form.value.resource_name
      }
      await createResource(payload)
      ElMessage.success('已记录并设置权责规则')
      router.push('/ledger')
      return
    }

    const payload = { ...form.value }
    if (payload.occur_at instanceof Date) payload.occur_at = payload.occur_at.toISOString()
    if (payload.type !== 'TRANSFER') delete payload.to_account_id
    payload.ext = payload.tags?.length ? { tags: payload.tags } : undefined
    delete payload.tags
    delete payload.enable_accrual
    delete payload.resource_name
    delete payload.amortize_type
    delete payload.period_range
    delete payload.amortize_days
    delete payload.start_use_at
    delete payload.total_qty
    delete payload.expected_days
    delete payload.include_start_gap
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
