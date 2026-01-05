<script lang="ts" setup>
import { ref, onMounted, watch } from 'vue'
import { GetLog, GetCommitDetail, Reset, Revert } from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

const props = defineProps<{
  hasRepository: boolean
}>()

const emit = defineEmits(['history-changed'])

const commits = ref<models.CommitInfo[]>([])
const isLoading = ref(false)
const selectedCommit = ref<string | null>(null)
const commitDetail = ref<{ hash: string; message: string; author: string; date: string } | null>(null)
const loadMoreCount = ref(20)

// 版本操作相关
const showResetDialog = ref(false)
const showRevertDialog = ref(false)
const resetType = ref<'soft' | 'mixed' | 'hard'>('soft')
const revertTargetHash = ref<string>('')
const isResetting = ref(false)
const isReverting = ref(false)
const operationResult = ref<{ success: boolean; message: string } | null>(null)

async function loadCommits() {
  if (!props.hasRepository) return

  isLoading.value = true
  try {
    const result = await GetLog(loadMoreCount.value)
    commits.value = result || []
  } catch (error: any) {
    console.error('Failed to load commits:', error)
    commits.value = []
  } finally {
    isLoading.value = false
  }
}

async function loadMore() {
  loadMoreCount.value += 20
  await loadCommits()
}

async function showCommitDetail(commit: models.CommitInfo) {
  selectedCommit.value = commit.hash
  try {
    const detail = await GetCommitDetail(commit.hash)
    if (detail) {
      commitDetail.value = {
        hash: detail.hash as string,
        message: detail.message as string,
        author: detail.author as string,
        date: detail.date as string
      }
    }
  } catch (error: any) {
    console.error('Failed to load commit detail:', error)
  }
}

function closeDetail() {
  selectedCommit.value = null
  commitDetail.value = null
}

// 撤销当前分支
async function performReset() {
  if (isResetting.value) return
  isResetting.value = true
  operationResult.value = null

  try {
    await Reset(resetType.value, 'HEAD~1')
    operationResult.value = { success: true, message: '撤销成功！' }
    showResetDialog.value = false
    await loadCommits()
    emit('history-changed')
  } catch (error: any) {
    console.error('Reset failed:', error)
    operationResult.value = { success: false, message: '撤销失败: ' + (error?.message || String(error)) }
  } finally {
    isResetting.value = false
  }
}

// 回滚到指定提交
async function performRevert() {
  if (!revertTargetHash.value || isReverting.value) return
  isReverting.value = true
  operationResult.value = null

  try {
    await Revert(revertTargetHash.value, false)
    operationResult.value = { success: true, message: '回滚成功！' }
    showRevertDialog.value = false
    revertTargetHash.value = ''
    await loadCommits()
    emit('history-changed')
  } catch (error: any) {
    console.error('Revert failed:', error)
    operationResult.value = { success: false, message: '回滚失败: ' + (error?.message || String(error)) }
  } finally {
    isReverting.value = false
  }
}

function openResetDialog() {
  showResetDialog.value = true
  operationResult.value = null
}

function openRevertDialog() {
  showRevertDialog.value = true
  operationResult.value = null
  if (selectedCommit.value) {
    revertTargetHash.value = selectedCommit.value
  }
}

onMounted(() => {
  if (props.hasRepository) {
    loadCommits()
  }
})

watch(() => props.hasRepository, (newVal) => {
  if (newVal) {
    loadCommits()
  } else {
    commits.value = []
  }
})

defineExpose({ loadCommits })
</script>

<template>
  <div class="history-panel">
    <div class="panel-header">
      <h2>提交历史</h2>
      <div class="header-actions">
        <button @click="openResetDialog" class="btn-action-small" :disabled="!hasRepository || commits.length === 0" title="撤销">
          ↩️ 撤销
        </button>
        <button @click="openRevertDialog" class="btn-action-small" :disabled="!hasRepository || !selectedCommit" title="回滚">
          🔄 回滚
        </button>
        <button @click="loadCommits" class="btn-refresh" :disabled="isLoading" title="刷新">
          <span v-if="isLoading">⟳</span>
          <span v-else>⟳</span>
        </button>
      </div>
    </div>

    <div v-if="!hasRepository" class="no-repo">
      <p>请先选择一个仓库</p>
    </div>

    <div v-else-if="isLoading && commits.length === 0" class="loading">加载中...</div>

    <div v-else-if="commits.length === 0" class="empty">暂无提交记录</div>

    <div v-else class="commit-list">
      <div
        v-for="commit in commits"
        :key="commit.hash"
        class="commit-item"
        :class="{ selected: selectedCommit === commit.hash }"
        @click="showCommitDetail(commit)"
      >
        <div class="commit-header">
          <span class="commit-hash">{{ commit.hash }}</span>
          <span class="commit-date">{{ commit.date }}</span>
        </div>
        <div class="commit-message">{{ commit.message }}</div>
        <div class="commit-meta">
          <span class="commit-author">{{ commit.author }}</span>
        </div>
      </div>

      <button @click="loadMore" class="btn-load-more" :disabled="isLoading">
        加载更多
      </button>
    </div>

    <!-- Operation Result -->
    <div v-if="operationResult" class="operation-result" :class="{ success: operationResult.success, error: !operationResult.success }">
      {{ operationResult.message }}
      <button @click="operationResult = null" class="close-result">✕</button>
    </div>

    <!-- Commit Detail Sidebar -->
    <div v-if="commitDetail" class="commit-detail">
      <div class="detail-header">
        <h3>提交详情</h3>
        <button @click="closeDetail" class="btn-close">✕</button>
      </div>
      <div class="detail-content">
        <div class="detail-row">
          <span class="label">Hash:</span>
          <span class="value hash">{{ commitDetail.hash }}</span>
        </div>
        <div class="detail-row">
          <span class="label">作者:</span>
          <span class="value">{{ commitDetail.author }}</span>
        </div>
        <div class="detail-row">
          <span class="label">时间:</span>
          <span class="value">{{ commitDetail.date }}</span>
        </div>
        <div class="detail-message">
          <span class="label">描述:</span>
          <p>{{ commitDetail.message }}</p>
        </div>
        <div class="detail-actions">
          <button @click="openRevertDialog" class="btn-detail-action" :disabled="!selectedCommit">
            🔄 回滚到此提交
          </button>
        </div>
      </div>
    </div>

    <!-- Reset Dialog -->
    <div v-if="showResetDialog" class="dialog-overlay" @click.self="showResetDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>撤销最近一次提交</h3>
          <button @click="showResetDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="warning-message">
            ⚠️ 撤销操作将回退到上一次提交，请确认操作。
          </div>
          <div class="form-group">
            <label>撤销类型</label>
            <div class="radio-group">
              <label class="radio-item">
                <input type="radio" v-model="resetType" value="soft" />
                <span>软撤销 - 保留变更在暂存区</span>
              </label>
              <label class="radio-item">
                <input type="radio" v-model="resetType" value="mixed" />
                <span>混合撤销 - 保留变更在工作区</span>
              </label>
              <label class="radio-item">
                <input type="radio" v-model="resetType" value="hard" />
                <span>硬撤销 - 丢弃所有变更（危险！）</span>
              </label>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showResetDialog = false" class="btn-cancel">取消</button>
          <button @click="performReset" class="btn-confirm btn-danger" :disabled="isResetting">
            <span v-if="isResetting">撤销中...</span>
            <span v-else>确认撤销</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Revert Dialog -->
    <div v-if="showRevertDialog" class="dialog-overlay" @click.self="showRevertDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>回滚到指定提交</h3>
          <button @click="showRevertDialog = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="warning-message">
            ⚠️ 回滚操作将创建一个新的提交来撤销指定提交的所有更改。
          </div>
          <div class="form-group">
            <label>目标提交</label>
            <input
              type="text"
              v-model="revertTargetHash"
              class="form-input"
              placeholder="输入提交 hash"
            />
            <p class="help-text" v-if="selectedCommit">
              当前选中: {{ selectedCommit }}
            </p>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showRevertDialog = false" class="btn-cancel">取消</button>
          <button @click="performRevert" class="btn-confirm" :disabled="!revertTargetHash || isReverting">
            <span v-if="isReverting">回滚中...</span>
            <span v-else>确认回滚</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
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
  align-items: center;
}

.btn-action-small {
  padding: 0.35rem 0.75rem;
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
  border-color: rgba(255, 255, 255, 0.3);
}

.btn-action-small:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-refresh {
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  color: #888;
  cursor: pointer;
  font-size: 1.2rem;
}

.btn-refresh:hover:not(:disabled) {
  color: #61dafb;
}

.no-repo, .loading, .empty {
  padding: 2rem;
  text-align: center;
  color: #888;
}

.commit-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.commit-item {
  padding: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  margin-bottom: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.commit-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.commit-item.selected {
  background: rgba(97, 218, 251, 0.1);
  border-color: #61dafb;
}

.commit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.25rem;
}

.commit-hash {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.8rem;
  color: #61dafb;
}

.commit-date {
  font-size: 0.75rem;
  color: #888;
}

.commit-message {
  font-size: 0.9rem;
  color: #e5e7eb;
  margin-bottom: 0.25rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.commit-meta {
  font-size: 0.75rem;
  color: #888;
}

.btn-load-more {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  margin-top: 0.5rem;
  transition: all 0.2s;
}

.btn-load-more:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.05);
}

.btn-load-more:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Commit Detail Sidebar */
.commit-detail {
  position: absolute;
  top: 0;
  right: 0;
  width: 300px;
  height: 100%;
  background: #1a1a2e;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  flex-direction: column;
  z-index: 10;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.detail-header h3 {
  margin: 0;
  font-size: 1rem;
  color: #fff;
}

.btn-close {
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  color: #888;
  cursor: pointer;
  font-size: 1rem;
}

.btn-close:hover {
  color: #fff;
}

.detail-content {
  padding: 1rem;
  flex: 1;
  overflow-y: auto;
}

.detail-row {
  display: flex;
  margin-bottom: 0.75rem;
}

.detail-row .label {
  width: 60px;
  color: #888;
  font-size: 0.85rem;
  flex-shrink: 0;
}

.detail-row .value {
  color: #ccc;
  font-size: 0.85rem;
  word-break: break-all;
}

.detail-row .value.hash {
  font-family: 'Consolas', 'Monaco', monospace;
  color: #61dafb;
}

.detail-message {
  margin-top: 1rem;
}

.detail-message .label {
  display: block;
  color: #888;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
}

.detail-message p {
  color: #e5e7eb;
  font-size: 0.9rem;
  line-height: 1.5;
  margin: 0;
}

.detail-actions {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-detail-action {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid rgba(97, 218, 251, 0.3);
  border-radius: 4px;
  background: rgba(97, 218, 251, 0.1);
  color: #61dafb;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.btn-detail-action:hover:not(:disabled) {
  background: rgba(97, 218, 251, 0.2);
}

.btn-detail-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Operation Result */
.operation-result {
  position: absolute;
  bottom: 1rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  z-index: 20;
  min-width: 200px;
  justify-content: center;
}

.operation-result.success {
  background: rgba(34, 197, 94, 0.15);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.operation-result.error {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.operation-result .close-result {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  opacity: 0.7;
  font-size: 0.9rem;
  padding: 0;
}

.operation-result .close-result:hover {
  opacity: 1;
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
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
}

.dialog-body {
  padding: 1.25rem;
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
  font-size: 0.9rem;
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

.btn-danger {
  border-color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.1) !important;
  color: #f87171 !important;
}

.btn-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2) !important;
}

.warning-message {
  padding: 0.75rem;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid #fbbf24;
  border-radius: 6px;
  color: #fbbf24;
  font-size: 0.9rem;
  margin-bottom: 1rem;
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

.help-text {
  font-size: 0.8rem;
  color: #888;
  margin-top: 0.5rem;
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.radio-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  color: #ccc;
}

.radio-item:hover {
  background: rgba(255, 255, 255, 0.05);
}
</style>
