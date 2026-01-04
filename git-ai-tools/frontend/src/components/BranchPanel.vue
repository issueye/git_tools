<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetBranches, CheckoutBranch, CreateBranch } from '/wailsjs/go/main/App'
import type { services } from '/wailsjs/go/models'

const emit = defineEmits(['branchChanged'])
const branches = ref<services.Branch[]>([])
const isLoading = ref(false)
const showCreateDialog = ref(false)
const newBranchName = ref('')
const checkoutAfterCreate = ref(true)

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
    if (!confirm('This is a remote branch. Do you want to create a local tracking branch?')) {
      return
    }
  }

  try {
    await CheckoutBranch(branchName)
    emit('branchChanged')
  } catch (error: any) {
    alert('Failed to checkout branch: ' + error.message)
  }
}

function isRemoteBranch(branch: services.Branch): boolean {
  return branch.name.startsWith('remotes/')
}

async function createBranch() {
  if (!newBranchName.value.trim()) {
    alert('Please enter a branch name')
    return
  }

  try {
    await CreateBranch(newBranchName.value, checkoutAfterCreate.value)
    showCreateDialog.value = false
    newBranchName.value = ''
    await loadBranches()
    emit('branchChanged')
  } catch (error: any) {
    alert('Failed to create branch: ' + error.message)
  }
}

onMounted(() => {
  loadBranches()
})

defineExpose({
  loadBranches
})
</script>

<template>
  <div class="branch-panel">
    <div class="panel-header">
      <h2>Branches</h2>
      <button @click="loadBranches" class="btn-refresh" :disabled="isLoading" title="Refresh">
        <span v-if="isLoading">⟳</span>
        <span v-else>⟳</span>
      </button>
    </div>

    <div class="branch-list">
      <div v-if="isLoading" class="loading">Loading branches...</div>

      <div v-else-if="branches.length === 0" class="empty">No branches found</div>

      <div v-else>
        <!-- Local Branches -->
        <div class="branch-group">
          <div class="group-label">Local</div>
          <div
            v-for="branch in branches.filter(b => !b.name.startsWith('remotes/'))"
            :key="branch.name"
            class="branch-item"
            :class="{ current: branch.isCurrent }"
            @click="switchToBranch(branch.name)"
          >
            <span class="branch-icon">{{ branch.isCurrent ? '●' : '○' }}</span>
            <span class="branch-name">{{ branch.name }}</span>
          </div>
        </div>

        <!-- Remote Branches -->
        <div v-if="branches.some(b => isRemoteBranch(b))" class="branch-group">
          <div class="group-label">Remote</div>
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

    <button @click="showCreateDialog = true" class="btn-create">+ New Branch</button>

    <!-- Create Branch Dialog -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>Create New Branch</h3>
          <button @click="showCreateDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>Branch Name</label>
            <input
              v-model="newBranchName"
              type="text"
              placeholder="feature/my-new-feature"
              @keyup.enter="createBranch"
            />
          </div>
          <div class="form-group">
            <label class="checkbox">
              <input type="checkbox" v-model="checkoutAfterCreate" />
              <span>Checkout after creating</span>
            </label>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showCreateDialog = false" class="btn-cancel">Cancel</button>
          <button @click="createBranch" class="btn-confirm">Create</button>
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
</style>
