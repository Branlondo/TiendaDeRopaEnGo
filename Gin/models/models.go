package models

// Define los modelos de datos para la aplicación.
type Categoria struct {
	ID          int
	Nombre      string
	Descripcion string
}

type Producto struct {
	ID           int
	Nombre       string
	Descripcion  string
	Precio       float64
	Talla        string
	Tallas       []string
	Imagen       string
	CategoriaID  int
	Subcategoria string
}

var Categorias = []Categoria{
	{111, "Hombre", "Ropa para hombres"},
	{222, "Mujer", "Ropa para mujeres"},
}

// ItemCarrito representa un producto dentro del carrito de compras.
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

// BuscarProductoPorID recorre el slice de productos y devuelve un
// puntero al producto con el ID indicado, o nil si no existe.
// El paquete carrito lo usa para construir el ItemCarrito al agregar.
func BuscarProductoPorID(id int) *Producto {
	for i, p := range Productos {
		if p.ID == id {
			return &Productos[i]
		}
	}
	return nil
}

// Para agregar una nueva subcategoría basta con crear un producto con el campo
// Subcategoria deseado; el menú de filtros se genera dinámicamente desde los datos.
var Productos = []Producto{
	{1, "Camisa", "Camisa de algodón para hombre", 29.99, "M", []string{"M", "L"}, "/static/assets/camisa.jpg", 111, "Camisa"},
	{2, "Pantalón", "Pantalón de mezclilla para hombre", 49.99, "L", []string{"L"}, "/static/assets/pantalon.jpg", 111, "Pantalón"},
	{3, "Blusa", "Blusa de algodón para mujer", 39.99, "S", []string{"S"}, "/static/assets/blusa.jpg", 222, "Blusa"},
	{4, "Falda", "Falda de flores para mujer", 24.99, "M", []string{"M"}, "/static/assets/falda.jpg", 222, "Falda"},
	{5, "Chaqueta", "Chaqueta de cuero para hombre", 89.99, "L", []string{"L"}, "/static/assets/chaqueta.jpg", 111, "Chaqueta"},
}
