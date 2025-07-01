package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
	e.GET("/", func(c echo.Context) error {
		path := c.QueryParam("path")
		if path == "" {
			return c.Redirect(http.StatusMovedPermanently, "/playlist")
		}

		var subtitles []Subtitle
		if path != "" {
			filename := filepath.Join(path, "subtitle.vtt")
			subtitleFilename := filepath.Join(segmentsDir, filename)
			if _, err := os.Stat(subtitleFilename); err == nil {
				subtitles = []Subtitle{
					{
						Src:       filename,
						Lang:      "en",
						Label:     "English",
						IsDefault: true,
					},
				}
			}
		}

		return c.Render(http.StatusOK, "index.html", Render{
			Path:      fmt.Sprintf("%s/playlist.m3u8", path),
			Subtitles: subtitles,
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

type M = map[string]any

type Render struct {
	Dirs      []string
	Path      string
	Subtitles []Subtitle
}

type Subtitle struct {
	Src       string
	Lang      string
	Label     string
	IsDefault bool
}
