/**
 * 文件列表渲染模块
 * 负责文件列表的渲染和交互
 */

import { formatSize, escapeHtml } from './utils.js';
import { openImagePreview } from './preview.js';

let thumbDelegated = false;

// 缩略图点击事件委托（挂在持久容器上，重渲染不丢失）
function ensureThumbDelegation(fileListEl) {
    if (thumbDelegated) return;
    thumbDelegated = true;
    fileListEl.addEventListener('click', e => {
        const link = e.target.closest('.file-thumb-link');
        if (!link) return;
        const workDir = document.getElementById('pathText').textContent;
        openImagePreview(workDir, link.dataset.path, link.dataset.name);
    });
}

function buildThumbHtml(file, workDir) {
    if (file.type !== 'image') {
        return '<div class="file-thumb-placeholder"></div>';
    }
    const src = `/api/file/image?work_dir=${encodeURIComponent(workDir)}&path=${encodeURIComponent(file.path)}`;
    return `<span class="file-thumb-link" data-path="${escapeHtml(file.path)}" data-name="${escapeHtml(file.name)}">
                <img class="file-thumb" src="${src}" loading="lazy" alt="">
            </span>`;
}

// 渲染文件列表
export function renderFileList(fileList, state) {
    const fileListEl = document.getElementById('fileList');
    ensureThumbDelegation(fileListEl);

    fileListEl.style.display = 'block';

    // 如果文件列表为空,显示空提示
    if (fileList.length === 0) {
        fileListEl.innerHTML = '<div class="empty-result">暂无符合条件的文件</div>';
        return;
    }

    const workDir = document.getElementById('pathText').textContent;

    const maxSize = Math.max(...fileList.map(f => f.size || 0));

    fileListEl.innerHTML = `
        <div class="file-list">
            <div class="file-list-header">
                <input type="checkbox" id="selectAllCheckbox" class="file-checkbox">
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
                                ${buildThumbHtml(file, workDir)}
                                <span class="file-name-text">${escapeHtml(file.name)}</span>
                                ${typeLabels[file.type] || typeLabels['other']}
                            </div>
                            <div class="file-meta">${escapeHtml(file.path)} | 修改时间: ${escapeHtml(file.mod_time)}</div>
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

    // 添加全选复选框事件
    const selectAllCb = document.getElementById('selectAllCheckbox');
    if (selectAllCb) {
        selectAllCb.addEventListener('change', function() {
            toggleSelectAll(this, state);
        });
    }

    // 添加单项复选框事件
    document.querySelectorAll('.file-checkbox[data-index]').forEach(cb => {
        cb.addEventListener('change', function() {
            const index = parseInt(this.dataset.index);
            if (state.filteredFiles[index]) {
                state.filteredFiles[index].selected = this.checked;
            }
            updateCompressPanel(state);
            
            // 更新全选框状态
            const allItems = state.filteredFiles.filter(f => f.type === 'image');
            const selectedItems = allItems.filter(f => f.selected);
            if (selectAllCb) {
                selectAllCb.checked = allItems.length > 0 && allItems.length === selectedItems.length;
                selectAllCb.indeterminate = selectedItems.length > 0 && selectedItems.length < allItems.length;
            }
        });
    });
}

// 全选/取消全选
export function toggleSelectAll(checkbox, state) {
    state.filteredFiles.forEach((file, index) => {
        if (file.type !== 'image') return;
        file.selected = checkbox.checked;
        const cb = document.querySelector(`.file-checkbox[data-index="${index}"]`);
        if (cb) cb.checked = checkbox.checked;
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
