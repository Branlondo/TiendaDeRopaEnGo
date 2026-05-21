// helpers_test.go — pruebas unitarias para el paquete helpers.
//
// Este archivo cubre todas las funciones utilitarias que son usadas tanto
// en el motor de templates de Gin como en la lógica de la tienda:
//
//   - FormatCOP    → formatea precios en pesos colombianos ($1.500.000)
//   - Add / Sub / FSub → aritmética para templates y carrito
//   - Join         → une slices de strings (tallas, etc.)
//   - Contiene     → busca string en slice (checkboxes de tallas)
//   - ContieneInt  → busca int en slice (checkboxes de relacionados)
//   - StrSlice     → crea slice de strings desde variádicos
//   - ToLower      → convierte a minúsculas (construcción de URLs)
//   - ToUpper      → convierte a mayúsculas (etiquetas)
//   - FormatFecha  → formatea time.Time como "DD/MM/YYYY HH:MM"
//
// Para ejecutar:
//
//	go test Gin/tests -run TestFormat -v
//	go test Gin/tests -run TestJoin -v
package tests

import (
	"testing"
	"time"

	"Gin/helpers"
)

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.FormatCOP
//
// FormatCOP convierte un float64 a cadena con separador de miles (punto)
// y prefijo "$". Los centavos se truncan.
// ─────────────────────────────────────────────────────────────────────────────

func TestFormatCOP(t *testing.T) {
	casos := []struct {
		nombre   string
		valor    float64
		esperado string
	}{
		// Caso 1: cero → solo prefijo y cifra
		{"cero", 0, "$0"},
		// Caso 2: menor a 1000 → sin separador
		{"novecientos", 900, "$900"},
		// Caso 3: exactamente 1000 → primer separador
		{"mil", 1000, "$1.000"},
		// Caso 4: diez mil
		{"diez_mil", 10000, "$10.000"},
		// Caso 5: cien mil
		{"cien_mil", 100000, "$100.000"},
		// Caso 6: un millón
		{"un_millon", 1000000, "$1.000.000"},
		// Caso 7: precio típico de prenda (39 990)
		{"precio_tipico", 39990, "$39.990"},
		// Caso 8: precio premium
		{"precio_premium", 1500000, "$1.500.000"},
		// Caso 9: los decimales se truncan (no redondean)
		{"decimal_trunca", 29999.99, "$29.999"},
		// Caso 10: precio de envío (15 000)
		{"costo_envio", 15000, "$15.000"},
		// Caso 11: precio exacto dos separadores
		{"dos_separadores", 12345678, "$12.345.678"},
		// Caso 12: precio mínimo (1 peso)
		{"un_peso", 1, "$1"},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := helpers.FormatCOP(c.valor)
			if got != c.esperado {
				t.Errorf("FormatCOP(%.2f) = %q; esperado %q", c.valor, got, c.esperado)
			}
		})
	}
}

// TestFormatCOP_PrefijoDolar verifica que SIEMPRE empiece con "$".
func TestFormatCOP_PrefijoDolar(t *testing.T) {
	resultado := helpers.FormatCOP(55000)
	if len(resultado) == 0 || resultado[0] != '$' {
		t.Errorf("FormatCOP debe empezar con '$', got %q", resultado)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.Add
// ─────────────────────────────────────────────────────────────────────────────

func TestAdd(t *testing.T) {
	casos := []struct{ a, b, esperado int }{
		{0, 0, 0},
		{1, 1, 2},
		{5, 3, 8},
		{-1, 1, 0},
		{100, -50, 50},
	}
	for _, c := range casos {
		got := helpers.Add(c.a, c.b)
		if got != c.esperado {
			t.Errorf("Add(%d, %d) = %d; esperado %d", c.a, c.b, got, c.esperado)
		}
	}
}

// TestAdd_Conmutativity verifica que a+b = b+a.
func TestAdd_Conmutativity(t *testing.T) {
	if helpers.Add(3, 7) != helpers.Add(7, 3) {
		t.Error("Add debe ser conmutativa: Add(3,7) != Add(7,3)")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.Sub
//
// Sub(a, b) devuelve a-b, pero nunca menos de 1.
// Esto evita que el carrito tenga cantidad 0 o negativa.
// ─────────────────────────────────────────────────────────────────────────────

func TestSub(t *testing.T) {
	casos := []struct {
		nombre   string
		a, b     int
		esperado int
	}{
		// Resultado positivo normal
		{"resta_normal", 5, 2, 3},
		// Resultado exactamente 1
		{"resultado_exactamente_1", 3, 2, 1},
		// a == b → resultado sería 0 → devuelve 1
		{"a_igual_b_devuelve_1", 4, 4, 1},
		// a < b → resultado negativo → devuelve 1
		{"negativo_devuelve_1", 2, 5, 1},
		// a = 1, b = 1 → devuelve 1 (mínimo carrito)
		{"minimo_carrito", 1, 1, 1},
	}
	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := helpers.Sub(c.a, c.b)
			if got != c.esperado {
				t.Errorf("Sub(%d, %d) = %d; esperado %d", c.a, c.b, got, c.esperado)
			}
		})
	}
}

// TestSub_NuncaMenorQueUno garantiza que Sub nunca devuelva < 1.
func TestSub_NuncaMenorQueUno(t *testing.T) {
	entradas := [][2]int{{1, 1}, {1, 100}, {0, 1}, {-5, 3}}
	for _, par := range entradas {
		got := helpers.Sub(par[0], par[1])
		if got < 1 {
			t.Errorf("Sub(%d, %d) = %d; nunca debe ser < 1", par[0], par[1], got)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.FSub
// ─────────────────────────────────────────────────────────────────────────────

func TestFSub(t *testing.T) {
	casos := []struct {
		nombre   string
		a, b     float64
		esperado float64
	}{
		{"resta_float_normal", 190000.0, 15000.0, 175000.0},
		{"resultado_negativo", 10000.0, 50000.0, -40000.0},
		{"resta_cero", 75000.0, 0.0, 75000.0},
		{"ambos_iguales", 30000.0, 30000.0, 0.0},
	}
	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := helpers.FSub(c.a, c.b)
			if got != c.esperado {
				t.Errorf("FSub(%.2f, %.2f) = %.2f; esperado %.2f", c.a, c.b, got, c.esperado)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.Join
// ─────────────────────────────────────────────────────────────────────────────

func TestJoin(t *testing.T) {
	casos := []struct {
		nombre    string
		slice     []string
		sep       string
		esperado  string
	}{
		{"slice_vacio", []string{}, ",", ""},
		{"un_elemento", []string{"M"}, " · ", "M"},
		{"tallas_tipicas", []string{"S", "M", "L"}, " · ", "S · M · L"},
		{"separador_coma", []string{"XS", "S", "M", "L", "XL"}, ", ", "XS, S, M, L, XL"},
		{"separador_vacio", []string{"a", "b", "c"}, "", "abc"},
		{"separador_guion", []string{"Rojo", "Azul"}, " - ", "Rojo - Azul"},
	}
	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := helpers.Join(c.slice, c.sep)
			if got != c.esperado {
				t.Errorf("Join(%v, %q) = %q; esperado %q", c.slice, c.sep, got, c.esperado)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.Contiene
// ─────────────────────────────────────────────────────────────────────────────

func TestContiene(t *testing.T) {
	tallas := []string{"XS", "S", "M", "L", "XL"}

	casos := []struct {
		nombre   string
		item     string
		esperado bool
	}{
		{"primer_elemento", "XS", true},
		{"elemento_medio", "M", true},
		{"ultimo_elemento", "XL", true},
		{"no_existe", "XXL", false},
		{"string_vacio", "", false},
		{"case_sensitive_lower", "m", false},  // "m" != "M"
		{"case_sensitive_mixed", "Xs", false}, // "Xs" != "XS"
	}
	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := helpers.Contiene(tallas, c.item)
			if got != c.esperado {
				t.Errorf("Contiene(%v, %q) = %v; esperado %v", tallas, c.item, got, c.esperado)
			}
		})
	}
}

// TestContiene_SliceVacio verifica que un slice vacío siempre devuelve false.
func TestContiene_SliceVacio(t *testing.T) {
	if helpers.Contiene([]string{}, "M") {
		t.Error("Contiene en slice vacío debe devolver false")
	}
}

// TestContiene_SliceNil verifica que nil se trate como vacío.
func TestContiene_SliceNil(t *testing.T) {
	if helpers.Contiene(nil, "M") {
		t.Error("Contiene en nil debe devolver false")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.ContieneInt
// ─────────────────────────────────────────────────────────────────────────────

func TestContieneInt(t *testing.T) {
	ids := []int{1, 5, 10, 23, 99}

	casos := []struct {
		nombre   string
		item     int
		esperado bool
	}{
		{"primer_elemento", 1, true},
		{"elemento_medio", 10, true},
		{"ultimo_elemento", 99, true},
		{"no_existe", 50, false},
		{"cero_no_existe", 0, false},
		{"negativo_no_existe", -1, false},
	}
	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			got := helpers.ContieneInt(ids, c.item)
			if got != c.esperado {
				t.Errorf("ContieneInt(%v, %d) = %v; esperado %v", ids, c.item, got, c.esperado)
			}
		})
	}
}

// TestContieneInt_SliceVacio verifica que un slice vacío siempre devuelve false.
func TestContieneInt_SliceVacio(t *testing.T) {
	if helpers.ContieneInt([]int{}, 5) {
		t.Error("ContieneInt en slice vacío debe devolver false")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.StrSlice
// ─────────────────────────────────────────────────────────────────────────────

func TestStrSlice(t *testing.T) {
	t.Run("sin_argumentos", func(t *testing.T) {
		got := helpers.StrSlice()
		if len(got) != 0 {
			t.Errorf("StrSlice() debería ser vacío, got %v", got)
		}
	})

	t.Run("un_argumento", func(t *testing.T) {
		got := helpers.StrSlice("XS")
		if len(got) != 1 || got[0] != "XS" {
			t.Errorf("StrSlice('XS') = %v; esperado ['XS']", got)
		}
	})

	t.Run("multiples_argumentos", func(t *testing.T) {
		got := helpers.StrSlice("XS", "S", "M", "L", "XL", "XXL")
		if len(got) != 6 {
			t.Errorf("StrSlice con 6 args debería tener len=6, got %d", len(got))
		}
	})

	t.Run("preserva_orden", func(t *testing.T) {
		got := helpers.StrSlice("primero", "segundo", "tercero")
		if got[0] != "primero" || got[1] != "segundo" || got[2] != "tercero" {
			t.Errorf("StrSlice no preserva el orden: %v", got)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.ToLower
// ─────────────────────────────────────────────────────────────────────────────

func TestToLower(t *testing.T) {
	casos := []struct{ entrada, esperado string }{
		{"Mujer", "mujer"},
		{"HOMBRE", "hombre"},
		{"mUjEr", "mujer"},
		{"", ""},
		{"mujer", "mujer"}, // ya está en minúsculas
	}
	for _, c := range casos {
		got := helpers.ToLower(c.entrada)
		if got != c.esperado {
			t.Errorf("ToLower(%q) = %q; esperado %q", c.entrada, got, c.esperado)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.ToUpper
// ─────────────────────────────────────────────────────────────────────────────

func TestToUpper(t *testing.T) {
	casos := []struct{ entrada, esperado string }{
		{"mujer", "MUJER"},
		{"HOMBRE", "HOMBRE"},
		{"hombre", "HOMBRE"},
		{"", ""},
	}
	for _, c := range casos {
		got := helpers.ToUpper(c.entrada)
		if got != c.esperado {
			t.Errorf("ToUpper(%q) = %q; esperado %q", c.entrada, got, c.esperado)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: helpers.FormatFecha
// ─────────────────────────────────────────────────────────────────────────────

// TestFormatFecha_FormatoDD_MM_YYYY verifica que la salida use DD/MM/YYYY HH:MM.
func TestFormatFecha_FormatoDD_MM_YYYY(t *testing.T) {
	// 14 de mayo de 2026 a las 11:30
	fecha := time.Date(2026, 5, 14, 11, 30, 0, 0, time.UTC)
	esperado := "14/05/2026 11:30"
	got := helpers.FormatFecha(fecha)
	if got != esperado {
		t.Errorf("FormatFecha = %q; esperado %q", got, esperado)
	}
}

// TestFormatFecha_DiaCeroRelleno verifica que el día < 10 lleve cero inicial.
func TestFormatFecha_DiaCeroRelleno(t *testing.T) {
	fecha := time.Date(2026, 1, 5, 9, 5, 0, 0, time.UTC)
	esperado := "05/01/2026 09:05"
	got := helpers.FormatFecha(fecha)
	if got != esperado {
		t.Errorf("FormatFecha = %q; esperado %q", got, esperado)
	}
}

// TestFormatFecha_LongitudFija verifica que la cadena resultante tenga
// exactamente 16 caracteres ("DD/MM/YYYY HH:MM").
func TestFormatFecha_LongitudFija(t *testing.T) {
	fecha := time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC)
	got := helpers.FormatFecha(fecha)
	if len(got) != 16 {
		t.Errorf("FormatFecha debería tener 16 chars, got %d: %q", len(got), got)
	}
}

// TestFormatFecha_Separadores verifica que usa '/' como separador de fecha
// y ' ' entre fecha y hora.
func TestFormatFecha_Separadores(t *testing.T) {
	fecha := time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC)
	got := helpers.FormatFecha(fecha)
	if got[2] != '/' || got[5] != '/' || got[10] != ' ' || got[13] != ':' {
		t.Errorf("FormatFecha separadores incorrectos: %q", got)
	}
}
