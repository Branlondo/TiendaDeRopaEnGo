// Define las rutas de la app y importa el paquete gin y define las funciones que se encargaran de procesar las solicitudes HTTP.
package routes

import (
	//se necesita para trabajar con el protocolo HTTP y manejar las respuestas
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configura las rutas para la aplicación Gin
func SetupRoutes(r *gin.Engine) {
	//esto funciona para servir archivos estáticos desde el directorio "./static" cuando se accede a la ruta "/static"
	r.Static("/static", "./static")
	//esto indica a Gin que cargue las plantillas HTML desde el directorio "templates" con la extensión ".html"
	r.LoadHTMLGlob("templates/*.html")

	//se define una ruta para la raíz ("/") que renderiza la plantilla "index.html" cuando se accede a ella
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	//se define una ruta para "/product" que renderiza la plantilla "product.html" cuando se accede a ella
	r.GET("/product", func(c *gin.Context) {
		c.HTML(200, "product.html", nil)
	})
	//se define una ruta para cualquier página ("/:page") que verifica si la plantilla correspondiente existe y la renderiza, o muestra una página de error 404 si no existe
	r.GET("/:page", func(c *gin.Context) {
		page := c.Param("page")
		//si el nombre de la página no termina con ".html", se le agrega esa extensión
		if !strings.HasSuffix(page, ".html") {
			//si el nombre de la página no termina con ".html", se le agrega esa extensión
			page += ".html"
		}
		//se verifica si el archivo de plantilla existe en el directorio "templates". Si existe, se renderiza la página; de lo contrario, se muestra una página de error 404
		if _, err := os.Stat("templates/" + page); err == nil {
			//si el archivo existe, se renderiza la página
			c.HTML(http.StatusOK, page, nil)
			//si el archivo no existe, se muestra una página de error 404
		} else {
			//si el archivo no existe, se muestra una página de error 404
			c.HTML(http.StatusNotFound, "404.html", nil)
		}
	})

}
