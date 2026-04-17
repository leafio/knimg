/**
 * 主应用模块
 * 负责应用初始化、状态管理和全局函数导出
 */

import { browseDirectory, loadFiles as apiLoadFiles, exportFiles as apiExportFiles } from './api.js';
import { showMessage, showProgress, hideProgress, debounce } from './utils.js';
import { applyFilters, removeFilterTag, applySizePreset, toggleCustomSize, setSort, setSortOrder, selectExtension, clearFilters } from './filters.js';
import { renderFileList, toggleSelectAll, updateCompressPanel } from './file-list.js';
import { showCompressDialog, closeCompressModal, confirmCompress, executeCompression } from './compress.js';

// 全局状态
export const state = {
    allFiles: [],
    filteredFiles: [],
    currentBrowsePath: '',
    activeDirectoryInput: null,
    currentExportFormat: null,
    sortBy: 'size',
    sortOrder: 'desc',
    compressFormat: '',
    outputDir: ''
};

// DOM 元素缓存
const elements = {
    searchInput: document.getElementById('searchInput'),
    minSize: document.getElementById('minSize'),
    maxSize: document.getElementById('maxSize'),
    fileExt: document.getElementById('fileExt'),
    fileList: document.getElementById('fileList'),
    emptyState: document.getElementById('emptyState'),
    compressPanel: document.getElementById('compressPanel'),
    quality: document.getElementById('quality'),
    qualityValue: document.getElementById('qualityValue'),
    outputDir: document.getElementById('outputDir'),
    directoryModal: document.getElementById('directoryModal'),
    currentPath: document.getElementById('currentPath'),
    directoryList: document.getElementById('directoryList'),
    messageContainer: document.getElementById('messageContainer'),
    progressOverlay: document.getElementById('progressOverlay')
};

// 更新统计信息
export function updateStats(stats) {
    if (!stats) return;
    
    // 更新导航栏统计
    const navbarStats = document.getElementById('navbarStats');
    navbarStats.style.display = 'flex';
    document.getElementById('navStatTotal').textContent = stats.total || 0;
    document.getElementById('navStatImage').textContent = stats.image_count || 0;
    document.getElementById('navStatDoc').textContent = stats.doc_count || 0;
    document.getElementById('navStatVideo').textContent = stats.video_count || 0;
}

// 加载文件列表
export async function loadFiles(workDir) {
    const path = workDir || document.getElementById('pathText').textContent;
    
    if (path === '未选择目录') {
        showMessage('请先选择工作目录', 'error');
        return;
    }

    // 更新当前浏览路径状态
    state.currentBrowsePath = path;

    showProgress('正在加载文件列表...');
    try {
        const result = await apiLoadFiles(path);
        state.allFiles = result.files;
        updateStats(result.stats);
        await applyFilters(state, elements);
    } catch (error) {
        showMessage('加载失败：' + error.message, 'error');
    } finally {
        hideProgress();
    }
}

// 目录浏览器
export async function openDirectoryBrowser(inputId) {
    console.log('📁 openDirectoryBrowser 被调用, inputId:', inputId);
    state.activeDirectoryInput = inputId;
    const isExport = inputId === 'outputDir';
    document.getElementById('modalTitle').textContent = isExport ? '选择输出目录' : '选择工作目录';
    await browseDirectoryInModal('');
    elements.directoryModal.classList.add('show');
    console.log('✅ 模态框已显示');
}

// 在模态框中浏览目录
async function browseDirectoryInModal(path) {
    try {
        const result = await browseDirectory(path);
        state.currentBrowsePath = result.current_path;
        elements.currentPath.textContent = state.currentBrowsePath;
        renderDirectoryList(result.directories);
    } catch (error) {
        console.error('浏览目录失败:', error);
    }
}

// 渲染目录列表
function renderDirectoryList(directories) {
    elements.directoryList.innerHTML = directories.map(dir => `
        <div class="directory-item ${dir === '..' ? 'parent' : ''}" onclick="navigateToDirectory('${dir}')">
            <span class="directory-icon">
                ${dir === '..' ? 
                    '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#007AFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 14l-4-4 4-4"/><path d="M5 10h11a4 4 0 1 1 0 8h-1"/></svg>..' : 
                    '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#007AFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>'}
            </span>
            <span class="directory-name">${dir === '..' ? '上级目录' : dir}</span>
        </div>
    `).join('');
}

// 导航到目录
export async function navigateToDirectory(dirName) {
    const newPath = dirName === '..' 
        ? state.currentBrowsePath.split('/').slice(0, -1).join('/') || '/'
        : state.currentBrowsePath === '/' ? '/' + dirName : state.currentBrowsePath + '/' + dirName;
    await browseDirectoryInModal(newPath);
}

// 关闭目录浏览器
export function closeDirectoryBrowser() {
    elements.directoryModal.classList.remove('show');
    state.activeDirectoryInput = null;
}

// 选择目录
export function selectDirectory() {
    if (!state.activeDirectoryInput || !state.currentBrowsePath) return;
    
    if (state.activeDirectoryInput === 'outputDir') {
        if (state.currentExportFormat) {
            // 导出场景
            closeDirectoryBrowser();
            performExport(state.currentBrowsePath);
        } else {
            elements.outputDir.value = state.currentBrowsePath;
            closeDirectoryBrowser();
        }
    } else {
        // 工作目录
        document.getElementById('pathText').textContent = state.currentBrowsePath;
        closeDirectoryBrowser();
        
        console.log('📁 选择目录 - 隐藏欢迎界面');
        console.log('📁 welcomeState:', document.getElementById('welcomeState'));
        console.log('📁 fileList:', document.getElementById('fileList'));
        
        // 直接控制DOM显示/隐藏
        const welcomeEl = document.getElementById('welcomeState');
        const fileListEl = document.getElementById('fileList');
        
        if (welcomeEl) {
            welcomeEl.style.display = 'none';
            console.log('✅ 已隐藏欢迎界面');
        } else {
            console.error('❌ 找不到 welcomeState 元素');
        }
        
        if (fileListEl) {
            fileListEl.style.display = 'block';
            console.log('✅ 已显示文件列表');
        } else {
            console.error('❌ 找不到 fileList 元素');
        }
        
        // 直接传递路径给loadFiles
        console.log('📁 开始加载文件:', state.currentBrowsePath);
        loadFiles(state.currentBrowsePath);
    }
}

// 执行导出
async function performExport(exportDir) {
    const format = state.currentExportFormat;
    if (!format) return;

    showProgress(`正在导出 ${format.toUpperCase()} 文件...`);

    const params = new URLSearchParams();
    const path = document.getElementById('pathText').textContent;
    params.append('work_dir', path);
    if (elements.searchInput.value) params.append('search', elements.searchInput.value);
    if (elements.fileExt.value) params.append('file_ext', elements.fileExt.value);
    const minSizeMB = parseFloat(elements.minSize.value) || 0;
    const maxSizeMB = parseFloat(elements.maxSize.value) || 0;
    if (minSizeMB > 0) params.append('min_size', minSizeMB * 1024 * 1024);
    if (maxSizeMB > 0) params.append('max_size', maxSizeMB * 1024 * 1024);
    params.append('sort_by', state.sortBy);
    params.append('sort_order', state.sortOrder);
    params.append('format', format);
    if (exportDir) params.append('export_dir', exportDir);

    try {
        const filePath = await apiExportFiles(params);
        showMessage(`导出成功：${filePath}`, 'success');
    } catch (error) {
        showMessage('导出失败：' + error.message, 'error');
    } finally {
        hideProgress();
    }
}

// 切换导出菜单
export function toggleExportMenu() {
    const menu = document.getElementById('exportMenu');
    menu.classList.toggle('show');
}

// 导出文件
export function exportFiles(format) {
    state.currentExportFormat = format;
    state.activeDirectoryInput = 'outputDir';
    document.getElementById('modalTitle').textContent = '选择导出目录';
    browseDirectoryInModal('').then(() => {
        elements.directoryModal.classList.add('show');
        document.getElementById('exportMenu').classList.remove('show');
    });
}

// 设置压缩格式
export function setFormat(format) {
    state.compressFormat = format;
    document.getElementById('compressFormatSelect').value = format;
}

// 初始化应用
export function init() {
    // 等待 DOM 完全加载
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initializeApp);
    } else {
        initializeApp();
    }
}

// 实际的应用初始化逻辑
function initializeApp() {
    // 搜索输入实时筛选
    elements.searchInput.addEventListener('input', debounce(() => applyFilters(state, elements), 300));
    elements.fileExt.addEventListener('input', debounce(() => applyFilters(state, elements), 300));
    elements.minSize.addEventListener('input', debounce(() => applyFilters(state, elements), 300));
    elements.maxSize.addEventListener('input', debounce(() => applyFilters(state, elements), 300));

    // 质量滑块
    elements.quality.addEventListener('input', function() {
        document.getElementById('qualityValue').textContent = this.value + '%';
    });

    // 监听下拉框变化
    document.getElementById('compressFormatSelect').addEventListener('change', function() {
        state.compressFormat = this.value;
    });

    // 监听压缩选项变化
    const compressOptions = document.querySelectorAll('input[name="compressOption"]');
    compressOptions.forEach(option => {
        option.addEventListener('change', function() {
            const outputDirSelector = document.getElementById('outputDirSelector');
            if (this.value === 'newdir') {
                outputDirSelector.style.display = 'block';
            } else {
                outputDirSelector.style.display = 'none';
            }
        });
    });

    // 点击模态框外部关闭
    elements.directoryModal.addEventListener('click', function(e) {
        if (e.target === elements.directoryModal) closeDirectoryBrowser();
    });

    // 点击外部关闭导出菜单
    document.addEventListener('click', function(e) {
        const exportMenu = document.getElementById('exportMenu');
        if (!e.target.closest('.dropdown')) {
            exportMenu.classList.remove('show');
        }
    });

    // 将函数暴露到全局作用域,供 HTML 中的 onclick 调用
    window.openDirectoryBrowser = openDirectoryBrowser;
    window.closeDirectoryBrowser = closeDirectoryBrowser;
    window.navigateToDirectory = navigateToDirectory;
    window.selectDirectory = selectDirectory;
    window.loadFiles = loadFiles;
    window.applySizePreset = (sizeMB) => applySizePreset(sizeMB, state, elements);
    window.toggleCustomSize = toggleCustomSize;
    window.setSort = (field) => setSort(field, state, elements);
    window.setSortOrder = (order) => setSortOrder(order, state, elements);
    window.selectExtension = (ext) => selectExtension(ext, elements);
    window.clearFilters = () => clearFilters(state, elements);
    window.removeFilterTag = (type) => removeFilterTag(type, state, elements);
    window.toggleSelectAll = (cb) => toggleSelectAll(cb, state);
    window.toggleExportMenu = toggleExportMenu;
    window.exportFiles = exportFiles;
    window.setFormat = setFormat;
    window.showCompressDialog = showCompressDialog;
    window.closeCompressModal = closeCompressModal;
    window.confirmCompress = () => confirmCompress(state);
    window.compressFiles = () => executeCompression(state);
    
    console.log('✅ KnImg 应用初始化完成');
}

// 启动应用
init();
