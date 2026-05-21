// routes.go — configura todas las rutas HTTP de la aplicación.
package routes

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Gin/auth"
	carritoPkg "Gin/carrito"
	"Gin/db"
	"Gin/helpers"
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
// cartCount, usuario autenticado, flash messages y subcategorías para el navbar.
func datosBase(c *gin.Context) gin.H {
	flashTipo, flashMsg := auth.GetFlash(c)
	return gin.H{
		"cartCount":    contarItemsCarrito(c),
		"usuario":      auth.UsuarioActual(c),
		"flashTipo":    flashTipo,
		"flashMsg":     flashMsg,
		"navSubcatsH":  db.ListarSubcategoriasPorCategoria(1), // Hombre — para el mega-menú
		"navSubcatsM":  db.ListarSubcategoriasPorCategoria(2), // Mujer  — para el mega-menú
	}
}

func contarItemsCarrito(c *gin.Context) int {
	return carritoPkg.ContarItems(carritoPkg.Obtener(c))
}

// SetupRoutes registra todas las rutas en el router de Gin.
func SetupRoutes(r *gin.Engine) {

	r.SetFuncMap(template.FuncMap{
		// Aritmética
		"add":  helpers.Add,
		"sub":  helpers.Sub,
		"fsub": helpers.FSub,
		// Strings
		"join":     helpers.Join,
		"toLower":  helpers.ToLower,
		"toUpper":  helpers.ToUpper,
		// Búsqueda en slices
		"contiene":    helpers.Contiene,
		"contieneInt": helpers.ContieneInt,
		"strSlice":    helpers.StrSlice,
		// Formateo
		"formatFecha": helpers.FormatFecha,
		"formatCOP":   helpers.FormatCOP,
	})

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*.html")

	// ── Tienda ───────────────────────────────────────────────────────────────

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", datosBase(c))
	})

	// handlerSubcats muestra el grid de tarjetas de subcategorías de una categoría.
	// /mujer y /hombre llevan aquí; solo muestran subcats con productos.
	handlerSubcats := func(categoriaID int, nombreCat string) gin.HandlerFunc {
		return func(c *gin.Context) {
			subcats := db.ListarSubcategoriasPorCategoria(categoriaID)
			data := datosBase(c)
			data["subcategorias"] = subcats
			data["categoriaActual"] = nombreCat
			data["categoriaID"] = categoriaID
			c.HTML(http.StatusOK, "subcategorias.html", data)
		}
	}

	r.GET("/hombre", handlerSubcats(1, "Hombre"))
	r.GET("/mujer", handlerSubcats(2, "Mujer"))

	// /mujer/:id y /hombre/:id → catálogo de productos de una subcategoría
	handlerCatalogo := func(categoriaID int, nombreCat string) gin.HandlerFunc {
		return func(c *gin.Context) {
			subcatID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.Redirect(http.StatusSeeOther, "/"+strings.ToLower(nombreCat))
				return
			}
			// Verificar que la subcategoría pertenece a esta categoría
			subcats := db.ListarSubcategoriasPorCategoria(categoriaID)
			var subcatNombre string
			for _, s := range subcats {
				if s.ID == subcatID {
					subcatNombre = s.Nombre
					break
				}
			}
			if subcatNombre == "" {
				c.Redirect(http.StatusSeeOther, "/"+strings.ToLower(nombreCat))
				return
			}
			talla := c.Query("talla") // filtro opcional por talla (?talla=M)
			productos := productoPkg.ListarPorSubcategoriaID(subcatID, talla)
			tallas := productoPkg.TallasDeSubcategoria(subcatID)
			data := datosBase(c)
			data["productos"] = productos
			data["tallas"] = tallas
			data["tallaActual"] = talla
			data["subcatNombre"] = subcatNombre
			data["subcatID"] = subcatID
			data["categoriaActual"] = nombreCat
			data["categoriaID"] = categoriaID
			c.HTML(http.StatusOK, "catalogo.html", data)
		}
	}

	r.GET("/hombre/:id", handlerCatalogo(1, "Hombre"))
	r.GET("/mujer/:id", handlerCatalogo(2, "Mujer"))

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
		data["relacionados"] = productoPkg.ObtenerRelacionados(id)
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
		data["producto"] = &models.Producto{CategoriaID: 1}
		data["esNuevo"] = true
		data["categorias"] = db.ListarCategorias()
		data["todosProductos"] = productoPkg.Listar()
		data["relacionadosIDs"] = []int{}
		data["imagenesGaleria"] = []string{}
		// subcatsActuales: subcategorías de la categoría por defecto (Hombre = 1)
		data["subcatsActuales"] = db.ListarSubcategoriasPorCategoria(1)
		c.HTML(http.StatusOK, "admin_producto_form.html", data)
	})

	// Guardar producto nuevo
	admin.POST("/productos/crear", func(c *gin.Context) {
		nombre := c.PostForm("nombre")
		descripcion := c.PostForm("descripcion")
		precio, _ := strconv.ParseFloat(c.PostForm("precio"), 64)
		categoriaID, _ := strconv.Atoi(c.PostForm("categoriaID"))
		subcategoriaID, _ := strconv.Atoi(c.PostForm("subcategoriaID"))
		// Obtener el nombre de texto de la subcategoría desde la DB
		subcategoria := db.ObtenerNombreSubcategoria(subcategoriaID)
		tallas := c.PostFormArray("tallas")

		// Imágenes: primera imagen (portada) + adicionales
		imagenes := c.PostFormArray("imagenes")
		portadaURL := c.PostForm("portada")
		if portadaURL == "" {
			for _, u := range imagenes {
				if u != "" { portadaURL = u; break }
			}
		}

		productoID, err := productoPkg.Crear(nombre, descripcion, precio, portadaURL, categoriaID, subcategoria, subcategoriaID, tallas)
		if err != nil {
			auth.SetFlash(c, "danger", "Error al crear producto: "+err.Error())
		} else {
			productoPkg.GuardarImagenes(productoID, imagenes, portadaURL)
			var relIDs []int
			for _, s := range c.PostFormArray("relacionados") {
				rid, _ := strconv.Atoi(s)
				if rid > 0 { relIDs = append(relIDs, rid) }
			}
			productoPkg.GuardarRelaciones(productoID, relIDs)
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
		data["todosProductos"] = productoPkg.Listar()
		data["relacionadosIDs"] = productoPkg.ObtenerRelacionadosIDs(id)
		// subcatsActuales: subcategorías de la categoría actual del producto
		data["subcatsActuales"] = db.ListarSubcategoriasPorCategoria(p.CategoriaID)
		// Si no hay imágenes en la galería, usar la portada existente
		imgs := productoPkg.ObtenerImagenes(id)
		if len(imgs) == 0 && p.Imagen != "" {
			imgs = []string{p.Imagen}
		}
		data["imagenesGaleria"] = imgs
		c.HTML(http.StatusOK, "admin_producto_form.html", data)
	})

	// Guardar cambios de un producto existente
	admin.POST("/productos/:id/actualizar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		nombre := c.PostForm("nombre")
		descripcion := c.PostForm("descripcion")
		precio, _ := strconv.ParseFloat(c.PostForm("precio"), 64)
		categoriaID, _ := strconv.Atoi(c.PostForm("categoriaID"))
		subcategoriaID, _ := strconv.Atoi(c.PostForm("subcategoriaID"))
		subcategoria := db.ObtenerNombreSubcategoria(subcategoriaID)
		tallas := c.PostFormArray("tallas")

		imagenes := c.PostFormArray("imagenes")
		portadaURL := c.PostForm("portada")
		if portadaURL == "" {
			for _, u := range imagenes {
				if u != "" { portadaURL = u; break }
			}
		}

		if err := productoPkg.Actualizar(id, nombre, descripcion, precio, portadaURL, categoriaID, subcategoria, subcategoriaID, tallas); err != nil {
			auth.SetFlash(c, "danger", "Error al actualizar: "+err.Error())
		} else {
			productoPkg.GuardarImagenes(id, imagenes, portadaURL)
			var relIDs []int
			for _, s := range c.PostFormArray("relacionados") {
				rid, _ := strconv.Atoi(s)
				if rid > 0 { relIDs = append(relIDs, rid) }
			}
			productoPkg.GuardarRelaciones(id, relIDs)
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
			NumeroGuia    string
			FechaEnvio    time.Time
			FechaEntrega  time.Time
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
			       COALESCE(p.costo_envio,0), p.Estado,
			       COALESCE(p.numero_guia,''),
			       COALESCE(p.fecha_envio,'0001-01-01'::timestamp),
			       COALESCE(p.fecha_entrega,'0001-01-01'::timestamp)
			FROM pedidos p
			LEFT JOIN usuarios u ON p.UsuarioID = u.ID_Usuario
			WHERE p.ID_Pedido = $1`, pedidoID).
			Scan(&p.ID_Pedido, &p.NombreCliente, &p.EmailContacto,
				&p.NombreEnvio, &p.ApellidoEnvio, &p.Cedula,
				&p.Telefono, &p.Direccion, &p.Direccion2,
				&p.Ciudad, &p.Departamento, &p.CodigoPostal,
				&p.MetodoPago, &p.Fecha, &p.Total,
				&p.CostoEnvio, &p.Estado,
				&p.NumeroGuia, &p.FechaEnvio, &p.FechaEntrega)
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

	// ── API pública: subcategorías por categoría (para el formulario de producto) ──
	r.GET("/api/subcategorias/:categoriaID", func(c *gin.Context) {
		catID, _ := strconv.Atoi(c.Param("categoriaID"))
		subcats := db.ListarSubcategoriasPorCategoria(catID)
		type Item struct {
			ID     int    `json:"id"`
			Nombre string `json:"nombre"`
		}
		var items []Item
		for _, s := range subcats {
			items = append(items, Item{ID: s.ID, Nombre: s.Nombre})
		}
		c.JSON(http.StatusOK, gin.H{"subcategorias": items})
	})

	// ── Admin: gestión de subcategorías ──────────────────────────────────────

	admin.GET("/subcategorias", func(c *gin.Context) {
		subcatsH := db.ListarSubcategoriasPorCategoria(1)
		subcatsM := db.ListarSubcategoriasPorCategoria(2)
		data := datosBase(c)
		data["paginaActual"] = "subcategorias"
		data["subcatsHombre"] = subcatsH
		data["subcatsMujer"] = subcatsM
		data["categorias"] = db.ListarCategorias()
		c.HTML(http.StatusOK, "admin_subcategorias.html", data)
	})

	admin.POST("/subcategorias/crear", func(c *gin.Context) {
		nombre := strings.TrimSpace(c.PostForm("nombre"))
		categoriaID, _ := strconv.Atoi(c.PostForm("categoriaID"))
		if nombre == "" {
			auth.SetFlash(c, "danger", "El nombre no puede estar vacío.")
		} else {
			_, err := db.CrearSubcategoria(nombre, categoriaID)
			if err != nil {
				auth.SetFlash(c, "danger", "Error al crear subcategoría: "+err.Error())
			} else {
				auth.SetFlash(c, "success", "Subcategoría '"+nombre+"' creada.")
			}
		}
		c.Redirect(http.StatusSeeOther, "/admin/subcategorias")
	})

	admin.POST("/subcategorias/:id/actualizar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		nombre := strings.TrimSpace(c.PostForm("nombre"))
		if nombre == "" {
			auth.SetFlash(c, "danger", "El nombre no puede estar vacío.")
		} else if err := db.ActualizarSubcategoria(id, nombre); err != nil {
			auth.SetFlash(c, "danger", "Error al actualizar: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "Subcategoría actualizada.")
		}
		c.Redirect(http.StatusSeeOther, "/admin/subcategorias")
	})

	admin.POST("/subcategorias/:id/eliminar", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := db.EliminarSubcategoria(id); err != nil {
			auth.SetFlash(c, "danger", "Error al eliminar: "+err.Error())
		} else {
			auth.SetFlash(c, "success", "Subcategoría eliminada. Los productos quedaron sin subcategoría asignada.")
		}
		c.Redirect(http.StatusSeeOther, "/admin/subcategorias")
	})

	// ── API pública: municipios por departamento ─────────────────────────────
	// Devuelve JSON con la lista de municipios del departamento solicitado.
	// Lo consume el select dinámico del checkout vía fetch().
	r.GET("/api/municipios/:departamento", func(c *gin.Context) {
		depto := c.Param("departamento")
		munis, ok := municipiosColombia[depto]
		if !ok {
			c.JSON(http.StatusOK, gin.H{"municipios": []string{}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"municipios": munis})
	})

	// ── Mis pedidos (cliente autenticado) ────────────────────────────────────

	r.GET("/mis-pedidos", func(c *gin.Context) {
		u := auth.UsuarioActual(c)
		if u == nil {
			auth.SetFlash(c, "warning", "Debes iniciar sesión para ver tus pedidos.")
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		type FilaMisPedidos struct {
			ID_Pedido int
			Fecha     time.Time
			Total     float64
			Estado    string
			Ciudad    string
		}
		rows, err := db.DB.Query(`
			SELECT ID_Pedido, Fecha, Total, Estado, COALESCE(ciudad,'')
			FROM pedidos WHERE UsuarioID = $1 ORDER BY ID_Pedido DESC`, u.ID_Usuario)
		var pedidos []FilaMisPedidos
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var f FilaMisPedidos
				rows.Scan(&f.ID_Pedido, &f.Fecha, &f.Total, &f.Estado, &f.Ciudad)
				pedidos = append(pedidos, f)
			}
		}
		data := datosBase(c)
		data["pedidos"] = pedidos
		c.HTML(http.StatusOK, "mis_pedidos.html", data)
	})

	r.GET("/mis-pedidos/:id", func(c *gin.Context) {
		u := auth.UsuarioActual(c)
		if u == nil {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		pedidoID, _ := strconv.Atoi(c.Param("id"))

		var p models.Pedido
		err := db.DB.QueryRow(`
			SELECT ID_Pedido, COALESCE(nombre_envio,''), COALESCE(apellido_envio,''),
			       COALESCE(email_contacto,''), Total, Estado, COALESCE(metodo_pago,''),
			       COALESCE(costo_envio,0), COALESCE(direccion,''), COALESCE(ciudad,''),
			       COALESCE(departamento,''), Fecha,
			       COALESCE(numero_guia,''),
			       COALESCE(fecha_envio, '0001-01-01'::timestamp),
			       COALESCE(fecha_entrega,'0001-01-01'::timestamp)
			FROM pedidos WHERE ID_Pedido = $1 AND UsuarioID = $2`, pedidoID, u.ID_Usuario).
			Scan(&p.ID_Pedido, &p.NombreEnvio, &p.ApellidoEnvio,
				&p.EmailContacto, &p.Total, &p.Estado, &p.MetodoPago,
				&p.CostoEnvio, &p.Direccion, &p.Ciudad, &p.Departamento, &p.Fecha,
				&p.NumeroGuia, &p.FechaEnvio, &p.FechaEntrega)
		if err != nil {
			auth.SetFlash(c, "danger", "Pedido no encontrado.")
			c.Redirect(http.StatusSeeOther, "/mis-pedidos")
			return
		}

		type ItemMiPedido struct {
			Nombre         string
			Talla          string
			Cantidad       int
			PrecioUnitario float64
			Subtotal       float64
			Imagen         string
		}
		rows, _ := db.DB.Query(`
			SELECT pi.Nombre, COALESCE(pi.Talla,''), pi.Cantidad,
			       pi.PrecioUnitario, pi.Subtotal, COALESCE(pr.ImagenURL,'')
			FROM pedido_items pi
			LEFT JOIN productos pr ON pi.ProductoID = pr.ID_Producto
			WHERE pi.PedidoID = $1 ORDER BY pi.ID_Item`, pedidoID)
		var items []ItemMiPedido
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var it ItemMiPedido
				rows.Scan(&it.Nombre, &it.Talla, &it.Cantidad,
					&it.PrecioUnitario, &it.Subtotal, &it.Imagen)
				items = append(items, it)
			}
		}

		data := datosBase(c)
		data["pedido"] = p
		data["items"] = items
		c.HTML(http.StatusOK, "mi_pedido_detalle.html", data)
	})

	// El cliente confirma que recibió el pedido
	r.POST("/pedidos/:id/confirmar-entrega", func(c *gin.Context) {
		u := auth.UsuarioActual(c)
		if u == nil {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		pedidoID, _ := strconv.Atoi(c.Param("id"))
		// Solo permite confirmar si el pedido pertenece al usuario y está "enviado"
		_, err := db.DB.Exec(`
			UPDATE pedidos SET Estado = 'entregado', fecha_entrega = $1
			WHERE ID_Pedido = $2 AND UsuarioID = $3 AND Estado = 'enviado'`,
			time.Now(), pedidoID, u.ID_Usuario)
		if err != nil {
			auth.SetFlash(c, "danger", "No se pudo confirmar la entrega.")
		} else {
			auth.SetFlash(c, "success", "¡Gracias! Has confirmado la recepción del pedido.")
		}
		c.Redirect(http.StatusSeeOther, "/mis-pedidos/"+strconv.Itoa(pedidoID))
	})

	// ── Cambio de estado del pedido (admin) ──────────────────────────────────

	admin.POST("/pedidos/:id/marcar-pagado", func(c *gin.Context) {
		pedidoID, _ := strconv.Atoi(c.Param("id"))
		db.DB.Exec(`UPDATE pedidos SET Estado = 'pagado' WHERE ID_Pedido = $1 AND Estado = 'pendiente'`, pedidoID)
		auth.SetFlash(c, "success", "Pedido #"+strconv.Itoa(pedidoID)+" marcado como pagado.")
		c.Redirect(http.StatusSeeOther, "/admin/pedidos/"+strconv.Itoa(pedidoID))
	})

	admin.POST("/pedidos/:id/marcar-enviado", func(c *gin.Context) {
		pedidoID, _ := strconv.Atoi(c.Param("id"))
		guia := strings.TrimSpace(c.PostForm("numero_guia"))
		db.DB.Exec(`
			UPDATE pedidos SET Estado = 'enviado', numero_guia = $1, fecha_envio = $2
			WHERE ID_Pedido = $3 AND Estado = 'pagado'`,
			guia, time.Now(), pedidoID)
		auth.SetFlash(c, "success", "Pedido #"+strconv.Itoa(pedidoID)+" marcado como enviado.")
		c.Redirect(http.StatusSeeOther, "/admin/pedidos/"+strconv.Itoa(pedidoID))
	})

	admin.POST("/pedidos/:id/marcar-cancelado", func(c *gin.Context) {
		pedidoID, _ := strconv.Atoi(c.Param("id"))
		db.DB.Exec(`UPDATE pedidos SET Estado = 'cancelado' WHERE ID_Pedido = $1 AND Estado NOT IN ('entregado','cancelado')`, pedidoID)
		auth.SetFlash(c, "warning", "Pedido #"+strconv.Itoa(pedidoID)+" cancelado.")
		c.Redirect(http.StatusSeeOther, "/admin/pedidos/"+strconv.Itoa(pedidoID))
	})

	// ── Subida de imágenes ───────────────────────────────────────────────────
	// Recibe un archivo multipart, lo guarda en static/uploads/ y devuelve
	// la URL pública como JSON. Lo usa el formulario de producto vía fetch().
	admin.POST("/uploads/imagen", func(c *gin.Context) {
		file, err := c.FormFile("archivo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No se recibió ningún archivo."})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		permitidos := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
		if !permitidos[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato no permitido. Usa JPG, PNG, WEBP o GIF."})
			return
		}

		// Nombre único basado en timestamp para evitar colisiones
		nombreArchivo := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
		destino := filepath.Join("static", "uploads", nombreArchivo)

		os.MkdirAll(filepath.Join("static", "uploads"), 0755)

		if err := c.SaveUploadedFile(file, destino); err != nil {
			log.Printf("uploads/imagen ERROR: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el archivo."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": "/static/uploads/" + nombreArchivo})
	})
}
