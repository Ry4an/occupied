package occupied

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"encoding/json"
	"net/http"
	"text/template"
	"time"
)

const RecordKey = "record"

type Record struct {
	Occupied bool
	Date     time.Time
}

func init() {
	http.HandleFunc("/", latest_html)
	http.HandleFunc("/latest.json", latest_json)
	http.HandleFunc("/record/opened", opened)
	http.HandleFunc("/record/closed", closed)
}

func set_record_into_memcache(c appengine.Context, record Record) error {

	recordJson, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// now put the json into memcached
	item := &memcache.Item{
		Key:   RecordKey,
		Value: []byte(recordJson),
	}

	if err := memcache.Set(c, item); err != nil {
		return err
	}

	c.Infof("added %v", string(recordJson))

	return nil
}

func get_latest_record_cached(r *http.Request) (rec Record, err error) {
	c := appengine.NewContext(r)
	var record Record
	if recordItem, err := memcache.Get(c, RecordKey); err == memcache.ErrCacheMiss {
		if record, err := get_latest_record(r); err != nil {
			return record, err
		}

		if err := set_record_into_memcache(c, record); err != nil {
			c.Infof("error during set_record_into_cache %v", err)
			return Record{}, err
		}

	} else {
		c.Infof("got item from memcached: %v", string(recordItem.Value))
		// unmarshal the json into record
		if err := json.Unmarshal(recordItem.Value, &record); err != nil {
			c.Infof("error during unmarshal %v", err)
			return Record{}, err
		}
	}

	c.Infof("get_latest_record_cached returning %v", record)

	return record, nil
}

func get_latest_record(r *http.Request) (rec Record, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Record").Order("-Date").Limit(1)
	records := make([]Record, 0, 1)
	if _, err := q.GetAll(c, &records); err != nil {
		return Record{}, err
	}
	if len(records) == 0 {
		return Record{false, time.Now()}, nil
	}
	record := records[0]

	return record, nil
}

func latest_json(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	add_standard_headers(w)
	var rec Record
	var err error
	if rec, err = get_latest_record_cached(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := latestJsonTemplate.Execute(w, rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func add_standard_headers(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func latest_html(w http.ResponseWriter, r *http.Request) {
	add_standard_headers(w)
	var rec Record
	var err error
	if rec, err = get_latest_record_cached(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := latestHtmlTemplate.Execute(w, rec); err != nil {
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
<link rel="icon"
      type="image/png"
      href="/static/img/{{if .Occupied}}df-poop-occupied{{else}}df-poop-vacant{{end}}.png">
</head><body>
<img src="/static/img/{{if .Occupied}}occupied{{else}}vacant{{end}}.jpg">
</body>
`

func opened(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	rec := Record{
		Occupied: false,
		Date:     time.Now(),
	}

	if err := set_record_into_memcache(c, rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		Date:     time.Now(),
	}
	if err := set_record_into_memcache(c, rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Record", nil), &rec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
