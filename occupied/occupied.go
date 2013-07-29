package occupied

import (
	"appengine"
	"appengine/datastore"
	"html/template"
	"net/http"
	"time"
)

type Record struct {
	Occupied bool
	Date  time.Time
}

func init() {
	http.HandleFunc("/", latest)
	http.HandleFunc("/record/opened", opened)
	http.HandleFunc("/record/closed", closed)
}

func latest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Record").Order("-Date").Limit(1)
	records := make([]Record, 0, 1)
	if _, err := q.GetAll(c, &records); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := latestTemplate.Execute(w, records); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var latestTemplate = template.Must(template.New("book").Parse(latestTemplateHTML))

const latestTemplateHTML = `
<html>
  <body>
    {{range .}}
      {"state": {{.Occupied}}}
    {{end}}
  </body>
</html>
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
