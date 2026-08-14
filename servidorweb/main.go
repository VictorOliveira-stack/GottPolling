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

func InsertDb(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}

	/*titulo := "inserindo no sqlite online"
	texto := "ola mundo estamos online"
	autor := "inserido online"*/

	titulo := r.FormValue("titulo")
	texto := r.FormValue("texto")
	autor := r.FormValue("autor")

	if titulo == "" || texto == "" || autor == "" {
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO webservidor (titulo, texto, autor) VALUES (?,?,?)`

	_, err := db.Exec(query, titulo, texto, autor)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

//enviar para o cliente
//do servidor para o cliente

func EnviarParaClienteLocal(w http.ResponseWriter, r *http.Request) {

	type EnviarAoCliente struct {
		ID     int    `json:"id"`
		Titulo string `json:"titulo"`
		Texto  string `json:"texto"`
		Autor  string `json:"autor"`
	}

	rows, err := db.Query(`SELECT id, titulo, texto, autor FROM webservidor`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var registros []EnviarAoCliente

	for rows.Next() {

		var registro EnviarAoCliente

		err := rows.Scan(
			&registro.ID,
			&registro.Titulo,
			&registro.Texto,
			&registro.Autor,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		registros = append(registros, registro)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(registros)

}

// do cliente para o servidor
func ReceberDoClienteNoServidorWeb(w http.ResponseWriter, r *http.Request) {

	/*ticker := time.NewTicker(5 * time.Second) // Tenta a cada 5 segundos
	defer ticker.Stop()*/

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido (use POST)", http.StatusMethodNotAllowed)
		return
	}

	type DadosRecebido struct {
		ID     int    `json:"id"`
		Titulo string `json:"titulo"`
		Texto  string `json:"texto"`
		Autor  string `json:"autor"`
	}

	var registros []DadosRecebido

	err := json.NewDecoder(r.Body).Decode(&registros)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
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
			log.Println("[Servidor] Erro ao salvar registro no banco:", err)
			http.Error(w, "Erro ao salvar no banco local", http.StatusInternalServerError)
			return
		}

	}

	//Responder ao cliente indicando sucesso
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "sucesso",
		"message": "Dados sicronizados no servidor com sucesso!",
	})

}

// recebendo do cliente
//do cliente para o servidor

//INSERT INTO servidorcliente //vou manter pra poder manter a vizualização no servidor//EnviarParaServidor()
//aqui só ta vindo para servidorcliente no render mas vindo de webservidor no cliente não escrevendo mais no servidorcliente local

func Receberdocliente(w http.ResponseWriter, r *http.Request) {

	/*ticker := time.NewTicker(5 * time.Second) // Tenta a cada 5 segundos
	defer ticker.Stop()*/

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido (use POST)", http.StatusMethodNotAllowed)
		return
	}

	type DadosRecebido struct {
		ID     int    `json:"id"`
		Titulo string `json:"titulo"`
		Texto  string `json:"texto"`
		Autor  string `json:"autor"`
	}

	var registros []DadosRecebido

	err := json.NewDecoder(r.Body).Decode(&registros)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	for _, r := range registros {
		_, err := db.Exec(`
			INSERT INTO servidorcliente
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
			log.Println("[Servidor] Erro ao salvar registro no banco:", err)
			http.Error(w, "Erro ao salvar no banco local", http.StatusInternalServerError)
			return
		}

	}

	//Responder ao cliente indicando sucesso
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "sucesso",
		"message": "Dados sicronizados no servidor com sucesso!",
	})

}

func Indexhtml(w http.ResponseWriter, r *http.Request) {

	type RegistroCliente struct {
		ID     int
		Titulo string
		Texto  string
		Autor  string
	}

	type RegistroServidorweb struct {
		ID     int
		Titulo string
		Texto  string
		Autor  string
	}

	type PaginaData struct {
		RegistroCliente     []RegistroCliente
		RegistroServidorweb []RegistroServidorweb
	}

	rows, err := db.Query(`SELECT id, titulo, texto, autor FROM servidorcliente ORDER BY id DESC`)
	if err != nil {
		http.Error(w, "Erro ao buscar dados do banco: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listaRegistros []RegistroCliente

	for rows.Next() {
		var reg RegistroCliente
		err := rows.Scan(&reg.ID, &reg.Titulo, &reg.Texto, &reg.Autor)
		if err != nil {
			http.Error(w, "Erro ao ler registros: "+err.Error(), http.StatusInternalServerError)
			return
		}
		listaRegistros = append(listaRegistros, reg)
	}

	rows2, err := db.Query(`SELECT id, titulo, texto, autor FROM webservidor ORDER BY id DESC`)
	if err != nil {
		http.Error(w, "Erro ao buscar dados do banco: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	var listaRegistros2 []RegistroServidorweb

	for rows2.Next() {
		var reg2 RegistroServidorweb
		err := rows2.Scan(&reg2.ID, &reg2.Titulo, &reg2.Texto, &reg2.Autor)
		if err != nil {
			http.Error(w, "Erro ao ler registros: "+err.Error(), http.StatusInternalServerError)
			return
		}
		listaRegistros2 = append(listaRegistros2, reg2)
	}

	dadosDaPagina := PaginaData{
		RegistroCliente:     listaRegistros,
		RegistroServidorweb: listaRegistros2,
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Erro ao carregar template"+err.Error(), http.StatusInternalServerError)
		return
	}
	//err = tmpl.Execute(w, nil)
	//fmt.Println("err", err)
	//tmpl.Execute(w, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//err = tmpl.Execute(w, listaRegistros)
	err = tmpl.Execute(w, dadosDaPagina)
	if err != nil {
		http.Error(w, "Erro ao renderizar HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

}

func main() {

	port := os.Getenv("PORT")

	DbSqlite()

	//rotas de direcionamento
	http.HandleFunc("/", Indexhtml)

	//rotas de direcionamento

	//rotas com form
	http.HandleFunc("/postar", InsertDb) //postar pelo form post

	//rotas com form

	fs := http.FileServer(http.Dir("/static/"))
	http.Handle("/static/", http.StripPrefix("static/", fs))

	//rotas de comunicaçao externa
	http.HandleFunc("/receberdocliente", EnviarParaClienteLocal)
	http.HandleFunc("/enviarparaservidor", Receberdocliente)
	http.HandleFunc("/receberdoclienteservidorweb", ReceberDoClienteNoServidorWeb)

	//rotas de comunicaçao externa

	//go IniciarPollingCliente()

	log.Printf("Servidor rodando na porta %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	/*fmt.Println("Servidor rodando em :8080")
	fmt.Println("acesse: http:localhost:8080")

	if err := http.ListenAndServe(":8080", nil); //aqui é o servidor
	err != nil {
		panic(err)
	}*/

}

//estava construindo a struct para enviar para o cliente, sai para criar o cliente
