// db.go — inicializa la conexión a SQLite y crea las tablas si no existen.
// Solo se llama una vez desde main.go con db.Init("./tienda.db").
package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // driver SQLite puro Go, sin CGO ni GCC en Windows
)

// DB es la conexión global compartida por todos los paquetes de la app.
// Se inicializa en Init() y nunca se cierra (vive mientras el servidor corre).
var DB *sql.DB

// Init abre el archivo SQLite en la ruta indicada y crea las tablas
// si todavía no existen. Llama log.Fatalf si no puede conectar.
func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("db: no se pudo abrir '%s': %v", path, err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("db: no se pudo conectar a '%s': %v", path, err)
	}
	crearTablas()
	log.Printf("db: base de datos lista → %s", path)
}

// crearTablas ejecuta CREATE TABLE IF NOT EXISTS para cada entidad.
// Si ya creaste las tablas con DB Browser, estas sentencias no hacen nada.
func crearTablas() {
	sentencias := []string{

		// ── usuarios ────────────────────────────────────────────────────────
		// Almacena clientes y administradores de la tienda.
		// Email tiene restricción UNIQUE: no se puede registrar dos veces.
		// Password guarda el hash bcrypt, nunca la contraseña en texto plano.
		// Rol puede ser 'cliente' (por defecto) o 'admin'.
		`CREATE TABLE IF NOT EXISTS usuarios (
			ID_Usuario  INTEGER PRIMARY KEY AUTOINCREMENT,
			Nombre      TEXT    NOT NULL,
			Email       TEXT    NOT NULL UNIQUE,
			Password    TEXT    NOT NULL,
			Rol         TEXT    NOT NULL DEFAULT 'cliente'
		);`,

		// ── pedidos ─────────────────────────────────────────────────────────
		// Historial de compras confirmadas.
		// El carrito sigue viviendo en la sesión HTTP; solo se convierte
		// en pedido cuando el cliente presiona "Proceder al pago".
		// UsuarioID es clave foránea que referencia a usuarios.ID_Usuario.
		`CREATE TABLE IF NOT EXISTS pedidos (
			ID_Pedido   INTEGER PRIMARY KEY AUTOINCREMENT,
			UsuarioID   INTEGER NOT NULL REFERENCES usuarios(ID_Usuario),
			Fecha       TEXT    NOT NULL,
			Total       REAL    NOT NULL DEFAULT 0,
			Estado      TEXT    NOT NULL DEFAULT 'pendiente'
		);`,
	}

	for _, stmt := range sentencias {
		if _, err := DB.Exec(stmt); err != nil {
			log.Fatalf("db: error creando tablas: %v", err)
		}
	}
}
