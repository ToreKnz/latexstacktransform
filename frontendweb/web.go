package frontendweb

import (
	"html/template"
	"log"
	"net/http"
	"stacklatex/latex"
)

type pageData struct {
	InputText     string
	OutputText    string
	Log 		  string
	ErrorMessage  string
	Info		  string
	Success       bool
}

var tmpl = template.Must(template.ParseFiles("template.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Request received:", r.Method, r.URL.Path)
    data := pageData{}
    if r.Method == http.MethodPost {
        r.ParseForm()
        input := r.FormValue("latex_input")
        result := latex.TransformLatex(input)
        data.InputText = input
        data.Success = result.Success
        if result.Success {
            data.OutputText = result.Transformed
        } else {
            data.ErrorMessage = result.ErrorMessage
        }
    }
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println("Template execution error:", err)
    }
}

func ServeWeb(port_num string) {
	http.HandleFunc("/", indexHandler)
	log.Println("Listening on port " + port_num)
	go http.ListenAndServe("0.0.0.0:" + port_num, nil) // IPv4
    go http.ListenAndServe("[::]:" + port_num, nil)    // IPv6
    select {}
}