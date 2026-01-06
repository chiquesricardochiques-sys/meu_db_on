package main

import (
	"log"
	"meu-provedor/config"
	"meu-provedor/routes"
	"os"
)

func main() {
	// 1️⃣ Carrega variáveis de ambiente
	config.LoadEnv()

	// 2️⃣ Conecta ao banco master
	config.ConnectMaster()

	// 3️⃣ Define porta do servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // porta padrão
	}

	// 4️⃣ Inicia servidor HTTP
	log.Println("🌐 Servidor rodando na porta", port)
	routes.StartServer(port)
}
