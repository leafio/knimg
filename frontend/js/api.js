/**
 * API 调用模块
 * 封装所有后端 API 请求
 */

const DEFAULT_TIMEOUT = 60000;

// 带超时的 fetch，防止请求挂起导致界面一直转圈
async function fetchJSON(url, options = {}, timeout = DEFAULT_TIMEOUT) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), timeout);
    try {
        const response = await fetch(url, { ...options, signal: controller.signal });
        const data = await response.json();
        return { response, data };
    } finally {
        clearTimeout(timer);
    }
}

function timeoutError() {
    return new Error('请求超时，目录可能过大或服务无响应');
}

// 浏览目录
export async function browseDirectory(path) {
    const params = new URLSearchParams();
    if (path) params.append('path', path);
    try {
        const { data } = await fetchJSON('/api/directory/browse?' + params.toString());
        if (data.success) {
            return {
                current_path: data.current_path,
                directories: data.directories
            };
        }
        throw new Error(data.message || '浏览目录失败');
    } catch (error) {
        if (error.name === 'AbortError') throw timeoutError();
        throw error;
    }
}

// 加载文件列表
export async function loadFiles(workDir) {
    const params = new URLSearchParams();
    params.append('work_dir', workDir);
    try {
        const { data } = await fetchJSON('/api/files?' + params.toString());
        if (data.success) {
            return {
                files: data.data,
                stats: data.stats
            };
        }
        throw new Error(data.message || '加载文件失败');
    } catch (error) {
        if (error.name === 'AbortError') throw timeoutError();
        throw error;
    }
}

// 筛选文件
export async function filterFiles(params) {
    try {
        const { data } = await fetchJSON('/api/files?' + params.toString());
        if (data.success) {
            return {
                files: Array.isArray(data.data) ? data.data : [],
                stats: data.stats
            };
        }
        throw new Error(data.message || '筛选失败');
    } catch (error) {
        if (error.name === 'AbortError') throw timeoutError();
        throw error;
    }
}

// 压缩图片
export async function compressImages(files, quality, format, workDir, outputDir) {
    let response, data;
    try {
        const result = await fetchJSON('/api/compress', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                files: files,
                quality: quality,
                format: format,
                work_dir: workDir,
                output_dir: outputDir
            })
        }, 600000); // 压缩大图耗时较长，放宽到 10 分钟
        response = result.response;
        data = result.data;
    } catch (error) {
        if (error.name === 'AbortError') throw new Error('压缩超时，请减少单次压缩数量');
        throw error;
    }

    if (!response.ok || !data.success) {
        throw new Error(data.message || '压缩失败');
    }
    return data.data;
}

// 导出文件
export async function exportFiles(params) {
    try {
        const { data } = await fetchJSON('/api/files/export?' + params.toString());
        if (data.success) {
            return data.file_path;
        }
        throw new Error(data.message || '导出失败');
    } catch (error) {
        if (error.name === 'AbortError') throw timeoutError();
        throw error;
    }
}
