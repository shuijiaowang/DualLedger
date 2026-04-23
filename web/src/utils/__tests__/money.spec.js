import { describe, it, expect } from 'vitest'
import {
  normalizeAmount,
  signedAmount,
  isCashflowOnly,
  validateTxForm,
  defaultDirection,
  groupCategories
} from '@/utils/money.js'

describe('normalizeAmount', () => {
  it('保留两位小数并兼容字符串/数字/非法输入', () => {
    expect(normalizeAmount('15')).toBe('15.00')
    expect(normalizeAmount(15.5)).toBe('15.50')
    expect(normalizeAmount('abc')).toBe('0.00')
    expect(normalizeAmount(null)).toBe('0.00')
    expect(normalizeAmount(undefined)).toBe('0.00')
    expect(normalizeAmount('')).toBe('0.00')
  })
})

describe('signedAmount', () => {
  it('EXPENSE/OUT 负号', () => {
    expect(signedAmount({ type: 'EXPENSE', direction: 'OUT', amount: '15.00' })).toBe('-15.00')
  })
  it('INCOME/IN 正号', () => {
    expect(signedAmount({ type: 'INCOME', direction: 'IN', amount: '4500' })).toBe('+4500.00')
  })
  it('TRANSFER 不带符号', () => {
    expect(signedAmount({ type: 'TRANSFER', direction: 'BOTH', amount: '100' })).toBe('100.00')
  })
  it('空输入', () => {
    expect(signedAmount(null)).toBe('0.00')
  })
})

describe('isCashflowOnly', () => {
  it('转账/借钱/押金/退款不计入权责视图合计', () => {
    expect(isCashflowOnly('TRANSFER')).toBe(true)
    expect(isCashflowOnly('LOAN')).toBe(true)
    expect(isCashflowOnly('DEPOSIT')).toBe(true)
    expect(isCashflowOnly('REFUND')).toBe(true)
    expect(isCashflowOnly('EXPENSE')).toBe(false)
    expect(isCashflowOnly('INCOME')).toBe(false)
  })
})

describe('validateTxForm', () => {
  const valid = {
    type: 'EXPENSE',
    amount: 15,
    account_id: 1
  }
  it('合法表单 errors 为空', () => {
    expect(validateTxForm(valid)).toEqual([])
  })
  it('amount <= 0 报错（对齐后端 §15.1）', () => {
    expect(validateTxForm({ ...valid, amount: 0 })).toContain('金额必须 > 0')
    expect(validateTxForm({ ...valid, amount: -1 })).toContain('金额必须 > 0')
  })
  it('缺少 type', () => {
    expect(validateTxForm({ ...valid, type: '' })).toContain('请选择类型')
  })
  it('缺少账户', () => {
    expect(validateTxForm({ ...valid, account_id: null })).toContain('请选择账户')
  })
  it('TRANSFER 必须有目标账户且不同', () => {
    expect(
      validateTxForm({ type: 'TRANSFER', amount: 100, account_id: 1 })
    ).toContain('转账需要选择目标账户')
    expect(
      validateTxForm({ type: 'TRANSFER', amount: 100, account_id: 1, to_account_id: 1 })
    ).toContain('出账和入账账户不能相同')
    expect(
      validateTxForm({ type: 'TRANSFER', amount: 100, account_id: 1, to_account_id: 2 })
    ).toEqual([])
  })
  it('空表单', () => {
    expect(validateTxForm(null)).toEqual(['表单为空'])
  })
})

describe('defaultDirection', () => {
  it('按 type 推导', () => {
    expect(defaultDirection('INCOME')).toBe('IN')
    expect(defaultDirection('EXPENSE')).toBe('OUT')
    expect(defaultDirection('TRANSFER')).toBe('BOTH')
    expect(defaultDirection('REFUND')).toBe('IN')
    expect(defaultDirection('LOAN')).toBe('')
  })
})

describe('groupCategories', () => {
  it('按 parent_code 组装为父-子结构', () => {
    const list = [
      { code: 'food', name: '餐饮', kind: 'EXPENSE', sort: 10 },
      { code: 'food.lunch', parent_code: 'food', name: '午餐', kind: 'EXPENSE', sort: 12 },
      { code: 'food.breakfast', parent_code: 'food', name: '早餐', kind: 'EXPENSE', sort: 11 },
      { code: 'income.salary', name: '工资', kind: 'INCOME', sort: 110 }
    ]
    const tree = groupCategories(list)
    expect(tree.map((r) => r.code)).toEqual(['food', 'income.salary'])
    const foodChildren = tree.find((r) => r.code === 'food').children.map((c) => c.code)
    expect(foodChildren).toEqual(['food.breakfast', 'food.lunch'])
  })
  it('非数组 → 空数组', () => {
    expect(groupCategories(null)).toEqual([])
  })
})
