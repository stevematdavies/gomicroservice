package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "test.page.gohtml")
	})

	fmt.Println("Starting service on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}

}

//go:embed templates
var templateFs embed.FS

func render(w http.ResponseWriter, t string) {
	partials := []string{
		"templates/base.layout.gohtml",
		"templates/header.partial.gohtml",
		"templates/footer.partial.gohtml",
	}

	var templatesSlice []string
	templatesSlice = append(templatesSlice, fmt.Sprintf("templates/%s", t))

	for _, x := range partials {
		templatesSlice = append(templatesSlice, x)
	}

	tmpl, err := template.ParseFS(templateFs, templatesSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
