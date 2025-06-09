package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Use(
		middleware.RequestID(),
		middleware.Logger(),
	)
	e.Static("js", "public/js")
	e.Static("", "segments")
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("public/template/*.html")),
	}
	e.GET("/", func(c echo.Context) error {
		path := c.QueryParam("path")
		dirs, err := os.ReadDir("segments")
		if err != nil {
			return err
		}
		paths := make([]string, len(dirs))
		for i := range dirs {
			paths[i] = dirs[i].Name()
		}
		return c.Render(http.StatusOK, "index.html", Render{
			Dirs:      paths,
			Path:      fmt.Sprintf("%s/playlist.m3u8", path),
			Thumbnail: fmt.Sprintf("%s/thumbnail.png", path),
			Subtitles: []Subtitle{
				{
					Src:       fmt.Sprintf("%s/subtitle.vtt", path),
					Lang:      "en",
					Label:     "English",
					IsDefault: true,
				},
			},
		})
	})
	e.Start("127.0.0.1:54321")
}

type M = map[string]any

type Render struct {
	Dirs      []string
	Path      string
	Thumbnail string
	Subtitles []Subtitle
}

type Subtitle struct {
	Src       string
	Lang      string
	Label     string
	IsDefault bool
}
