package main

import (
	"calculator-main/pkg/handlers"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlers.MainHandler)
	http.HandleFunc("/calc", handlers.CalcHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
