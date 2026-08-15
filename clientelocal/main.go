package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func DbSqlite() {

	dbDir := "./db"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic("Error creating database directory: " + err.Error())
	}

	var err error

	dbPath := filepath.Join(dbDir, "servidorweb.db")

	db, err = sql.Open("sqlite", dbPath+"?_timeout=5000")
	if err != nil {
		panic("Error opening database: " + err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS webserver(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			text TEXT,
			author TEXT
	);`)
	if err != nil {
		panic("Error creating table: " + err.Error())
	}

}

// recieving from server
// from the server to the client //client is here
func SentToLocalClient() {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	type RecievingFromServe struct {
		ID     int    `json:"id"`
		Tittle string `json:"tittle"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}

	for range ticker.C {

		resp, err := http.Get("https://yourDomainOrLocalhost/senttolocalclient")
		if err != nil {
			continue
		}

		var records []RecievingFromServe
		err = json.NewDecoder(resp.Body).Decode(&records)
		resp.Body.Close()
		if err != nil {
			fmt.Println("[Cliente] Erro ao decodificar JSON do servidor:", err)
			continue
		}

		for _, r := range records {

			_, err := db.Exec(`
				INSERT INTO webserver
				(id, tittle, text, author)
				VALUES (?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET titulo=excluded.titulo,
				 texto=excluded.texto, autor=excluded.autor
				`,
				r.ID,
				r.Tittle,
				r.Text,
				r.Author,
			)

			if err != nil {
				fmt.Println("[Client] Error saving server data: ", err)
			}
		}
	}
}

// sent to server
// from the client to the server // client is here
func RecievingFromLocalClient() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	type SentToClient struct {
		ID     int    `json:"id"`
		Tittle string `json:"tittle"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}

	for range ticker.C {
		// 1. Busca dados de 'servidorcliente' no banco local
		rows, err := db.Query(`SELECT id, tittle, text, author FROM webserver`)
		if err != nil {
			fmt.Println("[Client-WebServer] Error in local query:", err)
			continue
		}

		var records []SentToClient
		for rows.Next() {
			var r SentToClient
			if err := rows.Scan(&r.ID, &r.Tittle, &r.Text, &r.Author); err == nil {
				records = append(records, r)
			}
		}
		rows.Close()

		if len(records) == 0 {
			continue
		}

		// Convert to JSON and send via POST to Render.
		jsonData, err := json.Marshal(records)
		if err != nil {
			fmt.Println("[Client-WebServer] Error generating JSON:", err)
			continue
		}

		resp, err := http.Post("https://yourDomainOrLocalhost/recievingfromlocalclient", "application/json",
			bytes.NewBuffer(jsonData),
		)

		if err != nil {
			fmt.Println("[Client-WebServer] HTTP request error:", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("[Client-WebServer] Server responded with status: %s\n", resp.Status)
		} else {
			fmt.Println("[Client-WebServer] Synchronization completed successfully!")
		}

		resp.Body.Close()

	}
}

// Just to render the database table in the application and check if communication is working properly.
//from the client to the server // client is here

func RenderOnServer() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	type SentToClient struct {
		ID     int    `json:"id"`
		Tittle string `json:"tittle"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}

	for range ticker.C {
		// 1. Busca dados de 'servidorcliente' no banco local
		rows, err := db.Query(`SELECT id, tittle, text, author FROM webserver`)
		if err != nil {
			continue
		}

		var records []SentToClient
		for rows.Next() {
			var r SentToClient
			rows.Scan(&r.ID, &r.Tittle, &r.Text, &r.Author)
			records = append(records, r)
		}
		rows.Close()

		if len(records) == 0 {
			continue
		}

		// 2. Transforma em JSON e envia via POST para o Render
		jsonData, _ := json.Marshal(records)
		resp, err := http.Post("https://yourDomainOrLocalhost/renderonserver", "application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			resp.Body.Close()
		}
	}
}

// serving html
func HtmlClient(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Error loading template:", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

func main() {

	DbSqlite()

	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", HtmlClient)

	// Start polling in the background
	//polling endpoint
	go SentToLocalClient()
	go RecievingFromLocalClient()
	go RenderOnServer()
	//polling endpoint

	fmt.Println("Server running on port :9090")
	fmt.Println("access: http://localhost:9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		panic(err)
	}

}
