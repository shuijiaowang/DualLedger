<template>
  <div class="ledger-page">
    <header class="top-nav">
      <router-link to="/home" class="back">← 返回工作台</router-link>
      <span class="title">流水</span>
      <router-link to="/record" class="new-btn">＋ 记一笔</router-link>
    </header>

    <main class="content">
      <el-card>
        <div class="filters">
          <el-radio-group v-model="viewMode" @change="reload">
            <el-radio-button value="cashflow">现金流视图</el-radio-button>
            <el-radio-button value="accrual">权责视图</el-radio-button>
          </el-radio-group>

          <el-radio-group v-model="displayMode">
            <el-radio-button value="table">表格</el-radio-button>
            <el-radio-button value="calendar">日历</el-radio-button>
          </el-radio-group>

          <el-date-picker
            v-model="range"
            type="daterange"
            value-format="YYYY-MM-DD"
            range-separator="→"
            start-placeholder="开始"
            end-placeholder="结束"
            @change="reload"
          />

          <el-checkbox v-if="viewMode === 'accrual'" v-model="includeCashOnly" @change="reload">
            全部视图（含转账/借贷/押金/退款）
          </el-checkbox>

          <el-button @click="reload">刷新</el-button>
        </div>

        <!-- 现金流视图 -->
        <div v-if="viewMode === 'cashflow'">
          <el-table v-if="displayMode === 'table'" :data="cashRows" stripe size="small">
            <el-table-column label="时间" width="170">
              <template #default="{ row }">{{ fmt(row.occur_at) }}</template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="90" />
            <el-table-column label="分类" width="180">
              <template #default="{ row }">{{ categoryLabel(row.category_code) }}</template>
            </el-table-column>
            <el-table-column prop="title" label="描述" />
            <el-table-column prop="counterparty" label="对手方" width="120" />
            <el-table-column label="账户" width="140">
              <template #default="{ row }">
                {{ accountName(row.account_id) }}
                <span v-if="row.to_account_id">
                  → {{ accountName(row.to_account_id) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="金额" width="120" align="right">
              <template #default="{ row }">
                <span :class="amountClass(row)">{{ signed(row) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button size="small" link type="danger" @click="del(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-calendar v-else v-model="calendarDate">
            <template #date-cell="{ data }">
              <div class="day-cell" @click="jumpToDay(data.day)">
                <div class="day-num">{{ data.day.slice(-2) }}</div>
                <div v-if="showDayTotal(data.day)" class="day-sum">总 {{ dayStat(data.day).total }}</div>
                <div v-if="showDayPos(data.day)" class="day-pos">+ {{ dayStat(data.day).pos }}</div>
                <div v-if="showDayNeg(data.day)" class="day-neg">- {{ dayStat(data.day).neg }}</div>
              </div>
            </template>
          </el-calendar>
          <div class="summary">
            共 {{ total }} 条 · 区间净流 {{ netCashflow }}
          </div>
        </div>

        <!-- 权责视图 -->
        <div v-else>
          <el-table v-if="displayMode === 'table'" :data="accrualRows" stripe size="small">
            <el-table-column label="时间" width="170">
              <template #default="{ row }">{{ fmt(row.accrue_at) }}</template>
            </el-table-column>
            <el-table-column label="来源" width="140">
              <template #default="{ row }">
                <el-tag :type="sourceTag(row.source)" size="small">{{ sourceLabel(row.source) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="分类" width="180">
              <template #default="{ row }">{{ categoryLabel(row.category_code) }}</template>
            </el-table-column>
            <el-table-column label="描述">
              <template #default="{ row }">{{ row.title || row.note }}</template>
            </el-table-column>
            <el-table-column label="标签" width="200">
              <template #default="{ row }">
                <el-tag v-for="t in row.tags" :key="t" size="small" class="tag-chip">{{ t }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="金额" width="110" align="right">
              <template #default="{ row }">
                <span :class="accrualAmountClass(row)">{{ accrualAmountText(row) }}</span>
              </template>
            </el-table-column>
          </el-table>
          <el-calendar v-else v-model="calendarDate">
            <template #date-cell="{ data }">
              <div class="day-cell" @click="jumpToDay(data.day)">
                <div class="day-num">{{ data.day.slice(-2) }}</div>
                <div v-if="showDayTotal(data.day)" class="day-sum">总 {{ dayStat(data.day).total }}</div>
                <div v-if="showDayPos(data.day)" class="day-pos">+ {{ dayStat(data.day).pos }}</div>
                <div v-if="showDayNeg(data.day)" class="day-neg">- {{ dayStat(data.day).neg }}</div>
              </div>
            </template>
          </el-calendar>
          <div class="summary">
            共 {{ accrualRows.length }} 条 · 合计 {{ accrualTotal }}
          </div>
        </div>
      </el-card>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listTransactions, deleteTransaction } from '@/api/transaction.js'
import { listAccounts } from '@/api/account.js'
import { accrualView } from '@/api/accrual.js'
import { signedAmount, isCashflowOnly, normalizeAmount } from '@/utils/money.js'
import { useMetaStore } from '@/stores/meta.js'

const viewMode = ref('cashflow')
const displayMode = ref('table')
const range = ref([])
const includeCashOnly = ref(false)
const calendarDate = ref(new Date())

const accounts = ref([])
const metaStore = useMetaStore()
const cashRows = ref([])
const accrualRows = ref([])
const total = ref(0)

const fmt = (s) => (s ? String(s).slice(0, 16).replace('T', ' ') : '')
const accountName = (id) => accounts.value.find((a) => a.id === id)?.name || `#${id ?? '-'}`
const categoryLabel = (code) =>
  metaStore.categories.find((c) => c.code === code)?.name || code || '-'
const signed = (row) => signedAmount(row)
const amountClass = (row) => ({
  pos: row.direction === 'IN' && row.type !== 'TRANSFER',
  neg: row.direction === 'OUT' && row.type !== 'TRANSFER'
})

const sourceLabel = (s) =>
  ({ DYNAMIC_VIRTUAL: '规则动态', ACCRUAL_REAL: '真实事件', TX_IMMEDIATE: '即买即耗' }[s] || s)
const sourceTag = (s) =>
  ({ DYNAMIC_VIRTUAL: 'info', ACCRUAL_REAL: 'success', TX_IMMEDIATE: '' }[s] || '')

const categoryKind = (code) => metaStore.categories.find((c) => c.code === code)?.kind
const accrualSignedNumber = (row) => {
  const abs = Math.abs(Number(row?.amount || 0))
  if (row?.direction === 'OUT') return -abs
  if (row?.direction === 'IN') return abs
  const kind = categoryKind(row?.category_code)
  if (kind === 'EXPENSE') return -abs
  if (kind === 'INCOME') return abs
  return Number(row?.amount || 0)
}
const accrualAmountText = (row) => {
  const signed = accrualSignedNumber(row)
  const text = normalizeAmount(Math.abs(signed))
  return signed < 0 ? `-${text}` : text
}
const accrualAmountClass = (row) => (accrualSignedNumber(row) < 0 ? 'neg' : 'pos')

const buildParams = () => {
  const params = {}
  if (range.value && range.value.length === 2) {
    params.from = range.value[0]
    params.to = range.value[1]
  }
  return params
}

const reload = async () => {
  if (viewMode.value === 'cashflow') {
    try {
      const res = await listTransactions({ ...buildParams(), limit: 200 })
      cashRows.value = res?.data?.rows || []
      total.value = res?.data?.total || 0
    } catch {
      /* noop */
    }
  } else {
    const p = buildParams()
    if (!p.from || !p.to) {
      // 权责视图必须指定区间
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), 1)
      const end = new Date(now.getFullYear(), now.getMonth() + 1, 1)
      p.from = start.toISOString().slice(0, 10)
      p.to = end.toISOString().slice(0, 10)
    }
    if (includeCashOnly.value) p.include_cashonly = 'true'
    try {
      const res = await accrualView(p)
      accrualRows.value = res?.data?.rows || []
    } catch {
      /* noop */
    }
  }
}

const netCashflow = computed(() => {
  let sum = 0
  for (const r of cashRows.value) {
    if (isCashflowOnly(r.type) && r.type !== 'REFUND' && r.type !== 'LOAN' && r.type !== 'DEPOSIT') continue
    const amt = Number(r.amount)
    if (r.direction === 'IN') sum += amt
    else if (r.direction === 'OUT') sum -= amt
  }
  return normalizeAmount(sum)
})

const accrualTotal = computed(() => {
  let sum = 0
  for (const r of accrualRows.value) sum += accrualSignedNumber(r)
  return normalizeAmount(sum)
})

const dayStats = computed(() => {
  const map = new Map()
  if (viewMode.value === 'cashflow') {
    for (const r of cashRows.value) {
      const day = String(r.occur_at || '').slice(0, 10)
      if (!day) continue
      if (!map.has(day)) map.set(day, { pos: 0, neg: 0 })
      const stat = map.get(day)
      const amt = Number(r.amount || 0)
      if (r.direction === 'IN') stat.pos += amt
      else if (r.direction === 'OUT') stat.neg += amt
    }
  } else {
    for (const r of accrualRows.value) {
      const day = String(r.accrue_at || '').slice(0, 10)
      if (!day) continue
      if (!map.has(day)) map.set(day, { pos: 0, neg: 0 })
      const stat = map.get(day)
      const amt = accrualSignedNumber(r)
      if (amt >= 0) stat.pos += amt
      else stat.neg += Math.abs(amt)
    }
  }
  return map
})

const dayStat = (day) => {
  const stat = dayStats.value.get(day) || { pos: 0, neg: 0 }
  return {
    pos: normalizeAmount(stat.pos),
    neg: normalizeAmount(stat.neg),
    total: normalizeAmount(stat.pos - stat.neg)
  }
}

const dayStatNumber = (day) => {
  const stat = dayStats.value.get(day) || { pos: 0, neg: 0 }
  return {
    pos: Number(stat.pos || 0),
    neg: Number(stat.neg || 0),
    total: Number(stat.pos || 0) - Number(stat.neg || 0)
  }
}

const showDayTotal = (day) => dayStatNumber(day).total !== 0
const showDayPos = (day) => dayStatNumber(day).pos !== 0
const showDayNeg = (day) => dayStatNumber(day).neg !== 0

const formatDateLocal = (d) => {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const nextDay = (day) => {
  const d = new Date(`${day}T00:00:00`)
  d.setDate(d.getDate() + 1)
  return formatDateLocal(d)
}

const jumpToDay = async (day) => {
  range.value = [day, nextDay(day)]
  displayMode.value = 'table'
  await reload()
}

const del = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除这条 ${row.type} ${row.amount}？`, '提示', { type: 'warning' })
    await deleteTransaction(row.id)
    ElMessage.success('已删除（如需回退余额请在工作台点重算）')
    await reload()
  } catch {
    /* cancel */
  }
}

onMounted(async () => {
  await metaStore.load()
  try {
    const res = await listAccounts({ include_archived: true })
    accounts.value = res?.data || []
  } catch {
    /* noop */
  }
  await reload()
})
</script>

<style scoped>
.ledger-page { min-height: 100vh; background: #f5f7fb; }
.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
}
.title { font-weight: 600; }
.back, .new-btn { color: #606266; text-decoration: none; }
.new-btn { color: #42b883; }
.content { max-width: 1080px; margin: 20px auto; padding: 0 16px; }
.filters {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 14px;
}
.summary { margin-top: 10px; color: #606266; text-align: right; }
.pos { color: #42b883; font-weight: 500; }
.neg { color: #f56c6c; font-weight: 500; }
.tag-chip { margin-right: 4px; }
.day-cell {
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: 2px;
  padding: 4px 6px;
  cursor: pointer;
}
.day-num {
  font-size: 12px;
  color: #606266;
}
.day-sum {
  font-size: 12px;
  font-weight: 600;
  color: #303133;
}
.day-pos {
  font-size: 11px;
  color: #42b883;
}
.day-neg {
  font-size: 11px;
  color: #f56c6c;
}
</style>
