<script lang="ts" setup>
import { ref, onMounted, watch } from 'vue'
import { GetLog, GetCommitDetail } from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

const props = defineProps<{
  hasRepository: boolean
}>()

const emit = defineEmits(['commit-selected'])

const commits = ref<models.CommitInfo[]>([])
const isLoading = ref(false)
const selectedCommit = ref<string | null>(null)
const commitDetail = ref<{ hash: string; message: string; author: string; date: string } | null>(null)
const loadMoreCount = ref(20)

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
      <button @click="loadCommits" class="btn-refresh" :disabled="isLoading" title="刷新">
        <span v-if="isLoading">⟳</span>
        <span v-else>⟳</span>
      </button>
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
</style>
