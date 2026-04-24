<template>
  <div class="category-page">
    <header class="top-nav">
      <router-link to="/home" class="back">← 返回工作台</router-link>
      <span class="title">分类管理</span>
      <el-button type="primary" size="small" @click="openCreate">新增分类</el-button>
    </header>

    <main class="content">
      <el-card>
        <el-table
          :data="treeRows"
          size="small"
          stripe
          row-key="code"
          :tree-props="{ children: 'children' }"
          :default-expand-all="false"
        >
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="kind" label="类型" width="110" />
          <el-table-column prop="parent_name" label="父级" min-width="120" />
          <el-table-column prop="sort" label="排序" width="90" />
          <el-table-column prop="source" label="来源" width="90" />
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button link size="small" type="primary" @click="openCreateChild(row)">新增子级</el-button>
              <el-button link size="small" @click="openEdit(row)">编辑</el-button>
              <el-button link size="small" type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </main>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑分类' : '新增分类'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.kind">
            <el-option label="支出" value="EXPENSE" />
            <el-option label="收入" value="INCOME" />
            <el-option label="内部流转" value="TRANSFER" />
            <el-option label="其他" value="OTHER" />
          </el-select>
        </el-form-item>
        <el-form-item label="父级">
          <el-select v-model="form.parent_name" clearable filterable placeholder="可空">
            <el-option v-for="item in parentOptions" :key="item.id" :label="item.name" :value="item.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input v-model.number="form.sort" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="可空" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createCategory, deleteCategory, listCategories, updateCategory } from '@/api/category.js'
import { useMetaStore } from '@/stores/meta.js'

const rows = ref([])
const showDialog = ref(false)
const editingId = ref(null)
const metaStore = useMetaStore()

const emptyForm = () => ({ name: '', kind: 'EXPENSE', parent_name: '', sort: 0, icon: '' })
const form = ref(emptyForm())
const parentOptions = ref([])

const withParentName = (list) => {
  const byCode = new Map(list.map((x) => [x.code, x.name]))
  return list.map((row) => ({
    ...row,
    parent_name: row.parent_code ? byCode.get(row.parent_code) || '' : ''
  }))
}

const treeRows = computed(() => {
  const byCode = new Map()
  const roots = []
  rows.value.forEach((row) => {
    byCode.set(row.code, { ...row, children: [] })
  })
  byCode.forEach((row) => {
    if (row.parent_code && byCode.has(row.parent_code)) {
      byCode.get(row.parent_code).children.push(row)
    } else {
      roots.push(row)
    }
  })
  return roots
})

const load = async () => {
  const res = await listCategories()
  const list = res?.data || []
  rows.value = withParentName(list)
  parentOptions.value = list
}

const openCreate = () => {
  editingId.value = null
  form.value = emptyForm()
  showDialog.value = true
}

const openCreateChild = (row) => {
  editingId.value = null
  form.value = {
    ...emptyForm(),
    kind: row.kind || 'EXPENSE',
    parent_name: row.name || ''
  }
  showDialog.value = true
}

const openEdit = (row) => {
  editingId.value = row.id
  form.value = {
    name: row.name || '',
    kind: row.kind || 'EXPENSE',
    parent_name: row.parent_name || '',
    sort: Number(row.sort || 0),
    icon: row.icon || ''
  }
  showDialog.value = true
}

const save = async () => {
  if (!form.value.name || !form.value.kind) {
    ElMessage.warning('名称 / 类型必填')
    return
  }
  const payload = {
    name: form.value.name.trim(),
    kind: form.value.kind,
    parent_name: form.value.parent_name?.trim() || '',
    sort: Number(form.value.sort || 0),
    icon: form.value.icon?.trim() || '',
    source: 'user'
  }
  if (editingId.value) {
    await updateCategory(editingId.value, {
      name: payload.name,
      kind: payload.kind,
      parent_name: payload.parent_name,
      sort: payload.sort,
      icon: payload.icon
    })
  } else {
    await createCategory(payload)
  }
  showDialog.value = false
  ElMessage.success('已保存')
  await Promise.all([load(), metaStore.load(true)])
}

const remove = async (row) => {
  await ElMessageBox.confirm(`确认删除分类 ${row.name}？`, '提示', { type: 'warning' })
  await deleteCategory(row.id)
  ElMessage.success('已删除')
  await Promise.all([load(), metaStore.load(true)])
}

onMounted(load)
</script>

<style scoped>
.category-page { min-height: 100vh; background: #f5f7fb; }
.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
}
.back { color: #606266; text-decoration: none; }
.title { font-weight: 600; }
.content { max-width: 1080px; margin: 20px auto; padding: 0 16px; }
</style>
