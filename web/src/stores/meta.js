import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getCategories, getTags } from '@/api/meta.js'
import { groupCategories } from '@/utils/money.js'

// 缓存分类/标签静态数据，避免重复请求。
export const useMetaStore = defineStore('meta', () => {
  const categories = ref([])
  const categoryTree = ref([])
  const tags = ref([])
  const loaded = ref(false)

  const load = async (force = false) => {
    if (loaded.value && !force) return
    try {
      const [cats, tgs] = await Promise.all([getCategories(), getTags()])
      categories.value = cats?.data || []
      categoryTree.value = groupCategories(categories.value)
      tags.value = tgs?.data || []
      loaded.value = true
    } catch (e) {
      // 失败容忍：前端照样能渲染空分类，避免阻塞主流程
      console.error('加载元数据失败', e)
    }
  }

  return { categories, categoryTree, tags, loaded, load }
})
