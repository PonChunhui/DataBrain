<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">菜单管理</h1>
      <div class="page-actions">
        <el-button type="primary" @click="showDialog">
          <el-icon><Plus /></el-icon>
          新增菜单
        </el-button>
      </div>
    </div>

    <el-card class="content-card" shadow="hover">
      <el-table :data="menus" style="width: 100%" row-key="id">
        <el-table-column prop="id" label="ID" width="60" align="center"></el-table-column>
        <el-table-column prop="name" label="菜单名称" min-width="120"></el-table-column>
        <el-table-column prop="path" label="路径" min-width="140"></el-table-column>
        <el-table-column prop="icon" label="图标" width="100"></el-table-column>
        <el-table-column prop="sort" label="排序" width="60" align="center"></el-table-column>
        <el-table-column prop="component" label="组件" min-width="140"></el-table-column>
        <el-table-column label="显示" width="70" align="center">
          <template #default="scope">
            <el-tag v-if="scope.row.is_show" type="success" size="small">显示</el-tag>
            <el-tag v-else type="info" size="small">隐藏</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="editMenu(scope.row)">编辑</el-button>
            <el-button type="success" size="small" link @click="showButtonsDialog(scope.row)">按钮权限</el-button>
            <el-button type="danger" size="small" link @click="deleteMenu(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="menuForm.id ? '编辑菜单' : '新增菜单'" width="560px">
      <el-form :model="menuForm" label-width="90px">
        <el-form-item label="菜单名称">
          <el-input v-model="menuForm.name" placeholder="请输入菜单名称"></el-input>
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="menuForm.path" placeholder="/path"></el-input>
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="menuForm.icon" placeholder="Menu"></el-input>
        </el-form-item>
        <el-form-item label="组件">
          <el-input v-model="menuForm.component" placeholder="views/Example"></el-input>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="menuForm.sort" :min="0" style="width: 120px"></el-input-number>
        </el-form-item>
        <el-form-item label="父级菜单">
          <el-select v-model="menuForm.parent_id" placeholder="请选择父级菜单" clearable style="width: 100%">
            <el-option label="顶级菜单" :value="0"></el-option>
            <el-option v-for="menu in parentMenus" :key="menu.id" :label="menu.name" :value="menu.id"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="是否显示">
          <el-switch v-model="menuForm.is_show"></el-switch>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveMenu">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="buttonsDialogVisible" title="按钮权限配置" width="800px">
      <div style="margin-bottom: 16px">
        <el-button type="primary" size="small" @click="showButtonForm">
          <el-icon><Plus /></el-icon>
          新增按钮权限
        </el-button>
      </div>
      <el-table :data="buttons" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" align="center"></el-table-column>
        <el-table-column prop="code" label="权限代码" min-width="140"></el-table-column>
        <el-table-column prop="name" label="按钮名称" min-width="120"></el-table-column>
        <el-table-column prop="description" label="描述" min-width="180"></el-table-column>
        <el-table-column label="操作" width="120" align="center">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="editButton(scope.row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="deleteButton(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog v-model="buttonFormVisible" :title="buttonForm.id ? '编辑按钮权限' : '新增按钮权限'" width="500px" append-to-body>
        <el-form :model="buttonForm" label-width="90px">
          <el-form-item label="权限代码">
            <el-input v-model="buttonForm.code" placeholder="user:create"></el-input>
          </el-form-item>
          <el-form-item label="按钮名称">
            <el-input v-model="buttonForm.name" placeholder="新增用户"></el-input>
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="buttonForm.description" placeholder="权限描述"></el-input>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="buttonFormVisible = false">取消</el-button>
          <el-button type="primary" @click="saveButton">保存</el-button>
        </template>
      </el-dialog>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '../utils/request'

const menus = ref([])
const parentMenus = ref([])
const dialogVisible = ref(false)
const menuForm = ref({
  name: '',
  path: '',
  icon: '',
  sort: 0,
  parent_id: 0,
  is_show: true,
  component: ''
})

const buttonsDialogVisible = ref(false)
const buttonFormVisible = ref(false)
const currentMenuId = ref(0)
const buttons = ref([])
const buttonForm = ref({
  menu_id: 0,
  code: '',
  name: '',
  description: ''
})

const fetchMenus = async () => {
  try {
    const res = await request.get('/menu')
    menus.value = res.data || []
    parentMenus.value = res.data.filter(m => m.parent_id === 0) || []
  } catch (error) {
    ElMessage.error('获取菜单列表失败')
  }
}

const showDialog = () => {
  menuForm.value = { name: '', path: '', icon: '', sort: 0, parent_id: 0, is_show: true, component: '' }
  dialogVisible.value = true
}

const editMenu = (menu) => {
  menuForm.value = { ...menu }
  dialogVisible.value = true
}

const saveMenu = async () => {
  try {
    if (menuForm.value.id) {
      await request.put(`/menu/${menuForm.value.id}`, menuForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/menu', menuForm.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchMenus()
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '操作失败')
  }
}

const deleteMenu = async (id) => {
  try {
    await request.delete(`/menu/${id}`)
    ElMessage.success('删除成功')
    fetchMenus()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const showButtonsDialog = async (menu) => {
  currentMenuId.value = menu.id
  await fetchButtons(menu.id)
  buttonsDialogVisible.value = true
}

const fetchButtons = async (menuId) => {
  try {
    const res = await request.get(`/menu-button/menu/${menuId}`)
    buttons.value = res.data || []
  } catch (error) {
    ElMessage.error('获取菜单按钮失败')
  }
}

const showButtonForm = () => {
  buttonForm.value = { menu_id: currentMenuId.value, code: '', name: '', description: '' }
  buttonFormVisible.value = true
}

const editButton = (button) => {
  buttonForm.value = { ...button, menu_id: currentMenuId.value }
  buttonFormVisible.value = true
}

const saveButton = async () => {
  try {
    if (buttonForm.value.id) {
      await request.put(`/menu-button/${buttonForm.value.id}`, buttonForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/menu-button', buttonForm.value)
      ElMessage.success('创建成功')
    }
    buttonFormVisible.value = false
    await fetchButtons(currentMenuId.value)
  } catch (error) {
    ElMessage.error(error.response?.data?.msg || '操作失败')
  }
}

const deleteButton = async (id) => {
  try {
    await request.delete(`/menu-button/${id}`)
    ElMessage.success('删除成功')
    await fetchButtons(currentMenuId.value)
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(fetchMenus)
</script>

<style scoped>
.el-table {
  --el-table-border-color: #e4e7ed;
}

.el-table th.el-table__cell {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

.el-dialog .el-form-item {
  margin-bottom: 20px;
}
</style>