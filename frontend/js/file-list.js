/**
 * 文件列表渲染模块
 * 负责文件列表的渲染和交互
 */

import { formatSize } from './utils.js';

// 渲染文件列表
export function renderFileList(fileList, state) {
    const fileListEl = document.getElementById('fileList');
    const emptyStateEl = document.getElementById('emptyState');

    if (fileList.length === 0) {
        // 判断是否已选择目录
        const hasSelectedDir = state.currentBrowsePath && state.currentBrowsePath !== '';
        
        if (!hasSelectedDir) {
            // 未选择目录,显示欢迎界面
            emptyStateEl.style.display = 'block';
            fileListEl.style.display = 'none';
        } else {
            // 已选择目录但无结果,显示空结果提示
            showEmptyResult(emptyStateEl);
            fileListEl.style.display = 'none';
        }
        return;
    }

    emptyStateEl.style.display = 'none';
    fileListEl.style.display = 'block';

    const maxSize = Math.max(...fileList.map(f => f.size || 0));

    fileListEl.innerHTML = `
        <div class="file-list">
            <div class="file-list-header">
                <input type="checkbox" class="file-checkbox" onchange="toggleSelectAll(this)">
                <div class="file-info">文件名</div>
                <div class="file-size">大小</div>
            </div>
            ${fileList.map((file, index) => {
                const isLarge = file.size > 10 * 1024 * 1024; // 大于10MB
                const sizePercent = maxSize > 0 ? (file.size / maxSize * 100) : 0;
                const typeLabels = {
                    'image': '<span class="type-badge type-image">图片</span>',
                    'document': '<span class="type-badge type-document">文档</span>',
                    'video': '<span class="type-badge type-video">视频</span>',
                    'other': '<span class="type-badge type-other">其他</span>'
                };
                
                return `
                    <div class="file-item ${isLarge ? 'large-file' : ''}">
                        <input type="checkbox" class="file-checkbox" data-index="${index}"
                               ${file.type === 'image' ? '' : 'disabled'} ${file.selected ? 'checked' : ''}>
                        <div class="file-info">
                            <div class="file-name">
                                ${file.name}
                                ${typeLabels[file.type] || typeLabels['other']}
                            </div>
                            <div class="file-meta">${file.path} | 修改时间: ${file.mod_time}</div>
                        </div>
                        <div class="file-size">
                            ${formatSize(file.size)}
                            <div class="size-bar">
                                <div class="size-bar-fill" style="width: ${sizePercent}%"></div>
                            </div>
                        </div>
                    </div>
                `;
            }).join('')}
        </div>
    `;

    // 添加复选框事件
    document.querySelectorAll('.file-checkbox:not([disabled])').forEach(cb => {
        cb.addEventListener('change', function() {
            const index = parseInt(this.dataset.index);
            if (state.filteredFiles[index]) {
                state.filteredFiles[index].selected = this.checked;
            }
            updateCompressPanel(state);
        });
    });
}

// 显示空结果提示
function showEmptyResult(emptyStateEl) {
    emptyStateEl.innerHTML = `
        <div class="empty-state-icon">
            <svg width="72" height="72" viewBox="0 0 24 24" fill="none" stroke="#86868B" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="opacity: 0.3;">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                <line x1="8" y1="11" x2="14" y2="11"></line>
            </svg>
        </div>
        <h3>暂无符合条件的文件</h3>
        <p>当前筛选条件下没有找到匹配的文件</p>
        <button class="btn btn-secondary" onclick="clearFilters()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 6px; vertical-align: middle;">
                <polyline points="23 4 23 10 17 10"></polyline>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
            </svg>
            清除筛选条件
        </button>
    `;
}

// 全选/取消全选
export function toggleSelectAll(checkbox, state) {
    document.querySelectorAll('.file-checkbox:not([disabled])').forEach((cb, index) => {
        cb.checked = checkbox.checked;
        if (state.filteredFiles[index]) {
            state.filteredFiles[index].selected = checkbox.checked;
        }
    });
    updateCompressPanel(state);
}

// 更新压缩面板状态
export function updateCompressPanel(state) {
    const selectedImages = state.filteredFiles.filter(f => f.selected && f.type === 'image');
    const count = selectedImages.length;
    
    const compressBtn = document.getElementById('compressBtn');
    compressBtn.textContent = `🗜️ 压缩图片 (${count})`;
    
    if (count > 0) {
        compressBtn.disabled = false;
        compressBtn.style.opacity = '1';
        compressBtn.style.cursor = 'pointer';
    } else {
        compressBtn.disabled = true;
        compressBtn.style.opacity = '0.5';
        compressBtn.style.cursor = 'not-allowed';
    }
}
