package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Calculation struct {
	Expression string
	Result     float64
	Session    string
}

var calculations []Calculation
var store = sessions.NewCookieStore([]byte("pass"))
var templates *template.Template

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/calc", calcHandler)
	http.HandleFunc("/login", loginHandler)
	templates = template.Must(template.ParseGlob("templates/*.html"))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username := session.Values["username"]
	if username == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/calc", http.StatusSeeOther)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "index.html", nil)
	} else if r.Method == "POST" {
		r.ParseForm()
		username := r.PostForm.Get("username")
		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.Execute(w, "calc.html")
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		session, _ := store.Get(r, "session")
		untyped, ok := session.Values["username"]
		if !ok {
			return
		}
		username, ok := untyped.(string)
		if !ok {
			return
		}
		fmt.Fprintf(w, "<p>Username: %s<p>", username)

		session.Values["expression"] = r.FormValue("expression")
		untypedExpression, ok := session.Values["expression"]
		if !ok {
			return
		}
		expression, ok := untypedExpression.(string)
		session.Values["result"], err = calculateExpression(expression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		untypedResult, ok := session.Values["result"]
		if !ok {
			return
		}
		result, ok := untypedResult.(float64)
		if !ok {
			http.Error(w, "http status bad request", http.StatusBadRequest)
			return
		}

		calculations = append(calculations, Calculation{
			Expression: expression,
			Result:     result,
			Session:    username,
		})

		fmt.Fprintf(w, "<p>Result: %s<p>", strconv.FormatFloat(result, 'f', -1, 64)) // result output
		templates.Execute(w, calculations)
		for _, calc := range calculations {
			if username == calc.Session {
				fmt.Fprintf(w, "<p>%s = %s, user: %s</p>", calc.Expression, strconv.FormatFloat(calc.Result, 'f', -1, 64), calc.Session) // calc history output
			}
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
