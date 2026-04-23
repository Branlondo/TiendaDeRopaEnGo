// producto.go — repositorio de productos: toda la lógica de acceso a la tabla
// 'productos' de PostgreSQL vive aquí. Los handlers de routes.go llaman estas
// funciones en lugar de acceder a la DB directamente.
// NOTA: PostgreSQL usa $1, $2... en lugar del ? de SQLite.
package producto

import (
	"log"
	"strings"

	"Gin/db"
	"Gin/models"
)

// cols es la lista de columnas que se seleccionan en cada query.
// Se define aquí para no repetirla en cada función.
const cols = `SELECT ID_Producto, Nombre, COALESCE(Descripcion,''), Precio,
	COALESCE(ImagenURL,''), CategoriaID, COALESCE(Subcategoria,''), Tallas
	FROM productos`

// poblar lee tallasStr (CSV "S,M,L") y rellena los campos Tallas y Talla del struct.
func poblar(p *models.Producto, tallasStr string) {
	if tallasStr != "" {
		p.Tallas = strings.Split(tallasStr, ",")
	}
	if len(p.Tallas) > 0 {
		p.Talla = p.Tallas[0] // la primera talla es el valor por defecto
	}
}

// ── Consultas ────────────────────────────────────────────────────────────────

// Listar devuelve todos los productos ordenados por ID.
func Listar() []models.Producto {
	rows, err := db.DB.Query(cols + ` ORDER BY ID_Producto`)
	if err != nil {
		log.Println("producto.Listar:", err)
		return nil
	}
	defer rows.Close()
	var result []models.Producto
	for rows.Next() {
		var p models.Producto
		var tallasStr string
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen, &p.CategoriaID, &p.Subcategoria, &tallasStr); err != nil {
			continue
		}
		poblar(&p, tallasStr)
		result = append(result, p)
	}
	return result
}

// ListarPorCategoria devuelve los productos de una categoría (1=Hombre, 2=Mujer).
func ListarPorCategoria(categoriaID int) []models.Producto {
	rows, err := db.DB.Query(cols+` WHERE CategoriaID = $1 ORDER BY ID_Producto`, categoriaID)
	if err != nil {
		log.Println("producto.ListarPorCategoria:", err)
		return nil
	}
	defer rows.Close()
	var result []models.Producto
	for rows.Next() {
		var p models.Producto
		var tallasStr string
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen, &p.CategoriaID, &p.Subcategoria, &tallasStr); err != nil {
			continue
		}
		poblar(&p, tallasStr)
		result = append(result, p)
	}
	return result
}

// BuscarPorID devuelve un producto por su ID, o nil si no existe.
// Lo usa el carrito para obtener nombre/precio al agregar un ítem.
func BuscarPorID(id int) *models.Producto {
	var p models.Producto
	var tallasStr string
	err := db.DB.QueryRow(cols+` WHERE ID_Producto = $1`, id).
		Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen, &p.CategoriaID, &p.Subcategoria, &tallasStr)
	if err != nil {
		return nil
	}
	poblar(&p, tallasStr)
	return &p
}

// SubcategoriasUnicas extrae las subcategorías únicas de una categoría.
func SubcategoriasUnicas(categoriaID int) []string {
	rows, err := db.DB.Query(
		`SELECT DISTINCT Subcategoria FROM productos
		 WHERE CategoriaID = $1 AND Subcategoria IS NOT NULL AND Subcategoria != ''
		 ORDER BY Subcategoria`,
		categoriaID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		result = append(result, s)
	}
	return result
}

// ── CRUD para el panel admin ─────────────────────────────────────────────────

// Crear inserta un nuevo producto en la base de datos.
// tallas es un slice de strings como ["S","M","L"], se guarda como CSV.
func Crear(nombre, descripcion string, precio float64, imagen string, categoriaID int, subcategoria string, tallas []string) error {
	_, err := db.DB.Exec(
		`INSERT INTO productos (Nombre, Descripcion, Precio, ImagenURL, CategoriaID, Subcategoria, Tallas)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nombre, descripcion, precio, imagen, categoriaID, subcategoria, strings.Join(tallas, ","),
	)
	return err
}

// Actualizar sobreescribe todos los campos de un producto existente.
func Actualizar(id int, nombre, descripcion string, precio float64, imagen string, categoriaID int, subcategoria string, tallas []string) error {
	_, err := db.DB.Exec(
		`UPDATE productos
		 SET Nombre=$1, Descripcion=$2, Precio=$3, ImagenURL=$4, CategoriaID=$5, Subcategoria=$6, Tallas=$7
		 WHERE ID_Producto=$8`,
		nombre, descripcion, precio, imagen, categoriaID, subcategoria, strings.Join(tallas, ","), id,
	)
	return err
}

// Eliminar borra un producto de la base de datos por su ID.
// El admin ve un modal de confirmación antes de que se ejecute esta función.
func Eliminar(id int) error {
	_, err := db.DB.Exec(`DELETE FROM productos WHERE ID_Producto = $1`, id)
	return err
}
