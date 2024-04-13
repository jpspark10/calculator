package main

import (
	"calculator-main/pkg/handlers"
	"fmt"
	"html/template"
	"net/http"
)

var templates *template.Template

func main() {
	http.HandleFunc("/", handlers.MainHandler)
	http.HandleFunc("/calc", handlers.CalcHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	templates = template.Must(template.ParseGlob("templates/*.html"))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
