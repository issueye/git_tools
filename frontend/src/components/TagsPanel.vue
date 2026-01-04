<script lang="ts" setup>
import { ref, onMounted, watch } from 'vue'
import { GetTags, CreateTag, DeleteTag, CheckoutTag, GetLog } from '/wailsjs/go/main/App'
import type { models, git } from '/wailsjs/go/models'

const props = defineProps<{
  hasRepository: boolean
}>()

const emit = defineEmits(['tag-changed'])

const tags = ref<git.Tag[]>([])
const isLoading = ref(false)
const showCreateDialog = ref(false)
const newTagName = ref('')
const newTagMessage = ref('')
const newTagCommit = ref('HEAD')
const commits = ref<models.CommitInfo[]>([])

async function loadTags() {
  if (!props.hasRepository) return

  isLoading.value = true
  try {
    const result = await GetTags()
    tags.value = result || []
  } catch (error: any) {
    console.error('Failed to load tags:', error)
    tags.value = []
  } finally {
    isLoading.value = false
  }
}

async function loadCommitsForTag() {
  try {
    const result = await GetLog(20)
    commits.value = result || []
  } catch (error: any) {
    console.error('Failed to load commits:', error)
    commits.value = []
  }
}

function showCreateTagDialog() {
  showCreateDialog.value = true
  newTagName.value = ''
  newTagMessage.value = ''
  newTagCommit.value = 'HEAD'
  loadCommitsForTag()
}

async function createTag() {
  if (!newTagName.value.trim()) {
    alert('请输入标签名称')
    return
  }

  try {
    await CreateTag(newTagName.value.trim(), newTagMessage.value.trim(), newTagCommit.value)
    showCreateDialog.value = false
    newTagName.value = ''
    newTagMessage.value = ''
    await loadTags()
    emit('tag-changed')
  } catch (error: any) {
    console.error('Failed to create tag:', error)
    alert('创建标签失败: ' + error.message)
  }
}

async function deleteTag(tagName: string) {
  if (!confirm(`确定要删除标签 "${tagName}" 吗？`)) {
    return
  }

  try {
    await DeleteTag(tagName)
    await loadTags()
    emit('tag-changed')
  } catch (error: any) {
    console.error('Failed to delete tag:', error)
    alert('删除标签失败: ' + error.message)
  }
}

async function checkoutTag(tagName: string) {
  try {
    await CheckoutTag(tagName)
    emit('tag-changed')
  } catch (error: any) {
    console.error('Failed to checkout tag:', error)
    alert('检出标签失败: ' + error.message)
  }
}

onMounted(() => {
  if (props.hasRepository) {
    loadTags()
  }
})

watch(() => props.hasRepository, (newVal) => {
  if (newVal) {
    loadTags()
  } else {
    tags.value = []
  }
})

defineExpose({ loadTags })
</script>

<template>
  <div class="tags-panel">
    <div class="panel-header">
      <h2>标签管理</h2>
      <div class="header-actions">
        <button @click="loadTags" class="btn-refresh" :disabled="isLoading" title="刷新">
          <span v-if="isLoading">⟳</span>
          <span v-else>⟳</span>
        </button>
        <button @click="showCreateTagDialog" class="btn-create" :disabled="!hasRepository">
          + 新建标签
        </button>
      </div>
    </div>

    <div v-if="!hasRepository" class="no-repo">
      <p>请先选择一个仓库</p>
    </div>

    <div v-else-if="isLoading && tags.length === 0" class="loading">加载中...</div>

    <div v-else-if="tags.length === 0" class="empty">
      <p>暂无标签</p>
      <p class="hint">点击"新建标签"创建一个新标签</p>
    </div>

    <div v-else class="tag-list">
      <div
        v-for="tag in tags"
        :key="tag.name"
        class="tag-item"
      >
        <div class="tag-info">
          <span class="tag-icon">🏷️</span>
          <div class="tag-details">
            <span class="tag-name">{{ tag.name }}</span>
            <span class="tag-commit">{{ tag.commitHash.substring(0, 7) }}</span>
          </div>
          <span v-if="tag.isAnnotated" class="tag-badge">附注</span>
        </div>
        <div class="tag-message" v-if="tag.message">{{ tag.message }}</div>
        <div class="tag-actions">
          <button @click="checkoutTag(tag.name)" class="btn-action" title="检出">
            检出
          </button>
          <button @click="deleteTag(tag.name)" class="btn-action btn-danger" title="删除">
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Create Tag Dialog -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>创建新标签</h3>
          <button @click="showCreateDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>标签名称</label>
            <input
              v-model="newTagName"
              type="text"
              placeholder="v1.0.0"
              class="form-input"
              @keyup.enter="createTag"
            />
          </div>
          <div class="form-group">
            <label>关联提交</label>
            <select v-model="newTagCommit" class="form-input">
              <option value="HEAD">当前 HEAD</option>
              <option v-for="commit in commits" :key="commit.hash" :value="commit.hash">
                {{ commit.hash.substring(0, 7) }} - {{ commit.message }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>标签说明（可选）</label>
            <textarea
              v-model="newTagMessage"
              placeholder="输入标签说明..."
              class="form-input"
              rows="3"
            ></textarea>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showCreateDialog = false" class="btn-cancel">取消</button>
          <button @click="createTag" class="btn-confirm" :disabled="!newTagName.trim()">
            创建
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tags-panel {
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

.panel-header h2 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
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

.no-repo, .loading, .empty {
  padding: 2rem;
  text-align: center;
  color: #888;
}

.empty .hint {
  font-size: 0.85rem;
  color: #666;
  margin-top: 0.5rem;
}

.tag-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.tag-item {
  padding: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  margin-bottom: 0.5rem;
}

.tag-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tag-icon {
  font-size: 1.2rem;
}

.tag-details {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tag-name {
  font-size: 0.95rem;
  color: #e5e7eb;
  font-weight: 500;
}

.tag-commit {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.75rem;
  color: #888;
}

.tag-badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  background: rgba(147, 51, 234, 0.2);
  color: #c084fc;
}

.tag-message {
  font-size: 0.85rem;
  color: #999;
  margin-top: 0.5rem;
  padding-left: 1.7rem;
}

.tag-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
  padding-left: 1.7rem;
}

.btn-action {
  padding: 0.3rem 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 4px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  font-size: 0.8rem;
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

.form-input select {
  cursor: pointer;
}

.form-input textarea {
  resize: vertical;
  font-family: inherit;
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
