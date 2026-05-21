// Package helpers contiene funciones utilitarias puras (sin dependencias de BD
// ni de HTTP) que son usadas por el motor de templates de Gin y por la lógica
// de negocio de la tienda. Al vivir en su propio paquete, pueden probarse de
// forma aislada con go test sin necesitar una conexión a PostgreSQL.
package helpers

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ── Aritmética ────────────────────────────────────────────────────────────────

// Add suma dos enteros. Usada en templates para calcular índices.
func Add(a, b int) int { return a + b }

// Sub resta b de a. Devuelve 1 si el resultado sería menor que 1,
// lo que garantiza que el carrito nunca tenga cantidad cero o negativa.
func Sub(a, b int) int {
	if a-b < 1 {
		return 1
	}
	return a - b
}

// FSub resta dos float64. Usada en totales de pedido (total - costo envío).
func FSub(a, b float64) float64 { return a - b }

// ── Strings ───────────────────────────────────────────────────────────────────

// Join une un slice de strings con un separador arbitrario.
// Ejemplo: Join(["S","M","L"], " · ") → "S · M · L"
func Join(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

// ToLower convierte un string a minúsculas (usado para construir URLs
// como "/mujer" y "/hombre" a partir del nombre de categoría).
func ToLower(s string) string { return strings.ToLower(s) }

// ToUpper convierte un string a mayúsculas.
func ToUpper(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return strings.ToUpper(string(r))
}

// ── Búsqueda en slices ────────────────────────────────────────────────────────

// Contiene comprueba si item está en slice de strings (case-sensitive).
// Usada en los checkboxes de tallas del formulario admin.
func Contiene(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContieneInt comprueba si item está en un slice de enteros.
// Usada para pre-marcar checkboxes de productos relacionados en el admin.
func ContieneInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// StrSlice crea un slice de strings a partir de argumentos variádicos.
// Permite usar {{ range $t := strSlice "XS" "S" "M" "L" }} en templates.
func StrSlice(items ...string) []string { return items }

// ── Formateo ──────────────────────────────────────────────────────────────────

// FormatCOP formatea un float64 como precio en pesos colombianos con
// separadores de miles (punto). Trunca los centavos.
// Ejemplos: 1500 → "$1.500" | 1500000 → "$1.500.000"
func FormatCOP(v float64) string {
	n := int64(v)
	s := strconv.FormatInt(n, 10)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return "$" + result
}

// FormatFecha formatea un time.Time al estilo "DD/MM/YYYY HH:MM"
// para mostrarlo en facturas y detalles de pedido.
func FormatFecha(t time.Time) string {
	return t.Format("02/01/2006 15:04")
}
