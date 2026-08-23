/**
 * 压缩功能模块
 * 处理图片压缩相关逻辑
 */

import { compressImages } from './api.js';
import { showMessage, showProgress, updateProgress, hideProgress } from './utils.js';
import { loadFiles } from './app.js';

// 执行压缩
export async function executeCompression(state) {
    const selectedImages = state.filteredFiles.filter(f => f.selected && f.type === 'image');

    if (selectedImages.length === 0) {
        showMessage('请先选择要压缩的图片', 'error');
        return;
    }

    const quality = parseInt(document.getElementById('quality').value);
    const format = state.compressFormat;
    const workDir = document.getElementById('pathText').textContent;
    const outputDir = document.getElementById('outputDir').value.trim() || '';

    if (workDir === '未选择目录') {
        showMessage('请先选择工作目录', 'error');
        return;
    }

    showMessage(`正在压缩 ${selectedImages.length} 个图片...`, 'info');
    showProgress('正在压缩图片...');
    updateProgress(50);

    try {
        const result = await compressImages(
            selectedImages.map(f => f.path),
            quality,
            format,
            workDir,
            outputDir
        );

        updateProgress(100);

        let ratio = 0;
        if (result.orig_size > 0) {
            ratio = (1 - result.new_size / result.orig_size) * 100;
        }

        let message = `${result.message}，节省 ${ratio.toFixed(1)}% 空间`;
        if (result.failed_count > 0) {
            message += `，失败 ${result.failed_count} 个`;
        }
        showMessage(message, 'success');

        // 压缩完成后刷新文件列表
        await loadFiles(workDir);
    } catch (error) {
        showMessage('压缩失败：' + error.message, 'error');
    } finally {
        hideProgress();
    }
}
