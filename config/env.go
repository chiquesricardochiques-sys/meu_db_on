package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

func LoadEnv() {
    // Tenta carregar um arquivo .env se existir
    err := godotenv.Load()
    if err != nil {
        log.Println("⚠️ Aviso: .env não encontrado, usando variáveis do sistema")
    }
}

func GetEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        log.Fatalf("❌ ERRO: Variável de ambiente %s não definida", key)
    }
    return value
}
