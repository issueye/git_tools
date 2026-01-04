<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetStatus, StageFiles, UnstageFiles, DiscardChanges } from '/wailsjs/go/main/App'
import type { services } from '/wailsjs/go/models'

const emit = defineEmits(['refresh'])
const props = defineProps<{
  status: services.GitStatus | null
}>()

const selectedStaged = ref<string[]>([])
const selectedUnstaged = ref<string[]>([])
const selectedUntracked = ref<string[]>([])

function toggleSelection(files: services.FileChange[], path: string, selected: string[]) {
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
    <div v-if="!status" class="loading">Loading...</div>

    <div v-else-if="!status.isRepo" class="no-repo">
      <p>Not a git repository</p>
    </div>

    <div v-else class="status-container">
      <!-- Branch Info -->
      <div class="branch-info">
        <span class="branch-label">Branch:</span>
        <span class="branch-name">{{ status.branch || 'No branch' }}</span>
      </div>

      <div v-if="!status.hasChanges" class="no-changes">
        <p>No changes</p>
      </div>

      <div v-else class="changes-container">
        <!-- Staged Changes -->
        <div v-if="status.staged.length > 0" class="section staged">
          <div class="section-header">
            <h3>Staged Changes ({{ status.staged.length }})</h3>
            <button @click="unstageAll" class="btn-small btn-secondary">Unstage All</button>
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
            Unstage Selected ({{ selectedStaged.length }})
          </button>
        </div>

        <!-- Unstaged Changes -->
        <div v-if="status.unstaged.length > 0" class="section unstaged">
          <div class="section-header">
            <h3>Unstaged Changes ({{ status.unstaged.length }})</h3>
            <button @click="stageAll" class="btn-small">Stage All</button>
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
              <button @click.stop="discardChanges(file.path)" class="btn-icon" title="Discard changes">✕</button>
            </div>
          </div>
          <button v-if="selectedUnstaged.length > 0" @click="stageSelected" class="btn-action">
            Stage Selected ({{ selectedUnstaged.length + selectedUntracked.length }})
          </button>
        </div>

        <!-- Untracked Files -->
        <div v-if="status.untracked.length > 0" class="section untracked">
          <div class="section-header">
            <h3>Untracked Files ({{ status.untracked.length }})</h3>
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
              <span class="file-status text-gray-400">Untracked</span>
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
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #888;
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
</style>
