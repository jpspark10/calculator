package main

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Calculation struct {
	Expression string
	Result     float64
}

var calculations []Calculation

var store = sessions.NewCookieStore([]byte("pass"))

func main() {
	http.HandleFunc("/", createHandler)
	err := http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
	}
	r.ParseForm()
	name := r.FormValue("name")
	if name != "" {
		session.Values["name"] = name
	}
	fmt.Println("session:", session)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	calculator(w, r)
}

func calculator(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
	}
	if r.Method == "GET" {
		err := tmpl.Execute(w, calculations)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		expression := r.FormValue("expression")

		result, err := calculateExpression(expression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		calculations = append(calculations, Calculation{
			Expression: expression,
			Result:     result,
		})

		fmt.Fprintf(w, "<p>Result: %s<p>", strconv.FormatFloat(result, 'f', -1, 64))
		err = tmpl.Execute(w, calculations)
		if err != nil {
			fmt.Printf(err.Error())
		}
		fmt.Fprintf(w, "<h2>Calculation History</h2>")
		for _, calc := range calculations {
			fmt.Fprintf(w, "<p>%s = %s</p>", calc.Expression, strconv.FormatFloat(calc.Result, 'f', -1, 64))
		}
	}
}

func calculateExpression(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	expression = strings.ReplaceAll(expression, ",", ".")

	result, err := eval(expression)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func eval(expression string) (float64, error) {
	result, err := strconv.ParseFloat(expression, 64)
	if err == nil {
		return result, nil
	}

	operations := []string{"+", "-", "*", "/"}

	for _, op := range operations {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) != 2 {
				return 0, fmt.Errorf("Invalid expression")
			}

			left, err := eval(parts[0])
			if err != nil {
				return 0, err
			}

			right, err := eval(parts[1])
			if err != nil {
				return 0, err
			}

			switch op {
			case "+":
				return left + right, nil
			case "-":
				return left - right, nil
			case "*":
				return left * right, nil
			case "/":
				if right == 0 {
					return 0, fmt.Errorf("Division by zero")
				}
				return left / right, nil
			}
		}
	}

	return 0, fmt.Errorf("Invalid expression")
}