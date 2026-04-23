import service from '@/utils/request.js'

export const listAccounts = (params = {}) =>
  service({ url: '/accounts', method: 'get', params })

export const createAccount = (data) =>
  service({ url: '/accounts', method: 'post', data })

export const updateAccount = (id, data) =>
  service({ url: `/accounts/${id}`, method: 'put', data })

export const deleteAccount = (id) =>
  service({ url: `/accounts/${id}`, method: 'delete' })

export const rebuildBalance = (id) =>
  service({ url: `/accounts/${id}/rebuild-balance`, method: 'post' })
