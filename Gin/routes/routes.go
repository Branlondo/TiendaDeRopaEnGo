// routes.go — configura todas las rutas HTTP de la aplicación.
package routes

import (
	"html/template"
	"net/http"
	"strconv"

	"Gin/auth"
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

// contarItemsCarrito devuelve el número total de artículos en el carrito
// de la sesión actual. Se usa para el badge del navbar.
func contarItemsCarrito(c *gin.Context) int {
	items := carritoPkg.Obtener(c)
	return carritoPkg.ContarItems(items)
}

// datosBase construye el mapa de datos que TODOS los templates necesitan:
//   - cartCount: número de artículos en el carrito (para el badge)
//   - usuario:   datos del usuario autenticado, o nil si no hay sesión
//   - flashTipo: tipo del mensaje flash ("success", "danger", etc.)
//   - flashMsg:  texto del mensaje flash (vacío si no hay ninguno)
//
// Cada handler llama datosBase(c) y luego agrega sus propios datos encima.
func datosBase(c *gin.Context) gin.H {
	flashTipo, flashMsg := auth.GetFlash(c)
	return gin.H{
		"cartCount": contarItemsCarrito(c),
		"usuario":   auth.UsuarioActual(c),
		"flashTipo": flashTipo,
		"flashMsg":  flashMsg,
	}
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

	// ── Página principal ──────────────────────────────────────────────────────
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", datosBase(c))
	})

	// ── Sección Hombre ────────────────────────────────────────────────────────
	r.GET("/hombre", func(c *gin.Context) {
		var filtrados []models.Producto
		for _, p := range models.Productos {
			if p.CategoriaID == 111 {
				filtrados = append(filtrados, p)
			}
		}
		data := datosBase(c)
		data["productos"] = filtrados
		data["subcategorias"] = subcategoriasUnicas(filtrados)
		c.HTML(http.StatusOK, "Hombre.html", data)
	})

	// ── Sección Mujer ─────────────────────────────────────────────────────────
	r.GET("/mujer", func(c *gin.Context) {
		var filtrados []models.Producto
		for _, p := range models.Productos {
			if p.CategoriaID == 222 {
				filtrados = append(filtrados, p)
			}
		}
		data := datosBase(c)
		data["productos"] = filtrados
		data["subcategorias"] = subcategoriasUnicas(filtrados)
		c.HTML(http.StatusOK, "Mujer.html", data)
	})

	// ── Detalle de producto ───────────────────────────────────────────────────
	r.GET("/producto/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.HTML(http.StatusNotFound, "product.html", gin.H{"error": "Producto no encontrado"})
			return
		}
		var encontrado *models.Producto
		for i, p := range models.Productos {
			if p.ID == id {
				encontrado = &models.Productos[i]
				break
			}
		}
		if encontrado == nil {
			c.HTML(http.StatusNotFound, "product.html", gin.H{"error": "Producto no encontrado"})
			return
		}
		data := datosBase(c)
		data["producto"] = encontrado
		c.HTML(http.StatusOK, "product.html", data)
	})

	// ── Carrito (GET) ─────────────────────────────────────────────────────────
	r.GET("/carrito", func(c *gin.Context) {
		items := carritoPkg.Obtener(c)
		total := carritoPkg.Total(items)
		data := datosBase(c)
		data["items"] = items
		data["total"] = total
		data["productos"] = models.Productos
		c.HTML(http.StatusOK, "Carrito.html", data)
	})

	// ── Carrito (POST) — Agregar producto ─────────────────────────────────────
	r.POST("/carrito/agregar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		talla := c.PostForm("talla")
		cantidad, err := strconv.Atoi(c.PostForm("cantidad"))
		if err != nil || cantidad < 1 {
			cantidad = 1
		}
		carritoPkg.Agregar(c, id, talla, cantidad)
		referer := c.Request.Referer()
		if referer == "" {
			referer = "/carrito"
		}
		c.Redirect(http.StatusSeeOther, referer)
	})

	// ── Carrito (POST) — Eliminar producto ────────────────────────────────────
	r.POST("/carrito/eliminar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		talla := c.PostForm("talla")
		carritoPkg.Eliminar(c, id, talla)
		c.Redirect(http.StatusSeeOther, "/carrito")
	})

	// ── Carrito (POST) — Actualizar cantidad ──────────────────────────────────
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

	// ── Autenticación ─────────────────────────────────────────────────────────

	// POST /auth/registro — crea una nueva cuenta con rol "cliente"
	r.POST("/auth/registro", func(c *gin.Context) {
		nombre := c.PostForm("nombre")
		email := c.PostForm("email")
		password := c.PostForm("password")

		_, err := auth.Registrar(nombre, email, password)
		if err != nil {
			auth.SetFlash(c, "danger", "Error al registrarse: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "¡Cuenta creada correctamente! Ya puedes iniciar sesión.")
		}
		// POST → Redirect → GET para evitar reenvío del formulario al recargar
		referer := c.Request.Referer()
		if referer == "" {
			referer = "/"
		}
		c.Redirect(http.StatusSeeOther, referer)
	})

	// POST /auth/login — verifica credenciales y guarda la sesión
	r.POST("/auth/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		u, err := auth.IniciarSesion(c, email, password)
		if err != nil {
			auth.SetFlash(c, "danger", "Correo o contraseña incorrectos.")
		} else {
			auth.SetFlash(c, "success", "¡Bienvenido, "+u.Nombre+"!")
		}
		referer := c.Request.Referer()
		if referer == "" {
			referer = "/"
		}
		c.Redirect(http.StatusSeeOther, referer)
	})

	// GET /auth/logout — cierra la sesión y redirige al inicio
	r.GET("/auth/logout", func(c *gin.Context) {
		auth.CerrarSesion(c)
		auth.SetFlash(c, "success", "Has cerrado sesión correctamente.")
		c.Redirect(http.StatusSeeOther, "/")
	})
}
