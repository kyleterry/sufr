package server

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/data"
	"github.com/oxtoacart/bpool"
)

var templateBuffers = bpool.NewBufferPool(64)

type templates struct {
	templates map[string]*template.Template
}

type templateWriterFunc func(*templateWriter) error

func (t *templates) withWriter(path string, fn templateWriterFunc) error {
	tmpl, ok := t.templates[path]
	if !ok {
		return fmt.Errorf("no such template %s", path)
	}

	if err := fn(&templateWriter{t: tmpl}); err != nil {
		return err
	}

	return nil
}

type templateWriter struct {
	t *template.Template
}

func (tw *templateWriter) write(w http.ResponseWriter, r *http.Request, td interface{}) error {
	buf := templateBuffers.Get()
	defer templateBuffers.Put(buf)

	if err := tw.t.ExecuteTemplate(buf, "base", td); err != nil {
		return err
	}

	buf.WriteTo(w)

	return nil
}

type templateData struct {
	User       *api.User
	PinnedTags []string
	Paginator  data.URLPaginator
	Count      int
	Flashes    map[string][]interface{}
	Title      string
}

type timelineData struct {
	templateData
	URLs  []*api.UserURL
	Count int
}

func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("pairs required: parameters for dict must be passed in multiples of 2")
	}

	dict := make(map[string]interface{}, len(values)/2)

	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("invalid key type: dict keys must be strings")
		}

		dict[key] = values[i+1]
	}

	return dict, nil
}

func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC1123)
}

func tagNames(tl *api.TagList) string {
	names := []string{}

	for _, t := range tl.Items {
		names = append(names, t.Name)
	}

	return strings.Join(names, ",")
}
