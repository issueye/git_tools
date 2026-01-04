<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import {
    GetAllRepositories,
    GetRepository,
    AddRepository,
    UpdateRepository,
    UpdateRepositoryAlias,
    DeleteRepository,
    SearchRepositories,
    SelectDirectory,
    IsValidGitRepository,
    CloneRepository,
    SelectRepository,
} from "/wailsjs/go/main/App";
import type { models } from "/wailsjs/go/models";

const emit = defineEmits(["repo-selected", "repo-cloned"]);

const repositories = ref<models.Repository[]>([]);
const isLoading = ref(false);
const showEditor = ref(false);
const showCloneDialog = ref(false);
const editingRepo = ref<models.Repository | null>(null);
const searchKeyword = ref("");
const selectedRepoPath = ref("");

const form = ref({
    path: "",
    alias: "",
    description: "",
});

const cloneForm = ref({
    url: "",
    path: "",
    branch: "",
});

async function loadRepositories() {
    isLoading.value = true;
    try {
        repositories.value = await GetAllRepositories();
    } catch (error: any) {
        console.error("Failed to load repositories:", error);
        repositories.value = [];
    } finally {
        isLoading.value = false;
    }
}

async function searchRepositories() {
    isLoading.value = true;
    try {
        if (searchKeyword.value.trim()) {
            repositories.value = await SearchRepositories(
                searchKeyword.value.trim(),
            );
        } else {
            repositories.value = await GetAllRepositories();
        }
    } catch (error: any) {
        console.error("Failed to search repositories:", error);
        repositories.value = [];
    } finally {
        isLoading.value = false;
    }
}

async function openSelectDialog() {
    const path = await SelectDirectory();
    if (path) {
        const isValid = await IsValidGitRepository(path);
        if (!isValid) {
            alert("所选路径不是有效的 Git 仓库");
            return;
        }
        // 通知父组件选择仓库
        emit("repo-selected", path);
    }
}

async function openRepository() {
    const path = await SelectDirectory();
    if (path && path !== "") {
        const isValid = await IsValidGitRepository(path);
        if (!isValid) {
            alert("所选路径不是有效的 Git 仓库");
            return;
        }
        await selectRepository(path);
    }
}

async function selectRepository(path: string) {
    try {
        await SelectRepository(path);
        selectedRepoPath.value = path;
        emit("repo-selected", path);
    } catch (error: any) {
        console.error("Failed to select repository:", error);
        alert("选择仓库失败: " + error.message);
    }
}

function openCreateDialog() {
    editingRepo.value = null;
    form.value = { path: "", alias: "", description: "" };
    showEditor.value = true;
}

function openCloneDialog() {
    cloneForm.value = { url: "", path: "", branch: "" };
    showCloneDialog.value = true;
}

async function performClone() {
    if (!cloneForm.value.url.trim()) {
        alert("请输入仓库 URL");
        return;
    }

    try {
        await CloneRepository(
            cloneForm.value.url.trim(),
            cloneForm.value.path.trim(),
            cloneForm.value.branch.trim(),
        );
        showCloneDialog.value = false;
        await loadRepositories();
        emit("repo-cloned");
    } catch (error: any) {
        console.error("Failed to clone repository:", error);
        alert("克隆失败: " + error.message);
    }
}

function openEditDialog(repo: models.Repository) {
    editingRepo.value = repo;
    form.value = {
        path: repo.path,
        alias: repo.alias,
        description: repo.description,
    };
    showEditor.value = true;
}

async function saveRepository() {
    if (!form.value.path.trim()) {
        alert("请选择仓库路径");
        return;
    }

    // Validate path is a git repository
    const isValid = await IsValidGitRepository(form.value.path);
    if (!isValid) {
        alert("所选路径不是有效的 Git 仓库");
        return;
    }

    try {
        if (editingRepo.value) {
            await UpdateRepository(
                editingRepo.value.id,
                form.value.alias.trim(),
                form.value.description.trim(),
            );
        } else {
            await AddRepository(
                form.value.path,
                form.value.alias.trim(),
                form.value.description.trim(),
            );
        }
        showEditor.value = false;
        await loadRepositories();
    } catch (error: any) {
        console.error("Failed to save repository:", error);
        alert("保存失败: " + error.message);
    }
}

async function deleteRepository(repo: models.Repository) {
    if (
        !confirm(
            `确定要从管理列表中删除 "${repo.alias || repo.path}" 吗？\n这不会删除实际的仓库文件。`,
        )
    ) {
        return;
    }

    try {
        await DeleteRepository(repo.id);
        await loadRepositories();
    } catch (error: any) {
        console.error("Failed to delete repository:", error);
        alert("删除失败: " + error.message);
    }
}

function getDisplayName(repo: models.Repository): string {
    return repo.alias || repo.path.split("/").pop() || repo.path;
}

function truncatePath(path: string, maxLength: number = 40): string {
    if (path.length <= maxLength) return path;
    return "..." + path.slice(-maxLength);
}

const totalRepos = computed(() => repositories.value.length);

onMounted(() => {
    loadRepositories();
});

defineExpose({ loadRepositories });
</script>

<template>
    <div class="repositories-panel">
        <div class="panel-header">
            <div class="header-left">
                <h2>仓库管理</h2>
                <span class="count">{{ totalRepos }} 个仓库</span>
            </div>
            <div class="header-actions">
                <button
                    @click="loadRepositories"
                    class="btn-refresh"
                    :disabled="isLoading"
                    title="刷新"
                >
                    <span v-if="isLoading">⟳</span>
                    <span v-else>⟳</span>
                </button>
            </div>
        </div>

        <!-- Action Buttons -->
        <div class="action-row">
            <button
                @click="openRepository"
                class="action-btn primary"
                title="打开本地仓库"
            >
                <span class="action-icon">📂</span>
                <span>打开</span>
            </button>
            <button
                @click="openCloneDialog"
                class="action-btn"
                title="克隆远程仓库"
            >
                <span class="action-icon">📥</span>
                <span>克隆</span>
            </button>
            <div class="action-divider"></div>
            <button
                @click="openCreateDialog"
                class="action-btn"
                title="添加现有仓库到管理"
            >
                <span class="action-icon">➕</span>
                <span>添加</span>
            </button>
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
            <button
                v-if="searchKeyword"
                @click="
                    searchKeyword = '';
                    loadRepositories();
                "
                class="btn-clear"
            >
                ✕
            </button>
        </div>

        <div v-if="isLoading && repositories.length === 0" class="loading">
            加载中...
        </div>

        <div v-else-if="repositories.length === 0" class="empty">
            <p>暂无仓库</p>
            <p class="hint">点击上方按钮打开或添加仓库</p>
        </div>

        <div v-else class="repo-list">
            <div
                v-for="repo in repositories"
                :key="repo.id"
                class="repo-item"
                :class="{ selected: repo.path === selectedRepoPath }"
            >
                <div class="repo-header">
                    <div class="repo-info">
                        <span class="repo-icon">📁</span>
                        <div class="repo-details">
                            <span class="repo-name">{{
                                getDisplayName(repo)
                            }}</span>
                            <span class="repo-path" :title="repo.path">{{
                                truncatePath(repo.path)
                            }}</span>
                        </div>
                    </div>
                    <div class="repo-actions">
                        <button
                            @click="selectRepository(repo.path)"
                            class="btn-action"
                            :class="{
                                'btn-select active':
                                    repo.path === selectedRepoPath,
                            }"
                            :title="
                                repo.path === selectedRepoPath
                                    ? '当前仓库'
                                    : '选择操作此仓库'
                            "
                        >
                            {{
                                repo.path === selectedRepoPath ? "当前" : "选择"
                            }}
                        </button>
                        <button
                            @click="openEditDialog(repo)"
                            class="btn-action"
                            title="编辑仓库信息"
                        >
                            编辑
                        </button>
                        <button
                            @click="deleteRepository(repo)"
                            class="btn-action btn-danger"
                            title="从管理列表删除"
                        >
                            删除
                        </button>
                    </div>
                </div>
                <div class="repo-description" v-if="repo.description">
                    {{ repo.description }}
                </div>
                <div class="repo-meta">
                    <span class="repo-id"
                        >ID: {{ repo.id.substring(0, 8) }}</span
                    >
                    <span class="repo-updated"
                        >更新: {{ repo.updatedAt.split("T")[0] }}</span
                    >
                </div>
            </div>
        </div>

        <!-- Editor Dialog -->
        <div
            v-if="showEditor"
            class="dialog-overlay"
            @click.self="showEditor = false"
        >
            <div class="dialog">
                <div class="dialog-header">
                    <h3>{{ editingRepo ? "编辑仓库" : "添加仓库" }}</h3>
                    <button @click="showEditor = false" class="btn-close">
                        ✕
                    </button>
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
                            <button
                                v-if="!editingRepo"
                                @click="openSelectDialog"
                                class="btn-browse"
                            >
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
                    <button @click="showEditor = false" class="btn-cancel">
                        取消
                    </button>
                    <button
                        @click="saveRepository"
                        class="btn-confirm"
                        :disabled="!form.path.trim()"
                    >
                        保存
                    </button>
                </div>
            </div>
        </div>

        <!-- Clone Dialog -->
        <div
            v-if="showCloneDialog"
            class="dialog-overlay"
            @click.self="showCloneDialog = false"
        >
            <div class="dialog">
                <div class="dialog-header">
                    <h3>克隆仓库</h3>
                    <button @click="showCloneDialog = false" class="btn-close">
                        ✕
                    </button>
                </div>
                <div class="dialog-body">
                    <div class="form-group">
                        <label>仓库 URL <span class="required">*</span></label>
                        <input
                            v-model="cloneForm.url"
                            type="text"
                            placeholder="https://github.com/user/repo.git"
                            class="form-input"
                        />
                    </div>
                    <div class="form-group">
                        <label>本地路径</label>
                        <input
                            v-model="cloneForm.path"
                            type="text"
                            placeholder="留空则自动生成"
                            class="form-input"
                        />
                    </div>
                    <div class="form-group">
                        <label>分支（可选）</label>
                        <input
                            v-model="cloneForm.branch"
                            type="text"
                            placeholder="主分支"
                            class="form-input"
                        />
                    </div>
                </div>
                <div class="dialog-footer">
                    <button @click="showCloneDialog = false" class="btn-cancel">
                        取消
                    </button>
                    <button
                        @click="performClone"
                        class="btn-confirm"
                        :disabled="!cloneForm.url.trim()"
                    >
                        克隆
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

.action-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.action-row .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    padding: 0.5rem 0.85rem;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    color: #ccc;
    cursor: pointer;
    font-size: 0.8rem;
    transition: all 0.2s;
}

.action-row .action-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(97, 218, 251, 0.3);
    color: #fff;
}

.action-row .action-btn.primary {
    background: rgba(97, 218, 251, 0.1);
    border-color: rgba(97, 218, 251, 0.3);
    color: #61dafb;
}

.action-row .action-btn.primary:hover {
    background: rgba(97, 218, 251, 0.2);
    border-color: #61dafb;
}

.action-row .action-icon {
    font-size: 1rem;
}

.action-divider {
    width: 1px;
    height: 24px;
    background: rgba(255, 255, 255, 0.1);
    margin: 0 0.25rem;
}

.btn-refresh,
.btn-create {
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

.loading,
.empty {
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

.repo-item.selected {
    border-color: #61dafb;
    background: rgba(97, 218, 251, 0.05);
}

.repo-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
    justify-content: space-between;
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
    font-family: "Consolas", "Monaco", monospace;
    font-size: 0.75rem;
    color: #888;
}

.repo-actions {
    display: inline-flex;
    gap: 0.35rem;
    flex-direction: row;
}

.btn-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.3rem 0.65rem;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 4px;
    background: transparent;
    color: #999;
    cursor: pointer;
    font-size: 0.72rem;
    transition: all 0.2s;
    white-space: nowrap;
}

.btn-action:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #e5e7eb;
    border-color: rgba(255, 255, 255, 0.25);
}

.btn-action.btn-select {
    border-color: rgba(97, 218, 251, 0.4);
    color: #61dafb;
}

.btn-action.btn-select:hover {
    background: rgba(97, 218, 251, 0.1);
}

.btn-action.btn-select.active {
    background: rgba(97, 218, 251, 0.15);
    border-color: #61dafb;
    color: #fff;
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

.btn-cancel,
.btn-confirm {
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
