import service from '@/utils/request.js'

export const resetDevData = () =>
  service({ url: '/dev-data/reset', method: 'post' })
