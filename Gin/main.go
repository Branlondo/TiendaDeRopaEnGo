// main.go — punto de entrada de la aplicación.
// Carga las variables de entorno desde .env, inicializa PostgreSQL y arranca el servidor.
package main

import (
	"log"
	"os"

	"Gin/db"
	"Gin/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno desde el archivo .env.
	// Si el archivo no existe se omite (en producción las vars vienen del sistema).
	if err := godotenv.Load(); err != nil {
		log.Println("main: archivo .env no encontrado, usando variables del sistema")
	}

	// Conectar a PostgreSQL y crear las tablas si no existen.
	db.Init()

	r := gin.Default()

	// Leer la clave secreta de sesión desde el entorno.
	// Si no está definida usa un fallback (solo para desarrollo local).
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "clave-de-desarrollo-local-no-usar-en-produccion"
		log.Println("ADVERTENCIA: SESSION_SECRET no está definida, usando clave de desarrollo")
	}

	// cookie.NewStore crea un almacén de sesiones basado en cookies firmadas.
	store := cookie.NewStore([]byte(secret))
	r.Use(sessions.Sessions("rff_session", store))

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8181"
	}
	log.Printf("main: servidor escuchando en http://localhost:%s", port)
	r.Run(":" + port)
}
