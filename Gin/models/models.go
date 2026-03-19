package models

// Define los modelos de datos para la aplicación, incluyendo las estructuras para las categorías y productos, con sus respectivos campos y tipos de datos.
type Categoria struct {
	ID          int
	Nombre      string
	Descripcion string
}

type Producto struct {
	ID          int
	Nombre      string
	Descripcion string
	Precio      float64
	Talla       string
	Imagen      string
	CategoriaID int
}

//Datos de ejemplo para las categorías y productos, que se pueden utilizar
// para probar la aplicación antes de conectar con una base de datos real.
var Categorias = []Categoria{
	{1, "Hombre", "Ropa para hombres"},
	{2, "Mujer", "Ropa para mujeres"},
}

//Datos de ejemplo para los productos, que se pueden utilizar para probar la aplicación antes de conectar con una base de datos real.
var Productos = []Producto{
	{1, "Camisa", "Camisa de algodón para hombre", 29.99, "M", "/static/assets/camisa.jpg", 1},
	{2, "Pantalón", "Pantalón de mezclilla para hombre", 49.99, "L", "/static/assets/pantalon.jpg", 1},
	{3, "Blusa", "Blusa de algodón para mujer", 39.99, "S", "/static/assets/blusa.jpg", 2},
	{4, "Falda", "Falda de flores para mujer", 24.99, "M", "/static/assets/falda.jpg", 2},
}
