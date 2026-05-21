// db.go — inicializa la conexión a PostgreSQL y crea las tablas si no existen.
// Las credenciales se leen de variables de entorno (archivo .env),
// así nunca quedan escritas directamente en el código.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DB es la conexión global a la base de datos.
// La usan todos los paquetes que necesiten hacer queries (auth, producto, etc.).
var DB *sql.DB

// Init abre la conexión con PostgreSQL usando las variables de entorno y
// crea todas las tablas del esquema si aún no existen.
func Init() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "tienda_user"),
		getenv("DB_PASSWORD", ""),
		getenv("DB_NAME", "tienda_db"),
		getenv("DB_SSLMODE", "disable"),
	)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db: no se pudo abrir la conexión: %v", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("db: no se pudo conectar a PostgreSQL: %v\n"+
			"  → Asegúrate de que PostgreSQL esté corriendo y que .env tenga las credenciales correctas.", err)
	}

	crearTablas()
	migrarPedidos()
	migrarSubcategorias()
	seedCategorias()
	seedProductos()

	log.Printf("db: PostgreSQL listo → %s@%s/%s",
		getenv("DB_USER", "tienda_user"),
		getenv("DB_HOST", "localhost"),
		getenv("DB_NAME", "tienda_db"),
	)
}

// getenv devuelve el valor de una variable de entorno o un valor por defecto.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// migrarPedidos agrega las columnas de envío/pago y permite pedidos de invitados
// (UsuarioID nullable). Usa IF NOT EXISTS → seguro de correr múltiples veces.
func migrarPedidos() {
	migraciones := []string{
		// Permitir invitados: UsuarioID pasa a ser opcional
		"ALTER TABLE pedidos ALTER COLUMN UsuarioID DROP NOT NULL",
		// Columnas de datos de envío y pago
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS email_contacto  VARCHAR(100)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS nombre_envio    VARCHAR(100)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS apellido_envio  VARCHAR(50)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS cedula          VARCHAR(20)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS direccion       VARCHAR(255)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS direccion2      VARCHAR(100)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS ciudad          VARCHAR(100)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS departamento    VARCHAR(100)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS codigo_postal   VARCHAR(10)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS telefono        VARCHAR(20)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS metodo_pago     VARCHAR(50)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS costo_envio     FLOAT DEFAULT 15000",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS newsletter      BOOLEAN DEFAULT FALSE",
		"ALTER TABLE pedidos ALTER COLUMN Fecha TYPE TIMESTAMP USING Fecha::timestamp",
		// Columnas del sistema de seguimiento de envío
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS numero_guia   VARCHAR(100)",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS fecha_envio   TIMESTAMP",
		"ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS fecha_entrega TIMESTAMP",
		// Ampliar CHECK para admitir los nuevos estados del flujo de envío
		"ALTER TABLE pedidos DROP CONSTRAINT IF EXISTS pedidos_estado_check",
		"ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_check CHECK (Estado IN ('pendiente','pagado','enviado','entregado','cancelado'))",
	}
	for _, stmt := range migraciones {
		if _, err := DB.Exec(stmt); err != nil {
			log.Printf("db: migración pedidos (ignorado): %v", err)
		}
	}
}

// migrarSubcategorias popula la tabla subcategorias a partir del campo de texto
// libre Subcategoria de productos, y vincula cada producto a su fila con FK.
// Es idempotente: ON CONFLICT DO NOTHING evita duplicados al reiniciar.
func migrarSubcategorias() {
	// 1) Añadir columna SubcategoriaID si aún no existe
	DB.Exec(`ALTER TABLE productos ADD COLUMN IF NOT EXISTS SubcategoriaID INTEGER REFERENCES subcategorias(ID_Subcategoria)`)

	// 2) Insertar subcategorías únicas que ya existen como texto
	DB.Exec(`
		INSERT INTO subcategorias (CategoriaID, Nombre)
		SELECT DISTINCT CategoriaID, Subcategoria
		FROM productos
		WHERE Subcategoria IS NOT NULL AND Subcategoria != ''
		ON CONFLICT (CategoriaID, Nombre) DO NOTHING
	`)

	// 3) Vincular productos a su subcategoría por nombre
	DB.Exec(`
		UPDATE productos p
		SET SubcategoriaID = s.ID_Subcategoria
		FROM subcategorias s
		WHERE p.CategoriaID = s.CategoriaID
		  AND p.Subcategoria = s.Nombre
		  AND p.SubcategoriaID IS NULL
	`)
}

// crearTablas crea todas las tablas del esquema si no existen todavía.
// El orden importa: las tablas con FK deben crearse después de sus referenciadas.
func crearTablas() {
	sentencias := []string{

		// ── Categorias ──────────────────────────────────────────────────────
		// Secciones de la tienda (Hombre, Mujer, etc.)
		`CREATE TABLE IF NOT EXISTS categorias (
			ID_Categoria  SERIAL       PRIMARY KEY,
			Nombre        VARCHAR(50)  NOT NULL UNIQUE,
			Descripcion   VARCHAR(255)
		);`,

		// ── Usuarios ────────────────────────────────────────────────────────
		// Clientes y administradores. El rol controla el acceso al panel admin.
		`CREATE TABLE IF NOT EXISTS usuarios (
			ID_Usuario  SERIAL        PRIMARY KEY,
			Nombre      VARCHAR(30)   NOT NULL,
			Email       VARCHAR(100)  NOT NULL UNIQUE,
			Password    VARCHAR(100)  NOT NULL,
			Rol         VARCHAR(10)   NOT NULL DEFAULT 'cliente'
				CHECK (Rol IN ('cliente', 'admin'))
		);`,

		// ── Productos ───────────────────────────────────────────────────────
		// Catálogo de la tienda, gestionado desde el panel admin.
		// Tallas se guarda como CSV ("S,M,L") para soportar múltiples tallas
		// con un esquema simple sin tabla extra.
		`CREATE TABLE IF NOT EXISTS productos (
			ID_Producto   SERIAL        PRIMARY KEY,
			CategoriaID   INTEGER       NOT NULL REFERENCES categorias(ID_Categoria),
			Nombre        VARCHAR(100)  NOT NULL,
			Descripcion   VARCHAR(300),
			Precio        FLOAT         NOT NULL DEFAULT 0,
			ImagenURL     VARCHAR(255),
			Subcategoria  VARCHAR(50),
			Tallas        TEXT          NOT NULL DEFAULT ''
		);`,

		// ── Carrito ─────────────────────────────────────────────────────────
		// Un carrito por usuario. Total se recalcula al modificar sus ítems.
		`CREATE TABLE IF NOT EXISTS carrito (
			ID_Carrito  SERIAL   PRIMARY KEY,
			UsuarioID   INTEGER  NOT NULL REFERENCES usuarios(ID_Usuario) ON DELETE CASCADE,
			Total       FLOAT    NOT NULL DEFAULT 0,
			UNIQUE (UsuarioID)
		);`,

		// ── Item_Carrito ─────────────────────────────────────────────────────
		// Cada fila es un producto dentro de un carrito, con su talla y cantidad.
		`CREATE TABLE IF NOT EXISTS item_carrito (
			ID_ItemCarrito  SERIAL   PRIMARY KEY,
			CarritoID       INTEGER  NOT NULL REFERENCES carrito(ID_Carrito) ON DELETE CASCADE,
			ProductoID      INTEGER  NOT NULL REFERENCES productos(ID_Producto) ON DELETE CASCADE,
			Talla           VARCHAR(10),
			Cantidad        INTEGER  NOT NULL DEFAULT 1 CHECK (Cantidad > 0),
			Subtotal        FLOAT    NOT NULL DEFAULT 0
		);`,

		// ── Pedidos ─────────────────────────────────────────────────────────
		// Historial de compras confirmadas. Estado: pendiente/pagado/cancelado.
		`CREATE TABLE IF NOT EXISTS pedidos (
			ID_Pedido   SERIAL       PRIMARY KEY,
			UsuarioID   INTEGER      NOT NULL REFERENCES usuarios(ID_Usuario),
			Fecha       DATE         NOT NULL DEFAULT CURRENT_DATE,
			Total       FLOAT        NOT NULL DEFAULT 0,
			Estado      VARCHAR(15)  NOT NULL DEFAULT 'pendiente'
				CHECK (Estado IN ('pendiente', 'pagado', 'cancelado'))
		);`,

		// ── Pedido_Items ────────────────────────────────────────────────────
		// Guarda cada producto comprado dentro de un pedido.
		// Se almacenan nombre y precio al momento de la compra para que
		// cambios futuros en el catálogo no afecten el historial.
		`CREATE TABLE IF NOT EXISTS pedido_items (
			ID_Item         SERIAL        PRIMARY KEY,
			PedidoID        INTEGER       NOT NULL REFERENCES pedidos(ID_Pedido) ON DELETE CASCADE,
			ProductoID      INTEGER       REFERENCES productos(ID_Producto) ON DELETE SET NULL,
			Nombre          VARCHAR(100)  NOT NULL,
			Talla           VARCHAR(10),
			Cantidad        INTEGER       NOT NULL DEFAULT 1,
			PrecioUnitario  FLOAT         NOT NULL,
			Subtotal        FLOAT         NOT NULL
		);`,

		// ── Producto_Imagenes ─────────────────────────────────────────────────
		// Galería de imágenes de un producto. EsPortada = imagen principal.
		`CREATE TABLE IF NOT EXISTS producto_imagenes (
			ID_Imagen   SERIAL        PRIMARY KEY,
			ProductoID  INTEGER       NOT NULL REFERENCES productos(ID_Producto) ON DELETE CASCADE,
			ImagenURL   VARCHAR(500)  NOT NULL,
			EsPortada   BOOLEAN       NOT NULL DEFAULT FALSE
		);`,

		// ── Producto_Relaciones ───────────────────────────────────────────────
		// Relaciona productos para la sección "Completa tu outfit".
		// La relación es bidireccional: A→B implica B→A en las queries.
		`CREATE TABLE IF NOT EXISTS producto_relaciones (
			ProductoID    INTEGER  NOT NULL REFERENCES productos(ID_Producto) ON DELETE CASCADE,
			RelacionadoID INTEGER  NOT NULL REFERENCES productos(ID_Producto) ON DELETE CASCADE,
			PRIMARY KEY (ProductoID, RelacionadoID)
		);`,

		// ── Subcategorias ─────────────────────────────────────────────────────
		// Entidad propia: existe aunque no tenga productos asociados.
		// UNIQUE(CategoriaID, Nombre) evita duplicados por categoría.
		`CREATE TABLE IF NOT EXISTS subcategorias (
			ID_Subcategoria SERIAL       PRIMARY KEY,
			CategoriaID     INTEGER      NOT NULL REFERENCES categorias(ID_Categoria) ON DELETE CASCADE,
			Nombre          VARCHAR(100) NOT NULL,
			UNIQUE(CategoriaID, Nombre)
		);`,
	}

	for _, stmt := range sentencias {
		if _, err := DB.Exec(stmt); err != nil {
			log.Fatalf("db: error creando tablas: %v\nSQL: %s", err, stmt)
		}
	}
}

// seedCategorias inserta las categorías base solo si la tabla está vacía.
func seedCategorias() {
	var count int
	DB.QueryRow(`SELECT COUNT(*) FROM categorias`).Scan(&count)
	if count > 0 {
		return
	}
	DB.Exec(`INSERT INTO categorias (Nombre, Descripcion) VALUES
		('Hombre', 'Ropa y accesorios para hombres'),
		('Mujer',  'Ropa y accesorios para mujeres')
	`)
	log.Println("db: categorías base insertadas (Hombre=1, Mujer=2)")
}

// seedProductos inserta 5 productos de ejemplo solo si la tabla está vacía.
// Así el admin tiene datos para ver desde el primer arranque.
func seedProductos() {
	var count int
	DB.QueryRow(`SELECT COUNT(*) FROM productos`).Scan(&count)
	if count > 0 {
		return
	}
	iniciales := []struct {
		nombre, descripcion, imagen, subcategoria, tallas string
		precio                                             float64
		categoriaID                                        int // 1=Hombre, 2=Mujer
	}{
		{"Camisa",   "Camisa de algodón para hombre",      "/static/assets/camisa.jpg",   "Camisa",   "M,L",   29.99, 1},
		{"Pantalón", "Pantalón de mezclilla para hombre",  "/static/assets/pantalon.jpg", "Pantalón", "L",     49.99, 1},
		{"Blusa",    "Blusa de algodón para mujer",        "/static/assets/blusa.jpg",    "Blusa",    "S",     39.99, 2},
		{"Falda",    "Falda de flores para mujer",         "/static/assets/falda.jpg",    "Falda",    "M",     24.99, 2},
		{"Chaqueta", "Chaqueta de cuero para hombre",      "/static/assets/chaqueta.jpg", "Chaqueta", "L",     89.99, 1},
	}
	for _, p := range iniciales {
		DB.Exec(
			`INSERT INTO productos (CategoriaID, Nombre, Descripcion, Precio, ImagenURL, Subcategoria, Tallas)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			p.categoriaID, p.nombre, p.descripcion, p.precio, p.imagen, p.subcategoria, p.tallas,
		)
	}
	log.Println("db: productos de ejemplo insertados")
}

// ContarTabla devuelve el número de filas de una tabla.
// Se usa en el dashboard del admin para mostrar estadísticas.
func ContarTabla(tabla string) int {
	var n int
	DB.QueryRow("SELECT COUNT(*) FROM " + tabla).Scan(&n)
	return n
}

// ── Subcategorías ────────────────────────────────────────────────────────────

// ListarSubcategoriasPorCategoria devuelve las subcategorías de una categoría
// con la imagen del producto más reciente como portada y el total de productos.
// Las subcategorías sin productos también se devuelven (Total=0); el template
// decide si mostrarlas o no.
func ListarSubcategoriasPorCategoria(categoriaID int) []struct {
	ID          int
	CategoriaID int
	Nombre      string
	Portada     string
	Total       int
} {
	rows, err := DB.Query(`
		SELECT s.ID_Subcategoria, s.CategoriaID, s.Nombre,
		       COALESCE((
		           SELECT ImagenURL FROM productos
		           WHERE SubcategoriaID = s.ID_Subcategoria
		             AND ImagenURL IS NOT NULL AND ImagenURL != ''
		           ORDER BY ID_Producto DESC LIMIT 1
		       ), '') AS Portada,
		       (SELECT COUNT(*) FROM productos WHERE SubcategoriaID = s.ID_Subcategoria) AS Total
		FROM subcategorias s
		WHERE s.CategoriaID = $1
		ORDER BY s.Nombre`, categoriaID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []struct {
		ID          int
		CategoriaID int
		Nombre      string
		Portada     string
		Total       int
	}
	for rows.Next() {
		var s struct {
			ID          int
			CategoriaID int
			Nombre      string
			Portada     string
			Total       int
		}
		rows.Scan(&s.ID, &s.CategoriaID, &s.Nombre, &s.Portada, &s.Total)
		result = append(result, s)
	}
	return result
}

// ObtenerNombreSubcategoria devuelve el Nombre de una subcategoría por su ID.
// Devuelve "" si no existe (subcategoriaID = 0 o no encontrado).
func ObtenerNombreSubcategoria(id int) string {
	if id <= 0 {
		return ""
	}
	var nombre string
	DB.QueryRow(`SELECT Nombre FROM subcategorias WHERE ID_Subcategoria = $1`, id).Scan(&nombre)
	return nombre
}

// CrearSubcategoria inserta una nueva subcategoría y devuelve su ID.
func CrearSubcategoria(nombre string, categoriaID int) (int, error) {
	var id int
	err := DB.QueryRow(
		`INSERT INTO subcategorias (CategoriaID, Nombre) VALUES ($1, $2)
		 ON CONFLICT (CategoriaID, Nombre) DO UPDATE SET Nombre=EXCLUDED.Nombre
		 RETURNING ID_Subcategoria`,
		categoriaID, nombre,
	).Scan(&id)
	return id, err
}

// ActualizarSubcategoria cambia el nombre de una subcategoría.
func ActualizarSubcategoria(id int, nombre string) error {
	_, err := DB.Exec(`UPDATE subcategorias SET Nombre=$1 WHERE ID_Subcategoria=$2`, nombre, id)
	return err
}

// EliminarSubcategoria borra una subcategoría. Los productos quedan con
// SubcategoriaID = NULL (no se eliminan en cascada).
func EliminarSubcategoria(id int) error {
	DB.Exec(`UPDATE productos SET SubcategoriaID=NULL WHERE SubcategoriaID=$1`, id)
	_, err := DB.Exec(`DELETE FROM subcategorias WHERE ID_Subcategoria=$1`, id)
	return err
}

// ListarCategorias devuelve todas las categorías de la tienda desde la DB.
func ListarCategorias() []struct {
	ID          int
	Nombre      string
	Descripcion string
} {
	rows, err := DB.Query(`SELECT ID_Categoria, Nombre, COALESCE(Descripcion,'') FROM categorias ORDER BY ID_Categoria`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var cats []struct {
		ID          int
		Nombre      string
		Descripcion string
	}
	for rows.Next() {
		var c struct {
			ID          int
			Nombre      string
			Descripcion string
		}
		rows.Scan(&c.ID, &c.Nombre, &c.Descripcion)
		cats = append(cats, c)
	}
	return cats
}
