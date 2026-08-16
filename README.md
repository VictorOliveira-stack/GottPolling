# GottPolling

* [Objetivo](#objetivo)
* [Tecnologias](#tecnologias)

# sync/offline-first

  O GottPolling foi desenvolvido na intenção de conectar um servidor online
a um servidor local sem precisar lhe dá com a camada de rede NAT, permitindo
a comunicação sem aberturas de portas no roteador, mas sim atravéz de polling,
sendo esse metodo adotado em vez de websocket.

# Objetivo
## O objetivo do projeto é não precisar pagar um banco de dados...
  Buscando uma solução para poder conectar dois servidores sendo um local e outro em uma VPS
de baixo custo ou grátis, o GottPolling propõe um solução para manter renderizando na aplicação
em uma VPS grátis os dados efemeros de um banco de dados SQLite, que podem ser apagados após um
novo deploy ou tempo de inatividade que leva a instância a hibernar e apaga arquivos como databases
em SQLite ou Json do fileSystem. Como falado acima a intenção é não precisar pagar um banco de dados
caro logo de inicio (obviamente observando a nescessidade de cada projeto), permitindo escalar um pouco
mais ou até o nivel que a empresa preferir, notoriamente o SQLite na maquina local não é a única opção,
porém nessa camada de proposta inicial a idéia é poder renderizar os dados de forma leve na internet,
atravéz do SQLite.
  Nessa abordagem há dois bancos de dados sicronizando entre si, um em uma VPS e outro localmente, havendo
a exclusão de qualquer um ambos se restauram pois contém os mesmos dados e se comunicam entre si, por usar
a abordagem de offline-first.
## existe a nescessidade de um servidor local rodando porém com engenhosidade pode se encontrar algumas soluções interessantes que permitam desligar o servidor local, pois é essa uma das propostas do esquema apresentado.
Não me deterei em falar sobre essas soluções.

# Tecnologias

* **Golang:**
* **SQLite:**

## Utilizando

Antes de ser modificado para ser uma biblioteca deve se usar o GottPolling assim:

* **Importando** dependencias padrões do Golang

* **Servidor Web**


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



* **Servidor Local**
  
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
  

* **Servidor Web** crie o main com essas rotas e funções:



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


* **Servidor Local** crie o main com essas rotas e funções:

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


* **Servidor Web** crie essas funções de polling:

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


* **Servidor Local** crie essas funções de polling:

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

