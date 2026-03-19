// routes.go — configura todas las rutas HTTP de la aplicación.
// Cada ruta define qué template renderiza y qué datos le pasa.
package routes

import (
	"net/http"
	"strconv"

	"Gin/models"

	"github.com/gin-gonic/gin"
)

// SetupRoutes registra todas las rutas en el router de Gin.
// Se llama desde main.go pasando el router ya creado.
func SetupRoutes(r *gin.Engine) {

	// Archivos estáticos: CSS, JS, imágenes
	r.Static("/static", "./static")

	// Gin carga todos los archivos .html de la carpeta templates
	r.LoadHTMLGlob("templates/*.html")

	// ── Página principal ──
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// ── Sección Hombre ──
	// Filtra los productos con CategoriaID == 1 y los pasa al template
	r.GET("/hombre", func(c *gin.Context) {
		var filtrados []models.Producto
		for _, p := range models.Productos {
			if p.CategoriaID == 1 {
				filtrados = append(filtrados, p)
			}
		}
		c.HTML(http.StatusOK, "Hombre.html", gin.H{
			"productos": filtrados,
		})
	})

	// ── Sección Mujer ──
	// Filtra los productos con CategoriaID == 2 y los pasa al template
	r.GET("/mujer", func(c *gin.Context) {
		var filtrados []models.Producto
		for _, p := range models.Productos {
			if p.CategoriaID == 2 {
				filtrados = append(filtrados, p)
			}
		}
		c.HTML(http.StatusOK, "Mujer.html", gin.H{
			"productos": filtrados,
		})
	})

	// ── Detalle de producto ──
	// La ruta dinámica :id permite acceder a /producto/1, /producto/2, etc.
	r.GET("/producto/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusNotFound, "product.html", gin.H{"error": "Producto no encontrado"})
			return
		}
		var productoEncontrado *models.Producto
		for i, p := range models.Productos {
			if p.ID == id {
				productoEncontrado = &models.Productos[i]
				break
			}
		}
		if productoEncontrado == nil {
			c.HTML(http.StatusNotFound, "product.html", gin.H{"error": "Producto no encontrado"})
			return
		}
		c.HTML(http.StatusOK, "product.html", gin.H{
			"producto": productoEncontrado,
		})
	})

	// ── Carrito ──
	// Por ahora solo renderiza el template vacio
	r.GET("/carrito", func(c *gin.Context) {
		c.HTML(http.StatusOK, "Carrito.html", nil)
	})
}
