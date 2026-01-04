<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue'
import {
  GetAllRepositories, GetRepository, AddRepository, UpdateRepository,
  UpdateRepositoryAlias, DeleteRepository, SearchRepositories,
  SelectDirectory, IsValidGitRepository
} from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

const repositories = ref<models.Repository[]>([])
const isLoading = ref(false)
const showEditor = ref(false)
const editingRepo = ref<models.Repository | null>(null)
const searchKeyword = ref('')

const form = ref({
  path: '',
  alias: '',
  description: ''
})

async function loadRepositories() {
  isLoading.value = true
  try {
    repositories.value = await GetAllRepositories()
  } catch (error: any) {
    console.error('Failed to load repositories:', error)
    repositories.value = []
  } finally {
    isLoading.value = false
  }
}

async function searchRepositories() {
  isLoading.value = true
  try {
    if (searchKeyword.value.trim()) {
      repositories.value = await SearchRepositories(searchKeyword.value.trim())
    } else {
      repositories.value = await GetAllRepositories()
    }
  } catch (error: any) {
    console.error('Failed to search repositories:', error)
    repositories.value = []
  } finally {
    isLoading.value = false
  }
}

async function openSelectDialog() {
  const path = await SelectDirectory()
  if (path) {
    const isValid = await IsValidGitRepository(path)
    if (!isValid) {
      alert('所选路径不是有效的 Git 仓库')
      return
    }
    form.value.path = path
    showEditor.value = true
    editingRepo.value = null
  }
}

function openCreateDialog() {
  editingRepo.value = null
  form.value = { path: '', alias: '', description: '' }
  showEditor.value = true
}

function openEditDialog(repo: models.Repository) {
  editingRepo.value = repo
  form.value = {
    path: repo.path,
    alias: repo.alias,
    description: repo.description
  }
  showEditor.value = true
}

async function saveRepository() {
  if (!form.value.path.trim()) {
    alert('请选择仓库路径')
    return
  }

  // Validate path is a git repository
  const isValid = await IsValidGitRepository(form.value.path)
  if (!isValid) {
    alert('所选路径不是有效的 Git 仓库')
    return
  }

  try {
    if (editingRepo.value) {
      await UpdateRepository(editingRepo.value.id, form.value.alias.trim(), form.value.description.trim())
    } else {
      await AddRepository(form.value.path, form.value.alias.trim(), form.value.description.trim())
    }
    showEditor.value = false
    await loadRepositories()
  } catch (error: any) {
    console.error('Failed to save repository:', error)
    alert('保存失败: ' + error.message)
  }
}

async function deleteRepository(repo: models.Repository) {
  if (!confirm(`确定要从管理列表中删除 "${repo.alias || repo.path}" 吗？\n这不会删除实际的仓库文件。`)) {
    return
  }

  try {
    await DeleteRepository(repo.id)
    await loadRepositories()
  } catch (error: any) {
    console.error('Failed to delete repository:', error)
    alert('删除失败: ' + error.message)
  }
}

function getDisplayName(repo: models.Repository): string {
  return repo.alias || repo.path.split('/').pop() || repo.path
}

function truncatePath(path: string, maxLength: number = 40): string {
  if (path.length <= maxLength) return path
  return '...' + path.slice(-maxLength)
}

const totalRepos = computed(() => repositories.value.length)

onMounted(() => {
  loadRepositories()
})

defineExpose({ loadRepositories })
</script>

<template>
  <div class="repositories-panel">
    <div class="panel-header">
      <div class="header-left">
        <h2>仓库管理</h2>
        <span class="count">{{ totalRepos }} 个仓库</span>
      </div>
      <div class="header-actions">
        <button @click="loadRepositories" class="btn-refresh" :disabled="isLoading" title="刷新">
          <span v-if="isLoading">⟳</span>
          <span v-else>⟳</span>
        </button>
        <button @click="openCreateDialog" class="btn-create">
          + 添加仓库
        </button>
      </div>
    </div>

    <!-- Search -->
    <div class="search-bar">
      <input
        v-model="searchKeyword"
        type="text"
        placeholder="搜索仓库..."
        class="search-input"
        @input="searchRepositories"
      />
      <button v-if="searchKeyword" @click="searchKeyword = ''; loadRepositories()" class="btn-clear">
        ✕
      </button>
    </div>

    <div v-if="isLoading && repositories.length === 0" class="loading">加载中...</div>

    <div v-else-if="repositories.length === 0" class="empty">
      <p>暂无仓库</p>
      <p class="hint">点击"添加仓库"将本地 Git 仓库添加到管理列表</p>
    </div>

    <div v-else class="repo-list">
      <div
        v-for="repo in repositories"
        :key="repo.id"
        class="repo-item"
      >
        <div class="repo-header">
          <div class="repo-info">
            <span class="repo-icon">📁</span>
            <div class="repo-details">
              <span class="repo-name">{{ getDisplayName(repo) }}</span>
              <span class="repo-path" :title="repo.path">{{ truncatePath(repo.path) }}</span>
            </div>
          </div>
          <div class="repo-actions">
            <button @click="openEditDialog(repo)" class="btn-action" title="编辑">
              编辑
            </button>
            <button @click="deleteRepository(repo)" class="btn-action btn-danger" title="删除">
              删除
            </button>
          </div>
        </div>
        <div class="repo-description" v-if="repo.description">
          {{ repo.description }}
        </div>
        <div class="repo-meta">
          <span class="repo-id">ID: {{ repo.id.substring(0, 8) }}</span>
          <span class="repo-updated">更新: {{ repo.updatedAt.split('T')[0] }}</span>
        </div>
      </div>
    </div>

    <!-- Editor Dialog -->
    <div v-if="showEditor" class="dialog-overlay" @click.self="showEditor = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ editingRepo ? '编辑仓库' : '添加仓库' }}</h3>
          <button @click="showEditor = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>仓库路径 <span class="required">*</span></label>
            <div class="path-input-group">
              <input
                v-model="form.path"
                type="text"
                placeholder="选择或输入仓库路径"
                class="form-input"
                :disabled="!!editingRepo"
              />
              <button v-if="!editingRepo" @click="openSelectDialog" class="btn-browse">
                浏览...
              </button>
            </div>
          </div>
          <div class="form-group">
            <label>仓库别名</label>
            <input
              v-model="form.alias"
              type="text"
              placeholder="自定义显示名称"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>描述</label>
            <textarea
              v-model="form.description"
              placeholder="添加描述信息..."
              class="form-input"
              rows="3"
            ></textarea>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showEditor = false" class="btn-cancel">取消</button>
          <button @click="saveRepository" class="btn-confirm" :disabled="!form.path.trim()">
            保存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.repositories-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-left h2 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
}

.count {
  font-size: 0.8rem;
  color: #888;
  background: rgba(255, 255, 255, 0.05);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-refresh, .btn-create {
  padding: 0.4rem 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.btn-create {
  background: rgba(97, 218, 251, 0.1);
  border-color: #61dafb;
  color: #61dafb;
}

.btn-refresh:hover:not(:disabled),
.btn-create:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
}

.btn-refresh:disabled,
.btn-create:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.search-bar {
  display: flex;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.search-input {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  color: #e5e7eb;
  font-size: 0.9rem;
}

.search-input:focus {
  outline: none;
  border-color: #61dafb;
}

.btn-clear {
  padding: 0.4rem 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  background: transparent;
  color: #888;
  cursor: pointer;
}

.btn-clear:hover {
  color: #fff;
}

.loading, .empty {
  padding: 2rem;
  text-align: center;
  color: #888;
}

.empty .hint {
  font-size: 0.85rem;
  color: #666;
  margin-top: 0.5rem;
}

.repo-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.repo-item {
  padding: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  margin-bottom: 0.75rem;
  transition: all 0.2s;
}

.repo-item:hover {
  border-color: rgba(255, 255, 255, 0.2);
}

.repo-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.5rem;
}

.repo-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.repo-icon {
  font-size: 1.5rem;
}

.repo-details {
  display: flex;
  flex-direction: column;
}

.repo-name {
  font-size: 1rem;
  color: #e5e7eb;
  font-weight: 500;
}

.repo-path {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.75rem;
  color: #888;
}

.repo-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  padding: 0.3rem 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 4px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  font-size: 0.75rem;
  transition: all 0.2s;
}

.btn-action:hover {
  background: rgba(255, 255, 255, 0.1);
}

.btn-action.btn-danger:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.repo-description {
  font-size: 0.85rem;
  color: #999;
  margin-bottom: 0.5rem;
  padding-left: 2rem;
}

.repo-meta {
  display: flex;
  gap: 1rem;
  padding-left: 2rem;
  font-size: 0.7rem;
  color: #666;
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

.required {
  color: #f87171;
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

.form-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.form-input textarea {
  resize: vertical;
  font-family: inherit;
}

.path-input-group {
  display: flex;
  gap: 0.5rem;
}

.path-input-group .form-input {
  flex: 1;
}

.btn-browse {
  padding: 0.6rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  color: #ccc;
  cursor: pointer;
  font-size: 0.85rem;
  white-space: nowrap;
  transition: all 0.2s;
}

.btn-browse:hover {
  background: rgba(255, 255, 255, 0.1);
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
