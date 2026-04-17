/**
 * 筛选功能模块
 * 处理文件筛选、排序和过滤逻辑
 */

import { filterFiles } from './api.js';
import { renderFileList } from './file-list.js';
import { updateStats } from './app.js';

// 构建筛选参数
export function buildFilterParams(state, elements) {
    const path = document.getElementById('pathText').textContent;
    const minSizeMB = parseFloat(elements.minSize.value) || 0;
    const maxSizeMB = parseFloat(elements.maxSize.value) || 0;

    const params = new URLSearchParams();
    params.append('work_dir', path);
    if (elements.searchInput.value) params.append('search', elements.searchInput.value);
    if (elements.fileExt.value) params.append('file_ext', elements.fileExt.value);
    if (minSizeMB > 0) params.append('min_size', minSizeMB * 1024 * 1024);
    if (maxSizeMB > 0) params.append('max_size', maxSizeMB * 1024 * 1024);
    params.append('sort_by', state.sortBy);
    params.append('sort_order', state.sortOrder);
    
    return params;
}

// 应用筛选
export async function applyFilters(state, elements) {
    const params = buildFilterParams(state, elements);
    try {
        const result = await filterFiles(params);
        state.filteredFiles = result.files;
        renderFileList(state.filteredFiles, state);
        updateStats(result.stats);
        updateFilterTags(elements);
    } catch (error) {
        console.error('筛选失败:', error);
    }
}

// 更新筛选标签显示
export function updateFilterTags(elements) {
    const container = document.getElementById('filterTags');
    if (!container) return;
    
    const tags = [];

    if (elements.searchInput.value) {
        tags.push({ label: `名称: ${elements.searchInput.value}`, type: 'search' });
    }
    if (elements.fileExt.value) {
        tags.push({ label: `类型: ${elements.fileExt.value}`, type: 'ext' });
    }
    if (elements.minSize.value) {
        tags.push({ label: `> ${elements.minSize.value}MB`, type: 'minSize' });
    }
    if (elements.maxSize.value) {
        tags.push({ label: `< ${elements.maxSize.value}MB`, type: 'maxSize' });
    }

    container.innerHTML = tags.map(tag => `
        <span class="filter-tag">
            ${tag.label}
            <span class="close" onclick="removeFilterTag('${tag.type}')">×</span>
        </span>
    `).join('');
}

// 移除筛选标签
export function removeFilterTag(type, state, elements) {
    switch(type) {
        case 'search': elements.searchInput.value = ''; break;
        case 'ext': elements.fileExt.value = ''; break;
        case 'minSize': elements.minSize.value = ''; break;
        case 'maxSize': elements.maxSize.value = ''; break;
    }
    applyFilters(state, elements);
}

// 快速筛选预设
export function applySizePreset(sizeMB, state, elements) {
    // 清除其他预设的 active 状态
    document.querySelectorAll('.preset-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');

    elements.minSize.value = sizeMB;
    elements.maxSize.value = '';
    
    // 确保自定义输入框显示
    document.getElementById('customSizeInput').style.display = 'block';
    
    applyFilters(state, elements);
}

// 切换自定义大小输入框
export function toggleCustomSize() {
    const customInput = document.getElementById('customSizeInput');
    customInput.style.display = customInput.style.display === 'none' ? 'block' : 'none';
}

// 设置排序字段
export function setSort(field, state, elements) {
    state.sortBy = field;
    // 只清除排序字段组的按钮状态
    document.querySelectorAll('#sortFieldGroup .sort-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    applyFilters(state, elements);
}

// 设置排序顺序
export function setSortOrder(order, state, elements) {
    state.sortOrder = order;
    // 只清除排序方向组的按钮状态
    document.querySelectorAll('#sortOrderGroup .sort-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    applyFilters(state, elements);
}

// 选择扩展名快捷方式
export function selectExtension(extensions, elements) {
    elements.fileExt.value = extensions;
    applyFilters(null, elements);
}

// 清除所有筛选
export function clearFilters(state, elements) {
    elements.searchInput.value = '';
    elements.fileExt.value = '';
    elements.minSize.value = '';
    elements.maxSize.value = '';
    document.getElementById('customSizeInput').style.display = 'none';
    document.querySelectorAll('.preset-btn').forEach(btn => btn.classList.remove('active'));
    state.sortBy = 'name';
    state.sortOrder = 'asc';
    
    // 重置排序按钮
    document.querySelectorAll('.sort-btn').forEach((btn, index) => {
        btn.classList.toggle('active', index === 0 || index === 4);
    });
    
    applyFilters(state, elements);
}
