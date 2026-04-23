import service from '@/utils/request.js'

export const listTransactions = (params = {}) =>
  service({ url: '/transactions', method: 'get', params })

export const getTransaction = (id) =>
  service({ url: `/transactions/${id}`, method: 'get' })

export const createTransaction = (data) =>
  service({ url: '/transactions', method: 'post', data })

export const deleteTransaction = (id) =>
  service({ url: `/transactions/${id}`, method: 'delete' })
