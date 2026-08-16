# GottPolling

### https://github.com/VictorOliveira-stack/GottPolling

* [Objective](#objective)
  
* [Tecnologies](#tecnologies)
  
* [Usage](#Usage)
  

## sync/offline-first

GottPolling was developed with the intention of connecting an online server
to a local server without having to deal with the NAT network layer, allowing
communication without opening ports on the router, but rather through polling,
with this method being adopted instead of WebSocket.

## Objective

### The goal of the project is to avoid having to pay for a database...

Looking for a solution to connect two servers, one local and the other on a VPS
at low cost or free, GottPolling proposes a solution to keep the application rendering
on a free VPS the ephemeral data of an SQLite database, which may be deleted after a
new deployment or a period of inactivity that causes the instance to hibernate and delete files such as databases
in SQLite or JSON from the file system. As mentioned above, the intention is to avoid having to pay for a database
that is expensive from the beginning (obviously considering the needs of each project), allowing it to scale a little
more, or even to the level the company prefers. Notably, SQLite on the local machine is not the only option,
however, at this initial proposal level, the idea is to render the data on the internet in a lightweight way,
through SQLite.
In this approach, there are two databases synchronizing with each other, one on a VPS and the other locally; if
either one is deleted, both can restore themselves because they contain the same data and communicate with each other, by using
the offline-first approach.

### There is a need to have a local server running; however, with some ingenuity, interesting solutions can be found that allow the web server to be turned off, since this is one of the proposals of the presented approach.

## Technologies

**Golang:**

**SQLite:**

## Usage

Before being modified into a library, GottPolling should be used as follows:

Importing standard Golang dependencies.

**Web Server**

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



**Local Server**

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

**Web Server create the main with these routes and functions:**



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

Local Server create the main with these routes and functions:

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

**Web Server create these polling functions:**

// sent to client

// from the server to the client // server is here

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

// receiving from the client

// from the client to the server
// server is here

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

**Local Server create these polling functions:**

// receiving from server

// from the server to the client // client is here

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

// from the client to the server
// client is here

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
	// 1. Fetch data from 'servidorcliente' in the local database
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

**Note that:**
note that there are four functions communicating with each other. They could be two, but they were divided to make visualization easier.
**You** can copy these functions for other communications that require polling.
On the web server, copy the contents of:
**SentToLocalClient** e **RecievingFromLocalClient** which will be on the web server side of the code.
On the local server, copy the contents of:
**SentToLocalClient** e **RecievingFromLocalClient** which will be written on the local server side.

**the function names remain the same to make it easier to find your way around**



* **to check online whether it is working, use these two functions:**



**Web Server create this function:**


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
	http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
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

* **Local Server create this function:**

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
	// 1. Fetch data from 'servidorcliente' in the local database
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

	// 2. Convert to JSON and send via POST to Render
	jsonData, _ := json.Marshal(records)
	resp, err := http.Post("https://yourDomainOrLocalhost/renderonserver", "application/json", bytes.NewBuffer(jsonData))
	if err == nil {
		resp.Body.Close()
	}
}

}

**For a better understanding, you can check the code in the repository**
