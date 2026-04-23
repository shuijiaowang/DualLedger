import service from '@/utils/request.js'

export const login = (data) => {
  return service({
    url: '/user/login',
    method: 'post',
    data
  })
}

export const register = (data) => {
  return service({
    url: '/user/register',
    method: 'post',
    data
  })
}
