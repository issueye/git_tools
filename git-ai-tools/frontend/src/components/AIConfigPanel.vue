<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetAIConfig, SetAIConfig, TestAIConnection } from '/wailsjs/go/main/App'
import type { services } from '/wailsjs/go/models'

type AIProvider = 'openai' | 'claude' | 'ollama'

const emit = defineEmits(['configSaved'])

const config = ref<services.AIConfig>({
  provider: 'openai',
  apiKey: '',
  baseUrl: '',
  model: ''
})

const isLoading = ref(false)
const isTesting = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)
const showApiKey = ref(false)

const providerPresets: Record<string, { baseUrl: string; model: string; label: string }> = {
  openai: {
    baseUrl: 'https://api.openai.com/v1',
    model: 'gpt-4',
    label: 'OpenAI'
  },
  claude: {
    baseUrl: 'https://api.anthropic.com/v1',
    model: 'claude-3-sonnet-20240229',
    label: 'Anthropic Claude'
  },
  ollama: {
    baseUrl: 'http://localhost:11434',
    model: 'llama2',
    label: 'Ollama (Local)'
  }
}

async function loadConfig() {
  isLoading.value = true
  try {
    const result = await GetAIConfig()
    config.value = {
      provider: (result.provider as AIProvider) || 'openai',
      apiKey: result.apiKey || '',
      baseUrl: result.baseUrl || '',
      model: result.model || ''
    }
  } catch (error: any) {
    console.error('Failed to load config:', error)
  } finally {
    isLoading.value = false
  }
}

async function saveConfig() {
  try {
    await SetAIConfig({
      provider: config.value.provider,
      apiKey: config.value.apiKey,
      baseUrl: config.value.baseUrl,
      model: config.value.model
    })
    testResult.value = { success: true, message: 'Configuration saved successfully!' }
    emit('configSaved')
    setTimeout(() => { testResult.value = null }, 3000)
  } catch (error: any) {
    console.error('Save config error:', error)
    const errorMessage = error?.message || error?.toString?.() || 'Unknown error'
    testResult.value = { success: false, message: 'Failed to save: ' + errorMessage }
  }
}

async function testConnection() {
  if (isTesting.value) return

  isTesting.value = true
  testResult.value = null

  try {
    // Test with current configuration from the form
    await TestAIConnection({
      provider: config.value.provider,
      apiKey: config.value.apiKey,
      baseUrl: config.value.baseUrl,
      model: config.value.model
    })
    testResult.value = { success: true, message: 'Connection successful!' }
  } catch (error: any) {
    console.error('Test connection error:', error)
    const errorMessage = error?.message || error?.toString?.() || 'Unknown error'
    testResult.value = { success: false, message: 'Connection failed: ' + errorMessage }
  } finally {
    isTesting.value = false
  }
}

function applyProviderPreset() {
  const preset = providerPresets[config.value.provider]
  if (preset) {
    config.value.baseUrl = preset.baseUrl
    config.value.model = preset.model
  }
}

function onProviderChange() {
  applyProviderPreset()
  testResult.value = null
}

onMounted(() => {
  loadConfig()
})
</script>

<template>
  <div class="ai-config-panel">
    <div class="panel-header">
      <h2>AI Configuration</h2>
      <button @click="loadConfig" class="btn-refresh" :disabled="isLoading" title="Reload">
        <span v-if="isLoading">⟳</span>
        <span v-else>⟳</span>
      </button>
    </div>

    <div v-if="isLoading" class="loading">Loading configuration...</div>

    <div v-else class="config-form">
      <!-- Provider Selection -->
      <div class="form-group">
        <label>AI Provider</label>
        <div class="provider-options">
          <button
            v-for="(preset, key) in providerPresets"
            :key="key"
            @click="config.provider = key as AIProvider; onProviderChange()"
            class="provider-btn"
            :class="{ active: config.provider === key }"
          >
            {{ preset.label }}
          </button>
        </div>
      </div>

      <!-- API Key (not for Ollama) -->
      <div v-if="config.provider !== 'ollama'" class="form-group">
        <label>API Key</label>
        <div class="input-with-toggle">
          <input
            v-model="config.apiKey"
            :type="showApiKey ? 'text' : 'password'"
            placeholder="Enter your API key"
            class="form-input"
          />
          <button @click="showApiKey = !showApiKey" class="btn-toggle">
            {{ showApiKey ? '🙈' : '👁️' }}
          </button>
        </div>
      </div>

      <!-- Base URL -->
      <div class="form-group">
        <label>Base URL</label>
        <input
          v-model="config.baseUrl"
          type="text"
          placeholder="API base URL"
          class="form-input"
        />
        <span class="help-text">
          Default: {{ providerPresets[config.provider].baseUrl }}
        </span>
      </div>

      <!-- Model -->
      <div class="form-group">
        <label>Model</label>
        <input
          v-model="config.model"
          type="text"
          placeholder="Model name"
          class="form-input"
        />
        <span class="help-text">
          Default: {{ providerPresets[config.provider].model }}
        </span>
      </div>

      <!-- Test Result -->
      <div v-if="testResult" class="test-result" :class="{ success: testResult.success, error: !testResult.success }">
        {{ testResult.message }}
      </div>

      <!-- Actions -->
      <div class="actions">
        <button
          @click="testConnection"
          :disabled="isTesting || !config.model"
          class="btn-test"
        >
          <span v-if="isTesting">Testing...</span>
          <span v-else>Test Connection</span>
        </button>
        <button @click="saveConfig" class="btn-save">Save Configuration</button>
      </div>

      <!-- Provider Info -->
      <div class="provider-info">
        <h4>Provider Information</h4>
        <div v-if="config.provider === 'openai'" class="info-content">
          <p><strong>OpenAI</strong> requires an API key from platform.openai.com</p>
          <p>Recommended models: gpt-4, gpt-4-turbo, gpt-3.5-turbo</p>
        </div>
        <div v-else-if="config.provider === 'claude'" class="info-content">
          <p><strong>Anthropic Claude</strong> requires an API key from console.anthropic.com</p>
          <p>Recommended models: claude-3-opus-20240229, claude-3-sonnet-20240229</p>
        </div>
        <div v-else-if="config.provider === 'ollama'" class="info-content">
          <p><strong>Ollama</strong> runs locally and doesn't require an API key</p>
          <p>Make sure Ollama is running at the specified base URL</p>
          <p>Download from: ollama.ai</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-config-panel {
  height: 100%;
  overflow-y: auto;
  padding: 1rem;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.panel-header h2 {
  margin: 0;
  font-size: 1.2rem;
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

.loading {
  text-align: center;
  padding: 2rem;
  color: #888;
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-size: 0.9rem;
  color: #ccc;
  font-weight: 500;
}

.provider-options {
  display: flex;
  gap: 0.5rem;
}

.provider-btn {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: transparent;
  color: #ccc;
  cursor: pointer;
  transition: all 0.2s;
}

.provider-btn:hover {
  background: rgba(255, 255, 255, 0.05);
}

.provider-btn.active {
  border-color: #61dafb;
  background: rgba(97, 218, 251, 0.1);
  color: #61dafb;
}

.form-input {
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

.input-with-toggle {
  display: flex;
  gap: 0.5rem;
}

.input-with-toggle .form-input {
  flex: 1;
}

.btn-toggle {
  padding: 0 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  background: transparent;
  color: #888;
  cursor: pointer;
}

.btn-toggle:hover {
  background: rgba(255, 255, 255, 0.05);
}

.help-text {
  font-size: 0.75rem;
  color: #666;
}

.test-result {
  padding: 0.75rem;
  border-radius: 6px;
  text-align: center;
  font-size: 0.9rem;
}

.test-result.success {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid #22c55e;
  color: #4ade80;
}

.test-result.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid #ef4444;
  color: #f87171;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.btn-test, .btn-save {
  flex: 1;
  padding: 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.btn-test {
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: transparent;
  color: #ccc;
}

.btn-test:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
}

.btn-test:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-save {
  border: 1px solid #22c55e;
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}

.btn-save:hover {
  background: rgba(34, 197, 94, 0.2);
}

.provider-info {
  padding: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
}

.provider-info h4 {
  margin: 0 0 0.5rem 0;
  font-size: 0.9rem;
  color: #888;
}

.info-content {
  font-size: 0.85rem;
  color: #999;
  line-height: 1.5;
}

.info-content p {
  margin: 0.25rem 0;
}

.info-content strong {
  color: #ccc;
}
</style>
