import service from '@/utils/request.js'

/** 需登录：对应后端 POST /api/example/test */
export const pingExampleApi = () => {
  return service({
    url: '/example/test',
    method: 'post',
    data: {}
  })
}
