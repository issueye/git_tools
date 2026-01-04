<script lang="ts" setup>
import { ref } from 'vue'
import { Commit, GenerateCommitMessage } from '/wailsjs/go/main/App'

const emit = defineEmits(['committed'])
const props = defineProps<{
  hasStagedChanges: boolean
}>()

const message = ref('')
const isGenerating = ref(false)
const isCommitting = ref(false)

async function generateMessage() {
  if (isGenerating.value) return

  isGenerating.value = true
  try {
    const result = await GenerateCommitMessage()
    if (result) {
      message.value = result
    }
  } catch (error: any) {
    console.error('Failed to generate message:', error)
    alert('Failed to generate commit message: ' + error.message)
  } finally {
    isGenerating.value = false
  }
}

async function commit() {
  if (!message.value.trim()) {
    alert('Please enter a commit message')
    return
  }

  if (isCommitting.value) return

  isCommitting.value = true
  try {
    await Commit(message.value)
    message.value = ''
    emit('committed')
  } catch (error: any) {
    console.error('Failed to commit:', error)
    alert('Failed to commit: ' + error.message)
  } finally {
    isCommitting.value = false
  }
}

function handleKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    commit()
  }
}
</script>

<template>
  <div class="commit-panel">
    <div class="panel-header">
      <h2>Commit</h2>
      <button
        @click="generateMessage"
        :disabled="!hasStagedChanges || isGenerating"
        class="btn-generate"
        :class="{ loading: isGenerating }"
      >
        <span v-if="isGenerating">Generating...</span>
        <span v-else>✨ AI Generate</span>
      </button>
    </div>

    <div class="message-area">
      <textarea
        v-model="message"
        @keydown="handleKeydown"
        placeholder="Enter commit message...
Conventional Commits format: type(scope): description

Types: feat, fix, docs, style, refactor, test, chore"
        rows="8"
        :disabled="isCommitting"
      />
      <div class="message-footer">
        <span class="hint">Ctrl/Cmd + Enter to commit</span>
        <span class="char-count">{{ message.length }}</span>
      </div>
    </div>

    <button
      @click="commit"
      :disabled="!message.trim() || isCommitting"
      class="btn-commit"
      :class="{ loading: isCommitting }"
    >
      <span v-if="isCommitting">Committing...</span>
      <span v-else>Commit Changes</span>
    </button>
  </div>
</template>

<style scoped>
.commit-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  height: 100%;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-header h2 {
  margin: 0;
  font-size: 1.2rem;
  color: #fff;
}

.btn-generate {
  padding: 0.5rem 1rem;
  border: 1px solid #9333ea;
  border-radius: 6px;
  background: rgba(147, 51, 234, 0.1);
  color: #c084fc;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.btn-generate:hover:not(:disabled) {
  background: rgba(147, 51, 234, 0.2);
}

.btn-generate:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-generate.loading {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.message-area {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.message-area textarea {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  color: #e5e7eb;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.9rem;
  resize: none;
  line-height: 1.5;
}

.message-area textarea:focus {
  outline: none;
  border-color: #61dafb;
}

.message-area textarea:disabled {
  opacity: 0.6;
}

.message-area textarea::placeholder {
  color: #666;
}

.message-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  font-size: 0.8rem;
  color: #666;
}

.btn-commit {
  padding: 0.75rem;
  border: 1px solid #22c55e;
  border-radius: 6px;
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 1rem;
  font-weight: 600;
}

.btn-commit:hover:not(:disabled) {
  background: rgba(34, 197, 94, 0.2);
}

.btn-commit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-commit.loading {
  animation: pulse 1.5s infinite;
}
</style>
