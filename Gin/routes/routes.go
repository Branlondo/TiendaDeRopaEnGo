// routes.go — configura todas las rutas HTTP de la aplicación.
package routes

import (
	"html/template"
	"net/http"
	"strconv"

	// carritoPkg contiene toda la lógica de negocio del carrito:
	// agregar, eliminar, actualizar cantidad, calcular total.
	carritoPkg "Gin/carrito"
	"Gin/models"

	"github.com/gin-gonic/gin"
)

// subcategoriasUnicas extrae las subcategorías únicas de un slice de productos,
// respetando el orden de aparición. Así el menú de filtros se genera
// dinámicamente: agregar un nuevo producto con Subcategoria="Chaqueta" hace
// que "Chaqueta" aparezca sola en el dropdown sin tocar el template.
func subcategoriasUnicas(productos []models.Producto) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, p := range productos {
		if p.Subcategoria != "" && !seen[p.Subcategoria] {
			seen[p.Subcategoria] = true
			result = append(result, p.Subcategoria)
		}
	}
	return result
}

// contarItemsCarrito es un helper que los handlers usan para saber
// cuántos artículos hay en el carrito y pasarlos al template como
// "cartCount". Así el badge del navbar siempre muestra el número real.
func contarItemsCarrito(c *gin.Context) int {
	items := carritoPkg.Obtener(c)
	return carritoPkg.ContarItems(items)
}

// SetupRoutes registra todas las rutas en el router de Gin.
func SetupRoutes(r *gin.Engine) {

	// FuncMap permite usar funciones personalizadas dentro de los templates.
	// "add" y "sub" nos permiten escribir {{ sub .Cantidad 1 }} en el HTML
	// para calcular la nueva cantidad en los botones + y - del carrito.
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int {
			if a-b < 1 {
				return 1
			}
			return a - b
		},
	})

	r.Static("/static", "./static")
	// IMPORTANTE: LoadHTMLGlob debe ir DESPUÉS de SetFuncMap
	r.LoadHTMLGlob("templates/*.html")

	// ── Página principal ──
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// ── Sección Hombre ──
	r.GET("/hombre", func(c *gin.Context) {
		var filtrados []models.Producto
		for _, p := range models.Productos {
			if p.CategoriaID == 111 {
				filtrados = append(filtrados, p)
			}
		}
		c.HTML(http.StatusOK, "Hombre.html", gin.H{
			"productos":     filtrados,
			"subcategorias": subcategoriasUnicas(filtrados),
			// cartCount se pasa a TODOS los templates que tienen navbar
			// para que el badge siempre muestre el número correcto.
			"cartCount": contarItemsCarrito(c),
		})
	})

	// ── Sección Mujer ──
	r.GET("/mujer", func(c *gin.Context) {
		var filtrados []models.Producto
		for _, p := range models.Productos {
			if p.CategoriaID == 222 {
				filtrados = append(filtrados, p)
			}
		}
		c.HTML(http.StatusOK, "Mujer.html", gin.H{
			"productos":     filtrados,
			"subcategorias": subcategoriasUnicas(filtrados),
			"cartCount":     contarItemsCarrito(c),
		})
	})

	// ── Detalle de producto ──
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
			"producto":  productoEncontrado,
			"cartCount": contarItemsCarrito(c),
		})
	})

	// ── Carrito (GET) — Go renderiza los items desde la sesión ──
	r.GET("/carrito", func(c *gin.Context) {
		items := carritoPkg.Obtener(c)
		total := carritoPkg.Total(items)
		c.HTML(http.StatusOK, "Carrito.html", gin.H{
			"items":     items,
			"total":     total,
			"productos": models.Productos, // productos relacionados (sección inferior)
			"cartCount": carritoPkg.ContarItems(items),
		})
	})

	// ── Carrito (POST) — Agregar producto ──
	// Recibe: productoID (int), talla (string), cantidad (int)
	// El formulario en product.html envía estos campos con method="POST".
	r.POST("/carrito/agregar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		talla := c.PostForm("talla")
		cantidad, err := strconv.Atoi(c.PostForm("cantidad"))
		if err != nil || cantidad < 1 {
			cantidad = 1
		}
		carritoPkg.Agregar(c, id, talla, cantidad)

		// Redirigimos a la página anterior (la del producto).
		// Si no hay Referer (p.ej. llamada directa), vamos al carrito.
		referer := c.Request.Referer()
		if referer == "" {
			referer = "/carrito"
		}
		// 303 See Other es el código correcto después de un POST exitoso:
		// le dice al navegador que haga un GET a la URL de destino.
		c.Redirect(http.StatusSeeOther, referer)
	})

	// ── Carrito (POST) — Eliminar producto ──
	// Recibe: productoID (int), talla (string)
	r.POST("/carrito/eliminar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		talla := c.PostForm("talla")
		carritoPkg.Eliminar(c, id, talla)
		c.Redirect(http.StatusSeeOther, "/carrito")
	})

	// ── Carrito (POST) — Actualizar cantidad ──
	// Recibe: productoID (int), talla (string), cantidad (int)
	// Si cantidad llega como 0 o negativo, CambiarCantidad lo elimina.
	r.POST("/carrito/actualizar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		talla := c.PostForm("talla")
		cantidad, err := strconv.Atoi(c.PostForm("cantidad"))
		if err != nil {
			cantidad = 1
		}
		carritoPkg.CambiarCantidad(c, id, talla, cantidad)
		c.Redirect(http.StatusSeeOther, "/carrito")
	})
}
