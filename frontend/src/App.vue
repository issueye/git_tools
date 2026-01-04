<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import GitStatus from './components/GitStatus.vue'
import CommitPanel from './components/CommitPanel.vue'
import BranchPanel from './components/BranchPanel.vue'
import AIConfigPanel from './components/AIConfigPanel.vue'
import {
  GetStatus,
  GetRecentRepositories,
  SelectRepository,
  GetRepositoryInfo,
  SelectDirectory,
  CloneRepository
} from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

type TabType = 'status' | 'branches' | 'ai-config'

const currentTab = ref<TabType>('status')
const status = ref<models.GitStatus | null>(null)
const recentRepos = ref<string[]>([])
const currentRepo = ref('')
const isLoading = ref(false)

// Clone dialog state
const showCloneDialog = ref(false)
const cloneUrl = ref('')
const clonePath = ref('')
const cloneBranch = ref('')
const cloneLoading = ref(false)

const branchPanelRef = ref<InstanceType<typeof BranchPanel> | null>(null)

async function loadStatus() {
  try {
    const result = await GetStatus()
    status.value = result
  } catch (error: any) {
    // No repository selected - this is expected, don't log error
    const msg = error?.message || String(error) || ''
    if (!msg.toLowerCase().includes('repository')) {
      console.error('Failed to get status:', error)
    }
    status.value = null
  }
}

async function loadRecentRepos() {
  try {
    const result = await GetRecentRepositories()
    recentRepos.value = result || []
  } catch (error) {
    console.error('Failed to load recent repos:', error)
  }
}

async function loadRepoInfo() {
  try {
    const info = await GetRepositoryInfo()
    currentRepo.value = info?.path || ''
  } catch (error) {
    currentRepo.value = ''
  }
}

async function selectRecentRepo(path: string) {
  isLoading.value = true
  try {
    await SelectRepository(path)
    await loadStatus()
    await loadRepoInfo()
    if (branchPanelRef.value) {
      branchPanelRef.value.loadBranches()
    }
  } catch (error: any) {
    alert('Failed to open repository: ' + error.message)
  } finally {
    isLoading.value = false
  }
}

function onRefresh() {
  loadStatus()
  if (branchPanelRef.value) {
    branchPanelRef.value.loadBranches()
  }
}

function onCommitted() {
  loadStatus()
  if (branchPanelRef.value) {
    branchPanelRef.value.loadBranches()
  }
}

function onBranchChanged() {
  loadStatus()
  loadRepoInfo()
}

// Open local repository
async function openRepository() {
  try {
    const path = await SelectDirectory()
    if (!path) return // User cancelled

    // Try to select the repository
    try {
      await SelectRepository(path)
      await loadStatus()
      await loadRepoInfo()
      await loadRecentRepos()
      if (branchPanelRef.value) {
        branchPanelRef.value.loadBranches()
      }
    } catch (err: any) {
      // If selection fails, check if it's a valid git repository
      if (err?.message?.includes('not a git repository')) {
        alert('The selected directory is not a git repository')
      } else {
        throw err
      }
    }
  } catch (error: any) {
    console.error('Failed to open repository:', error)
    alert('Failed to open repository: ' + (error?.message || error))
  }
}

// Show clone dialog
function showCloneRepository() {
  cloneUrl.value = ''
  clonePath.value = ''
  cloneBranch.value = ''
  showCloneDialog.value = true
}

// Clone repository
async function cloneRepository() {
  if (!cloneUrl.value.trim()) {
    alert('Please enter a repository URL')
    return
  }

  if (!clonePath.value.trim()) {
    alert('Please select a destination directory')
    return
  }

  cloneLoading.value = true
  try {
    await CloneRepository(cloneUrl.value.trim(), clonePath.value.trim(), cloneBranch.value.trim())
    showCloneDialog.value = false
    // Load the newly cloned repository
    await selectRecentRepo(clonePath.value.trim())
  } catch (error: any) {
    console.error('Failed to clone repository:', error)
    alert('Failed to clone repository: ' + (error?.message || error))
  } finally {
    cloneLoading.value = false
  }
}

// Select clone destination path
async function selectClonePath() {
  try {
    const path = await SelectDirectory()
    if (path) {
      clonePath.value = path
    }
  } catch (error: any) {
    console.error('Failed to select path:', error)
  }
}

onMounted(async () => {
  await loadStatus()
  await loadRecentRepos()
  await loadRepoInfo()
})

const hasStagedChanges = ref(false)

function updateHasStagedChanges() {
  hasStagedChanges.value = (status.value?.staged?.length ?? 0) > 0
}

// Watch status changes
import { watch } from 'vue'
watch(status, () => {
  updateHasStagedChanges()
}, { immediate: true, deep: true })
</script>

<template>
  <div class="app-container">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1>Git AI 工具</h1>
        <div class="repo-info" v-if="currentRepo">
          <span class="repo-label">仓库路径:</span>
          <span class="repo-path" :title="currentRepo">{{ currentRepo.split('/').pop() || currentRepo }}</span>
        </div>
      </div>

      <!-- Navigation Tabs -->
      <nav class="nav-tabs">
        <button
          @click="currentTab = 'status'"
          :class="{ active: currentTab === 'status' }"
          class="nav-tab"
        >
          <span class="tab-icon">📋</span>
          <span>状态</span>
        </button>
        <button
          @click="currentTab = 'branches'"
          :class="{ active: currentTab === 'branches' }"
          class="nav-tab"
        >
          <span class="tab-icon">🌿</span>
          <span>分支</span>
        </button>
        <button
          @click="currentTab = 'ai-config'"
          :class="{ active: currentTab === 'ai-config' }"
          class="nav-tab"
        >
          <span class="tab-icon">🤖</span>
          <span>AI 配置</span>
        </button>
      </nav>

      <!-- Repository Actions -->
      <div class="repo-actions">
        <button @click="openRepository" class="action-btn">
          <span class="action-icon">📂</span>
          <span>打开仓库</span>
        </button>
        <button @click="showCloneRepository" class="action-btn">
          <span class="action-icon">📥</span>
          <span>克隆仓库</span>
        </button>
      </div>

      <!-- Recent Repositories -->
      <div class="recent-repos" v-if="recentRepos.length > 0">
        <h3>最近打开</h3>
        <div class="repo-list">
          <div
            v-for="repo in recentRepos.slice(0, 5)"
            :key="repo"
            @click="selectRecentRepo(repo)"
            class="repo-item"
            :class="{ current: repo === currentRepo }"
          >
            <span class="repo-icon">📁</span>
            <span class="repo-name" :title="repo">{{ repo.split('/').pop() || repo }}</span>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <div v-if="isLoading" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <!-- Status Tab -->
      <div v-show="currentTab === 'status'" class="tab-content">
        <div class="content-split">
          <div class="status-section">
            <GitStatus :status="status" @refresh="loadStatus" />
          </div>
          <div class="commit-section">
            <CommitPanel :has-staged-changes="hasStagedChanges" @committed="onCommitted" />
          </div>
        </div>
      </div>

      <!-- Branches Tab -->
      <div v-show="currentTab === 'branches'" class="tab-content">
        <BranchPanel ref="branchPanelRef" :has-repository="!!currentRepo" @branch-changed="onBranchChanged" />
      </div>

      <!-- AI Config Tab -->
      <div v-show="currentTab === 'ai-config'" class="tab-content">
        <AIConfigPanel />
      </div>
    </main>

    <!-- Clone Repository Dialog -->
    <div v-if="showCloneDialog" class="dialog-overlay" @click.self="showCloneDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>克隆仓库</h3>
          <button @click="showCloneDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>仓库地址</label>
            <input
              v-model="cloneUrl"
              type="text"
              placeholder="https://github.com/user/repo.git"
              class="form-input"
              @keyup.enter="cloneRepository"
            />
          </div>
          <div class="form-group">
            <label>本地目录</label>
            <div class="input-with-btn">
              <input
                v-model="clonePath"
                type="text"
                placeholder="请选择目录"
                class="form-input"
                readonly
              />
              <button @click="selectClonePath" class="btn-browse">浏览</button>
            </div>
          </div>
          <div class="form-group">
            <label>分支（可选）</label>
            <input
              v-model="cloneBranch"
              type="text"
              placeholder="留空则使用默认分支"
              class="form-input"
              @keyup.enter="cloneRepository"
            />
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showCloneDialog = false" class="btn-cancel">取消</button>
          <button @click="cloneRepository" class="btn-confirm" :disabled="cloneLoading || !cloneUrl || !clonePath">
            <span v-if="cloneLoading">克隆中...</span>
            <span v-else>克隆</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: #1a1a2e;
  color: #e5e7eb;
  overflow: hidden;
}

#app {
  height: 100vh;
  width: 100vw;
}

.app-container {
  display: flex;
  height: 100%;
}

/* Sidebar */
.sidebar {
  width: 260px;
  background: #16162a;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  flex-direction: column;
  padding: 1rem;
  gap: 1.5rem;
}

.sidebar-header h1 {
  font-size: 1.2rem;
  font-weight: 600;
  color: #fff;
  margin-bottom: 0.5rem;
}

.repo-info {
  padding: 0.5rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 6px;
  font-size: 0.75rem;
}

.repo-label {
  color: #888;
  display: block;
  margin-bottom: 0.25rem;
}

.repo-path {
  color: #61dafb;
  font-family: 'Consolas', 'Monaco', monospace;
  word-break: break-all;
}

/* Navigation Tabs */
.nav-tabs {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

/* Repository Actions */
.repo-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(97, 218, 251, 0.3);
  color: #fff;
}

.action-icon {
  font-size: 1.1rem;
}

.nav-tab {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.nav-tab:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}

.nav-tab.active {
  background: rgba(97, 218, 251, 0.1);
  border-color: #61dafb;
  color: #61dafb;
}

.tab-icon {
  font-size: 1.2rem;
}

/* Recent Repositories */
.recent-repos h3 {
  font-size: 0.8rem;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
}

.repo-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.repo-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.repo-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.repo-item.current {
  background: rgba(97, 218, 251, 0.1);
}

.repo-icon {
  font-size: 0.9rem;
}

.repo-name {
  font-size: 0.85rem;
  color: #ccc;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Main Content */
.main-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.tab-content {
  height: 100%;
  overflow: auto;
}

.content-split {
  display: flex;
  height: 100%;
  gap: 1px;
  background: rgba(255, 255, 255, 0.1);
}

.status-section {
  flex: 2;
  background: #1e1e3f;
  overflow: hidden;
}

.commit-section {
  flex: 1;
  background: #1e1e3f;
  overflow: hidden;
  min-width: 300px;
  max-width: 450px;
}

/* Loading Overlay */
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: #61dafb;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Scrollbar Styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.2);
}

::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
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
  min-width: 450px;
  max-width: 90vw;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
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
  padding: 1.25rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #ccc;
  font-size: 0.9rem;
}

.form-input {
  width: 100%;
  padding: 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  color: #e5e7eb;
  font-size: 0.9rem;
}

.form-input:focus {
  outline: none;
  border-color: #61dafb;
}

.form-input[readonly] {
  background: rgba(0, 0, 0, 0.3);
  cursor: not-allowed;
}

.input-with-btn {
  display: flex;
  gap: 0.5rem;
}

.input-with-btn .form-input {
  flex: 1;
}

.btn-browse {
  padding: 0.5rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.btn-browse:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-cancel, .btn-confirm {
  padding: 0.5rem 1.25rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
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

.btn-confirm:hover:not(:disabled) {
  background: rgba(97, 218, 251, 0.2);
}

.btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
