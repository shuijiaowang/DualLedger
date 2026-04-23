// 金额/交易展示工具。
// 后端返回 amount 恒 > 0 的字符串（DECIMAL(14,2)），正负由 type+direction 决定。
// 前端统一用这些函数，避免每个页面各自拼接。

/**
 * 标准化为两位小数字符串；非法输入返回 '0.00'。
 * @param {string|number|null|undefined} v
 * @returns {string}
 */
export function normalizeAmount(v) {
  if (v === null || v === undefined || v === '') return '0.00'
  const n = typeof v === 'number' ? v : Number(v)
  if (!Number.isFinite(n)) return '0.00'
  return n.toFixed(2)
}

/**
 * 将 transaction 对象转换为带符号金额。
 * 约定：EXPENSE/LOAN-OUT/DEPOSIT-OUT/ADJUST-OUT 展示 -；INCOME/REFUND/LOAN-IN/DEPOSIT-IN 展示 +；
 * TRANSFER 不带正负（内部流转）。
 * @param {{ type: string, direction: string, amount: string|number }} tx
 * @returns {string} 如 '-15.00' / '+4500.00' / '100.00'
 */
export function signedAmount(tx) {
  if (!tx) return '0.00'
  const abs = normalizeAmount(tx.amount)
  if (tx.type === 'TRANSFER') return abs
  if (tx.direction === 'IN') return '+' + abs
  if (tx.direction === 'OUT') return '-' + abs
  return abs
}

/**
 * 判断是否进入权责视图默认合计（TRANSFER/LOAN/DEPOSIT/REFUND 纯资金运输不计入）。
 * @param {string} type
 */
export function isCashflowOnly(type) {
  return ['TRANSFER', 'LOAN', 'DEPOSIT', 'REFUND'].includes(type)
}

/**
 * 表单最小校验 —— 返回错误信息数组（空数组即通过）。
 * 与后端 validateTxInput 保持一致，避免无效请求。
 * @param {object} form
 */
export function validateTxForm(form) {
  const errors = []
  if (!form) return ['表单为空']
  const amt = Number(form.amount)
  if (!Number.isFinite(amt) || amt <= 0) errors.push('金额必须 > 0')
  if (!form.type) errors.push('请选择类型')
  if (!form.account_id) errors.push('请选择账户')
  if (form.type === 'TRANSFER') {
    if (!form.to_account_id) errors.push('转账需要选择目标账户')
    if (form.to_account_id && form.to_account_id === form.account_id)
      errors.push('出账和入账账户不能相同')
  }
  return errors
}

/**
 * 根据 type 推导 direction；TRANSFER 恒为 BOTH，LOAN/DEPOSIT/ADJUST 由 UI 让用户选。
 * @param {string} type
 * @returns {string}
 */
export function defaultDirection(type) {
  switch (type) {
    case 'INCOME':
    case 'REFUND':
      return 'IN'
    case 'EXPENSE':
      return 'OUT'
    case 'TRANSFER':
      return 'BOTH'
    default:
      return ''
  }
}

/**
 * 把分类清单整理成父-子结构，便于展示/级联。
 * @param {Array<{code:string,parent_code?:string,name:string,kind:string,sort:number}>} list
 */
export function groupCategories(list) {
  if (!Array.isArray(list)) return []
  const byCode = new Map()
  for (const c of list) byCode.set(c.code, { ...c, children: [] })
  const roots = []
  for (const c of byCode.values()) {
    if (c.parent_code && byCode.has(c.parent_code)) {
      byCode.get(c.parent_code).children.push(c)
    } else {
      roots.push(c)
    }
  }
  const bySort = (a, b) => (a.sort || 0) - (b.sort || 0)
  roots.sort(bySort)
  for (const r of roots) r.children.sort(bySort)
  return roots
}
