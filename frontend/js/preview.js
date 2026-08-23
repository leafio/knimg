/**
 * 图片预览模块
 * 大图查看弹层
 */

let initialized = false;

// 打开大图预览
export function openImagePreview(workDir, path, name) {
    const overlay = document.getElementById('imagePreview');
    const img = document.getElementById('imagePreviewImg');
    const title = document.getElementById('imagePreviewTitle');
    if (!overlay || !img) return;

    title.textContent = name || '';
    img.src = `/api/file/image?work_dir=${encodeURIComponent(workDir)}&path=${encodeURIComponent(path)}&full=1`;
    overlay.classList.add('show');
}

// 关闭大图预览
export function closeImagePreview() {
    const overlay = document.getElementById('imagePreview');
    const img = document.getElementById('imagePreviewImg');
    if (overlay) overlay.classList.remove('show');
    if (img) img.src = '';
}

// 初始化事件（幂等，可重复调用）
export function initPreviewEvents() {
    if (initialized) return;
    initialized = true;

    const overlay = document.getElementById('imagePreview');

    document.getElementById('imagePreviewClose')?.addEventListener('click', closeImagePreview);

    overlay?.addEventListener('click', e => {
        if (e.target === overlay) closeImagePreview();
    });

    document.addEventListener('keydown', e => {
        if (e.key === 'Escape' && overlay?.classList.contains('show')) {
            closeImagePreview();
        }
    });
}
