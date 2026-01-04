<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue'
import {
  GetPrompts, GetDefaultPrompt, CreatePrompt, UpdatePrompt, DeletePrompt, SetDefaultPrompt
} from '/wailsjs/go/main/App'
import type { models } from '/wailsjs/go/models'

const prompts = ref<models.Prompt[]>([])
const isLoading = ref(false)
const showEditor = ref(false)
const editingPrompt = ref<models.Prompt | null>(null)
const defaultPromptId = ref<string>('')

const form = ref({
  name: '',
  description: '',
  template: '',
  isDefault: false
})

async function loadPrompts() {
  isLoading.value = true
  try {
    const [promptsData, defaultData] = await Promise.all([
      GetPrompts(),
      GetDefaultPrompt()
    ])
    prompts.value = promptsData || []
    defaultPromptId.value = defaultData?.id || ''
  } catch (error: any) {
    console.error('Failed to load prompts:', error)
    prompts.value = []
  } finally {
    isLoading.value = false
  }
}

function openCreateDialog() {
  editingPrompt.value = null
  form.value = { name: '', description: '', template: '', isDefault: false }
  showEditor.value = true
}

function openEditDialog(prompt: models.Prompt) {
  editingPrompt.value = prompt
  form.value = {
    name: prompt.name,
    description: prompt.description,
    template: prompt.template,
    isDefault: prompt.isDefault
  }
  showEditor.value = true
}

async function savePrompt() {
  if (!form.value.name.trim()) {
    alert('请输入提示词名称')
    return
  }
  if (!form.value.template.trim()) {
    alert('请输入模板内容')
    return
  }

  try {
    if (editingPrompt.value) {
      await UpdatePrompt(
        editingPrompt.value.id,
        form.value.name.trim(),
        form.value.description.trim(),
        form.value.template.trim(),
        form.value.isDefault
      )
    } else {
      await CreatePrompt(
        form.value.name.trim(),
        form.value.description.trim(),
        form.value.template.trim(),
        form.value.isDefault
      )
    }
    showEditor.value = false
    await loadPrompts()
  } catch (error: any) {
    console.error('Failed to save prompt:', error)
    alert('保存失败: ' + error.message)
  }
}

async function deletePrompt(prompt: models.Prompt) {
  if (!confirm(`确定要删除提示词 "${prompt.name}" 吗？`)) {
    return
  }

  try {
    await DeletePrompt(prompt.id)
    await loadPrompts()
  } catch (error: any) {
    console.error('Failed to delete prompt:', error)
    alert('删除失败: ' + error.message)
  }
}

async function setDefault(prompt: models.Prompt) {
  try {
    await SetDefaultPrompt(prompt.id)
    defaultPromptId.value = prompt.id
  } catch (error: any) {
    console.error('Failed to set default:', error)
    alert('设置默认失败: ' + error.message)
  }
}

function isDefault(prompt: models.Prompt): boolean {
  return prompt.id === defaultPromptId.value
}

const totalPrompts = computed(() => prompts.value.length)

onMounted(() => {
  loadPrompts()
})

defineExpose({ loadPrompts })
</script>

<template>
  <div class="prompts-panel">
    <div class="panel-header">
      <div class="header-left">
        <h2>提示词管理</h2>
        <span class="count">{{ totalPrompts }} 个提示词</span>
      </div>
      <div class="header-actions">
        <button @click="loadPrompts" class="btn-refresh" :disabled="isLoading" title="刷新">
          <span v-if="isLoading">⟳</span>
          <span v-else>⟳</span>
        </button>
        <button @click="openCreateDialog" class="btn-create">
          + 新建提示词
        </button>
      </div>
    </div>

    <div v-if="isLoading && prompts.length === 0" class="loading">加载中...</div>

    <div v-else-if="prompts.length === 0" class="empty">
      <p>暂无提示词</p>
      <p class="hint">点击"新建提示词"创建一个新模板</p>
    </div>

    <div v-else class="prompt-list">
      <div
        v-for="prompt in prompts"
        :key="prompt.id"
        class="prompt-item"
        :class="{ 'is-default': isDefault(prompt) }"
      >
        <div class="prompt-header">
          <div class="prompt-info">
            <span class="prompt-name">{{ prompt.name }}</span>
            <span v-if="isDefault(prompt)" class="default-badge">默认</span>
          </div>
          <div class="prompt-actions">
            <button
              v-if="!isDefault(prompt)"
              @click="setDefault(prompt)"
              class="btn-action btn-default"
              title="设为默认"
            >
              设为默认
            </button>
            <button @click="openEditDialog(prompt)" class="btn-action" title="编辑">
              编辑
            </button>
            <button @click="deletePrompt(prompt)" class="btn-action btn-danger" title="删除">
              删除
            </button>
          </div>
        </div>
        <div class="prompt-description" v-if="prompt.description">
          {{ prompt.description }}
        </div>
        <div class="prompt-template">
          <pre>{{ prompt.template }}</pre>
        </div>
      </div>
    </div>

    <!-- Editor Dialog -->
    <div v-if="showEditor" class="dialog-overlay" @click.self="showEditor = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ editingPrompt ? '编辑提示词' : '新建提示词' }}</h3>
          <button @click="showEditor = false" class="btn-close">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>名称 <span class="required">*</span></label>
            <input
              v-model="form.name"
              type="text"
              placeholder="例如：生成提交信息"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>描述</label>
            <input
              v-model="form.description"
              type="text"
              placeholder="简要描述这个提示词的用途"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>模板内容 <span class="required">*</span></label>
            <textarea
              v-model="form.template"
              placeholder="使用 {{.Diff}} 表示差异内容占位符"
              class="form-input template-input"
              rows="8"
            ></textarea>
            <p class="help-text">
              可用变量：<code>{{"{.Diff}"}}</code> - Git 差异内容
            </p>
          </div>
          <div class="form-group">
            <label class="checkbox">
              <input type="checkbox" v-model="form.isDefault" />
              <span>设为默认提示词</span>
            </label>
          </div>
        </div>
        <div class="dialog-footer">
          <button @click="showEditor = false" class="btn-cancel">取消</button>
          <button @click="savePrompt" class="btn-confirm" :disabled="!form.name.trim() || !form.template.trim()">
            保存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.prompts-panel {
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

.prompt-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.prompt-item {
  padding: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  margin-bottom: 0.75rem;
  transition: all 0.2s;
}

.prompt-item:hover {
  border-color: rgba(255, 255, 255, 0.2);
}

.prompt-item.is-default {
  border-color: rgba(97, 218, 251, 0.3);
  background: rgba(97, 218, 251, 0.05);
}

.prompt-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.5rem;
}

.prompt-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.prompt-name {
  font-size: 1rem;
  color: #e5e7eb;
  font-weight: 500;
}

.default-badge {
  font-size: 0.65rem;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  background: rgba(97, 218, 251, 0.2);
  color: #61dafb;
}

.prompt-actions {
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

.btn-action.btn-default {
  border-color: #61dafb;
  color: #61dafb;
}

.btn-action.btn-default:hover {
  background: rgba(97, 218, 251, 0.1);
}

.btn-action.btn-danger:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.prompt-description {
  font-size: 0.85rem;
  color: #999;
  margin-bottom: 0.75rem;
}

.prompt-template {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
  padding: 0.75rem;
  overflow-x: auto;
}

.prompt-template pre {
  margin: 0;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.8rem;
  color: #a5d6ff;
  white-space: pre-wrap;
  word-break: break-all;
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
  min-width: 500px;
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

.template-input {
  font-family: 'Consolas', 'Monaco', monospace;
  resize: vertical;
}

.help-text {
  margin: 0.5rem 0 0 0;
  font-size: 0.8rem;
  color: #888;
}

.help-text code {
  background: rgba(97, 218, 251, 0.1);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  color: #61dafb;
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
