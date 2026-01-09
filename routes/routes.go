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

	// ====================================================
	// ROTAS PROTEGIDAS (USO INTERNO / BACKEND)
	// ====================================================
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(security.InternalOnly)

	// ====================================================
	// CRUD GENÉRICO POR INSTÂNCIA (DADOS DO USUÁRIO FINAL)
	// ====================================================
	// Todas essas rotas:
	// - Exigem project_id
	// - Exigem id_instancia
	// - Nunca vazam dados entre clientes
	protected.HandleFunc("/data/insert", handlers.Insert).Methods("POST")
	protected.HandleFunc("/data/get", handlers.Get).Methods("POST")
	protected.HandleFunc("/data/update", handlers.Update).Methods("POST")
	protected.HandleFunc("/data/delete", handlers.Delete).Methods("POST")

	// Query string segura (uso interno)
	protected.HandleFunc("/data/getqs", handlers.GetQueryString).Methods("GET")

	// ====================================================
	// PROJETOS (TEMPLATES DE SISTEMA)
	// ====================================================
	protected.HandleFunc("/projects", handlers.ListProjects).Methods("GET")
	protected.HandleFunc("/projects", handlers.CreateProject).Methods("POST")
	protected.HandleFunc("/projects/{id}", handlers.UpdateProject).Methods("PUT")
	protected.HandleFunc("/projects/{id}", handlers.DeleteProject).Methods("DELETE")

	// ====================================================
	// INSTÂNCIAS (CLIENTES)
	// ====================================================
	protected.HandleFunc("/instances", handlers.ListInstances).Methods("GET")
	protected.HandleFunc("/instances", handlers.CreateInstance).Methods("POST")
	protected.HandleFunc("/instances/{id}", handlers.UpdateInstance).Methods("PUT")
	protected.HandleFunc("/instances/{id}", handlers.DeleteInstance).Methods("DELETE")

	// ====================================================
	// GERENCIAMENTO DE SCHEMA (ESTRUTURA DOS PROJETOS)
	// ====================================================
	// Criação de tabelas do projeto
	protected.HandleFunc("/schema/table", handlers.CreateProjectTable).Methods("POST")

	// Listagem de tabelas do projeto
	protected.HandleFunc("/schema/tables", handlers.ListProjectTables).Methods("GET")

	// Remoção de tabela
	protected.HandleFunc("/schema/table", handlers.DeleteProjectTable).Methods("DELETE")

	// Alteração de colunas
	protected.HandleFunc("/schema/column", handlers.AddColumn).Methods("POST")
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
