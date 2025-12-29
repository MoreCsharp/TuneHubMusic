package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"yinyue/storage"

	"github.com/gin-gonic/gin"
)

const baseURL = "https://music-dl.sayqz.com"

var DownloadDir = "./downloads"

// InitLibrary 初始化音乐库（启动时调用）
func InitLibrary(dataDir string) {
	// 从存储加载设置
	settings := storage.GetSettings()
	if settings.DownloadDir != "" {
		DownloadDir = settings.DownloadDir
	}

	// 从存储加载音乐库
	songs := storage.GetLibrary()
	libMutex.Lock()
	downloadedSongs = make([]DownloadedSong, len(songs))
	for i, s := range songs {
		downloadedSongs[i] = DownloadedSong{
			ID:       s.ID,
			Name:     s.Name,
			Artist:   s.Artist,
			Album:    s.Album,
			Source:   s.Source,
			Filename: s.Filename,
			Path:     s.Path,
			Time:     s.Time,
		}
	}
	libMutex.Unlock()

	// 验证音乐库文件是否存在
	ValidateLibrary()
}

// ValidateLibrary 验证音乐库，移除不存在的文件
func ValidateLibrary() int {
	libMutex.Lock()
	defer libMutex.Unlock()

	validSongs := make([]DownloadedSong, 0, len(downloadedSongs))
	removed := 0

	for _, song := range downloadedSongs {
		if _, err := os.Stat(song.Path); err == nil {
			validSongs = append(validSongs, song)
		} else {
			removed++
		}
	}

	if removed > 0 {
		downloadedSongs = validSongs
		// 同步到存储
		syncLibraryToStorage()
	}

	return removed
}

// syncLibraryToStorage 同步音乐库到存储（调用前需持有libMutex锁）
func syncLibraryToStorage() {
	songs := make([]storage.DownloadedSong, len(downloadedSongs))
	for i, s := range downloadedSongs {
		songs[i] = storage.DownloadedSong{
			ID:       s.ID,
			Name:     s.Name,
			Artist:   s.Artist,
			Album:    s.Album,
			Source:   s.Source,
			Filename: s.Filename,
			Path:     s.Path,
			Time:     s.Time,
		}
	}
	storage.SetLibrary(songs)
}

// 下载任务
type DownloadTask struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Source   string `json:"source"`
	Status   string `json:"status"` // pending, downloading, success, failed
	Progress int    `json:"progress"`
	Error    string `json:"error,omitempty"`
}

// progressWriter 追踪下载进度的 writer
type progressWriter struct {
	task    *DownloadTask
	total   int64
	written int64
	file    *os.File
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.file.Write(p)
	if err != nil {
		return n, err
	}
	pw.written += int64(n)
	if pw.total > 0 {
		progress := int(float64(pw.written) / float64(pw.total) * 100)
		taskMutex.Lock()
		pw.task.Progress = progress
		taskMutex.Unlock()
	}
	return n, nil
}

// 已下载歌曲
type DownloadedSong struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Source   string `json:"source"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Time     string `json:"time"`
}

var (
	downloadTasks   = make(map[string]*DownloadTask)
	downloadedSongs = make([]DownloadedSong, 0)
	taskMutex       sync.RWMutex
	libMutex        sync.RWMutex
)

func SearchMusic(c *gin.Context) {
	source := c.Query("source")
	keyword := c.Query("keyword")
	limit := c.DefaultQuery("limit", "20")

	if source == "" || keyword == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	params := url.Values{}
	params.Set("source", source)
	params.Set("type", "search")
	params.Set("keyword", keyword)
	params.Set("limit", limit)

	reqURL := baseURL + "/api/?" + params.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "请求失败"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// GetMusicURL 获取音乐文件URL
func GetMusicURL(c *gin.Context) {
	source := c.Query("source")
	id := c.Query("id")
	br := c.DefaultQuery("br", "320k")

	if source == "" || id == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	params := url.Values{}
	params.Set("source", source)
	params.Set("type", "url")
	params.Set("id", id)
	params.Set("br", br)

	reqURL := baseURL + "/api/?" + params.Encode()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(reqURL)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "请求失败"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		sourceSwitch := resp.Header.Get("X-Source-Switch")
		c.JSON(200, gin.H{
			"code":         200,
			"url":          location,
			"sourceSwitch": sourceSwitch,
		})
	} else {
		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", body)
	}
}

// DownloadMusic 下载音乐文件（异步）
func DownloadMusic(c *gin.Context) {
	source := c.Query("source")
	id := c.Query("id")
	name := c.Query("name")
	artist := c.Query("artist")
	album := c.Query("album")
	br := c.DefaultQuery("br", "320k")

	if source == "" || id == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	taskID := source + "_" + id

	// 检查是否已下载
	libMutex.RLock()
	for _, song := range downloadedSongs {
		if song.ID == id && song.Source == source {
			libMutex.RUnlock()
			c.JSON(200, gin.H{"code": 200, "message": "已下载", "taskId": taskID})
			return
		}
	}
	libMutex.RUnlock()

	// 创建下载任务
	taskMutex.Lock()
	if _, exists := downloadTasks[taskID]; exists {
		taskMutex.Unlock()
		c.JSON(200, gin.H{"code": 200, "message": "下载中", "taskId": taskID})
		return
	}
	task := &DownloadTask{
		ID:       taskID,
		Name:     name,
		Artist:   artist,
		Source:   source,
		Status:   "pending",
		Progress: 0,
	}
	downloadTasks[taskID] = task
	taskMutex.Unlock()

	// 异步下载
	go doDownload(task, source, id, name, artist, album, br)

	c.JSON(200, gin.H{"code": 200, "message": "已加入下载队列", "taskId": taskID})
}

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}

// doDownload 执行下载
func doDownload(task *DownloadTask, source, id, name, artist, album, br string) {
	taskMutex.Lock()
	task.Status = "downloading"
	task.Progress = 0
	taskMutex.Unlock()

	params := url.Values{}
	params.Set("source", source)
	params.Set("type", "url")
	params.Set("id", id)
	params.Set("br", br)

	reqURL := baseURL + "/api/?" + params.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		taskMutex.Lock()
		task.Status = "failed"
		task.Error = "请求失败"
		taskMutex.Unlock()
		return
	}
	defer resp.Body.Close()

	os.MkdirAll(DownloadDir, 0755)

	ext := ".mp3"
	if br == "flac" || br == "flac24bit" {
		ext = ".flac"
	}
	filename := sanitizeFilename(artist + " - " + name + ext)
	filePath := filepath.Join(DownloadDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		taskMutex.Lock()
		task.Status = "failed"
		task.Error = "创建文件失败"
		taskMutex.Unlock()
		return
	}
	defer file.Close()

	// 使用进度追踪 writer
	pw := &progressWriter{
		task:  task,
		total: resp.ContentLength,
		file:  file,
	}

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		taskMutex.Lock()
		task.Status = "failed"
		task.Error = "写入失败"
		taskMutex.Unlock()
		return
	}

	taskMutex.Lock()
	task.Status = "success"
	task.Progress = 100
	taskMutex.Unlock()

	// 添加到音乐库
	libMutex.Lock()
	downloadedSongs = append(downloadedSongs, DownloadedSong{
		ID:       id,
		Name:     name,
		Artist:   artist,
		Album:    album,
		Source:   source,
		Filename: filename,
		Path:     filePath,
		Time:     time.Now().Format("2006-01-02 15:04"),
	})
	// 持久化保存
	syncLibraryToStorage()
	libMutex.Unlock()
}

// GetSettings 获取设置
func GetSettings(c *gin.Context) {
	settings := storage.GetSettings()
	c.JSON(200, gin.H{
		"code": 200,
		"data": gin.H{
			"downloadDir": settings.DownloadDir,
			"quality":     settings.Quality,
		},
	})
}

// UpdateSettings 更新设置
func UpdateSettings(c *gin.Context) {
	var req struct {
		DownloadDir string `json:"downloadDir"`
		Quality     string `json:"quality"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.DownloadDir != "" {
		DownloadDir = req.DownloadDir
	}

	// 持久化保存设置
	err := storage.UpdateSettings(storage.Settings{
		DownloadDir: DownloadDir,
		Quality:     req.Quality,
	})
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "保存设置失败"})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "设置已保存",
	})
}

// GetDownloadTasks 获取下载任务列表
func GetDownloadTasks(c *gin.Context) {
	taskMutex.RLock()
	tasks := make([]*DownloadTask, 0, len(downloadTasks))
	for _, t := range downloadTasks {
		tasks = append(tasks, t)
	}
	taskMutex.RUnlock()

	c.JSON(200, gin.H{"code": 200, "data": tasks})
}

// GetLibrary 获取音乐库
func GetLibrary(c *gin.Context) {
	libMutex.RLock()
	songs := make([]DownloadedSong, len(downloadedSongs))
	copy(songs, downloadedSongs)
	libMutex.RUnlock()

	c.JSON(200, gin.H{"code": 200, "data": songs})
}

// RefreshLibrary 刷新音乐库，移除不存在的文件
func RefreshLibrary(c *gin.Context) {
	removed := ValidateLibrary()

	libMutex.RLock()
	songs := make([]DownloadedSong, len(downloadedSongs))
	copy(songs, downloadedSongs)
	libMutex.RUnlock()

	c.JSON(200, gin.H{
		"code":    200,
		"message": "音乐库已刷新",
		"removed": removed,
		"data":    songs,
	})
}

// IsDownloaded 检查歌曲是否已下载
func IsDownloaded(c *gin.Context) {
	ids := c.Query("ids")
	source := c.Query("source")

	idList := strings.Split(ids, ",")
	result := make(map[string]bool)

	libMutex.RLock()
	for _, id := range idList {
		for _, song := range downloadedSongs {
			if song.ID == id && song.Source == source {
				// 验证文件是否实际存在
				if _, err := os.Stat(song.Path); err == nil {
					result[id] = true
				}
				break
			}
		}
	}
	libMutex.RUnlock()

	c.JSON(200, gin.H{"code": 200, "data": result})
}

// GetToplists 获取排行榜列表
func GetToplists(c *gin.Context) {
	source := c.Query("source")
	if source == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	params := url.Values{}
	params.Set("source", source)
	params.Set("type", "toplists")

	reqURL := baseURL + "/api/?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "请求失败"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// GetToplistSongs 获取排行榜歌曲
func GetToplistSongs(c *gin.Context) {
	source := c.Query("source")
	id := c.Query("id")
	if source == "" || id == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	params := url.Values{}
	params.Set("source", source)
	params.Set("type", "toplist")
	params.Set("id", id)

	reqURL := baseURL + "/api/?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "请求失败"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// ImportPlaylist 导入歌单
func ImportPlaylist(c *gin.Context) {
	source := c.Query("source")
	id := c.Query("id")

	if source == "" || id == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	params := url.Values{}
	params.Set("source", source)
	params.Set("type", "playlist")
	params.Set("id", id)

	reqURL := baseURL + "/api/?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "请求失败"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 解析响应
	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				ID     string   `json:"id"`
				Name   string   `json:"name"`
				Artist string   `json:"artist"`
				Album  string   `json:"album"`
				Types  []string `json:"types"`
			} `json:"list"`
			Info struct {
				Name   string `json:"name"`
				Author string `json:"author"`
			} `json:"info"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "解析响应失败"})
		return
	}

	if result.Code != 200 {
		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	// 保存歌单到本地
	playlist := storage.Playlist{
		ID:     id,
		Source: source,
		Name:   result.Data.Info.Name,
		Author: result.Data.Info.Author,
		Songs:  make([]storage.PlaylistSong, len(result.Data.List)),
	}

	for i, song := range result.Data.List {
		playlist.Songs[i] = storage.PlaylistSong{
			ID:     song.ID,
			Name:   song.Name,
			Artist: song.Artist,
			Album:  song.Album,
			Types:  song.Types,
		}
	}

	if err := storage.AddPlaylist(playlist); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "保存歌单失败"})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "导入成功",
		"data":    playlist,
	})
}

// GetPlaylists 获取已导入的歌单列表
func GetPlaylists(c *gin.Context) {
	playlists := storage.GetPlaylists()
	c.JSON(200, gin.H{"code": 200, "data": playlists})
}

// DeletePlaylist 删除歌单
func DeletePlaylist(c *gin.Context) {
	source := c.Query("source")
	id := c.Query("id")

	if source == "" || id == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	if err := storage.DeletePlaylist(id, source); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}
