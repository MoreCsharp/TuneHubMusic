package routes

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"yinyue/controllers"
	"yinyue/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(staticFS, templatesFS embed.FS) *gin.Engine {
	r := gin.Default()

	// 从嵌入的文件系统加载 HTML 模板
	tmpl := template.Must(template.New("").ParseFS(templatesFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	// 从嵌入的文件系统提供静态文件
	staticSubFS, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticSubFS))

	// 全局中间件
	r.Use(middleware.Cors())

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		api.GET("/hello", controllers.Hello)
		api.GET("/search", controllers.SearchMusic)
		api.GET("/url", controllers.GetMusicURL)
		api.GET("/download", controllers.DownloadMusic)
		api.GET("/downloads", controllers.GetDownloadTasks)
		api.GET("/library", controllers.GetLibrary)
		api.POST("/library/refresh", controllers.RefreshLibrary)
		api.GET("/downloaded", controllers.IsDownloaded)
		api.GET("/settings", controllers.GetSettings)
		api.POST("/settings", controllers.UpdateSettings)
		api.GET("/toplists", controllers.GetToplists)
		api.GET("/toplist", controllers.GetToplistSongs)
		api.GET("/playlists", controllers.GetPlaylists)
		api.GET("/playlist/import", controllers.ImportPlaylist)
		api.DELETE("/playlist", controllers.DeletePlaylist)
	}

	return r
}
