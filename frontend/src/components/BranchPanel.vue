<script lang="ts" setup>
import { ref, onMounted, watch, computed } from 'vue'
import { GetBranches, CheckoutBranch, CreateBranch, DeleteBranch, MergeBranch } from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

const props = defineProps<{
  hasRepository: boolean
}>()

const emit = defineEmits(['branchChanged'])
const branches = ref<models.Branch[]>([])
const isLoading = ref(false)
const showCreateDialog = ref(false)
const showMergeDialog = ref(false)
const newBranchName = ref('')
const checkoutAfterCreate = ref(true)
const mergeSourceBranch = ref('')
const mergeLoading = ref(false)

// Context menu state
const contextMenu = ref<{ x: number; y: number; branch: models.Branch | null }>({
  x: 0,
  y: 0,
  branch: null
})

// Computed property for current branch name
const currentBranchName = computed(() => {
  const current = branches.value.find(b => b.isCurrent)
  return current ? current.name : '未知'
})

async function loadBranches() {
  isLoading.value = true
  try {
    const result = await GetBranches()
    branches.value = result || []
  } catch (error: any) {
    console.error('Failed to load branches:', error)
  } finally {
    isLoading.value = false
  }
}

async function switchToBranch(branchName: string) {
  if (branchName.startsWith('remotes/')) {
    if (!confirm('这是远程分支，是否创建本地跟踪分支？')) {
      return
    }
  }

  try {
    await CheckoutBranch(branchName)
    emit('branchChanged')
  } catch (error: any) {
    alert('切换分支失败: ' + error.message)
  }
}

function isRemoteBranch(branch: models.Branch): boolean {
  return branch.name.startsWith('remotes/')
}

function isCurrentBranch(branch: models.Branch): boolean {
  return branch.isCurrent
}

async function createBranch() {
  if (!newBranchName.value.trim()) {
    alert('请输入分支名称')
    return
  }

  try {
    await CreateBranch(newBranchName.value, checkoutAfterCreate.value)
    showCreateDialog.value = false
    newBranchName.value = ''
    await loadBranches()
    emit('branchChanged')
  } catch (error: any) {
    alert('创建分支失败: ' + error.message)
  }
}

async function deleteLocalBranch(branchName: string) {
  if (isCurrentBranch({ name: branchName, isCurrent: true })) {
    alert('不能删除当前所在分支')
    return
  }

  if (!confirm(`确定要删除本地分支 "${branchName}" 吗？`)) {
    return
  }

  try {
    await DeleteBranch(branchName, false)
    await loadBranches()
    emit('branchChanged')
  } catch (error: any) {
    console.error('Failed to delete branch:', error)
    alert('删除分支失败: ' + error.message)
  }
}

async function mergeBranch() {
  if (!mergeSourceBranch.value) {
    alert('请选择要合并的分支')
    return
  }

  mergeLoading.value = true
  try {
    await MergeBranch(mergeSourceBranch.value, false)
    showMergeDialog.value = false
    mergeSourceBranch.value = ''
    emit('branchChanged')
  } catch (error: any) {
    console.error('Failed to merge branch:', error)
    alert('合并分支失败: ' + error.message)
  } finally {
    mergeLoading.value = false
  }
}

// Context menu functions
function showContextMenu(e: MouseEvent, branch: models.Branch) {
  if (isRemoteBranch(branch)) return // No context menu for remote branches
  if (isCurrentBranch(branch)) return // No context menu for current branch

  e.preventDefault()
  contextMenu.value = {
    x: e.clientX,
    y: e.clientY,
    branch: branch
  }
}

function hideContextMenu() {
  contextMenu.value.branch = null
}

async function handleContextMenuDelete() {
  if (contextMenu.value.branch) {
    await deleteLocalBranch(contextMenu.value.branch.name)
    hideContextMenu()
  }
}

function openMergeDialog() {
  showMergeDialog.value = true
  mergeSourceBranch.value = ''
  hideContextMenu()
}

onMounted(() => {
  if (props.hasRepository) {
    loadBranches()
  }

  // Close context menu on click outside
  document.addEventListener('click', hideContextMenu)
})

watch(() => props.hasRepository, (newVal) => {
  if (newVal) {
    loadBranches()
  } else {
    branches.value = []
  }
})

defineExpose({
  loadBranches
})
</script>

<template>
  <div class="branch-panel">
    <div class="panel-header">
      <h2>分支</h2>
      <div class="header-actions">
        <button @click="openMergeDialog" class="btn-action-small" :disabled="!hasRepository" title="合并分支">
          合并
        </button>
        <button @click="loadBranches" class="btn-refresh" :disabled="isLoading" title="刷新">
          <span v-if="isLoading">⟳</span>
          <span v-else>⟳</span>
        </button>
      </div>
    </div>

    <div class="branch-list">
      <div v-if="isLoading" class="loading">加载中...</div>

      <div v-else-if="branches.length === 0" class="empty">暂无分支</div>

      <div v-else>
        <!-- Local Branches -->
        <div class="branch-group">
          <div class="group-label">本地分支</div>
          <div
            v-for="branch in branches.filter(b => !b.name.startsWith('remotes/'))"
            :key="branch.name"
            class="branch-item"
            :class="{ current: branch.isCurrent }"
            @click="switchToBranch(branch.name)"
            @contextmenu="showContextMenu($event, branch)"
          >
            <span class="branch-icon">{{ branch.isCurrent ? '●' : '○' }}</span>
            <span class="branch-name">{{ branch.name }}</span>
            <span v-if="branch.isCurrent" class="current-badge">当前</span>
          </div>
        </div>

        <!-- Remote Branches -->
        <div v-if="branches.some(b => isRemoteBranch(b))" class="branch-group">
          <div class="group-label">远程分支</div>
          <div
            v-for="branch in branches.filter(b => isRemoteBranch(b))"
            :key="branch.name"
            class="branch-item"
            @click="switchToBranch(branch.name)"
          >
            <span class="branch-icon">○</span>
            <span class="branch-name">{{ branch.name.replace('remotes/', '') }}</span>
          </div>
        </div>
      </div>
    </div>

    <button @click="showCreateDialog = true" class="btn-create">+ 新建分支</button>

    <!-- Context Menu -->
    <div
      v-if="contextMenu.branch"
      class="context-menu"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
    >
      <div class="context-item" @click="handleContextMenuDelete">删除分支</div>
    </div>

    <!-- Create Branch Dialog -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>创建新分支</h3>
          <button @click="showCreateDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>分支名称</label>
            <input
              v-model="newBranchName"
              type="text"
              placeholder="feature/xxx"
              @keyup.enter="createBranch"
            />
          </div>
          <div class="form-group">
            <label class="checkbox">
              <input type="checkbox" v-model="checkoutAfterCreate" />
              <span>创建后切换到新分支</span>
            </label>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showCreateDialog = false" class="btn-cancel">取消</button>
          <button @click="createBranch" class="btn-confirm">创建</button>
        </div>
      </div>
    </div>

    <!-- Merge Branch Dialog -->
    <div v-if="showMergeDialog" class="dialog-overlay" @click.self="showMergeDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>合并分支到当前分支</h3>
          <button @click="showMergeDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>选择要合并的分支</label>
            <select v-model="mergeSourceBranch" class="form-input">
              <option value="">请选择...</option>
              <option
                v-for="branch in branches.filter(b => !b.isCurrent && !isRemoteBranch(b))"
                :key="branch.name"
                :value="branch.name"
              >
                {{ branch.name }}
              </option>
            </select>
          </div>
          <p class="help-text">将选中的分支合并到当前分支 ({{ currentBranchName }})</p>
        </div>
        <div class="dialog-footer">
          <button @click="showMergeDialog = false" class="btn-cancel">取消</button>
          <button @click="mergeBranch" class="btn-confirm" :disabled="!mergeSourceBranch || mergeLoading">
            <span v-if="mergeLoading">合并中...</span>
            <span v-else>确认合并</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.branch-panel {
  display: flex;
  flex-direction: column;
  padding: 1rem;
  height: 100%;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.panel-header h2 {
  margin: 0;
  font-size: 1.2rem;
  color: #fff;
}

.btn-refresh {
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  color: #888;
  cursor: pointer;
  font-size: 1.2rem;
  transition: color 0.2s;
}

.btn-refresh:hover:not(:disabled) {
  color: #61dafb;
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.branch-list {
  flex: 1;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 0.5rem;
}

.loading, .empty {
  padding: 2rem;
  text-align: center;
  color: #888;
}

.branch-group {
  margin-bottom: 1rem;
}

.branch-group:last-child {
  margin-bottom: 0;
}

.group-label {
  font-size: 0.75rem;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.25rem 0.5rem;
  margin-bottom: 0.25rem;
}

.branch-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.branch-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.branch-item.current {
  background: rgba(97, 218, 251, 0.1);
}

.branch-icon {
  color: #61dafb;
  font-size: 0.8rem;
}

.branch-name {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.9rem;
  color: #e5e7eb;
}

.btn-create {
  margin-top: 1rem;
  padding: 0.5rem;
  border: 1px dashed rgba(255, 255, 255, 0.3);
  border-radius: 6px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-create:hover {
  border-color: #61dafb;
  color: #61dafb;
}

/* Dialog Styles */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  min-width: 400px;
  max-width: 90vw;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
}

.btn-close {
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  color: #888;
  cursor: pointer;
  font-size: 1.2rem;
  line-height: 1;
}

.btn-close:hover {
  color: #fff;
}

.dialog-body {
  padding: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #ccc;
  font-size: 0.9rem;
}

.form-group input[type="text"] {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.2);
  color: #e5e7eb;
  font-family: 'Consolas', 'Monaco', monospace;
}

.form-group input[type="text"]:focus {
  outline: none;
  border-color: #61dafb;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.checkbox input[type="checkbox"] {
  margin: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-cancel, .btn-confirm {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel {
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: transparent;
  color: #ccc;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.1);
}

.btn-confirm {
  border: 1px solid #61dafb;
  background: rgba(97, 218, 251, 0.1);
  color: #61dafb;
}

.btn-confirm:hover {
  background: rgba(97, 218, 251, 0.2);
}

.btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-action-small {
  padding: 0.3rem 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  font-size: 0.8rem;
  transition: all 0.2s;
}

.btn-action-small:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
}

.btn-action-small:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Context Menu */
.context-menu {
  position: fixed;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  padding: 0.25rem 0;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.context-item {
  padding: 0.5rem 1rem;
  cursor: pointer;
  color: #e5e7eb;
  font-size: 0.85rem;
  transition: background 0.2s;
}

.context-item:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

/* Current badge */
.current-badge {
  font-size: 0.65rem;
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  background: rgba(97, 218, 251, 0.2);
  color: #61dafb;
  margin-left: auto;
}

/* Help text */
.help-text {
  font-size: 0.8rem;
  color: #888;
  margin: 0.5rem 0 0 0;
}

/* Form input select */
.form-group select.form-input {
  cursor: pointer;
}

.form-group select.form-input:focus {
  outline: none;
  border-color: #61dafb;
}
</style>
