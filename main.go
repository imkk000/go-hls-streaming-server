package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

const segmentsDir = "segments"

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(
		middleware.RequestID(),
		middleware.Recover(),
		middleware.Logger(),
	)
	e.Static("js", "public/js")
	e.Static("", segmentsDir)
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("public/template/*.html")),
	}
	e.GET("/playlist", func(c echo.Context) error {
		dirs, err := os.ReadDir(segmentsDir)
		if err != nil {
			return err
		}
		paths := make([]string, len(dirs))
		for i := range dirs {
			paths[i] = dirs[i].Name()
		}
		return c.Render(http.StatusOK, "list.html", echo.Map{
			"Dirs": paths,
		})
	})
	e.GET("js/hls.render.js", func(c echo.Context) error {
		path := c.QueryParam("path")
		nowNano := time.Now().UnixNano()

		content := fmt.Sprintf(hlsScriptTemplate, path, nowNano, path, nowNano)
		c.Response().Header().Set("Content-Type", "text/javascript; charset=utf-8")
		return c.String(http.StatusOK, content)
	})

	e.GET("/", func(c echo.Context) error {
		path := c.QueryParam("path")
		if path == "" {
			return c.Redirect(http.StatusMovedPermanently, "/playlist")
		}
		return c.Render(http.StatusOK, "index.html", echo.Map{
			"Path":      path,
			"Timestamp": time.Now().UnixNano(),
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "54321"
	}
	addr := fmt.Sprintf("127.0.0.1:%s", port)
	log.Info().Msgf("start server %s", addr)
	if err := e.StartTLS(addr, "cert.pem", "key.pem"); err != nil {
		log.Fatal().Err(err).Msg("start server")
	}
}

const hlsScriptTemplate = `
const video = document.getElementById('video');
const hls = new Hls();
hls.loadSource('/%s/playlist.m3u8?v=%d');
hls.attachMedia(video);
hls.on(Hls.Events.MANIFEST_LOADED, () => {
	video.appendChild(Object.assign(document.createElement('track'), {
		kind: 'subtitles',
		src: '/%s/subtitles.vtt?v=%d',
		srclang: 'en',
		label: 'English',
		default: true,
	}));
});
hls.on(Hls.Events.ERROR, (event, data) => {
	console.log('HLS Error:', data);
});
`
