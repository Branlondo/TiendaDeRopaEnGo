// Package tests contiene las pruebas unitarias del proyecto RFF (Ropa Flash Fashion).
// Este archivo prueba el paquete models: structs de datos y sus métodos.
//
// Para ejecutar SOLO estas pruebas:
//
//	go test Gin/tests -run TestItemCarrito -v
//
// Para ejecutar TODAS las pruebas del proyecto:
//
//	go test ./tests/... -v
package tests

import (
	"testing"

	"Gin/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: models.ItemCarrito.Total()
//
// El método Total() multiplica Precio × Cantidad.
// Es la base del cálculo del subtotal de cada línea del carrito.
// ─────────────────────────────────────────────────────────────────────────────

// TestItemCarritoTotal_CasosBase verifica los casos fundamentales del método
// Total() usando una tabla de casos para máxima cobertura.
func TestItemCarritoTotal_CasosBase(t *testing.T) {
	casos := []struct {
		nombre   string
		precio   float64
		cantidad int
		esperado float64
	}{
		// Caso 1: valores típicos de una compra normal
		{"precio_normal_cantidad_2", 50000.0, 2, 100000.0},
		// Caso 2: precio cero → total siempre cero sin importar cantidad
		{"precio_cero", 0.0, 5, 0.0},
		// Caso 3: cantidad cero → total siempre cero sin importar precio
		{"cantidad_cero", 80000.0, 0, 0.0},
		// Caso 4: cantidad unitaria → total igual al precio
		{"cantidad_uno", 75000.0, 1, 75000.0},
		// Caso 5: precio y cantidad grandes (envío masivo)
		{"cantidad_grande", 15000.0, 100, 1500000.0},
		// Caso 6: precio con centavos (el struct guarda float64)
		{"precio_decimal", 29999.99, 2, 59999.98},
		// Caso 7: precio millonario (ropa de lujo)
		{"precio_millones", 1500000.0, 3, 4500000.0},
		// Caso 8: cantidad diez
		{"cantidad_diez", 25000.0, 10, 250000.0},
		// Caso 9: ambos uno → resultado igual al precio
		{"precio_y_cantidad_uno", 123456.0, 1, 123456.0},
		// Caso 10: precio mínimo de mercado
		{"precio_minimo", 1000.0, 1, 1000.0},
	}

	for _, c := range casos {
		c := c // captura de variable para goroutines paralelas
		t.Run(c.nombre, func(t *testing.T) {
			item := models.ItemCarrito{
				Precio:   c.precio,
				Cantidad: c.cantidad,
			}
			got := item.Total()
			if got != c.esperado {
				t.Errorf("Total() = %.2f; esperado %.2f (precio=%.2f, cantidad=%d)",
					got, c.esperado, c.precio, c.cantidad)
			}
		})
	}
}

// TestItemCarritoTotal_Multiplicacion verifica la propiedad algebraica:
// precio × cantidad = cantidad × precio (conmutatividad).
func TestItemCarritoTotal_Multiplicacion(t *testing.T) {
	precio := 37500.0
	cantidad := 4
	item := models.ItemCarrito{Precio: precio, Cantidad: cantidad}
	esperado := precio * float64(cantidad)
	if item.Total() != esperado {
		t.Errorf("Total() = %.2f; esperado %.2f", item.Total(), esperado)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: models.ItemCarrito — integridad de campos del struct
// ─────────────────────────────────────────────────────────────────────────────

// TestItemCarritoStruct_CamposVacios verifica que un ItemCarrito vacío tenga
// los valores cero para cada tipo (comportamiento de Go por defecto).
func TestItemCarritoStruct_CamposVacios(t *testing.T) {
	var item models.ItemCarrito

	t.Run("ProductoID_cero", func(t *testing.T) {
		if item.ProductoID != 0 {
			t.Errorf("ProductoID vacío debería ser 0, got %d", item.ProductoID)
		}
	})

	t.Run("Precio_cero", func(t *testing.T) {
		if item.Precio != 0.0 {
			t.Errorf("Precio vacío debería ser 0.0, got %.2f", item.Precio)
		}
	})

	t.Run("Cantidad_cero", func(t *testing.T) {
		if item.Cantidad != 0 {
			t.Errorf("Cantidad vacía debería ser 0, got %d", item.Cantidad)
		}
	})

	t.Run("Talla_vacia", func(t *testing.T) {
		if item.Talla != "" {
			t.Errorf("Talla vacía debería ser '', got '%s'", item.Talla)
		}
	})

	t.Run("Total_cero_cuando_vacio", func(t *testing.T) {
		if item.Total() != 0.0 {
			t.Errorf("Total() de item vacío debería ser 0, got %.2f", item.Total())
		}
	})
}

// TestItemCarritoStruct_AsignacionCampos verifica que los campos se asignen
// correctamente al construir el struct con valores explícitos.
func TestItemCarritoStruct_AsignacionCampos(t *testing.T) {
	item := models.ItemCarrito{
		ProductoID:  42,
		Talla:       "M",
		Cantidad:    3,
		Nombre:      "Camisa Lino",
		Descripcion: "Camisa de lino premium",
		Precio:      89000.0,
		Imagen:      "https://example.com/img.jpg",
	}

	t.Run("ProductoID_correcto", func(t *testing.T) {
		if item.ProductoID != 42 {
			t.Errorf("esperado 42, got %d", item.ProductoID)
		}
	})

	t.Run("Talla_correcta", func(t *testing.T) {
		if item.Talla != "M" {
			t.Errorf("esperado 'M', got '%s'", item.Talla)
		}
	})

	t.Run("Nombre_correcto", func(t *testing.T) {
		if item.Nombre != "Camisa Lino" {
			t.Errorf("esperado 'Camisa Lino', got '%s'", item.Nombre)
		}
	})

	t.Run("Total_calculado", func(t *testing.T) {
		esperado := 267000.0 // 89000 × 3
		if item.Total() != esperado {
			t.Errorf("esperado %.2f, got %.2f", esperado, item.Total())
		}
	})
}
