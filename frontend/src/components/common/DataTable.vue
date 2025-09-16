<template>
  <div class="data-table">
    <el-table
      :data="tableData"
      :loading="loading"
      stripe
      border
      style="width: 100%"
      :height="height"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <!-- 选择列 -->
      <el-table-column
        v-if="selectable"
        type="selection"
        width="55"
        align="center"
      />
      
      <!-- 序号列 -->
      <el-table-column
        v-if="showIndex"
        type="index"
        label="序号"
        width="60"
        align="center"
      />
      
      <!-- 动态列 -->
      <el-table-column
        v-for="column in columns"
        :key="column.prop"
        :prop="column.prop"
        :label="column.label"
        :width="column.width"
        :min-width="column.minWidth"
        :align="column.align || 'left'"
        :sortable="column.sortable"
        :formatter="column.formatter"
        :show-overflow-tooltip="column.showOverflowTooltip !== false"
      >
        <template #default="{ row, column: col, $index }">
          <!-- 自定义插槽 -->
          <slot 
            v-if="column.slot" 
            :name="column.slot" 
            :row="row" 
            :column="col" 
            :index="$index"
          />
          
          <!-- 状态标签 -->
          <el-tag
            v-else-if="column.type === 'status'"
            :type="getStatusType(row[column.prop])"
            size="small"
          >
            {{ getStatusText(row[column.prop]) }}
          </el-tag>
          
          <!-- 操作按钮 -->
          <div v-else-if="column.type === 'actions'" class="table-actions">
            <el-button
              v-for="action in column.actions"
              :key="action.name"
              :type="action.type || 'primary'"
              :size="action.size || 'small'"
              :icon="action.icon"
              :disabled="action.disabled && action.disabled(row)"
              @click="handleAction(action.name, row, $index)"
            >
              {{ action.label }}
            </el-button>
          </div>
          
          <!-- 默认显示 -->
          <span v-else>
            {{ column.formatter ? column.formatter(row, col, row[column.prop], $index) : row[column.prop] }}
          </span>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页 -->
    <div v-if="showPagination" class="table-pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="pageSizes"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface TableColumn {
  prop: string
  label: string
  width?: number
  minWidth?: number
  align?: 'left' | 'center' | 'right'
  sortable?: boolean
  formatter?: Function
  showOverflowTooltip?: boolean
  type?: 'status' | 'actions'
  slot?: string
  actions?: Array<{
    name: string
    label: string
    type?: string
    size?: string
    icon?: string
    disabled?: (row: any) => boolean
  }>
}

interface Props {
  data: any[]
  columns: TableColumn[]
  loading?: boolean
  height?: string | number
  selectable?: boolean
  showIndex?: boolean
  showPagination?: boolean
  pageSize?: number
  pageSizes?: number[]
  total?: number
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  columns: () => [],
  loading: false,
  height: 'auto',
  selectable: false,
  showIndex: false,
  showPagination: false,
  pageSize: 10,
  pageSizes: () => [10, 20, 50, 100],
  total: 0
})

const emit = defineEmits<{
  selectionChange: [selection: any[]]
  sortChange: [sort: { column: any; prop: string; order: string }]
  action: [actionName: string, row: any, index: number]
  pageChange: [page: number, size: number]
}>()

// 分页相关
const currentPage = ref(1)
const pageSize = ref(props.pageSize)

// 表格数据
const tableData = computed(() => {
  if (props.showPagination) {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return props.data.slice(start, end)
  }
  return props.data
})

// 状态类型映射
const statusTypeMap: Record<string, string> = {
  'online': 'success',
  'offline': 'danger',
  'warning': 'warning',
  'normal': 'success',
  'error': 'danger',
  'running': 'success',
  'stopped': 'info',
  'pending': 'warning'
}

// 状态文本映射
const statusTextMap: Record<string, string> = {
  'online': '在线',
  'offline': '离线',
  'warning': '警告',
  'normal': '正常',
  'error': '错误',
  'running': '运行中',
  'stopped': '已停止',
  'pending': '等待中'
}

// 获取状态类型
const getStatusType = (status: string) => {
  return statusTypeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  return statusTextMap[status] || status
}

// 处理选择变化
const handleSelectionChange = (selection: any[]) => {
  emit('selectionChange', selection)
}

// 处理排序变化
const handleSortChange = (sort: { column: any; prop: string; order: string }) => {
  emit('sortChange', sort)
}

// 处理操作按钮点击
const handleAction = (actionName: string, row: any, index: number) => {
  emit('action', actionName, row, index)
}

// 处理页面大小变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  emit('pageChange', currentPage.value, pageSize.value)
}

// 处理当前页变化
const handleCurrentChange = (page: number) => {
  currentPage.value = page
  emit('pageChange', currentPage.value, pageSize.value)
}

// 监听props变化
watch(() => props.pageSize, (newSize) => {
  pageSize.value = newSize
})
</script>

<style scoped>
.data-table {
  width: 100%;
}

.table-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.table-pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table th) {
  background-color: #fafafa;
  color: #262626;
  font-weight: 600;
}

:deep(.el-table td) {
  padding: 12px 0;
}

:deep(.el-table--border) {
  border: 1px solid #f0f0f0;
}

:deep(.el-table--border th) {
  border-right: 1px solid #f0f0f0;
}

:deep(.el-table--border td) {
  border-right: 1px solid #f0f0f0;
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background-color: #fafafa;
}

:deep(.el-pagination) {
  --el-pagination-font-size: 14px;
}
</style>
