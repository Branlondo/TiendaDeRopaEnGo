// Package carrito gestiona el carrito de compras usando la sesión de Gin.
// Los ítems se serializan como JSON en una cookie firmada (gin-contrib/sessions).
// Toda la lógica de negocio (agregar, quitar, cambiar cantidad, calcular total)
// vive aquí en Go, sin depender de JavaScript.
package carrito

import (
	"encoding/json"

	"Gin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const sessionKey = "carrito"

// Obtener devuelve los ítems actuales del carrito desde la sesión.
func Obtener(c *gin.Context) []models.ItemCarrito {
	session := sessions.Default(c)
	raw := session.Get(sessionKey)
	if raw == nil {
		return []models.ItemCarrito{}
	}
	var items []models.ItemCarrito
	var data []byte
	switch v := raw.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return []models.ItemCarrito{}
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return []models.ItemCarrito{}
	}
	return items
}

// guardar persiste el slice de ítems en la sesión.
func guardar(c *gin.Context, items []models.ItemCarrito) error {
	session := sessions.Default(c)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	session.Set(sessionKey, data)
	return session.Save()
}

// Agregar añade un producto al carrito (o incrementa la cantidad si ya existe).
func Agregar(c *gin.Context, productoID int, talla string, cantidad int) error {
	items := Obtener(c)
	for i := range items {
		if items[i].ProductoID == productoID && items[i].Talla == talla {
			items[i].Cantidad += cantidad
			return guardar(c, items)
		}
	}
	p := models.BuscarProductoPorID(productoID)
	if p == nil {
		return nil
	}
	items = append(items, models.ItemCarrito{
		ProductoID:  p.ID,
		Talla:       talla,
		Cantidad:    cantidad,
		Nombre:      p.Nombre,
		Descripcion: p.Descripcion,
		Precio:      p.Precio,
		Imagen:      p.Imagen,
	})
	return guardar(c, items)
}

// Eliminar quita un ítem del carrito por productoID + talla.
func Eliminar(c *gin.Context, productoID int, talla string) error {
	items := Obtener(c)
	nuevo := []models.ItemCarrito{}
	for _, item := range items {
		if !(item.ProductoID == productoID && item.Talla == talla) {
			nuevo = append(nuevo, item)
		}
	}
	return guardar(c, nuevo)
}

// CambiarCantidad actualiza la cantidad de un ítem; si queda <= 0 lo elimina.
func CambiarCantidad(c *gin.Context, productoID int, talla string, cantidad int) error {
	if cantidad <= 0 {
		return Eliminar(c, productoID, talla)
	}
	items := Obtener(c)
	for i := range items {
		if items[i].ProductoID == productoID && items[i].Talla == talla {
			items[i].Cantidad = cantidad
			return guardar(c, items)
		}
	}
	return nil
}

// Total calcula el precio total de todos los ítems del carrito.
func Total(items []models.ItemCarrito) float64 {
	var t float64
	for _, item := range items {
		t += item.Total()
	}
	return t
}

// ContarItems devuelve la cantidad total de unidades en el carrito.
func ContarItems(items []models.ItemCarrito) int {
	n := 0
	for _, item := range items {
		n += item.Cantidad
	}
	return n
}
