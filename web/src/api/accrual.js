import service from '@/utils/request.js'

export const accrualView = (params = {}) =>
  service({ url: '/accrual-view', method: 'get', params })

export const listAccrualEntries = (params = {}) =>
  service({ url: '/accrual-entries', method: 'get', params })

export const createAccrualEntry = (data) =>
  service({ url: '/accrual-entries', method: 'post', data })

export const deleteAccrualEntry = (id) =>
  service({ url: `/accrual-entries/${id}`, method: 'delete' })
