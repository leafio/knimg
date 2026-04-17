/**
 * 工具函数模块
 * 包含通用工具函数和辅助方法
 */

// 防抖函数
export function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 格式化文件大小
export function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
}

// 显示消息提示
export function showMessage(text, type = 'success') {
    const messageContainer = document.getElementById('messageContainer');
    const message = document.createElement('div');
    message.className = `message message-${type}`;
    message.textContent = text;
    messageContainer.appendChild(message);

    setTimeout(() => {
        message.remove();
    }, 5000);
}

// 显示进度覆盖层
export function showProgress(title) {
    document.getElementById('progressTitle').textContent = title;
    document.getElementById('progressBarFill').style.width = '0%';
    document.getElementById('progressText').textContent = '0%';
    document.getElementById('progressOverlay').classList.add('show');
}

// 更新进度
export function updateProgress(percent) {
    document.getElementById('progressBarFill').style.width = percent + '%';
    document.getElementById('progressText').textContent = percent + '%';
}

// 隐藏进度覆盖层
export function hideProgress() {
    setTimeout(() => {
        document.getElementById('progressOverlay').classList.remove('show');
    }, 1000);
}
