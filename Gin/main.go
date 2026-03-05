//se encarga de crear el router se encarga de iniciar el serve y importar route y usar las funciones necesarias para registrar rutas en el router

package main

//importa el paquete "routes" que contiene la configuración de las rutas de la aplicación y el paquete "github.com/gin-gonic/gin" para trabajar con el framework Gin en Go
import (
	//Manda a importar las rutas
	"Gin/routes"
	// se trabaja el servidor en gin
	"github.com/gin-gonic/gin"
)

// func main es el punto de entrada de la aplicación, donde se crea el router, se configuran las rutas y se inicia el servidor en el puerto 8181
func main() {
	//sirve para crear un nuevo router de Gin con la configuración predeterminada, que incluye middleware para el registro de solicitudes y la recuperación de pánicos
	r := gin.Default()
	//se llama a la función SetupRoutes del paquete routes, pasando el router r como argumento para configurar las rutas de la aplicación
	routes.SetupRoutes(r)
	//el servidor 8080 el cual es por defecto tiene un problema asi que se cambia a 8181 para evitar conflictos con otros servidores que puedan estar corriendo en el puerto 8080
	r.Run(":8181")
}
