/**
 * 文件列表渲染模块
 * 负责文件列表的渲染和交互
 */

import { formatSize } from './utils.js';

// 渲染文件列表
export function renderFileList(fileList, state) {
    const fileListEl = document.getElementById('fileList');

    console.log('🎨 renderFileList 被调用');
    console.log('🎨 fileList 长度:', fileList.length);
    console.log('🎨 fileListEl:', fileListEl);

    // 显示文件列表
    fileListEl.style.display = 'block';
    console.log('✅ fileListEl.style.display 设置为 block');

    // 如果文件列表为空,显示空提示
    if (fileList.length === 0) {
        fileListEl.innerHTML = '<div class="empty-result">暂无符合条件的文件</div>';
        console.log('📝 文件列表为空，显示空提示');
        return;
    }

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
