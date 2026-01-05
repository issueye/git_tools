<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetStatus, StageFiles, UnstageFiles, DiscardChanges, Push, Pull, GetRemoteNames } from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

const emit = defineEmits(['refresh'])
const props = defineProps<{
  status: models.GitStatus | null
}>()

const selectedStaged = ref<string[]>([])
const selectedUnstaged = ref<string[]>([])
const selectedUntracked = ref<string[]>([])

// 远程操作相关
const isPushing = ref(false)
const isPulling = ref(false)
const remoteNames = ref<string[]>(['origin'])
const selectedRemote = ref('origin')
const operationResult = ref<{ success: boolean; message: string } | null>(null)

async function loadRemotes() {
  try {
    const remotes = await GetRemoteNames()
    remoteNames.value = remotes.length > 0 ? remotes : ['origin']
    if (!remoteNames.value.includes(selectedRemote.value)) {
      selectedRemote.value = remoteNames.value[0]
    }
  } catch (error) {
    console.error('Failed to load remotes:', error)
  }
}

async function pushToRemote() {
  if (isPushing.value) return
  isPushing.value = true
  operationResult.value = null

  try {
    await Push(selectedRemote.value)
    operationResult.value = { success: true, message: '推送成功！' }
    emit('refresh')
  } catch (error: any) {
    console.error('Push failed:', error)
    operationResult.value = { success: false, message: '推送失败: ' + (error?.message || String(error)) }
  } finally {
    isPushing.value = false
  }
}

async function pullFromRemote() {
  if (isPulling.value) return
  isPulling.value = true
  operationResult.value = null

  try {
    await Pull(selectedRemote.value, '')
    operationResult.value = { success: true, message: '拉取成功！' }
    emit('refresh')
  } catch (error: any) {
    console.error('Pull failed:', error)
    operationResult.value = { success: false, message: '拉取失败: ' + (error?.message || String(error)) }
  } finally {
    isPulling.value = false
  }
}

onMounted(() => {
  loadRemotes()
})

function toggleSelection(files: models.FileChange[], path: string, selected: string[]) {
  const index = selected.indexOf(path)
  if (index === -1) {
    selected.push(path)
  } else {
    selected.splice(index, 1)
  }
}

function toggleUntracked(path: string) {
  const index = selectedUntracked.value.indexOf(path)
  if (index === -1) {
    selectedUntracked.value.push(path)
  } else {
    selectedUntracked.value.splice(index, 1)
  }
}

async function stageSelected() {
  const files = [...selectedUnstaged.value, ...selectedUntracked.value]
  if (files.length > 0) {
    await StageFiles(files)
    selectedUnstaged.value = []
    selectedUntracked.value = []
    emit('refresh')
  }
}

async function stageAll() {
  const unstaged = props.status?.unstaged.map(f => f.path) || []
  const allFiles = [...unstaged, ...(props.status?.untracked || [])]
  if (allFiles.length > 0) {
    await StageFiles(allFiles)
    emit('refresh')
  }
}

async function unstageSelected() {
  if (selectedStaged.value.length > 0) {
    await UnstageFiles(selectedStaged.value)
    selectedStaged.value = []
    emit('refresh')
  }
}

async function unstageAll() {
  const staged = props.status?.staged.map(f => f.path) || []
  if (staged.length > 0) {
    await UnstageFiles(staged)
    emit('refresh')
  }
}

async function discardChanges(filePath: string) {
  if (confirm(`Discard changes to ${filePath}?`)) {
    await DiscardChanges(filePath)
    emit('refresh')
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'Staged': return 'text-green-500'
    case 'Modified': return 'text-yellow-500'
    case 'Deleted': return 'text-red-500'
    case 'Added': return 'text-green-400'
    case 'Renamed': return 'text-blue-400'
    default: return 'text-gray-400'
  }
}
</script>

<template>
  <div class="git-status">
    <div v-if="!status" class="no-repo">
      <p>请选择一个仓库以开始使用</p>
      <p class="hint">从侧边栏选择已有仓库或打开新仓库</p>
    </div>

    <div v-else-if="!status.isRepo" class="no-repo">
      <p>当前目录不是 git 仓库</p>
    </div>

    <div v-else class="status-container">
      <!-- Branch Info with Refresh -->
      <div class="branch-info">
        <div class="branch-left">
          <span class="branch-label">分支:</span>
          <span class="branch-name">{{ status.branch || '无分支' }}</span>
        </div>
        <button @click="$emit('refresh')" class="btn-refresh" title="刷新状态">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
          </svg>
        </button>
      </div>

      <!-- Remote Operations -->
      <div class="remote-section">
        <div class="remote-header">
          <span class="remote-label">远程:</span>
          <select v-model="selectedRemote" class="remote-select">
            <option v-for="remote in remoteNames" :key="remote" :value="remote">
              {{ remote }}
            </option>
          </select>
        </div>
        <div class="remote-actions">
          <button @click="pushToRemote" class="btn-remote" :disabled="isPushing">
            <span v-if="isPushing">推送中...</span>
            <span v-else>📤 推送</span>
          </button>
          <button @click="pullFromRemote" class="btn-remote" :disabled="isPulling">
            <span v-if="isPulling">拉取中...</span>
            <span v-else>📥 拉取</span>
          </button>
        </div>
        <div v-if="operationResult" class="operation-result" :class="{ success: operationResult.success, error: !operationResult.success }">
          {{ operationResult.message }}
        </div>
      </div>

      <div v-if="!status.hasChanges" class="no-changes">
        <p>暂无变更</p>
      </div>

      <div v-else class="changes-container">
        <!-- Staged Changes -->
        <div v-if="status.staged.length > 0" class="section staged">
          <div class="section-header">
            <h3>已暂存 ({{ status.staged.length }})</h3>
            <button @click="unstageAll" class="btn-small btn-secondary">取消全部暂存</button>
          </div>
          <div class="file-list">
            <div
              v-for="file in status.staged"
              :key="file.path"
              class="file-item"
              :class="{ selected: selectedStaged.includes(file.path) }"
              @click="toggleSelection(status.staged, file.path, selectedStaged)"
            >
              <input type="checkbox" :checked="selectedStaged.includes(file.path)" />
              <span :class="getStatusColor(file.status)" class="file-status">{{ file.status }}</span>
              <span class="file-path">{{ file.path }}</span>
            </div>
          </div>
          <button v-if="selectedStaged.length > 0" @click="unstageSelected" class="btn-action">
            取消暂存所选 ({{ selectedStaged.length }})
          </button>
        </div>

        <!-- Unstaged Changes -->
        <div v-if="status.unstaged.length > 0" class="section unstaged">
          <div class="section-header">
            <h3>未暂存 ({{ status.unstaged.length }})</h3>
            <button @click="stageAll" class="btn-small">全部暂存</button>
          </div>
          <div class="file-list">
            <div
              v-for="file in status.unstaged"
              :key="file.path"
              class="file-item"
              :class="{ selected: selectedUnstaged.includes(file.path) }"
              @click="toggleSelection(status.unstaged, file.path, selectedUnstaged)"
            >
              <input type="checkbox" :checked="selectedUnstaged.includes(file.path)" />
              <span :class="getStatusColor(file.status)" class="file-status">{{ file.status }}</span>
              <span class="file-path">{{ file.path }}</span>
              <button @click.stop="discardChanges(file.path)" class="btn-icon" title="放弃更改">✕</button>
            </div>
          </div>
          <button v-if="selectedUnstaged.length > 0" @click="stageSelected" class="btn-action">
            暂存所选 ({{ selectedUnstaged.length + selectedUntracked.length }})
          </button>
        </div>

        <!-- Untracked Files -->
        <div v-if="status.untracked.length > 0" class="section untracked">
          <div class="section-header">
            <h3>未跟踪文件 ({{ status.untracked.length }})</h3>
          </div>
          <div class="file-list">
            <div
              v-for="file in status.untracked"
              :key="file"
              class="file-item"
              :class="{ selected: selectedUntracked.includes(file) }"
              @click="toggleUntracked(file)"
            >
              <input type="checkbox" :checked="selectedUntracked.includes(file)" />
              <span class="file-status text-gray-400">未跟踪</span>
              <span class="file-path">{{ file }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.git-status {
  height: 100%;
  overflow-y: auto;
}

.loading, .no-repo, .no-changes {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #888;
  text-align: center;
  padding: 2rem;
}

.no-repo .hint {
  font-size: 0.85rem;
  color: #666;
  margin-top: 0.5rem;
}

.status-container {
  padding: 1rem;
}

.branch-info {
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  margin-bottom: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.branch-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.branch-label {
  color: #888;
  font-size: 0.9rem;
}

.branch-name {
  color: #61dafb;
  font-weight: 600;
  font-family: 'Consolas', 'Monaco', monospace;
}

.btn-refresh {
  padding: 0.4rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: transparent;
  color: #888;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-refresh:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #61dafb;
  border-color: rgba(97, 218, 251, 0.3);
}

.section {
  margin-bottom: 1.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.section-header h3 {
  margin: 0;
  font-size: 1rem;
  color: #ccc;
}

.file-list {
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  overflow: hidden;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  cursor: pointer;
  transition: background 0.2s;
}

.file-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.file-item.selected {
  background: rgba(97, 218, 251, 0.1);
}

.file-item:last-child {
  border-bottom: none;
}

.file-item input[type="checkbox"] {
  margin: 0;
  cursor: pointer;
}

.file-status {
  font-size: 0.8rem;
  min-width: 80px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.file-path {
  flex: 1;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.9rem;
  word-break: break-all;
}

.btn-small {
  padding: 0.25rem 0.75rem;
  font-size: 0.8rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-small:hover {
  background: rgba(255, 255, 255, 0.1);
}

.btn-secondary {
  border-color: rgba(255, 150, 100, 0.5);
  color: #ff9664;
}

.btn-action {
  margin-top: 0.5rem;
  padding: 0.5rem 1rem;
  border: 1px solid #61dafb;
  border-radius: 4px;
  background: rgba(97, 218, 251, 0.1);
  color: #61dafb;
  cursor: pointer;
  width: 100%;
  transition: all 0.2s;
}

.btn-action:hover {
  background: rgba(97, 218, 251, 0.2);
}

.btn-icon {
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  color: #ff6b6b;
  cursor: pointer;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.btn-icon:hover {
  opacity: 1;
}

.text-green-500 { color: #4ade80; }
.text-green-400 { color: #4ade80; }
.text-yellow-500 { color: #fbbf24; }
.text-red-500 { color: #ef4444; }
.text-blue-400 { color: #60a5fa; }
.text-gray-400 { color: #9ca3af; }

/* Remote Section */
.remote-section {
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  margin-bottom: 1rem;
}

.remote-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.remote-label {
  color: #888;
  font-size: 0.9rem;
}

.remote-select {
  flex: 1;
  padding: 0.35rem 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.2);
  color: #e5e7eb;
  font-size: 0.85rem;
}

.remote-select:focus {
  outline: none;
  border-color: #61dafb;
}

.remote-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-remote {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.btn-remote:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}

.btn-remote:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.operation-result {
  margin-top: 0.5rem;
  padding: 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  text-align: center;
}

.operation-result.success {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.operation-result.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}
</style>
