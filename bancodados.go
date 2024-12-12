package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type Usuario struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "./api.db")
	if err != nil {
		panic("Erro ao conectar ao banco de dados: " + err.Error())
	}
	defer db.Close()

	CriarTabela()

	fmt.Println("Conexão com SQLite estabelecida!")

	r := mux.NewRouter()

	r.HandleFunc("/usuarios", ListarUsuarios).Methods("GET")
	r.HandleFunc("/usuarios", CriarUsuario).Methods("POST")
	r.HandleFunc("/usuarios/{id}", AtualizarUsuario).Methods("PUT")
	r.HandleFunc("/usuarios/{id}", ExcluirUsuario).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	fmt.Printf("Servidor rodando na porta %s...\n", port)
	http.ListenAndServe(":"+port, r)
}

func CriarTabela() {
	query := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		panic("Erro ao criar tabela: " + err.Error())
	}
	fmt.Println("Tabela criada ou já existente!")
}

func CriarUsuario(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var u Usuario
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if u.Nome == "" {
		http.Error(w, "O nome não pode ser vazio", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO usuarios (nome) VALUES (?)", u.Nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	u.ID = int(id)
	json.NewEncoder(w).Encode(u)
}

func ListarUsuarios(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("SELECT id, nome FROM usuarios")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var usuarios []Usuario
	for rows.Next() {
		var u Usuario
		if err := rows.Scan(&u.ID, &u.Nome); err != nil {
			http.Error(w, "Erro ao ler os dados", http.StatusInternalServerError)
			return
		}
		usuarios = append(usuarios, u)
	}
	json.NewEncoder(w).Encode(usuarios)
}

func AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
    
    params := mux.Vars(r)
    id := params["id"]

    
    idInt, err := strconv.Atoi(id)
    if err != nil {
        http.Error(w, "ID inválido", http.StatusBadRequest)
        return
    }
   
    var u Usuario
    err = json.NewDecoder(r.Body).Decode(&u)
    if err != nil {
        http.Error(w, "Dados inválidos", http.StatusBadRequest)
        return
    }
     
    result, err := db.Exec("UPDATE usuarios SET nome = ? WHERE id = ?", u.Nome, idInt)
    if err != nil {
        return
    }
     
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Usuário não encontrado", http.StatusNotFound)
        return
    }

    u.ID = idInt // O ID permanece o mesmo
    json.NewEncoder(w).Encode(u)
}

func ExcluirUsuario(w http.ResponseWriter, r *http.Request) {

    params := mux.Vars(r)
    id := params["id"]

    result, err := db.Exec("DELETE FROM usuarios WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Usuário não encontrado", http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Usuário excluído com sucesso!"))
}