// carrito_test.go — pruebas unitarias para el paquete carrito.
//
// Se prueban ÚNICAMENTE las funciones puras que no necesitan DB ni sesión HTTP:
//   - Total(items)    → suma precio×cantidad de todos los ítems
//   - ContarItems(items) → suma de todas las cantidades
//
// Las funciones que dependen de gin.Context (Obtener, Agregar, Eliminar,
// CambiarCantidad, Vaciar) se validan mediante pruebas de integración HTTP
// y están excluidas de este archivo para mantenerlo sin dependencias externas.
//
// Para ejecutar:
//
//	go test Gin/tests -run TestTotal -v
//	go test Gin/tests -run TestContarItems -v
package tests

import (
	"testing"

	carritoPkg "Gin/carrito"
	"Gin/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers internos del test
// ─────────────────────────────────────────────────────────────────────────────

// nuevoItem crea un ItemCarrito con los campos mínimos para las pruebas.
func nuevoItem(id int, precio float64, cantidad int) models.ItemCarrito {
	return models.ItemCarrito{
		ProductoID: id,
		Nombre:     "Producto test",
		Precio:     precio,
		Cantidad:   cantidad,
		Talla:      "M",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: carrito.Total()
// ─────────────────────────────────────────────────────────────────────────────

// TestTotal_SliceVacio confirma que un carrito vacío tiene total 0.
func TestTotal_SliceVacio(t *testing.T) {
	total := carritoPkg.Total([]models.ItemCarrito{})
	if total != 0.0 {
		t.Errorf("Total de carrito vacío = %.2f; esperado 0.0", total)
	}
}

// TestTotal_NilSlice confirma que nil se trata como carrito vacío.
func TestTotal_NilSlice(t *testing.T) {
	total := carritoPkg.Total(nil)
	if total != 0.0 {
		t.Errorf("Total de nil = %.2f; esperado 0.0", total)
	}
}

// TestTotal_UnSoloItem verifica el cálculo con un único ítem.
func TestTotal_UnSoloItem(t *testing.T) {
	items := []models.ItemCarrito{nuevoItem(1, 50000.0, 2)}
	esperado := 100000.0
	if got := carritoPkg.Total(items); got != esperado {
		t.Errorf("Total = %.2f; esperado %.2f", got, esperado)
	}
}

// TestTotal_MultiplesItems verifica que la suma de múltiples ítems sea correcta.
func TestTotal_MultiplesItems(t *testing.T) {
	items := []models.ItemCarrito{
		nuevoItem(1, 50000.0, 2),  // subtotal 100 000
		nuevoItem(2, 30000.0, 1),  // subtotal  30 000
		nuevoItem(3, 20000.0, 3),  // subtotal  60 000
	}
	esperado := 190000.0
	if got := carritoPkg.Total(items); got != esperado {
		t.Errorf("Total = %.2f; esperado %.2f", got, esperado)
	}
}

// TestTotal_TablaVariantes usa tabla de casos para cubrir múltiples escenarios.
func TestTotal_TablaVariantes(t *testing.T) {
	casos := []struct {
		nombre   string
		items    []models.ItemCarrito
		esperado float64
	}{
		{
			"todos_precio_cero",
			[]models.ItemCarrito{nuevoItem(1, 0, 5), nuevoItem(2, 0, 3)},
			0.0,
		},
		{
			"un_item_cantidad_cero",
			[]models.ItemCarrito{nuevoItem(1, 80000, 0)},
			0.0,
		},
		{
			"precio_decimal",
			[]models.ItemCarrito{nuevoItem(1, 29999.99, 2)},
			59999.98,
		},
		{
			"cantidad_grande",
			[]models.ItemCarrito{nuevoItem(1, 15000.0, 100)},
			1500000.0,
		},
		{
			"dos_items_misma_cantidad",
			[]models.ItemCarrito{nuevoItem(1, 40000, 2), nuevoItem(2, 60000, 2)},
			200000.0,
		},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := carritoPkg.Total(c.items)
			if got != c.esperado {
				t.Errorf("Total() = %.2f; esperado %.2f", got, c.esperado)
			}
		})
	}
}

// TestTotal_ConsistenciaConItemTotal verifica que carrito.Total() coincide
// con la suma manual de item.Total() para cada ítem.
func TestTotal_ConsistenciaConItemTotal(t *testing.T) {
	items := []models.ItemCarrito{
		nuevoItem(1, 45000, 2),
		nuevoItem(2, 78000, 1),
		nuevoItem(3, 12500, 4),
	}
	var sumaManual float64
	for _, it := range items {
		sumaManual += it.Total()
	}
	got := carritoPkg.Total(items)
	if got != sumaManual {
		t.Errorf("carrito.Total() = %.2f; suma manual = %.2f (deberían ser iguales)", got, sumaManual)
	}
}

// TestTotal_OrdenNoImporta confirma que sumar en distinto orden da el mismo resultado.
func TestTotal_OrdenNoImporta(t *testing.T) {
	a := []models.ItemCarrito{nuevoItem(1, 30000, 2), nuevoItem(2, 70000, 1)}
	b := []models.ItemCarrito{nuevoItem(2, 70000, 1), nuevoItem(1, 30000, 2)}
	if carritoPkg.Total(a) != carritoPkg.Total(b) {
		t.Errorf("Total debería ser igual sin importar el orden de ítems")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: carrito.ContarItems()
// ─────────────────────────────────────────────────────────────────────────────

// TestContarItems_SliceVacio confirma que un carrito vacío tiene 0 unidades.
func TestContarItems_SliceVacio(t *testing.T) {
	if n := carritoPkg.ContarItems([]models.ItemCarrito{}); n != 0 {
		t.Errorf("ContarItems vacío = %d; esperado 0", n)
	}
}

// TestContarItems_NilSlice confirma que nil devuelve 0.
func TestContarItems_NilSlice(t *testing.T) {
	if n := carritoPkg.ContarItems(nil); n != 0 {
		t.Errorf("ContarItems nil = %d; esperado 0", n)
	}
}

// TestContarItems_TablaVariantes usa tabla de casos para diferentes configuraciones.
func TestContarItems_TablaVariantes(t *testing.T) {
	casos := []struct {
		nombre   string
		items    []models.ItemCarrito
		esperado int
	}{
		{"un_item_cantidad_1", []models.ItemCarrito{nuevoItem(1, 50000, 1)}, 1},
		{"un_item_cantidad_5", []models.ItemCarrito{nuevoItem(1, 50000, 5)}, 5},
		{"dos_items_suma", []models.ItemCarrito{nuevoItem(1, 50000, 3), nuevoItem(2, 30000, 2)}, 5},
		{"tres_items_mixto", []models.ItemCarrito{nuevoItem(1, 10000, 1), nuevoItem(2, 20000, 2), nuevoItem(3, 30000, 3)}, 6},
		{"cantidad_cero_ignorada", []models.ItemCarrito{nuevoItem(1, 50000, 0), nuevoItem(2, 30000, 3)}, 3},
		{"todos_cantidad_uno", []models.ItemCarrito{nuevoItem(1, 10000, 1), nuevoItem(2, 20000, 1), nuevoItem(3, 30000, 1)}, 3},
		{"cantidad_grande", []models.ItemCarrito{nuevoItem(1, 15000, 50), nuevoItem(2, 20000, 50)}, 100},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := carritoPkg.ContarItems(c.items)
			if got != c.esperado {
				t.Errorf("ContarItems() = %d; esperado %d", got, c.esperado)
			}
		})
	}
}

// TestContarItems_SumaManual verifica que ContarItems coincide con sumar
// manualmente las cantidades de cada ítem.
func TestContarItems_SumaManual(t *testing.T) {
	items := []models.ItemCarrito{
		nuevoItem(1, 50000, 3),
		nuevoItem(2, 30000, 7),
		nuevoItem(3, 20000, 2),
	}
	var sumaManual int
	for _, it := range items {
		sumaManual += it.Cantidad
	}
	got := carritoPkg.ContarItems(items)
	if got != sumaManual {
		t.Errorf("ContarItems() = %d; suma manual = %d", got, sumaManual)
	}
}

// TestContarItems_IndependienteDelPrecio confirma que el precio no afecta
// el conteo de unidades (solo importa la cantidad).
func TestContarItems_IndependienteDelPrecio(t *testing.T) {
	itemBarato := []models.ItemCarrito{nuevoItem(1, 1000, 5)}
	itemCaro := []models.ItemCarrito{nuevoItem(1, 9999999, 5)}

	if carritoPkg.ContarItems(itemBarato) != carritoPkg.ContarItems(itemCaro) {
		t.Error("ContarItems no debe depender del precio del ítem")
	}
}
