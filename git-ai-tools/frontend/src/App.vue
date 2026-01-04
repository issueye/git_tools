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
  GetRepositoryInfo
} from '/wailsjs/go/main/App'
import type { services } from '/wailsjs/go/models'

type TabType = 'status' | 'branches' | 'ai-config'

const currentTab = ref<TabType>('status')
const status = ref<services.GitStatus | null>(null)
const recentRepos = ref<string[]>([])
const currentRepo = ref('')
const isLoading = ref(false)

const branchPanelRef = ref<InstanceType<typeof BranchPanel> | null>(null)

async function loadStatus() {
  try {
    const result = await GetStatus()
    status.value = result
  } catch (error: any) {
    console.error('Failed to get status:', error)
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
        <h1>Git AI Tools</h1>
        <div class="repo-info" v-if="currentRepo">
          <span class="repo-label">Repository:</span>
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
          <span>Status</span>
        </button>
        <button
          @click="currentTab = 'branches'"
          :class="{ active: currentTab === 'branches' }"
          class="nav-tab"
        >
          <span class="tab-icon">🌿</span>
          <span>Branches</span>
        </button>
        <button
          @click="currentTab = 'ai-config'"
          :class="{ active: currentTab === 'ai-config' }"
          class="nav-tab"
        >
          <span class="tab-icon">🤖</span>
          <span>AI Config</span>
        </button>
      </nav>

      <!-- Recent Repositories -->
      <div class="recent-repos" v-if="recentRepos.length > 0">
        <h3>Recent Repositories</h3>
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
        <BranchPanel ref="branchPanelRef" @branch-changed="onBranchChanged" />
      </div>

      <!-- AI Config Tab -->
      <div v-show="currentTab === 'ai-config'" class="tab-content">
        <AIConfigPanel />
      </div>
    </main>
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
</style>
