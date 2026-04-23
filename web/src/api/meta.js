import service from '@/utils/request.js'

// 元数据：公开读取，无需登录态
export const getCategories = () =>
  service({ url: '/meta/categories', method: 'get' })

export const getTags = () =>
  service({ url: '/meta/tags', method: 'get' })

export const getEnums = () =>
  service({ url: '/meta/enums', method: 'get' })
