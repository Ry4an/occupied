package occupied

import (
	"appengine"
	"appengine/datastore"
	"text/template"
	"net/http"
	"time"
)

type Record struct {
	Occupied bool
	Date  time.Time
}

func init() {
	http.HandleFunc("/", latest_html)
	http.HandleFunc("/latest.json", latest_json)
	http.HandleFunc("/record/opened", opened)
	http.HandleFunc("/record/closed", closed)
}

func latest_json(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Record").Order("-Date").Limit(1)
	records := make([]Record, 0, 1)
	if _, err := q.GetAll(c, &records); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := latestJsonTemplate.Execute(w, records[0]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func latest_html(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Record").Order("-Date").Limit(1)
	records := make([]Record, 0, 1)
	if _, err := q.GetAll(c, &records); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := latestHtmlTemplate.Execute(w, records[0]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var latestJsonTemplate = template.Must(template.New("latest_json").Parse(latestJsonTemplateStr))
var latestHtmlTemplate = template.Must(template.New("latest_html").Parse(latestHtmlTemplateStr))

const latestJsonTemplateStr = `{"occupied": {{.Occupied}}}`
const latestHtmlTemplateStr = `
<html>
<head>
<meta http-equiv="refresh" content="5">
<title>{{if .Occupied}}Occupied{{else}}Available{{end}}</title>
</head><body>
<img src="/static/img/{{if .Occupied}}occupied{{else}}vacant{{end}}.jpg">
</body>
`

func opened(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	rec := Record{
		Occupied: false,
		Date:  time.Now(),
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Record", nil), &rec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func closed(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	rec := Record{
		Occupied: true,
		Date:  time.Now(),
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Record", nil), &rec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
