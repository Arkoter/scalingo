package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Message struct {
	ID      int
	Content string
}

type PageData struct {
	DBMessages []Message
	EditID     int
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

	fs := http.FileServer(http.Dir("static"))
	http.Handle("GET /static/", http.StripPrefix("/static/", fs))

	tmpl := template.Must(template.ParseFiles("static/index.html"))

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		savedMessages := getMessagesFromDB()
		editID, _ := strconv.Atoi(r.URL.Query().Get("edit_id"))

		data := PageData{
			DBMessages: savedMessages,
			EditID:     editID,
		}
		_ = tmpl.Execute(w, data)
	})

	http.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		nomUtilisateur := r.FormValue("mon_champ")

		if db != nil && nomUtilisateur != "" {
			_, err := db.Exec("INSERT INTO messages (content) VALUES (?)", nomUtilisateur)
			if err != nil {
				log.Printf("ERREUR : Impossible d'ajouter la tâche '%s' -> %v\n", nomUtilisateur, err)
			} else {
				log.Printf("SUCCES : Tâche ajoutée -> '%s'\n", nomUtilisateur)
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("POST /delete", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)

		if db != nil && err == nil {
			_, errExec := db.Exec("DELETE FROM messages WHERE id = ?", id)
			if errExec != nil {
				log.Printf("ERREUR : Impossible de supprimer la tâche (ID: %d) -> %v\n", id, errExec)
			} else {
				log.Printf("SUCCES : Tâche supprimée (ID: %d)\n", id)
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("POST /update", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		idStr := r.FormValue("id")
		content := r.FormValue("content")
		id, err := strconv.Atoi(idStr)

		if db != nil && err == nil && content != "" {
			_, errExec := db.Exec("UPDATE messages SET content = ? WHERE id = ?", content, id)
			if errExec != nil {
				log.Printf("ERREUR : Impossible de modifier la tâche (ID: %d) -> %v\n", id, errExec)
			} else {
				log.Printf("SUCCES : Tâche modifiée (ID: %d) -> Nouveau contenu : '%s'\n", id, content)
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	log.Printf("Serveur lancé sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getMessagesFromDB() []Message {
	var list []Message
	if db == nil {
		return list
	}

	rows, err := db.Query("SELECT id, content FROM messages ORDER BY id DESC")
	if err != nil {
		log.Printf("ERREUR : Impossible de récupérer les tâches depuis la base de données -> %v\n", err)
		return list
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Content); err == nil {
			list = append(list, m)
		} else {
			log.Printf("AVERTISSEMENT : Erreur lors de la lecture d'une ligne -> %v\n", err)
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
