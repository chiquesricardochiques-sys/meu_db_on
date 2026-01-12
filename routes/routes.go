package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"meu-provedor/handlers"
	"meu-provedor/security"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(security.InternalOnly)

	// ====== CRUD GENÉRICO ======
	protected.HandleFunc("/data/insert", handlers.Insert).Methods("POST")
	protected.HandleFunc("/data/get", handlers.Get).Methods("POST")
	protected.HandleFunc("/data/update", handlers.Update).Methods("POST")
	protected.HandleFunc("/data/delete", handlers.Delete).Methods("POST")
	//protected.HandleFunc("/data/getqs", handlers.GetQueryString).Methods("GET")

	// ====== PROJETOS ======
	protected.HandleFunc("/projects", handlers.ListProjects).Methods("GET")
	protected.HandleFunc("/projects", handlers.CreateProject).Methods("POST")
	protected.HandleFunc("/projects/{id}", handlers.UpdateProject).Methods("PUT")
	protected.HandleFunc("/projects/{id}", handlers.DeleteProject).Methods("DELETE")

	// ====== INSTÂNCIAS ======
	protected.HandleFunc("/instances", handlers.ListInstances).Methods("GET")
	protected.HandleFunc("/instances", handlers.CreateInstance).Methods("POST")
	protected.HandleFunc("/instances/{id}", handlers.UpdateInstance).Methods("PUT")
	protected.HandleFunc("/instances/{id}", handlers.DeleteInstance).Methods("DELETE")

	// ====== TABELAS (SCHEMA) ======
	// Criar tabela (aceita com ou sem índices)
	protected.HandleFunc("/schema/table", handlers.CreateProjectTable).Methods("POST")
	
	// Listar tabelas (?detailed=true para detalhes completos)
	protected.HandleFunc("/schema/tables", handlers.ListProjectTables).Methods("GET")
	
	// Detalhes de uma tabela específica
	protected.HandleFunc("/schema/table/details", handlers.GetTableDetails).Methods("GET")
	
	// Deletar tabela
	protected.HandleFunc("/schema/table", handlers.DeleteProjectTable).Methods("DELETE")

	// ====== COLUNAS ======
	protected.HandleFunc("/schema/column", handlers.AddColumn).Methods("POST")
	protected.HandleFunc("/schema/column", handlers.ModifyColumn).Methods("PUT")
	protected.HandleFunc("/schema/column", handlers.DropColumn).Methods("DELETE")

	// ====== ÍNDICES ======
	protected.HandleFunc("/schema/index", handlers.AddIndex).Methods("POST")
	protected.HandleFunc("/schema/index", handlers.DropIndex).Methods("DELETE")

	return r
}

func StartServer(port string) {
	r := SetupRouter()
	log.Println("🚀 Servidor iniciado na porta", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		log.Fatal("❌ Erro ao iniciar servidor:", err)
	}
}

