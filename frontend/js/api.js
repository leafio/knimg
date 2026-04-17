/**
 * API 调用模块
 * 封装所有后端 API 请求
 */

// 浏览目录
export async function browseDirectory(path) {
    try {
        const params = new URLSearchParams();
        if (path) params.append('path', path);
        const response = await fetch('/api/directory/browse?' + params.toString());
        const data = await response.json();
        if (data.success) {
            return {
                current_path: data.current_path,
                directories: data.directories
            };
        }
        throw new Error(data.message || '浏览目录失败');
    } catch (error) {
        console.error('浏览目录失败:', error);
        throw error;
    }
}

// 加载文件列表
export async function loadFiles(workDir) {
    try {
        const params = new URLSearchParams();
        params.append('work_dir', workDir);
        
        const response = await fetch('/api/files?' + params.toString());
        const data = await response.json();

        if (data.success) {
            return {
                files: data.data,
                stats: data.stats
            };
        } else {
            throw new Error(data.message || '加载文件失败');
        }
    } catch (error) {
        console.error('加载文件失败:', error);
        throw error;
    }
}

// 筛选文件
export async function filterFiles(params) {
    try {
        const response = await fetch('/api/files?' + params.toString());
        const data = await response.json();
        
        if (data.success) {
            return {
                files: Array.isArray(data.data) ? data.data : [],
                stats: data.stats
            };
        }
        throw new Error(data.message || '筛选失败');
    } catch (error) {
        console.error('筛选失败:', error);
        throw error;
    }
}

// 压缩图片
export async function compressImages(files, quality, format, outputDir) {
    try {
        const response = await fetch('/api/compress', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                files: files,
                quality: quality,
                format: format,
                output_dir: outputDir
            })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || '压缩失败');
        }

        return response;
    } catch (error) {
        console.error('压缩请求失败:', error);
        throw error;
    }
}

// 导出文件
export async function exportFiles(params) {
    try {
        const response = await fetch('/api/files/export?' + params.toString());
        const data = await response.json();

        if (data.success) {
            return data.file_path;
        } else {
            throw new Error(data.message || '导出失败');
        }
    } catch (error) {
        console.error('导出失败:', error);
        throw error;
    }
}
