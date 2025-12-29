package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// AppData 应用数据结构
type AppData struct {
	Settings  Settings         `json:"settings"`
	Library   []DownloadedSong `json:"library"`
	Playlists []Playlist       `json:"playlists"`
}

// Playlist 导入的歌单
type Playlist struct {
	ID     string         `json:"id"`
	Source string         `json:"source"`
	Name   string         `json:"name"`
	Author string         `json:"author"`
	Songs  []PlaylistSong `json:"songs"`
}

// PlaylistSong 歌单中的歌曲
type PlaylistSong struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Artist string   `json:"artist"`
	Album  string   `json:"album"`
	Types  []string `json:"types"`
}

// Settings 设置
type Settings struct {
	DownloadDir string `json:"downloadDir"`
	Quality     string `json:"quality"`
}

// DownloadedSong 已下载歌曲
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
	db      *sql.DB
	dbMu    sync.RWMutex
	baseDir string
)

// Init 初始化存储
func Init(dataDir string) error {
	baseDir = dataDir
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "app_data.db")
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	return initTables()
}

// initTables 创建数据库表
func initTables() error {
	// 设置表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)
	`)
	if err != nil {
		return err
	}

	// 初始化默认设置
	_, err = db.Exec(`
		INSERT OR IGNORE INTO settings (key, value) VALUES ('downloadDir', './downloads');
		INSERT OR IGNORE INTO settings (key, value) VALUES ('quality', '320k');
	`)
	if err != nil {
		return err
	}

	// 音乐库表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS library (
			id TEXT,
			source TEXT,
			name TEXT,
			artist TEXT,
			album TEXT,
			filename TEXT,
			path TEXT,
			time TEXT,
			PRIMARY KEY (id, source)
		)
	`)
	if err != nil {
		return err
	}

	// 歌单表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS playlists (
			id TEXT,
			source TEXT,
			name TEXT,
			author TEXT,
			PRIMARY KEY (id, source)
		)
	`)
	if err != nil {
		return err
	}

	// 歌单歌曲表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS playlist_songs (
			playlist_id TEXT,
			playlist_source TEXT,
			song_id TEXT,
			name TEXT,
			artist TEXT,
			album TEXT,
			types TEXT,
			PRIMARY KEY (playlist_id, playlist_source, song_id),
			FOREIGN KEY (playlist_id, playlist_source) REFERENCES playlists(id, source) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// 启用外键约束
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	return err
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetSettings 获取设置
func GetSettings() Settings {
	dbMu.RLock()
	defer dbMu.RUnlock()

	settings := Settings{
		DownloadDir: "./downloads",
		Quality:     "320k",
	}

	rows, err := db.Query("SELECT key, value FROM settings")
	if err != nil {
		return settings
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		switch key {
		case "downloadDir":
			settings.DownloadDir = value
		case "quality":
			settings.Quality = value
		}
	}
	return settings
}

// UpdateSettings 更新设置
func UpdateSettings(s Settings) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	_, err := db.Exec("UPDATE settings SET value = ? WHERE key = 'downloadDir'", s.DownloadDir)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE settings SET value = ? WHERE key = 'quality'", s.Quality)
	return err
}

// GetLibrary 获取音乐库
func GetLibrary() []DownloadedSong {
	dbMu.RLock()
	defer dbMu.RUnlock()

	rows, err := db.Query("SELECT id, source, name, artist, album, filename, path, time FROM library")
	if err != nil {
		return []DownloadedSong{}
	}
	defer rows.Close()

	var songs []DownloadedSong
	for rows.Next() {
		var song DownloadedSong
		err := rows.Scan(&song.ID, &song.Source, &song.Name, &song.Artist,
			&song.Album, &song.Filename, &song.Path, &song.Time)
		if err != nil {
			continue
		}
		songs = append(songs, song)
	}

	if songs == nil {
		return []DownloadedSong{}
	}
	return songs
}

// AddToLibrary 添加歌曲到音乐库
func AddToLibrary(song DownloadedSong) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	_, err := db.Exec(`
		INSERT OR REPLACE INTO library (id, source, name, artist, album, filename, path, time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, song.ID, song.Source, song.Name, song.Artist, song.Album, song.Filename, song.Path, song.Time)
	return err
}

// SetLibrary 设置整个音乐库
func SetLibrary(songs []DownloadedSong) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM library")
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO library (id, source, name, artist, album, filename, path, time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, song := range songs {
		_, err = stmt.Exec(song.ID, song.Source, song.Name, song.Artist,
			song.Album, song.Filename, song.Path, song.Time)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// IsInLibrary 检查歌曲是否在音乐库中
func IsInLibrary(id, source string) bool {
	dbMu.RLock()
	defer dbMu.RUnlock()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM library WHERE id = ? AND source = ?", id, source).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// ValidateLibrary 验证音乐库，移除不存在的文件
func ValidateLibrary() (removed int, err error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	rows, err := db.Query("SELECT id, source, path FROM library")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var toRemove []struct{ id, source string }
	for rows.Next() {
		var id, source, path string
		if err := rows.Scan(&id, &source, &path); err != nil {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			toRemove = append(toRemove, struct{ id, source string }{id, source})
		}
	}

	for _, item := range toRemove {
		_, err = db.Exec("DELETE FROM library WHERE id = ? AND source = ?", item.id, item.source)
		if err == nil {
			removed++
		}
	}

	return removed, nil
}

// GetPlaylists 获取所有歌单
func GetPlaylists() []Playlist {
	dbMu.RLock()
	defer dbMu.RUnlock()

	rows, err := db.Query("SELECT id, source, name, author FROM playlists")
	if err != nil {
		return []Playlist{}
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Source, &p.Name, &p.Author); err != nil {
			continue
		}
		p.Songs = getPlaylistSongs(p.ID, p.Source)
		playlists = append(playlists, p)
	}

	if playlists == nil {
		return []Playlist{}
	}
	return playlists
}

// getPlaylistSongs 获取歌单中的歌曲（内部函数，调用前需持有锁）
func getPlaylistSongs(playlistID, playlistSource string) []PlaylistSong {
	rows, err := db.Query(`
		SELECT song_id, name, artist, album, types
		FROM playlist_songs
		WHERE playlist_id = ? AND playlist_source = ?
	`, playlistID, playlistSource)
	if err != nil {
		return []PlaylistSong{}
	}
	defer rows.Close()

	var songs []PlaylistSong
	for rows.Next() {
		var s PlaylistSong
		var typesJSON string
		if err := rows.Scan(&s.ID, &s.Name, &s.Artist, &s.Album, &typesJSON); err != nil {
			continue
		}
		json.Unmarshal([]byte(typesJSON), &s.Types)
		songs = append(songs, s)
	}

	if songs == nil {
		return []PlaylistSong{}
	}
	return songs
}

// AddPlaylist 添加歌单
func AddPlaylist(playlist Playlist) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 删除旧歌单（如果存在）
	_, err = tx.Exec("DELETE FROM playlist_songs WHERE playlist_id = ? AND playlist_source = ?",
		playlist.ID, playlist.Source)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 插入或更新歌单
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO playlists (id, source, name, author)
		VALUES (?, ?, ?, ?)
	`, playlist.ID, playlist.Source, playlist.Name, playlist.Author)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 插入歌曲
	stmt, err := tx.Prepare(`
		INSERT INTO playlist_songs (playlist_id, playlist_source, song_id, name, artist, album, types)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, song := range playlist.Songs {
		typesJSON, _ := json.Marshal(song.Types)
		_, err = stmt.Exec(playlist.ID, playlist.Source, song.ID, song.Name,
			song.Artist, song.Album, string(typesJSON))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// DeletePlaylist 删除歌单
func DeletePlaylist(id, source string) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	// 先删除歌单歌曲
	_, err := db.Exec("DELETE FROM playlist_songs WHERE playlist_id = ? AND playlist_source = ?", id, source)
	if err != nil {
		return err
	}

	// 再删除歌单
	_, err = db.Exec("DELETE FROM playlists WHERE id = ? AND source = ?", id, source)
	return err
}
