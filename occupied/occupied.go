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
	http.HandleFunc("/", latest_json)
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

var latestJsonTemplate = template.Must(template.New("latest_json").Parse(latestJsonTemplateStr))

const latestJsonTemplateStr = `{"occupied": {{.Occupied}}}`

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
