package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type PageData struct {
	EnvMessage string
	FormResult string
	DBMessages []string
}

var db *sql.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("SCALINGO_MYSQL_URL")
	if dbURL != "" {
		dsn, err := parseURLtoDSN(dbURL)
		if err != nil {
			log.Fatal(err)
		}

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatal(err)
		}

		if err := db.Ping(); err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			log.Fatal(err)
		}
	}

	tmpl := template.Must(template.ParseFiles("static/index.html"))

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		message := os.Getenv("MESSAGE_ACCUEIL")
		if message == "" {
			message = "Aucune variable 'MESSAGE_ACCUEIL' détectée."
		}

		savedMessages := getMessagesFromDB()

		data := PageData{
			EnvMessage: message,
			DBMessages: savedMessages,
		}
		_ = tmpl.Execute(w, data)
	})

	http.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		w.Write([]byte("Bonjour " + name))
	})

	http.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		nomUtilisateur := r.FormValue("mon_champ")
		statusMsg := "Reçu mais non sauvegardé (Pas de DB)"

		if db != nil && nomUtilisateur != "" {
			_, err := db.Exec("INSERT INTO messages (content) VALUES (?)", nomUtilisateur)
			if err != nil {
				statusMsg = "Erreur lors de la sauvegarde."
			} else {
				statusMsg = "Sauvegardé en base de données : " + nomUtilisateur
			}
		}

		savedMessages := getMessagesFromDB()

		data := PageData{
			EnvMessage: os.Getenv("MESSAGE_ACCUEIL"),
			FormResult: statusMsg,
			DBMessages: savedMessages,
		}
		_ = tmpl.Execute(w, data)
	})

	log.Printf("Serveur lancé sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getMessagesFromDB() []string {
	var list []string
	if db == nil {
		return list
	}

	rows, err := db.Query("SELECT content FROM messages ORDER BY id DESC LIMIT 5")
	if err != nil {
		return list
	}
	defer rows.Close()

	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err == nil {
			list = append(list, content)
		}
	}
	return list
}

func parseURLtoDSN(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	password, _ := u.User.Password()
	host := u.Host
	if !strings.Contains(host, "(") {
		host = "tcp(" + host + ")"
	}
	return fmt.Sprintf("%s:%s@%s%s", u.User.Username(), password, host, u.Path), nil
}
