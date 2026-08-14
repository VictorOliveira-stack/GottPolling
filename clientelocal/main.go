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
		panic("Erro ao criar diretório db: " + err.Error())
	}

	var err error

	dbPath := filepath.Join(dbDir, "servidorweb.db")

	db, err = sql.Open("sqlite", dbPath+"?_timeout=5000")
	if err != nil {
		panic("Erro ao abrir banco: " + err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS webservidor(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			titulo TEXT,
			texto TEXT,
			autor TEXT
	);
	
	CREATE TABLE IF NOT EXISTS servidorcliente(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		titulo TEXT,
		texto TEXT,
		autor TEXT
	);`)
	if err != nil {
		panic("Erro ao criar tabela: " + err.Error())
	}

}

func InsertDb() {

	titulo := "servidorcliente"
	texto := "ta tudo louco tem que fazer cada tabela conversar com sua xará apenas"
	autor := "servidorcliente"

	query := `INSERT INTO servidorcliente (titulo, texto, autor) VALUES (?,?,?)`

	_, err := db.Exec(query, titulo, texto, autor)
	if err != nil {
		fmt.Println(err)
	}

}

func InsertDb2() {

	titulo := "teste de ajuste"
	texto := "dados vindo do cliente"
	autor := "vindo do servidorcliente observação. onde está renderizando?"

	query := `INSERT INTO webservidor (titulo, texto, autor) VALUES (?,?,?)`

	_, err := db.Exec(query, titulo, texto, autor)
	if err != nil {
		fmt.Println(err)
	}

}

// recebendo do servidor
// insert local
// do servidor para o cliente
func EnviarParaClienteLocal() {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	type RespostaVindaDoServidor struct {
		ID     int    `json:"id"`
		Titulo string `json:"titulo"`
		Texto  string `json:"texto"`
		Autor  string `json:"autor"`
	}

	for range ticker.C {

		resp, err := http.Get("https://dbsqlite.onrender.com/receberdocliente")
		if err != nil {
			// Silencia o erro se o servidor web estiver offline
			continue
			//panic(err)
		}
		//defer resp.Body.Close()
		//fmt.Println("erro em fetch http://localhost:8080/fullsync", err)

		var registros []RespostaVindaDoServidor
		err = json.NewDecoder(resp.Body).Decode(&registros)
		resp.Body.Close()
		if err != nil {
			fmt.Println("[Cliente] Erro ao decodificar JSON do servidor:", err)
			continue
		}

		for _, r := range registros {

			_, err := db.Exec(`
				INSERT INTO webservidor
				(id, titulo, texto, autor)
				VALUES (?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET titulo=excluded.titulo,
				 texto=excluded.texto, autor=excluded.autor
				`,
				r.ID,
				r.Titulo,
				r.Texto,
				r.Autor,
			)

			if err != nil {
				fmt.Println("[Cliente] Erro ao salvar dados do servidor:", err)
			}
		}
	}
}

// enviando para servidor
// do cliente para o servidor
func ReceberDoClienteNoServidorWeb() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	type EnviarAoCliente struct {
		ID     int    `json:"id"`
		Titulo string `json:"titulo"`
		Texto  string `json:"texto"`
		Autor  string `json:"autor"`
	}

	for range ticker.C {
		// 1. Busca dados de 'servidorcliente' no banco local
		rows, err := db.Query(`SELECT id, titulo, texto, autor FROM webservidor`)
		if err != nil {
			fmt.Println("[Cliente-WebServidor] Erro na query local:", err)
			continue
		}

		var registros []EnviarAoCliente
		for rows.Next() {
			var r EnviarAoCliente
			if err := rows.Scan(&r.ID, &r.Titulo, &r.Texto, &r.Autor); err == nil {
				registros = append(registros, r)
			}
		}
		rows.Close()

		if len(registros) == 0 {
			continue
		}

		// 2. Transforma em JSON e envia via POST para o Render
		jsonData, err := json.Marshal(registros)
		if err != nil {
			fmt.Println("[Cliente-WebServidor] Erro ao gerar JSON:", err)
			continue
		}

		resp, err := http.Post("https://dbsqlite.onrender.com/receberdoclienteservidorweb", "application/json",
			bytes.NewBuffer(jsonData),
		)

		if err != nil {
			fmt.Println("[Cliente-WebServidor] Erro no envio HTTP:", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("[Cliente-WebServidor] Servidor respondeu com status: %s\n", resp.Status)
		} else {
			fmt.Println("[Cliente-WebServidor] Sincronização realizada com sucesso!")
		}

		resp.Body.Close()

	}
}

func HtmlClient(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Erro ao carregar template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

// enviando para servidor
//do cliente para o servidor
//INSERT INTO servidorcliente //vou manter pra poder manter a vizualização no servidor
//aqui só ta indo não ta escrevendo na table servidorcliente local mais só no servidorcliente no render

func EnviarParaServidor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	type EnviarAoCliente struct {
		ID     int    `json:"id"`
		Titulo string `json:"titulo"`
		Texto  string `json:"texto"`
		Autor  string `json:"autor"`
	}

	for range ticker.C {
		// 1. Busca dados de 'servidorcliente' no banco local
		rows, err := db.Query(`SELECT id, titulo, texto, autor FROM webservidor`)
		if err != nil {
			continue
		}

		var registros []EnviarAoCliente
		for rows.Next() {
			var r EnviarAoCliente
			rows.Scan(&r.ID, &r.Titulo, &r.Texto, &r.Autor)
			registros = append(registros, r)
		}
		rows.Close()

		if len(registros) == 0 {
			continue
		}

		// 2. Transforma em JSON e envia via POST para o Render
		jsonData, _ := json.Marshal(registros)
		resp, err := http.Post("https://dbsqlite.onrender.com/enviarparaservidor", "application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			resp.Body.Close()
		}
	}
}

func main() {

	DbSqlite()
	InsertDb()
	InsertDb2()

	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", HtmlClient)

	//endpoint para polling
	//http.HandleFunc("/enviarparaservidor", EnviarParaServidor) //endpoint
	go EnviarParaServidor()
	//inicia o polling em backgroung
	go EnviarParaClienteLocal()

	go ReceberDoClienteNoServidorWeb()

	fmt.Println("servidor Cliente rodando :9090")
	fmt.Println("acesse: http://localhost:9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		panic(err)
	}

}
