// models.go — define los structs (tipos de datos) de la aplicación.
// Solo contiene estructuras de datos, sin lógica de base de datos.
// Las funciones CRUD de cada entidad viven en sus propios paquetes:
//   - Productos  → Gin/producto/producto.go
//   - Usuarios   → Gin/auth/auth.go
//   - Categorías → Gin/db/db.go (ListarCategorias)
package models

// Categoria representa una sección de la tienda (Hombre / Mujer).
// Los datos vienen de la tabla 'categorias' de PostgreSQL.
type Categoria struct {
	ID          int
	Nombre      string
	Descripcion string
}

// Producto representa un artículo del catálogo de la tienda.
// Los datos vienen de la tabla 'productos' de PostgreSQL.
type Producto struct {
	ID              int
	Nombre          string
	Descripcion     string
	Precio          float64
	Talla           string   // primera talla disponible (valor por defecto)
	Tallas          []string // todas las tallas disponibles, ej: ["S","M","L"]
	Imagen          string
	CategoriaID     int
	NombreCategoria string // obtenido por JOIN con la tabla categorias
	Subcategoria    string
}

// ItemCarrito representa un producto dentro del carrito de compras (en sesión).
// Guarda una copia de los datos del producto para que, si el precio
// cambia, el carrito refleje el precio al momento de agregar.
type ItemCarrito struct {
	ProductoID  int
	Talla       string
	Cantidad    int
	Nombre      string
	Descripcion string
	Precio      float64
	Imagen      string
}

// Total calcula el subtotal de este ítem (precio × cantidad).
// Al ser un método del struct, se puede llamar desde los templates
// de Go con {{ .Total }} sin necesidad de funciones extra.
func (i ItemCarrito) Total() float64 {
	return i.Precio * float64(i.Cantidad)
}

// Usuario representa a una persona registrada en la tienda.
// El campo Rol puede ser "cliente" (comprador) o "admin" (gestión de tienda).
// Password nunca se muestra: auth.go lo cifra con bcrypt antes de guardarlo.
type Usuario struct {
	ID_Usuario int
	Nombre     string
	Email      string
	Password   string // hash bcrypt — nunca texto plano
	Rol        string // "cliente" | "admin"
}

// Pedido representa una compra confirmada por un cliente.
// Estado: "pendiente" → "pagado" | "cancelado"
type Pedido struct {
	ID_Pedido     int
	UsuarioID     int
	NombreCliente string // campo de JOIN con la tabla usuarios
	Fecha         string
	Total         float64
	Estado        string // "pendiente" | "pagado" | "cancelado"
	// Datos de envío
	EmailContacto string
	NombreEnvio   string
	ApellidoEnvio string
	Cedula        string
	Direccion     string
	Direccion2    string
	Ciudad        string
	Departamento  string
	CodigoPostal  string
	Telefono      string
	MetodoPago    string
	CostoEnvio    float64
}
