package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

type PageData struct {
	EnvMessage string
	FormResult string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tmpl := template.Must(template.ParseFiles("static/index.html"))

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		message := os.Getenv("MESSAGE_ACCUEIL")
		if message == "" {
			message = "Aucune variable 'MESSAGE_ACCUEIL' détectée."
		}

		data := PageData{
			EnvMessage: message,
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		w.Write([]byte("Bonjour " + name))
	})

	http.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		nomUtilisateur := r.FormValue("mon_champ")

		data := PageData{
			EnvMessage: os.Getenv("MESSAGE_ACCUEIL"),
			FormResult: "Reçu côté serveur : " + nomUtilisateur,
		}
		tmpl.Execute(w, data)
	})

	log.Printf("Serveur lancé sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
