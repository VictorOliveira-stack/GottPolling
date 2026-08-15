package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func DbSqlite() {

	dbDir := "./db"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic("Error creating database directory: " + err.Error())
	}

	var err error
	dbPath := filepath.Join(dbDir, "webserver.db")
	db, err = sql.Open("sqlite", dbPath+"?_timeout=5000")
	if err != nil {
		panic("Error opening database: " + err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS webserver(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tittle TEXT,
		text TEXT,
		author TEXT
	);
	
	CREATE TABLE IF NOT EXISTS serverclient(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tittle TEXT,
		text TEXT,
		author TEXT
	);`)
	if err != nil {
		panic("Error creating table: " + err.Error())
	}

}

func InsertDb(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	/*titulo := "inserindo no sqlite online"
	texto := "ola mundo estamos online"
	autor := "inserido online"*/

	tittle := r.FormValue("tittle")
	text := r.FormValue("text")
	author := r.FormValue("author")

	if tittle == "" || text == "" || author == "" {
		http.Error(w, "All fields are required.", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO webserver (tittle, text, author) VALUES (?,?,?)`

	_, err := db.Exec(query, tittle, text, author)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

//sent to client
//from the server to the client //server is here

func SentToLocalClient(w http.ResponseWriter, r *http.Request) {

	type SentToClient struct {
		ID     int    `json:"id"`
		Tittle string `json:"tittle"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}

	rows, err := db.Query(`SELECT id, tittle, text, author FROM webserver`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []SentToClient

	for rows.Next() {

		var record SentToClient

		err := rows.Scan(
			&record.ID,
			&record.Tittle,
			&record.Text,
			&record.Author,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		records = append(records, record)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(records)

}

// recieving from the client
// from the client to the server // server is here
func RecievingFromLocalClient(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed (use POST)", http.StatusMethodNotAllowed)
		return
	}

	type DataRecivied struct {
		ID     int    `json:"id"`
		Tittle string `json:"tittle"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}

	var records []DataRecivied

	err := json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
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
			log.Println("[Server] Error saving record to database: ", err)
			http.Error(w, "Error saving to local database: ", http.StatusInternalServerError)
			return
		}

	}

	// Client response started successfully.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data synchronized with the server successfully! ",
	})

}

// Just to render the database table in the application and check if communication is working properly.
//from the client to the server // server is here

func RenderOnServer(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed (use POST)", http.StatusMethodNotAllowed)
		return
	}

	type DataRecivied struct {
		ID     int    `json:"id"`
		Tittle string `json:"tittle"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}

	var records []DataRecivied

	err := json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	for _, r := range records {
		_, err := db.Exec(`
			INSERT INTO serverclient
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
			log.Println("[Server] Error saving record to database:", err)
			http.Error(w, "Error saving to local database:", http.StatusInternalServerError)
			return
		}

	}

	// Answer to client started successfully.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data synchronized with the server successfully!",
	})

}

// serving html
func Indexhtml(w http.ResponseWriter, r *http.Request) {

	type ClientRecord struct {
		ID     int
		Tittle string
		Text   string
		Author string
	}

	type RecordsServerWeb struct {
		ID     int
		Tittle string
		Text   string
		Author string
	}

	type PagData struct {
		ClientRecord     []ClientRecord
		RecordsServerWeb []RecordsServerWeb
	}

	rows, err := db.Query(`SELECT id, tittle, text, author FROM servidorcliente ORDER BY id DESC`)
	if err != nil {
		http.Error(w, "Error fetching data from database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recordList []ClientRecord

	for rows.Next() {
		var rec ClientRecord
		err := rows.Scan(&rec.ID, &rec.Tittle, &rec.Text, &rec.Author)
		if err != nil {
			http.Error(w, "Error reading records: "+err.Error(), http.StatusInternalServerError)
			return
		}
		recordList = append(recordList, rec)
	}

	rows2, err := db.Query(`SELECT id, tittle, text, author FROM webserver ORDER BY id DESC`)
	if err != nil {
		http.Error(w, "Error fetching data from database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	var recordList2 []RecordsServerWeb

	for rows2.Next() {
		var rec2 RecordsServerWeb
		err := rows2.Scan(&rec2.ID, &rec2.Tittle, &rec2.Text, &rec2.Author)
		if err != nil {
			http.Error(w, "Error reading records: "+err.Error(), http.StatusInternalServerError)
			return
		}
		recordList2 = append(recordList2, rec2)
	}

	dataOfPag := PagData{
		ClientRecord:     recordList,
		RecordsServerWeb: recordList2,
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Error loading template:"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, dataOfPag)
	if err != nil {
		http.Error(w, "Error rendering HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

}

func main() {

	port := os.Getenv("PORT")

	DbSqlite()

	// Backend navigation routes
	http.HandleFunc("/", Indexhtml)

	// Backend navigation routes

	//routes with form
	http.HandleFunc("/postar", InsertDb) //postar pelo form post

	//routes with form

	fs := http.FileServer(http.Dir("/static/"))
	http.Handle("/static/", http.StripPrefix("static/", fs))

	// External communication routes

	http.HandleFunc("/senttolocalclient", SentToLocalClient)
	http.HandleFunc("/recievingfromlocalclient", RecievingFromLocalClient)
	http.HandleFunc("/renderonserver", RenderOnServer)

	// External communication routes

	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
