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
    const outputDir = document.getElementById('outputDir').value.trim() || state.outputDir;

    showMessage(`正在压缩 ${selectedImages.length} 个图片...`, 'info');
    showProgress('正在压缩图片...');

    try {
        const response = await compressImages(
            selectedImages.map(f => f.path),
            quality,
            format,
            outputDir
        );

        await handleCompressStream(response);
        hideProgress();
        
        // 压缩完成后刷新文件列表
        const pathText = document.getElementById('pathText').textContent;
        if (pathText !== '未选择目录') {
            await loadFiles(pathText, state);
        }
    } catch (error) {
        hideProgress();
        showMessage('压缩失败：' + error.message, 'error');
    }
}

// 处理压缩流式响应
async function handleCompressStream(response) {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        
        for (const line of lines) {
            if (line.startsWith('progress:')) {
                const progress = parseInt(line.split(':')[1].trim());
                if (!isNaN(progress)) {
                    updateProgress(progress);
                }
            } else if (line.startsWith('data:')) {
                try {
                    const data = JSON.parse(line.substring(5));
                    if (data.success) {
                        let message = `${data.message} | 压缩率：${data.ratio}`;
                        if (data.failed_count > 0) {
                            message += ` | 失败：${data.failed_count} 个`;
                        }
                        showMessage(message, 'success');
                    } else {
                        showMessage('压缩失败：' + data.message, 'error');
                    }
                } catch (e) {
                    console.error('解析响应失败:', e);
                }
            }
        }
        
        buffer = lines[lines.length - 1];
    }
}
