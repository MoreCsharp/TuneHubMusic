// ==================== 歌曲列表统一模板 ====================

/**
 * 渲染单个歌曲项
 * @param {Object} item - 歌曲数据
 * @param {number} index - 索引
 * @param {Object} options - 配置选项
 * @param {string} options.type - 列表类型: 'search' | 'toplist' | 'playlist' | 'library'
 * @param {string} options.source - 音乐源
 * @param {boolean} options.isDownloaded - 是否已下载
 * @param {boolean} options.showCheckbox - 是否显示复选框 (默认true)
 * @param {boolean} options.showArtist - 是否显示艺术家 (默认true)
 * @param {boolean} options.showAlbum - 是否显示专辑 (默认true)
 * @param {boolean} options.showTypes - 是否显示类型标签 (默认false)
 * @param {boolean} options.showDownloadBtn - 是否显示下载按钮 (默认true)
 */
function renderSongItem(item, index, options = {}) {
    const {
        type = 'search',
        source = '',
        isDownloaded = false,
        showCheckbox = true,
        showArtist = true,
        showAlbum = true,
        showTypes = false,
        showDownloadBtn = true
    } = options;

    // 复选框
    const checkboxHtml = showCheckbox
        ? `<input type="checkbox" class="song-checkbox" data-type="${type}" data-index="${index}"
            ${isDownloaded ? 'disabled' : ''} onchange="toggleSongSelect('${type}', ${index})">`
        : '';

    // 副标题：艺术家或类型
    let subtitleHtml = '';
    if (showTypes && item.types && item.types.length > 0) {
        subtitleHtml = `<div class="song-subtitle">${item.types.join(' / ')}</div>`;
    } else if (showArtist) {
        subtitleHtml = `<div class="song-subtitle">${item.artist || ''}</div>`;
    }

    // 专辑
    const albumHtml = showAlbum && item.album
        ? `<span class="album">${item.album}</span>`
        : '';

    // 下载按钮或已下载标签
    let actionHtml = '';
    if (isDownloaded) {
        actionHtml = '<span class="downloaded-tag">已下载</span>';
    } else if (showDownloadBtn) {
        const artist = (item.artist || '').replace(/'/g, "\\'");
        const name = (item.name || '').replace(/'/g, "\\'");
        const album = (item.album || '').replace(/'/g, "\\'");
        actionHtml = `<button class="download-btn" onclick="downloadSong('${source}', '${item.id}', '${name}', '${artist}', '${album}')">下载</button>`;
    }

    return `
        <div class="song-item">
            ${checkboxHtml}
            <span class="index">${index + 1}</span>
            <div class="song-info">
                <div class="song-name">${item.name}</div>
                ${subtitleHtml}
            </div>
            ${albumHtml}
            ${actionHtml}
        </div>
    `;
}

/**
 * 渲染歌曲列表
 * @param {Array} songs - 歌曲数组
 * @param {Object} downloaded - 已下载状态映射 {id: boolean}
 * @param {Object} options - 配置选项 (同 renderSongItem)
 */
function renderSongList(songs, downloaded = {}, options = {}) {
    if (!songs || songs.length === 0) {
        return '<div class="no-results">暂无歌曲</div>';
    }
    return songs.map((item, index) => {
        const isDownloaded = downloaded[item.id] || false;
        return renderSongItem(item, index, { ...options, isDownloaded });
    }).join('');
}

// ==================== 工具函数 ====================

// Toast 提示
function toast(message, type = 'info', duration = 3000) {
    const container = document.getElementById('toast-container');
    const el = document.createElement('div');
    el.className = `toast ${type}`;
    el.textContent = message;
    container.appendChild(el);

    setTimeout(() => {
        el.classList.add('hide');
        setTimeout(() => el.remove(), 300);
    }, duration);
}

// 自定义确认弹窗
function showConfirm(message, title = '确认') {
    return new Promise((resolve) => {
        const modal = document.getElementById('confirm-modal');
        const titleEl = document.getElementById('confirm-title');
        const messageEl = document.getElementById('confirm-message');
        const okBtn = document.getElementById('confirm-ok');
        const cancelBtn = document.getElementById('confirm-cancel');

        titleEl.textContent = title;
        messageEl.textContent = message;
        modal.classList.add('show');

        const cleanup = () => {
            modal.classList.remove('show');
            okBtn.removeEventListener('click', onOk);
            cancelBtn.removeEventListener('click', onCancel);
        };

        const onOk = () => { cleanup(); resolve(true); };
        const onCancel = () => { cleanup(); resolve(false); };

        okBtn.addEventListener('click', onOk);
        cancelBtn.addEventListener('click', onCancel);
    });
}

// 自定义下拉组件
function initCustomSelects() {
    document.querySelectorAll('.custom-select').forEach(select => {
        const trigger = select.querySelector('.custom-select-trigger');
        const options = select.querySelectorAll('.custom-select-option');
        const valueDisplay = select.querySelector('.custom-select-value');

        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            // 关闭其他下拉
            document.querySelectorAll('.custom-select.open').forEach(s => {
                if (s !== select) s.classList.remove('open');
            });
            select.classList.toggle('open');
        });

        options.forEach(option => {
            option.addEventListener('click', (e) => {
                e.stopPropagation();
                const value = option.dataset.value;
                const text = option.textContent;

                // 更新选中状态
                options.forEach(o => o.classList.remove('selected'));
                option.classList.add('selected');

                // 更新显示和值
                valueDisplay.textContent = text;
                select.dataset.value = value;
                select.classList.remove('open');

                // 触发自定义change事件
                select.dispatchEvent(new CustomEvent('change', { detail: { value, text } }));
            });
        });
    });

    // 点击外部关闭下拉
    document.addEventListener('click', () => {
        document.querySelectorAll('.custom-select.open').forEach(s => {
            s.classList.remove('open');
        });
    });
}

// 获取下拉组件的值
function getSelectValue(id) {
    const wrapper = document.getElementById(id);
    return wrapper ? wrapper.dataset.value : null;
}

// 设置下拉组件的值
function setSelectValue(id, value) {
    const wrapper = document.getElementById(id);
    if (!wrapper) return;

    const option = wrapper.querySelector(`.custom-select-option[data-value="${value}"]`);
    if (option) {
        wrapper.querySelectorAll('.custom-select-option').forEach(o => o.classList.remove('selected'));
        option.classList.add('selected');
        wrapper.querySelector('.custom-select-value').textContent = option.textContent;
        wrapper.dataset.value = value;
    }
}

document.addEventListener('DOMContentLoaded', function() {
    initCustomSelects();
    initSearch();
    initNav();
    initSettings();
    initLibrary();
    initBatchActions();
    initPlaylist();
    loadSettings();
    startDownloadPolling();
    loadToplists();

    // 切换音乐源时重新加载排行榜
    document.getElementById('source-select-wrapper').addEventListener('change', function() {
        const homeSection = document.getElementById('home-section');
        if (homeSection && homeSection.style.display !== 'none') {
            loadToplists();
        }
    });
});

// 初始化音乐库
function initLibrary() {
    const refreshBtn = document.getElementById('refresh-library-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', refreshLibrary);
    }
}

// 批量操作相关
let searchSelectedSongs = [];
let toplistSelectedSongs = [];
let currentSearchResults = [];
let currentToplistSongs = [];
let currentToplistSource = '';

function initBatchActions() {
    // 搜索结果批量操作
    document.getElementById('search-select-all').addEventListener('change', function() {
        toggleSelectAll('search', this.checked);
    });
    document.getElementById('search-batch-download').addEventListener('click', function() {
        batchDownload('search');
    });

    // 排行榜批量操作
    document.getElementById('toplist-select-all').addEventListener('change', function() {
        toggleSelectAll('toplist', this.checked);
    });
    document.getElementById('toplist-batch-download').addEventListener('click', function() {
        batchDownload('toplist');
    });
}

// 歌单相关变量
let importedPlaylists = [];
let currentPlaylistDetail = null;
let playlistSelectedSongs = [];
let currentPlaylistSongs = [];

// 初始化歌单功能
function initPlaylist() {
    const importBtn = document.getElementById('import-playlist-btn');
    const backBtn = document.getElementById('back-to-playlists');
    const idInput = document.getElementById('playlist-id-input');

    if (importBtn) {
        importBtn.addEventListener('click', importPlaylist);
    }
    if (backBtn) {
        backBtn.addEventListener('click', showPlaylistGrid);
    }
    if (idInput) {
        idInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') importPlaylist();
        });
    }

    // 歌单批量操作
    document.getElementById('playlist-select-all').addEventListener('change', function() {
        toggleSelectAll('playlist', this.checked);
    });
    document.getElementById('playlist-batch-download').addEventListener('click', function() {
        batchDownload('playlist');
    });
}

// 加载已导入的歌单列表
async function loadPlaylists() {
    try {
        const resp = await fetch('/api/v1/playlists');
        const data = await resp.json();
        if (data.code === 200) {
            importedPlaylists = data.data || [];
            renderPlaylists();
        }
    } catch (err) {
        console.error('加载歌单失败');
    }
}

// 渲染歌单列表
function renderPlaylists() {
    const grid = document.getElementById('playlist-grid');
    if (importedPlaylists.length === 0) {
        grid.innerHTML = '<div class="no-results">暂无导入的歌单，请输入歌单ID导入</div>';
        return;
    }
    grid.innerHTML = importedPlaylists.map(item => `
        <div class="playlist-card" onclick="openPlaylist('${item.id}', '${item.source}')">
            <div class="playlist-card-content">
                <div class="playlist-title">${item.name}</div>
                <div class="playlist-meta">${item.author} · ${item.songs.length} 首</div>
            </div>
            <button class="delete-playlist-btn" onclick="event.stopPropagation(); deletePlaylist('${item.id}', '${item.source}')">删除</button>
        </div>
    `).join('');
}

// 导入歌单
async function importPlaylist() {
    const source = getSelectValue('playlist-source-wrapper');
    const id = document.getElementById('playlist-id-input').value.trim();

    if (!id) {
        toast('请输入歌单ID', 'warning');
        return;
    }

    const btn = document.getElementById('import-playlist-btn');
    btn.textContent = '导入中...';
    btn.disabled = true;

    try {
        const resp = await fetch(`/api/v1/playlist/import?source=${source}&id=${id}`);
        const data = await resp.json();

        if (data.code === 200) {
            toast('歌单导入成功', 'success');
            document.getElementById('playlist-id-input').value = '';
            loadPlaylists();
        } else {
            toast('导入失败: ' + (data.message || '未知错误'), 'error');
        }
    } catch (err) {
        toast('请求失败', 'error');
    } finally {
        btn.textContent = '导入歌单';
        btn.disabled = false;
    }
}

// 打开歌单详情
function openPlaylist(id, source) {
    const playlist = importedPlaylists.find(p => p.id === id && p.source === source);
    if (!playlist) return;

    currentPlaylistDetail = playlist;
    document.getElementById('playlist-grid').style.display = 'none';
    document.querySelector('.playlist-import-form').style.display = 'none';
    document.getElementById('playlist-detail').style.display = 'block';
    document.getElementById('playlist-detail-name').textContent = playlist.name;
    document.getElementById('playlist-detail-author').textContent = playlist.author;

    renderPlaylistSongs(playlist);
}

// 渲染歌单歌曲列表
async function renderPlaylistSongs(playlist) {
    const list = document.getElementById('playlist-songs');
    list.innerHTML = '<div class="loading">加载中...</div>';

    // 重置选择状态
    playlistSelectedSongs = [];
    currentPlaylistSongs = playlist.songs.map(s => ({ ...s, source: playlist.source }));
    document.getElementById('playlist-select-all').checked = false;
    updateSelectedCount('playlist');

    // 检查已下载状态
    const ids = playlist.songs.map(s => s.id).join(',');
    let downloaded = {};
    try {
        const resp = await fetch(`/api/v1/downloaded?source=${playlist.source}&ids=${ids}`);
        const data = await resp.json();
        downloaded = data.data || {};
    } catch (e) {}

    list.innerHTML = renderSongList(playlist.songs, downloaded, {
        type: 'playlist',
        source: playlist.source,
        showCheckbox: true,
        showArtist: true,
        showAlbum: true,
        showDownloadBtn: true
    });
}

// 搜索功能
function initSearch() {
    const searchBtn = document.getElementById('search-btn');
    const searchInput = document.getElementById('search-input');

    searchBtn.addEventListener('click', doSearch);
    searchInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') doSearch();
    });
}

async function doSearch() {
    const source = getSelectValue('source-select-wrapper');
    const keyword = document.getElementById('search-input').value.trim();

    if (!keyword) return;

    const searchBtn = document.getElementById('search-btn');
    searchBtn.textContent = '搜索中...';
    searchBtn.disabled = true;

    try {
        const url = `/api/v1/search?source=${source}&keyword=${encodeURIComponent(keyword)}`;
        const resp = await fetch(url);
        const data = await resp.json();

        if (data.code === 200 && data.data) {
            showSearchResults(data.data.results || []);
        } else {
            toast('搜索失败: ' + (data.message || '未知错误'), 'error');
        }
    } catch (err) {
        toast('请求失败', 'error');
    } finally {
        searchBtn.textContent = '搜索';
        searchBtn.disabled = false;
    }
}

async function showSearchResults(results) {
    const section = document.getElementById('search-results-section');
    const list = document.getElementById('search-results');
    const source = getSelectValue('source-select-wrapper');
    const batchActions = document.getElementById('search-batch-actions');

    // 重置选择状态
    searchSelectedSongs = [];
    currentSearchResults = results.map(r => ({ ...r, source }));
    document.getElementById('search-select-all').checked = false;
    updateSelectedCount('search');

    if (results.length === 0) {
        list.innerHTML = '<div class="no-results">未找到相关歌曲</div>';
        batchActions.style.display = 'none';
        section.style.display = 'block';
        return;
    }

    // 检查已下载状态
    const ids = results.map(r => r.id).join(',');
    let downloaded = {};
    try {
        const resp = await fetch(`/api/v1/downloaded?source=${source}&ids=${ids}`);
        const data = await resp.json();
        downloaded = data.data || {};
    } catch (e) {}

    list.innerHTML = renderSongList(results, downloaded, {
        type: 'search',
        source: source,
        showCheckbox: true,
        showArtist: true,
        showAlbum: true,
        showDownloadBtn: true
    });

    batchActions.style.display = 'flex';
    section.style.display = 'block';
}

// 下载歌曲
async function downloadSong(source, id, name, artist, album) {
    const br = getSelectValue('quality-select-wrapper');
    const btn = event.target;
    btn.textContent = '已加入';
    btn.disabled = true;

    try {
        const params = new URLSearchParams({ source, id, name, artist, album: album || '', br });
        await fetch('/api/v1/download?' + params);
    } catch (err) {
        btn.textContent = '下载';
        btn.disabled = false;
    }
}

// 导航功能
function initNav() {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            const page = this.dataset.page;

            document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
            this.classList.add('active');

            showPage(page);
        });
    });
}

function showPage(page) {
    const sections = ['settings-section', 'playlist-section', 'search-results-section', 'downloads-section', 'library-section', 'home-section'];
    sections.forEach(id => {
        const el = document.getElementById(id);
        if (el) el.style.display = 'none';
    });

    if (page === 'settings') {
        document.getElementById('settings-section').style.display = 'block';
    } else if (page === 'downloads') {
        document.getElementById('downloads-section').style.display = 'block';
        loadDownloads();
    } else if (page === 'library') {
        document.getElementById('library-section').style.display = 'block';
        loadLibrary();
    } else if (page === 'playlist') {
        document.getElementById('playlist-section').style.display = 'block';
        showPlaylistGrid();
        loadPlaylists();
    } else if (page === 'home') {
        document.getElementById('home-section').style.display = 'block';
    } else {
        document.getElementById('home-section').style.display = 'block';
    }
}

// 设置功能
function initSettings() {
    document.getElementById('save-settings').addEventListener('click', saveSettings);
}

async function loadSettings() {
    try {
        const resp = await fetch('/api/v1/settings');
        const data = await resp.json();
        if (data.code === 200) {
            document.getElementById('download-dir').value = data.data.downloadDir || '';
            // 加载音质设置
            if (data.data.quality) {
                setSelectValue('quality-select-wrapper', data.data.quality);
            }
        }
    } catch (err) {
        console.error('加载设置失败');
    }
}

async function saveSettings() {
    const downloadDir = document.getElementById('download-dir').value;
    const quality = getSelectValue('quality-select-wrapper');

    try {
        const resp = await fetch('/api/v1/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ downloadDir, quality })
        });
        const data = await resp.json();
        toast(data.message || '保存成功', 'success');
    } catch (err) {
        toast('保存失败', 'error');
    }
}

// 下载管理
async function loadDownloads() {
    try {
        const resp = await fetch('/api/v1/downloads');
        const data = await resp.json();
        renderDownloads(data.data || []);
    } catch (err) {
        console.error('加载下载列表失败');
    }
}

function renderDownloads(tasks) {
    const list = document.getElementById('download-list');
    if (tasks.length === 0) {
        list.innerHTML = '<div class="no-results">暂无下载任务</div>';
        return;
    }
    list.innerHTML = tasks.map(t => `
        <div class="download-item">
            <div class="song-info">
                <div class="song-name">${t.name}</div>
                <div class="artist">${t.artist}</div>
            </div>
            <div class="progress-wrapper">
                <div class="progress-bar">
                    <div class="progress-fill" style="width:${t.progress}%"></div>
                </div>
            </div>
            <span class="status ${t.status}">${getStatusText(t.status)}</span>
        </div>
    `).join('');
}

function getStatusText(status) {
    const map = { pending: '等待中', downloading: '下载中', success: '已完成', failed: '失败' };
    return map[status] || status;
}

function startDownloadPolling() {
    setInterval(async () => {
        const section = document.getElementById('downloads-section');
        if (section && section.style.display !== 'none') {
            loadDownloads();
        }
    }, 2000);
}

// 排行榜功能
let currentToplistId = null;

async function loadToplists() {
    const source = getSelectValue('source-select-wrapper');
    const tabsContainer = document.getElementById('toplist-tabs');

    try {
        const resp = await fetch(`/api/v1/toplists?source=${source}`);
        const data = await resp.json();

        if (data.code === 200 && data.data && data.data.list) {
            renderToplistTabs(data.data.list);
            // 默认选择第一个排行榜
            if (data.data.list.length > 0) {
                selectToplist(data.data.list[0].id);
            }
        } else {
            tabsContainer.innerHTML = '<span class="no-results">加载排行榜失败</span>';
        }
    } catch (err) {
        tabsContainer.innerHTML = '<span class="no-results">加载排行榜失败</span>';
    }
}

function renderToplistTabs(list) {
    const tabsContainer = document.getElementById('toplist-tabs');
    tabsContainer.innerHTML = list.map(item => `
        <div class="toplist-tab" data-id="${item.id}" onclick="selectToplist('${item.id}')">
            <span class="tab-name">${item.name}</span>
            ${item.updateFrequency ? `<span class="tab-freq">${item.updateFrequency}</span>` : ''}
        </div>
    `).join('');
}

function selectToplist(id) {
    currentToplistId = id;
    // 更新tab选中状态
    document.querySelectorAll('.toplist-tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.id === id);
    });
    // 加载歌曲列表
    loadToplistSongs(id);
}

async function loadToplistSongs(id) {
    const source = getSelectValue('source-select-wrapper');
    const songsContainer = document.getElementById('toplist-songs');
    songsContainer.innerHTML = '<div class="loading">加载中...</div>';

    try {
        const resp = await fetch(`/api/v1/toplist?source=${source}&id=${id}`);
        const data = await resp.json();

        if (data.code === 200 && data.data && data.data.list) {
            await renderToplistSongs(data.data.list, data.data.source || source);
        } else {
            songsContainer.innerHTML = '<div class="no-results">加载歌曲失败</div>';
        }
    } catch (err) {
        songsContainer.innerHTML = '<div class="no-results">加载歌曲失败</div>';
    }
}

async function renderToplistSongs(songs, source) {
    const songsContainer = document.getElementById('toplist-songs');
    const batchActions = document.getElementById('toplist-batch-actions');

    // 重置选择状态
    toplistSelectedSongs = [];
    currentToplistSongs = songs.map(s => ({ ...s, source }));
    currentToplistSource = source;
    document.getElementById('toplist-select-all').checked = false;
    updateSelectedCount('toplist');

    if (songs.length === 0) {
        songsContainer.innerHTML = '<div class="no-results">暂无歌曲</div>';
        batchActions.style.display = 'none';
        return;
    }

    // 检查已下载状态
    const ids = songs.map(s => s.id).join(',');
    let downloaded = {};
    try {
        const resp = await fetch(`/api/v1/downloaded?source=${source}&ids=${ids}`);
        const data = await resp.json();
        downloaded = data.data || {};
    } catch (e) {}

    songsContainer.innerHTML = renderSongList(songs, downloaded, {
        type: 'toplist',
        source: source,
        showCheckbox: true,
        showArtist: true,
        showAlbum: true,
        showDownloadBtn: true
    });

    batchActions.style.display = 'flex';
}

// 音乐库
async function loadLibrary() {
    try {
        const resp = await fetch('/api/v1/library');
        const data = await resp.json();
        renderLibrary(data.data || []);
    } catch (err) {
        console.error('加载音乐库失败');
    }
}

// 刷新音乐库
async function refreshLibrary() {
    const btn = document.getElementById('refresh-library-btn');
    btn.textContent = '刷新中...';
    btn.disabled = true;

    try {
        const resp = await fetch('/api/v1/library/refresh', { method: 'POST' });
        const data = await resp.json();
        if (data.code === 200) {
            renderLibrary(data.data || []);
            if (data.removed > 0) {
                toast(`已移除 ${data.removed} 首不存在的歌曲`, 'info');
            } else {
                toast('音乐库已刷新', 'success');
            }
        } else {
            toast('刷新失败', 'error');
        }
    } catch (err) {
        toast('刷新失败', 'error');
    } finally {
        btn.textContent = '刷新';
        btn.disabled = false;
    }
}

function renderLibrary(songs) {
    const list = document.getElementById('library-list');
    if (songs.length === 0) {
        list.innerHTML = '<div class="no-results">暂无已下载歌曲</div>';
        return;
    }
    // 音乐库所有歌曲都是已下载的
    const downloaded = {};
    songs.forEach(s => downloaded[s.id] = true);

    list.innerHTML = renderSongList(songs, downloaded, {
        type: 'library',
        source: '',
        showCheckbox: false,
        showArtist: true,
        showAlbum: true,
        showDownloadBtn: false
    });
}

// 切换单个歌曲选择
function toggleSongSelect(type, index) {
    let selectedArr;
    if (type === 'search') {
        selectedArr = searchSelectedSongs;
    } else if (type === 'toplist') {
        selectedArr = toplistSelectedSongs;
    } else {
        selectedArr = playlistSelectedSongs;
    }
    const idx = selectedArr.indexOf(index);
    if (idx > -1) {
        selectedArr.splice(idx, 1);
    } else {
        selectedArr.push(index);
    }
    updateSelectedCount(type);
    updateSelectAllState(type);
}

// 全选/取消全选
function toggleSelectAll(type, checked) {
    let listId;
    if (type === 'search') {
        listId = 'search-results';
    } else if (type === 'toplist') {
        listId = 'toplist-songs';
    } else {
        listId = 'playlist-songs';
    }
    const checkboxes = document.querySelectorAll(`#${listId} .song-checkbox:not(:disabled)`);

    if (type === 'search') {
        searchSelectedSongs = [];
        if (checked) {
            checkboxes.forEach(cb => {
                cb.checked = true;
                searchSelectedSongs.push(parseInt(cb.dataset.index));
            });
        } else {
            checkboxes.forEach(cb => cb.checked = false);
        }
    } else if (type === 'toplist') {
        toplistSelectedSongs = [];
        if (checked) {
            checkboxes.forEach(cb => {
                cb.checked = true;
                toplistSelectedSongs.push(parseInt(cb.dataset.index));
            });
        } else {
            checkboxes.forEach(cb => cb.checked = false);
        }
    } else {
        playlistSelectedSongs = [];
        if (checked) {
            checkboxes.forEach(cb => {
                cb.checked = true;
                playlistSelectedSongs.push(parseInt(cb.dataset.index));
            });
        } else {
            checkboxes.forEach(cb => cb.checked = false);
        }
    }
    updateSelectedCount(type);
}

// 更新已选数量显示
function updateSelectedCount(type) {
    let count;
    if (type === 'search') {
        count = searchSelectedSongs.length;
    } else if (type === 'toplist') {
        count = toplistSelectedSongs.length;
    } else {
        count = playlistSelectedSongs.length;
    }
    const countEl = document.getElementById(`${type}-selected-count`);
    countEl.textContent = `已选 ${count} 首`;
}

// 更新全选框状态
function updateSelectAllState(type) {
    let listId, selectedArr;
    if (type === 'search') {
        listId = 'search-results';
        selectedArr = searchSelectedSongs;
    } else if (type === 'toplist') {
        listId = 'toplist-songs';
        selectedArr = toplistSelectedSongs;
    } else {
        listId = 'playlist-songs';
        selectedArr = playlistSelectedSongs;
    }
    const checkboxes = document.querySelectorAll(`#${listId} .song-checkbox:not(:disabled)`);
    const selectAllEl = document.getElementById(`${type}-select-all`);

    if (checkboxes.length === 0) {
        selectAllEl.checked = false;
    } else {
        selectAllEl.checked = selectedArr.length === checkboxes.length;
    }
}

// 批量下载
async function batchDownload(type) {
    let selectedArr, songsData, listId;
    if (type === 'search') {
        selectedArr = searchSelectedSongs;
        songsData = currentSearchResults;
        listId = 'search-results';
    } else if (type === 'toplist') {
        selectedArr = toplistSelectedSongs;
        songsData = currentToplistSongs;
        listId = 'toplist-songs';
    } else {
        selectedArr = playlistSelectedSongs;
        songsData = currentPlaylistSongs;
        listId = 'playlist-songs';
    }

    if (selectedArr.length === 0) {
        toast('请先选择要下载的歌曲', 'warning');
        return;
    }

    const btn = document.getElementById(`${type}-batch-download`);
    btn.textContent = '下载中...';
    btn.disabled = true;

    const br = getSelectValue('quality-select-wrapper');
    let successCount = 0;
    const downloadedIndexes = [];

    for (const index of selectedArr) {
        const song = songsData[index];
        if (!song) continue;

        try {
            const params = new URLSearchParams({
                source: song.source,
                id: song.id,
                name: song.name,
                artist: song.artist || '',
                album: song.album || '',
                br
            });
            await fetch('/api/v1/download?' + params);
            successCount++;
            downloadedIndexes.push(index);
        } catch (err) {
            console.error('下载失败:', song.name);
        }
    }

    // 禁用已下载歌曲的复选框并更新UI
    downloadedIndexes.forEach(index => {
        const checkbox = document.querySelector(
            `#${listId} .song-checkbox[data-index="${index}"]`
        );
        if (checkbox) {
            checkbox.checked = false;
            checkbox.disabled = true;
            // 更新下载按钮为已下载标签
            const songItem = checkbox.closest('.song-item');
            const downloadBtn = songItem.querySelector('.download-btn');
            if (downloadBtn) {
                const tag = document.createElement('span');
                tag.className = 'downloaded-tag';
                tag.textContent = '已下载';
                downloadBtn.replaceWith(tag);
            }
        }
    });

    // 清空选择状态
    if (type === 'search') {
        searchSelectedSongs = [];
    } else if (type === 'toplist') {
        toplistSelectedSongs = [];
    } else {
        playlistSelectedSongs = [];
    }
    updateSelectedCount(type);
    updateSelectAllState(type);

    toast(`已添加 ${successCount} 首歌曲到下载队列`, 'success');
    btn.textContent = '批量下载';
    btn.disabled = false;
}

// 返回歌单列表
function showPlaylistGrid() {
    document.getElementById('playlist-detail').style.display = 'none';
    document.getElementById('playlist-grid').style.display = 'grid';
    document.querySelector('.playlist-import-form').style.display = 'flex';
    currentPlaylistDetail = null;
}

// 删除歌单
async function deletePlaylist(id, source) {
    const confirmed = await showConfirm('确定要删除这个歌单吗？', '删除歌单');
    if (!confirmed) return;

    try {
        const resp = await fetch(`/api/v1/playlist?source=${source}&id=${id}`, {
            method: 'DELETE'
        });
        const data = await resp.json();
        if (data.code === 200) {
            toast('歌单已删除', 'success');
            loadPlaylists();
        } else {
            toast('删除失败', 'error');
        }
    } catch (err) {
        toast('删除失败', 'error');
    }
}

// 下载歌单中的歌曲
async function downloadPlaylistSong(source, id, name, artist, album, btn) {
    const br = getSelectValue('quality-select-wrapper');
    btn.textContent = '已加入';
    btn.disabled = true;

    try {
        const params = new URLSearchParams({ source, id, name, artist: artist || '', album: album || '', br });
        await fetch('/api/v1/download?' + params);
    } catch (err) {
        btn.textContent = '下载';
        btn.disabled = false;
    }
}
