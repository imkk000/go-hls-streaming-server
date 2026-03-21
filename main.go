package main

import (
	"crypto/subtle"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	var user, pass string
	var segmentsDir string
	var port int

	flag.StringVar(&user, "user", "", "set basic auth username")
	flag.StringVar(&pass, "pass", "", "set basic auth password")
	flag.StringVar(&segmentsDir, "segments", "segments", "set segments path")
	flag.IntVar(&port, "port", 54321, "set server port")
	flag.Parse()

	if len(user) == 0 || len(pass) == 0 {
		log.Fatal().Err(errors.New("empty basic auth")).Msg("check basic authentication")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("get current working directory")
	}
	segmentPath := filepath.Join(wd, segmentsDir)

	e := echo.New()
	e.HideBanner = true
	e.Use(
		middleware.RequestID(),
		middleware.Recover(),
		middleware.Logger(),
	)
	e.Static("js", "public/js")
	e.Static("css", "public/css")
	e.Static("fonts", "public/fonts")
	e.Static("", segmentsDir)
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("public/template/*.html")),
	}
	g := e.Group("/", middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(user)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(pass)) == 1 {
			return true, nil
		}
		return false, nil
	}))
	g.GET("", func(c echo.Context) error {
		path := c.QueryParam("path")
		if path == "" {
			return c.Redirect(http.StatusMovedPermanently, "/playlist")
		}
		return c.Render(http.StatusOK, "index.html", echo.Map{
			"Path":      path,
			"Timestamp": time.Now().UnixNano(),
		})
	})
	g.GET("playlist", func(c echo.Context) error {
		series := c.QueryParam("s")
		if series != "" {
			path, err := filepath.Abs(series)
			if err != nil {
				return err
			}
			if !strings.HasPrefix(path, wd) {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}
		}
		path := filepath.Join(segmentPath, series)
		dirs, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		list := make([]List, 0, len(dirs))
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			pathName := filepath.Join(series, dir.Name())
			item := List{
				Dir: dir.Name(),
				URL: fmt.Sprintf("/?path=%s", pathName),
			}
			next := filepath.Join(path, dir.Name(), "playlist.m3u8")
			if _, err := os.Stat(next); err != nil {
				item.URL = fmt.Sprintf("/playlist?s=%s", pathName)
			}
			list = append(list, item)
		}
		return c.Render(http.StatusOK, "list.html", echo.Map{
			"Dirs": slices.Clip(list),
		})
	})
	e.GET("js/hls.render.js", func(c echo.Context) error {
		path := c.QueryParam("path")
		nowNano := time.Now().UnixNano()

		content := fmt.Sprintf(hlsScriptTemplate, path, nowNano, path, nowNano)
		c.Response().Header().Set("Content-Type", "text/javascript; charset=utf-8")
		return c.String(http.StatusOK, content)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Info().Msgf("start server %d", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal().Err(err).Msg("start server")
	}
}

const hlsScriptTemplate = `
const video = document.getElementById("player");
const hls = new Hls();
hls.loadSource("/%s/playlist.m3u8?v=%d");
hls.attachMedia(video);
hls.on(Hls.Events.MANIFEST_LOADED, () => {
  video.appendChild(
    Object.assign(document.createElement("track"), {
      kind: "subtitles",
      src: "/%s/subtitles.vtt?v=%d",
      srclang: "en",
      label: "English",
      default: true,
    }),
  );
});
hls.on(Hls.Events.ERROR, (event, data) => {
  console.log("HLS Error:", data);
});
`

type List struct {
	URL string
	Dir string
}
