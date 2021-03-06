package photos

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/artursapek/artur.co/config"
	"github.com/julienschmidt/httprouter"
)

var (
	permalinkTemplate *template.Template
)

func init() {
	var permalinkTemplateErr error
	permalinkTemplate, permalinkTemplateErr = template.ParseFiles("templates/photos/permalink.html")
	if permalinkTemplateErr != nil {
		panic(permalinkTemplateErr)
	}
}

type Location struct {
	Lat, Lng float64
}

func (loc Location) String() string {
	return fmt.Sprintf("(%.5f, %.5f)", loc.Lat, loc.Lng)
}

func (loc Location) Valid() bool {
	return loc.Lat != 0 && loc.Lng != 0
}

type Permalink struct {
	ContentItem
	time.Time
	Location

	NextLink string
	PrevLink string
}

func PhotoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := albumsAuthWall(w, r); err != nil {
		return
	}

	if params.ByName("path") == "" || params.ByName("path") == "/" {
		http.Redirect(w, r, "/albums", 302)
		return
	}

	permalinkHandler("photo", w, r, params)
}

func VideoPermalinkHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := albumsAuthWall(w, r); err != nil {
		return
	}
	permalinkHandler("video", w, r, params)
}

func permalinkHandler(t string, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	item := ContentItem{
		Type: t,
		Src:  params.ByName("path")[1:],
	}
	if _, err := os.Stat(item.RawPath()); err != nil {
		http.Error(w, "No such "+t, 404)
		return
	}

	var (
		nextLink, prevLink string
		base                   = filepath.Join(filepath.Dir(item.RawPath()), "*")
		siblings, globErr      = filepath.Glob(base)
		index              int = -1
	)

	if globErr == nil {

		for i, fn := range siblings {
			if fn == item.RawPath() {
				index = i
				break
			}
		}

		if index > -1 {
			log.Println("found", item.RawPath(), index)

			if index > 0 {
				prevLink = strings.Replace(siblings[index-1], config.Config.RawRoot+"/"+item.Type+"s", "/"+item.Type+"s/permalink", 1)
			}
			if index < len(siblings)-1 {
				nextLink = strings.Replace(siblings[index+1], config.Config.RawRoot+"/"+item.Type+"s", "/"+item.Type+"s/permalink", 1)
			}
		}
	}

	renderErr := permalinkTemplate.Execute(w, Permalink{
		ContentItem: item,
		Time:        item.Timestamp(),
		Location:    item.Location(),

		NextLink: nextLink,
		PrevLink: prevLink,
	})
	if renderErr != nil {
		log.Fatal(renderErr)
	}
}
