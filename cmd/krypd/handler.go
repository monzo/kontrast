package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/sergi/go-diff/diffmatchpatch"
	log "github.com/sirupsen/logrus"
)

var templateFiles = []string{"assets/templates/main.tmpl"}

func handleDiffDisplay(dm *DiffManager, path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if dm.LastErr != nil {
			fmt.Fprintf(w, "Error running diff :( : %s", dm.LastErr.Error())
			log.Errorf("Error getting diff: %s", dm.LastErr.Error())
			return
		}

		if dm.LastRun == nil {
			fmt.Fprintf(w, "Diff has not been run yet - please try again soon")
			return
		}

		t, err := template.
			New("main.tmpl").
			Funcs(template.FuncMap{
				"humanizeTime": func(t time.Time) string {
					return humanize.Time(t)
				},
				"renderDiffHTML": renderDiffHTML,
			}).
			ParseFiles(templateFiles...)
		if err != nil {
			fmt.Fprintf(w, "Error parsing template :( : %s", err.Error())
			log.Errorf("Error parsing template: %s", err.Error())
			return
		}

		err = t.Execute(w, dm.LastRun)
		if err != nil {
			fmt.Fprintf(w, "Error rendering template :( : %s", err.Error())
			log.Errorf("Error rendering template: %s", err.Error())
			return
		}
	}
}

func renderDiffHTML(d Diff) template.HTML {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(d.Left, d.Right, false)
	return template.HTML(dmp.DiffPrettyHtml(diffs))
}
