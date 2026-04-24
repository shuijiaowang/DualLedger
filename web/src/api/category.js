import service from '@/utils/request.js'

export const listCategories = () =>
  service({ url: '/categories', method: 'get' })

export const createCategory = (data) =>
  service({ url: '/categories', method: 'post', data })

export const updateCategory = (id, data) =>
  service({ url: `/categories/${id}`, method: 'put', data })

export const deleteCategory = (id) =>
  service({ url: `/categories/${id}`, method: 'delete' })
