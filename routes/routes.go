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

	protected.HandleFunc("/insert", handlers.Insert).Methods("POST")
	protected.HandleFunc("/get", handlers.Get).Methods("POST")
	protected.HandleFunc("/update", handlers.Update).Methods("POST")
	protected.HandleFunc("/delete", handlers.Delete).Methods("POST")
	protected.HandleFunc("/getqs", handlers.GetQueryString).Methods("GET")

	return r
}

func StartServer(port string) {
	r := SetupRouter()
	log.Println("🚀 Servidor iniciado (modo interno) na porta", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		log.Fatal("❌ Erro ao iniciar servidor:", err)
	}
}

