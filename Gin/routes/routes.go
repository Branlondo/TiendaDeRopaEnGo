// routes.go — configura todas las rutas HTTP de la aplicación.
package routes

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"Gin/auth"
	carritoPkg "Gin/carrito"
	"Gin/db"
	"Gin/models"
	productoPkg "Gin/producto"

	"github.com/gin-gonic/gin"
)

// costoEnvio es el costo fijo de envío a nivel nacional (Colombia).
const costoEnvio = 15000.0

// departamentosColombia lista todos los departamentos de Colombia.
var departamentosColombia = []string{
	"Amazonas", "Antioquia", "Arauca", "Atlántico", "Bogotá D.C.",
	"Bolívar", "Boyacá", "Caldas", "Caquetá", "Casanare", "Cauca",
	"Cesar", "Chocó", "Córdoba", "Cundinamarca", "Guainía", "Guaviare",
	"Huila", "La Guajira", "Magdalena", "Meta", "Nariño",
	"Norte de Santander", "Putumayo", "Quindío", "Risaralda",
	"San Andrés y Providencia", "Santander", "Sucre", "Tolima",
	"Valle del Cauca", "Vaupés", "Vichada",
}

// Los métodos de pago se manejan directamente en checkout_pago.html (UI estática).

// datosBase construye el mapa que TODOS los templates necesitan:
// cartCount, usuario autenticado y flash messages.
func datosBase(c *gin.Context) gin.H {
	flashTipo, flashMsg := auth.GetFlash(c)
	return gin.H{
		"cartCount": contarItemsCarrito(c),
		"usuario":   auth.UsuarioActual(c),
		"flashTipo": flashTipo,
		"flashMsg":  flashMsg,
	}
}

func contarItemsCarrito(c *gin.Context) int {
	return carritoPkg.ContarItems(carritoPkg.Obtener(c))
}

// SetupRoutes registra todas las rutas en el router de Gin.
func SetupRoutes(r *gin.Engine) {

	r.SetFuncMap(template.FuncMap{
		// Aritmética para los botones +/- del carrito (enteros)
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int {
			if a-b < 1 {
				return 1
			}
			return a - b
		},
		// fsub resta dos float64 (usado en totales de pedido)
		"fsub": func(a, b float64) float64 { return a - b },
		// join une un slice de strings con un separador (ej: join .Tallas ", ")
		"join": func(slice []string, sep string) string {
			return strings.Join(slice, sep)
		},
		// contiene comprueba si un item está en un slice (para marcar checkboxes)
		"contiene": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
		// strSlice crea un slice de strings desde argumentos variádicos.
		// Se usa en el formulario de admin: {{ range $t := strSlice "XS" "S" "M" ... }}
		"strSlice": func(items ...string) []string { return items },
		// toLower / toUpper — para construir URLs y etiquetas a partir del nombre de categoría
		"toLower": strings.ToLower,
		"toUpper": func(s string) string {
			if s == "" {
				return s
			}
			r := []rune(s)
			r[0] = unicode.ToUpper(r[0])
			return strings.ToUpper(string(r))
		},
		"formatFecha": func(t time.Time) string {
			return t.Format("02/01/2006 15:04")
		},
		"formatCOP": func(v float64) string {
			n := int64(v)
			s := strconv.FormatInt(n, 10)
			result := ""
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					result += "."
				}
				result += string(c)
			}
			return "$" + result
		},
	})

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*.html")

	// ── Tienda ───────────────────────────────────────────────────────────────

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", datosBase(c))
	})

	// handlerCategoria es la función reutilizable para cualquier categoría de la tienda.
	// Recibe el ID de la categoría en la DB y el subtítulo del header.
	// Para agregar una nueva categoría en el futuro solo hace falta añadir una línea abajo
	// y crear el registro correspondiente en la tabla `categorias`.
	handlerCategoria := func(categoriaID int, subtitulo string) gin.HandlerFunc {
		return func(c *gin.Context) {
			// Buscar el nombre de la categoría desde la DB para no hardcodearlo aquí
			cats := db.ListarCategorias()
			var nombreActual string
			for _, cat := range cats {
				if cat.ID == categoriaID {
					nombreActual = cat.Nombre
					break
				}
			}
			data := datosBase(c)
			data["productos"] = productoPkg.ListarPorCategoria(categoriaID)
			data["subcategorias"] = productoPkg.SubcategoriasUnicas(categoriaID)
			data["categorias"] = cats
			data["categoriaActual"] = nombreActual
			data["subtitulo"] = subtitulo
			c.HTML(http.StatusOK, "categoria.html", data)
		}
	}

	r.GET("/hombre", handlerCategoria(1, "Ropa exclusivamente para hombres"))
	r.GET("/mujer", handlerCategoria(2, "Ropa exclusivamente para mujeres"))

	r.GET("/producto/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.HTML(http.StatusNotFound, "product.html", gin.H{"error": "Producto no encontrado"})
			return
		}
		p := productoPkg.BuscarPorID(id)
		if p == nil {
			c.HTML(http.StatusNotFound, "product.html", gin.H{"error": "Producto no encontrado"})
			return
		}
		data := datosBase(c)
		data["producto"] = p
		c.HTML(http.StatusOK, "product.html", data)
	})

	// ── Carrito ──────────────────────────────────────────────────────────────

	r.GET("/carrito", func(c *gin.Context) {
		items := carritoPkg.Obtener(c)
		data := datosBase(c)
		data["items"] = items
		data["total"] = carritoPkg.Total(items)
		data["productos"] = productoPkg.Listar()
		c.HTML(http.StatusOK, "Carrito.html", data)
	})

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

	r.POST("/carrito/eliminar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		carritoPkg.Eliminar(c, id, c.PostForm("talla"))
		c.Redirect(http.StatusSeeOther, "/carrito")
	})

	r.POST("/carrito/actualizar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("productoID"))
		cantidad, err := strconv.Atoi(c.PostForm("cantidad"))
		if err != nil {
			cantidad = 1
		}
		carritoPkg.CambiarCantidad(c, id, c.PostForm("talla"), cantidad)
		c.Redirect(http.StatusSeeOther, "/carrito")
	})

	// ── Autenticación ────────────────────────────────────────────────────────

	r.POST("/auth/registro", func(c *gin.Context) {
		_, err := auth.Registrar(c.PostForm("nombre"), c.PostForm("email"), c.PostForm("password"))
		if err != nil {
			auth.SetFlash(c, "danger", "Error al registrarse: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "¡Cuenta creada! Ya puedes iniciar sesión.")
		}
		referer := c.Request.Referer()
		if referer == "" {
			referer = "/"
		}
		c.Redirect(http.StatusSeeOther, referer)
	})

	r.POST("/auth/login", func(c *gin.Context) {
		u, err := auth.IniciarSesion(c, c.PostForm("email"), c.PostForm("password"))
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

	r.GET("/auth/logout", func(c *gin.Context) {
		auth.CerrarSesion(c)
		auth.SetFlash(c, "success", "Has cerrado sesión correctamente.")
		c.Redirect(http.StatusSeeOther, "/")
	})

	// ── Panel Admin ──────────────────────────────────────────────────────────
	// Todas las rutas bajo /admin requieren rol "admin" (middleware RequiereAdmin).

	admin := r.Group("/admin")
	admin.Use(auth.RequiereAdmin())

	// Dashboard — estadísticas generales
	admin.GET("", func(c *gin.Context) {
		data := datosBase(c)
		data["paginaActual"] = "dashboard"
		data["totalProductos"] = db.ContarTabla("productos")
		data["totalUsuarios"] = db.ContarTabla("usuarios")
		data["totalPedidos"] = db.ContarTabla("pedidos")
		data["productos"] = productoPkg.Listar() // tabla reciente
		c.HTML(http.StatusOK, "admin_dashboard.html", data)
	})

	// Lista de productos
	admin.GET("/productos", func(c *gin.Context) {
		data := datosBase(c)
		data["paginaActual"] = "productos"
		data["productos"] = productoPkg.Listar()
		c.HTML(http.StatusOK, "admin_productos.html", data)
	})

	// Formulario para crear un producto nuevo
	admin.GET("/productos/nuevo", func(c *gin.Context) {
		data := datosBase(c)
		data["paginaActual"] = "productos"
		data["producto"] = &models.Producto{CategoriaID: 1} // valores por defecto
		data["esNuevo"] = true
		data["categorias"] = db.ListarCategorias()
		c.HTML(http.StatusOK, "admin_producto_form.html", data)
	})

	// Guardar producto nuevo
	admin.POST("/productos/crear", func(c *gin.Context) {
		nombre := c.PostForm("nombre")
		descripcion := c.PostForm("descripcion")
		precio, _ := strconv.ParseFloat(c.PostForm("precio"), 64)
		imagen := c.PostForm("imagen")
		categoriaID, _ := strconv.Atoi(c.PostForm("categoriaID"))
		subcategoria := c.PostForm("subcategoria")
		tallas := c.PostFormArray("tallas") // checkboxes: ["S","M","L"]

		if err := productoPkg.Crear(nombre, descripcion, precio, imagen, categoriaID, subcategoria, tallas); err != nil {
			auth.SetFlash(c, "danger", "Error al crear producto: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "¡Producto '"+nombre+"' creado correctamente!")
		}
		c.Redirect(http.StatusSeeOther, "/admin/productos")
	})

	// Formulario para editar un producto existente (pre-cargado con sus datos)
	admin.GET("/productos/:id/editar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		p := productoPkg.BuscarPorID(id)
		if p == nil {
			auth.SetFlash(c, "danger", "Producto no encontrado.")
			c.Redirect(http.StatusSeeOther, "/admin/productos")
			return
		}
		data := datosBase(c)
		data["paginaActual"] = "productos"
		data["producto"] = p
		data["esNuevo"] = false
		data["categorias"] = db.ListarCategorias()
		c.HTML(http.StatusOK, "admin_producto_form.html", data)
	})

	// Guardar cambios de un producto existente
	admin.POST("/productos/:id/actualizar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		nombre := c.PostForm("nombre")
		descripcion := c.PostForm("descripcion")
		precio, _ := strconv.ParseFloat(c.PostForm("precio"), 64)
		imagen := c.PostForm("imagen")
		categoriaID, _ := strconv.Atoi(c.PostForm("categoriaID"))
		subcategoria := c.PostForm("subcategoria")
		tallas := c.PostFormArray("tallas")

		if err := productoPkg.Actualizar(id, nombre, descripcion, precio, imagen, categoriaID, subcategoria, tallas); err != nil {
			auth.SetFlash(c, "danger", "Error al actualizar: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "Producto actualizado correctamente.")
		}
		c.Redirect(http.StatusSeeOther, "/admin/productos")
	})

	// Eliminar un producto (POST desde modal de confirmación)
	admin.POST("/productos/:id/eliminar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := productoPkg.Eliminar(id); err != nil {
			auth.SetFlash(c, "danger", "Error al eliminar: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "Producto eliminado.")
		}
		c.Redirect(http.StatusSeeOther, "/admin/productos")
	})

	// ── Checkout ─────────────────────────────────────────────────────────────

	// Mostrar formulario de checkout
	r.GET("/checkout", func(c *gin.Context) {
		// Login es OPCIONAL — invitados también pueden comprar
		items := carritoPkg.Obtener(c)
		if len(items) == 0 {
			c.Redirect(http.StatusSeeOther, "/carrito")
			return
		}
		subtotal := carritoPkg.Total(items)
		data := datosBase(c)
		data["items"] = items
		data["subtotal"] = subtotal
		data["costoEnvio"] = costoEnvio
		data["total"] = subtotal + costoEnvio
		data["departamentos"] = departamentosColombia
		c.HTML(http.StatusOK, "checkout.html", data)
	})

	// PASO 2 — Recibe datos de envío, muestra página de selección de pago
	r.POST("/checkout/pago", func(c *gin.Context) {
		items := carritoPkg.Obtener(c)
		if len(items) == 0 {
			c.Redirect(http.StatusSeeOther, "/carrito")
			return
		}
		subtotal := carritoPkg.Total(items)
		data := datosBase(c)
		data["items"] = items
		data["subtotal"] = subtotal
		data["costoEnvio"] = costoEnvio
		data["total"] = subtotal + costoEnvio
		// Pasar todos los campos del formulario para reenviarlos como hidden inputs
		data["form"] = map[string]string{
			"email":        c.PostForm("email"),
			"newsletter":   c.PostForm("newsletter"),
			"nombre":       c.PostForm("nombre"),
			"apellido":     c.PostForm("apellido"),
			"cedula":       c.PostForm("cedula"),
			"direccion":    c.PostForm("direccion"),
			"direccion2":   c.PostForm("direccion2"),
			"ciudad":       c.PostForm("ciudad"),
			"departamento": c.PostForm("departamento"),
			"codigo_postal": c.PostForm("codigo_postal"),
			"telefono":     c.PostForm("telefono"),
		}
		c.HTML(http.StatusOK, "checkout_pago.html", data)
	})

	// PASO 3 — Crea el pedido (mock) y redirige a confirmación
	r.POST("/checkout/procesar", func(c *gin.Context) {
		items := carritoPkg.Obtener(c)
		if len(items) == 0 {
			c.Redirect(http.StatusSeeOther, "/carrito")
			return
		}
		subtotal := carritoPkg.Total(items)
		total := subtotal + costoEnvio

		// UsuarioID: verificar que el usuario de la sesión exista en PostgreSQL.
		// Si la sesión es de una DB anterior (SQLite) o el usuario no existe, se trata como invitado.
		u := auth.UsuarioActual(c)
		var usuarioID interface{} = nil
		if u != nil {
			var existe bool
			db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM usuarios WHERE ID_Usuario = $1)`, u.ID_Usuario).Scan(&existe)
			if existe {
				usuarioID = u.ID_Usuario
			}
		}

		newsletter := c.PostForm("newsletter") == "on"
		var pedidoID int
		err := db.DB.QueryRow(`
			INSERT INTO pedidos
				(UsuarioID, Fecha, Total, Estado,
				 email_contacto, nombre_envio, apellido_envio, cedula,
				 direccion, direccion2, ciudad, departamento,
				 codigo_postal, telefono, metodo_pago, costo_envio, newsletter)
			VALUES
				($1, $2, $3, 'pendiente',
				 $4, $5, $6, $7,
				 $8, $9, $10, $11,
				 $12, $13, $14, $15, $16)
			RETURNING ID_Pedido`,
			usuarioID,
			time.Now(),
			total,
			c.PostForm("email"),
			c.PostForm("nombre"),
			c.PostForm("apellido"),
			c.PostForm("cedula"),
			c.PostForm("direccion"),
			c.PostForm("direccion2"),
			c.PostForm("ciudad"),
			c.PostForm("departamento"),
			c.PostForm("codigo_postal"),
			c.PostForm("telefono"),
			c.PostForm("metodo_pago"),
			costoEnvio,
			newsletter,
		).Scan(&pedidoID)

		if err != nil {
			log.Printf("checkout/procesar ERROR: %v", err)
			auth.SetFlash(c, "danger", "Error al procesar el pedido: "+err.Error())
			c.Redirect(http.StatusSeeOther, "/checkout")
			return
		}

		// Guardar cada ítem del carrito en pedido_items
		for _, item := range items {
			db.DB.Exec(`
				INSERT INTO pedido_items (PedidoID, ProductoID, Nombre, Talla, Cantidad, PrecioUnitario, Subtotal)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				pedidoID, item.ProductoID, item.Nombre, item.Talla,
				item.Cantidad, item.Precio, item.Total(),
			)
		}

		carritoPkg.Vaciar(c)
		c.Redirect(http.StatusSeeOther, "/checkout/confirmacion/"+strconv.Itoa(pedidoID))
	})

	// Página de confirmación
	r.GET("/checkout/confirmacion/:id", func(c *gin.Context) {
		pedidoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		// Datos completos del pedido para el recibo
		var p models.Pedido
		db.DB.QueryRow(`
			SELECT ID_Pedido,
			       COALESCE(nombre_envio,''), COALESCE(apellido_envio,''),
			       COALESCE(email_contacto,''), Total, Estado,
			       COALESCE(metodo_pago,''), COALESCE(costo_envio,0),
			       COALESCE(direccion,''), COALESCE(ciudad,''),
			       COALESCE(departamento,''), Fecha
			FROM pedidos WHERE ID_Pedido = $1`, pedidoID).
			Scan(&p.ID_Pedido, &p.NombreEnvio, &p.ApellidoEnvio,
				&p.EmailContacto, &p.Total, &p.Estado,
				&p.MetodoPago, &p.CostoEnvio,
				&p.Direccion, &p.Ciudad, &p.Departamento, &p.Fecha)

		// Ítems del pedido para el recibo
		type ItemRecibo struct {
			Nombre         string
			Talla          string
			Cantidad       int
			PrecioUnitario float64
			Subtotal       float64
		}
		rows, _ := db.DB.Query(`
			SELECT Nombre, COALESCE(Talla,''), Cantidad, PrecioUnitario, Subtotal
			FROM pedido_items WHERE PedidoID = $1 ORDER BY ID_Item`, pedidoID)
		var items []ItemRecibo
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var it ItemRecibo
				rows.Scan(&it.Nombre, &it.Talla, &it.Cantidad, &it.PrecioUnitario, &it.Subtotal)
				items = append(items, it)
			}
		}

		data := datosBase(c)
		data["pedido"] = p
		data["items"] = items
		data["subtotalProductos"] = p.Total - p.CostoEnvio
		c.HTML(http.StatusOK, "checkout_confirmacion.html", data)
	})

	// Detalle completo de un pedido (productos + datos de envío)
	admin.GET("/pedidos/:id", func(c *gin.Context) {
		pedidoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/admin/pedidos")
			return
		}

		// Datos generales del pedido
		type DetallePedido struct {
			ID_Pedido     int
			NombreCliente string
			EmailContacto string
			NombreEnvio   string
			ApellidoEnvio string
			Cedula        string
			Telefono      string
			Direccion     string
			Direccion2    string
			Ciudad        string
			Departamento  string
			CodigoPostal  string
			MetodoPago    string
			Fecha         time.Time
			Total         float64
			CostoEnvio    float64
			Estado        string
		}
		var p DetallePedido
		err = db.DB.QueryRow(`
			SELECT p.ID_Pedido, COALESCE(u.Nombre,'Invitado'),
			       COALESCE(p.email_contacto,''), COALESCE(p.nombre_envio,''),
			       COALESCE(p.apellido_envio,''), COALESCE(p.cedula,''),
			       COALESCE(p.telefono,''), COALESCE(p.direccion,''),
			       COALESCE(p.direccion2,''), COALESCE(p.ciudad,''),
			       COALESCE(p.departamento,''), COALESCE(p.codigo_postal,''),
			       COALESCE(p.metodo_pago,''), p.Fecha, p.Total,
			       COALESCE(p.costo_envio,0), p.Estado
			FROM pedidos p
			LEFT JOIN usuarios u ON p.UsuarioID = u.ID_Usuario
			WHERE p.ID_Pedido = $1`, pedidoID).
			Scan(&p.ID_Pedido, &p.NombreCliente, &p.EmailContacto,
				&p.NombreEnvio, &p.ApellidoEnvio, &p.Cedula,
				&p.Telefono, &p.Direccion, &p.Direccion2,
				&p.Ciudad, &p.Departamento, &p.CodigoPostal,
				&p.MetodoPago, &p.Fecha, &p.Total,
				&p.CostoEnvio, &p.Estado)
		if err != nil {
			auth.SetFlash(c, "danger", "Pedido no encontrado.")
			c.Redirect(http.StatusSeeOther, "/admin/pedidos")
			return
		}

		// Ítems del pedido
		type ItemPedido struct {
			Nombre         string
			Talla          string
			Cantidad       int
			PrecioUnitario float64
			Subtotal       float64
			Imagen         string
		}
		rows, _ := db.DB.Query(`
			SELECT pi.Nombre, COALESCE(pi.Talla,''), pi.Cantidad,
			       pi.PrecioUnitario, pi.Subtotal,
			       COALESCE(pr.ImagenURL,'')
			FROM pedido_items pi
			LEFT JOIN productos pr ON pi.ProductoID = pr.ID_Producto
			WHERE pi.PedidoID = $1
			ORDER BY pi.ID_Item`, pedidoID)
		var itemsPedido []ItemPedido
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var it ItemPedido
				rows.Scan(&it.Nombre, &it.Talla, &it.Cantidad,
					&it.PrecioUnitario, &it.Subtotal, &it.Imagen)
				itemsPedido = append(itemsPedido, it)
			}
		}

		data := datosBase(c)
		data["paginaActual"] = "pedidos"
		data["pedido"] = p
		data["items"] = itemsPedido
		c.HTML(http.StatusOK, "admin_pedido_detalle.html", data)
	})

	// Eliminar un pedido (y sus ítems en cascada)
	admin.POST("/pedidos/:id/eliminar", func(c *gin.Context) {
		pedidoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			auth.SetFlash(c, "danger", "ID de pedido inválido.")
			c.Redirect(http.StatusSeeOther, "/admin/pedidos")
			return
		}
		// pedido_items se elimina en cascada (ON DELETE CASCADE en la FK)
		_, err = db.DB.Exec(`DELETE FROM pedidos WHERE ID_Pedido = $1`, pedidoID)
		if err != nil {
			log.Printf("eliminar pedido #%d ERROR: %v", pedidoID, err)
			auth.SetFlash(c, "danger", "No se pudo eliminar el pedido: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "Pedido #"+strconv.Itoa(pedidoID)+" eliminado correctamente.")
		}
		c.Redirect(http.StatusSeeOther, "/admin/pedidos")
	})

	// Lista de pedidos
	admin.GET("/pedidos", func(c *gin.Context) {
		type FilaPedido struct {
			ID_Pedido     int
			NombreCliente string
			Fecha         time.Time
			Total         float64
			Estado        string
			EmailContacto string
			NombreEnvio   string
			ApellidoEnvio string
			Ciudad        string
			Departamento  string
			MetodoPago    string
		}
		rows, err := db.DB.Query(`
			SELECT p.ID_Pedido, COALESCE(u.Nombre, COALESCE(p.nombre_envio,'Invitado')),
			       p.Fecha, p.Total, p.Estado,
			       COALESCE(p.email_contacto,''), COALESCE(p.nombre_envio,''),
			       COALESCE(p.apellido_envio,''), COALESCE(p.ciudad,''),
			       COALESCE(p.departamento,''), COALESCE(p.metodo_pago,'')
			FROM pedidos p
			LEFT JOIN usuarios u ON p.UsuarioID = u.ID_Usuario
			ORDER BY p.ID_Pedido DESC
		`)
		var pedidos []FilaPedido
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var f FilaPedido
				rows.Scan(&f.ID_Pedido, &f.NombreCliente, &f.Fecha, &f.Total, &f.Estado,
					&f.EmailContacto, &f.NombreEnvio, &f.ApellidoEnvio,
					&f.Ciudad, &f.Departamento, &f.MetodoPago)
				pedidos = append(pedidos, f)
			}
		}
		data := datosBase(c)
		data["paginaActual"] = "pedidos"
		data["pedidos"] = pedidos
		c.HTML(http.StatusOK, "admin_pedidos.html", data)
	})
}
