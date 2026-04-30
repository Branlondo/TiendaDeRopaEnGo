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
// Incluye JOIN con categorias para obtener el nombre de la categoría
// sin necesidad de hardcodear IDs en los templates.
const cols = `SELECT p.ID_Producto, p.Nombre, COALESCE(p.Descripcion,''), p.Precio,
	COALESCE(p.ImagenURL,''), p.CategoriaID, COALESCE(p.Subcategoria,''), p.Tallas,
	COALESCE(c.Nombre,'')
	FROM productos p
	LEFT JOIN categorias c ON p.CategoriaID = c.ID_Categoria`

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
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen, &p.CategoriaID, &p.Subcategoria, &tallasStr, &p.NombreCategoria); err != nil {
			continue
		}
		poblar(&p, tallasStr)
		result = append(result, p)
	}
	return result
}

// ListarPorCategoria devuelve los productos de una categoría (1=Hombre, 2=Mujer).
func ListarPorCategoria(categoriaID int) []models.Producto {
	rows, err := db.DB.Query(cols+` WHERE p.CategoriaID = $1 ORDER BY p.ID_Producto`, categoriaID)
	if err != nil {
		log.Println("producto.ListarPorCategoria:", err)
		return nil
	}
	defer rows.Close()
	var result []models.Producto
	for rows.Next() {
		var p models.Producto
		var tallasStr string
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen, &p.CategoriaID, &p.Subcategoria, &tallasStr, &p.NombreCategoria); err != nil {
			continue
		}
		poblar(&p, tallasStr)
		result = append(result, p)
	}
	return result
}

// BuscarPorID devuelve un producto por su ID, o nil si no existe.
// También carga la galería de imágenes desde producto_imagenes.
func BuscarPorID(id int) *models.Producto {
	var p models.Producto
	var tallasStr string
	err := db.DB.QueryRow(cols+` WHERE p.ID_Producto = $1`, id).
		Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen, &p.CategoriaID, &p.Subcategoria, &tallasStr, &p.NombreCategoria)
	if err != nil {
		return nil
	}
	poblar(&p, tallasStr)
	p.Imagenes = ObtenerImagenes(id)
	return &p
}

// ObtenerImagenes devuelve todas las URLs de la galería de un producto.
// La portada aparece primero (ORDER BY EsPortada DESC).
func ObtenerImagenes(productoID int) []string {
	rows, err := db.DB.Query(
		`SELECT ImagenURL FROM producto_imagenes
		 WHERE ProductoID = $1 ORDER BY EsPortada DESC, ID_Imagen ASC`,
		productoID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		rows.Scan(&u)
		urls = append(urls, u)
	}
	return urls
}

// GuardarImagenes reemplaza toda la galería de un producto.
// portadaURL es la URL que se marcará como EsPortada.
func GuardarImagenes(productoID int, urls []string, portadaURL string) {
	db.DB.Exec(`DELETE FROM producto_imagenes WHERE ProductoID = $1`, productoID)
	for _, u := range urls {
		if u == "" {
			continue
		}
		esPortada := u == portadaURL
		db.DB.Exec(
			`INSERT INTO producto_imagenes (ProductoID, ImagenURL, EsPortada) VALUES ($1, $2, $3)`,
			productoID, u, esPortada,
		)
	}
}

// ObtenerRelacionados devuelve los productos vinculados a uno dado.
// La relación es bidireccional: si A→B existe, B también devuelve A.
func ObtenerRelacionados(productoID int) []models.Producto {
	rows, err := db.DB.Query(
		cols+` WHERE p.ID_Producto IN (
			SELECT RelacionadoID FROM producto_relaciones WHERE ProductoID = $1
			UNION
			SELECT ProductoID    FROM producto_relaciones WHERE RelacionadoID = $1
		) AND p.ID_Producto != $1`,
		productoID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []models.Producto
	for rows.Next() {
		var p models.Producto
		var tallasStr string
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Imagen,
			&p.CategoriaID, &p.Subcategoria, &tallasStr, &p.NombreCategoria); err != nil {
			continue
		}
		poblar(&p, tallasStr)
		result = append(result, p)
	}
	return result
}

// ObtenerRelacionadosIDs devuelve solo los IDs de los relacionados (para el formulario admin).
func ObtenerRelacionadosIDs(productoID int) []int {
	rows, err := db.DB.Query(
		`SELECT RelacionadoID FROM producto_relaciones WHERE ProductoID = $1`,
		productoID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids
}

// GuardarRelaciones reemplaza las relaciones de un producto.
func GuardarRelaciones(productoID int, relacionadosIDs []int) {
	db.DB.Exec(`DELETE FROM producto_relaciones WHERE ProductoID = $1`, productoID)
	for _, relID := range relacionadosIDs {
		if relID == productoID {
			continue
		}
		db.DB.Exec(
			`INSERT INTO producto_relaciones (ProductoID, RelacionadoID) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			productoID, relID,
		)
	}
}

// SubcategoriasUnicas extrae las subcategorías únicas de una categoría.
func SubcategoriasUnicas(categoriaID int) []string {
	rows, err := db.DB.Query(
		`SELECT DISTINCT Subcategoria FROM productos
		 WHERE CategoriaID = $1 AND Subcategoria IS NOT NULL AND Subcategoria != ''
		 ORDER BY Subcategoria`,
		categoriaID,
	) // esta query no usa el JOIN, no necesita prefijo de tabla
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

// Crear inserta un nuevo producto y devuelve su ID generado.
// tallas es un slice de strings como ["S","M","L"], se guarda como CSV.
func Crear(nombre, descripcion string, precio float64, imagen string, categoriaID int, subcategoria string, tallas []string) (int, error) {
	var id int
	err := db.DB.QueryRow(
		`INSERT INTO productos (Nombre, Descripcion, Precio, ImagenURL, CategoriaID, Subcategoria, Tallas)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING ID_Producto`,
		nombre, descripcion, precio, imagen, categoriaID, subcategoria, strings.Join(tallas, ","),
	).Scan(&id)
	return id, err
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
