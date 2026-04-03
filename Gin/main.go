// main.go — punto de entrada de la aplicación.
// Crea el router, registra el middleware de sesiones y arranca el servidor.
package main

import (
	"Gin/routes"

	// sessions permite guardar datos entre peticiones HTTP usando cookies.
	// Así el carrito persiste mientras el usuario navega por la tienda.
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// cookie.NewStore crea un almacén de sesiones basado en cookies firmadas.
	// La clave secreta firma la cookie para que el cliente no pueda manipularla.
	// En producción usarías una clave larga y aleatoria guardada en una variable
	// de entorno, nunca escrita directamente en el código.
	store := cookie.NewStore([]byte("rff-clave-secreta-2026"))

	// sessions.Sessions registra el middleware: a partir de aquí, cada handler
	// puede llamar sessions.Default(c) para leer o escribir la sesión.
	r.Use(sessions.Sessions("rff_session", store))

	routes.SetupRoutes(r)
	r.Run(":8181")
}
