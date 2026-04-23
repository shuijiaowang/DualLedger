import service from '@/utils/request.js'

export const listResources = (params = {}) =>
  service({ url: '/resources', method: 'get', params })

export const getResource = (id) =>
  service({ url: `/resources/${id}`, method: 'get' })

export const createResource = (data) =>
  service({ url: '/resources', method: 'post', data })

export const punchResource = (id, data) =>
  service({ url: `/resources/${id}/punch`, method: 'post', data })

export const endResource = (id, data) =>
  service({ url: `/resources/${id}/end`, method: 'post', data })
