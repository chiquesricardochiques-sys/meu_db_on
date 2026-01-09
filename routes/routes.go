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

	// ===================== ROTAS PROTEGIDAS =====================
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(security.InternalOnly)

	// ===================== CRUD GENÉRICO (JÁ EXISTENTE) =====================
	protected.HandleFunc("/insert", handlers.Insert).Methods("POST")
	protected.HandleFunc("/get", handlers.Get).Methods("POST")
	protected.HandleFunc("/update", handlers.Update).Methods("POST")
	protected.HandleFunc("/delete", handlers.Delete).Methods("POST")
	protected.HandleFunc("/getqs", handlers.GetQueryString).Methods("GET")

	// ===================== PROJETOS =====================
	protected.HandleFunc("/projects", handlers.ListProjects).Methods("GET")
	protected.HandleFunc("/projects", handlers.CreateProject).Methods("POST")
	protected.HandleFunc("/projects/{id}", handlers.UpdateProject).Methods("PUT")
	protected.HandleFunc("/projects/{id}", handlers.DeleteProject).Methods("DELETE")

	// ===================== INSTÂNCIAS =====================
	protected.HandleFunc("/instances", handlers.ListInstances).Methods("GET")
	protected.HandleFunc("/instances", handlers.CreateInstance).Methods("POST")
	protected.HandleFunc("/instances/{id}", handlers.UpdateInstance).Methods("PUT")
	protected.HandleFunc("/instances/{id}", handlers.DeleteInstance).Methods("DELETE")

	// ===================== GERENCIAMENTO DE TABELAS (CORE) =====================
	// Criar tabela para um projeto
	protected.HandleFunc("/schema/table", handlers.CreateProjectTable).Methods("POST")

	// Listar tabelas de um projeto
	protected.HandleFunc("/schema/tables", handlers.ListProjectTables).Methods("GET")

	// Deletar tabela de um projeto
	protected.HandleFunc("/schema/table", handlers.DeleteProjectTable).Methods("DELETE")

	// Adicionar coluna em tabela
	protected.HandleFunc("/schema/column", handlers.AddColumn).Methods("POST")

	// Remover coluna de tabela
	protected.HandleFunc("/schema/column", handlers.DropColumn).Methods("DELETE")

	return r
}

func StartServer(port string) {
	r := SetupRouter()
	log.Println("🚀 Servidor iniciado (modo interno) na porta", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		log.Fatal("❌ Erro ao iniciar servidor:", err)
	}
}
