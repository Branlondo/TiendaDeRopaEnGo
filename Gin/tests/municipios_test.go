// municipios_test.go — pruebas unitarias para los datos de municipios de Colombia.
//
// Este archivo valida la integridad del mapa municipiosColombia usado en el
// checkout de la tienda para el select dinámico de ciudades:
//
//   - Que existan los 33 departamentos + Bogotá D.C. (total 33)
//   - Que cada departamento tenga al menos 1 municipio
//   - Que ningún slice de municipios sea nil
//   - Que ciudades clave estén en sus departamentos correctos
//   - Que departamentos clave existan con el nombre exacto del mapa
//
// Los datos se acceden mediante las funciones exportadas GetMunicipios()
// y GetDepartamentos() definidas en routes/municipios.go.
//
// Para ejecutar:
//
//	go test Gin/tests -run TestMunicipios -v
package tests

import (
	"testing"

	"Gin/routes"
)

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Existencia de departamentos clave
// ─────────────────────────────────────────────────────────────────────────────

// TestMunicipios_DepartamentosTotales verifica que el mapa tenga exactamente
// 33 entradas (32 departamentos + Bogotá D.C.).
func TestMunicipios_DepartamentosTotales(t *testing.T) {
	deps := routes.GetDepartamentos()
	esperado := 33
	if len(deps) != esperado {
		t.Errorf("Total departamentos = %d; esperado %d", len(deps), esperado)
	}
}

// TestMunicipios_DepartamentosExisten verifica la existencia de cada uno
// de los 33 departamentos/Bogotá en el mapa, con el nombre exacto.
func TestMunicipios_DepartamentosExisten(t *testing.T) {
	departamentos := []string{
		"Amazonas", "Antioquia", "Arauca", "Atlántico", "Bogotá D.C.",
		"Bolívar", "Boyacá", "Caldas", "Caquetá", "Casanare", "Cauca",
		"Cesar", "Chocó", "Córdoba", "Cundinamarca", "Guainía", "Guaviare",
		"Huila", "La Guajira", "Magdalena", "Meta", "Nariño",
		"Norte de Santander", "Putumayo", "Quindío", "Risaralda",
		"San Andrés y Providencia", "Santander", "Sucre", "Tolima",
		"Valle del Cauca", "Vaupés", "Vichada",
	}

	for _, dep := range departamentos {
		dep := dep
		t.Run(dep, func(t *testing.T) {
			munis := routes.GetMunicipios(dep)
			if munis == nil {
				t.Errorf("Departamento %q no existe en el mapa (GetMunicipios devolvió nil)", dep)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Integridad de municipios por departamento
// ─────────────────────────────────────────────────────────────────────────────

// TestMunicipios_TodosTienenAlMenosUno verifica que ningún departamento
// tenga un slice de municipios vacío.
func TestMunicipios_TodosTienenAlMenosUno(t *testing.T) {
	deps := routes.GetDepartamentos()
	for _, dep := range deps {
		dep := dep
		t.Run(dep, func(t *testing.T) {
			munis := routes.GetMunicipios(dep)
			if len(munis) == 0 {
				t.Errorf("Departamento %q tiene 0 municipios", dep)
			}
		})
	}
}

// TestMunicipios_NingunoNil verifica que GetMunicipios nunca devuelva nil
// para un departamento que existe en el mapa.
func TestMunicipios_NingunoNil(t *testing.T) {
	deps := routes.GetDepartamentos()
	for _, dep := range deps {
		if routes.GetMunicipios(dep) == nil {
			t.Errorf("GetMunicipios(%q) devolvió nil", dep)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Ciudades clave en sus departamentos correctos
// ─────────────────────────────────────────────────────────────────────────────

// TestMunicipios_CiudadesClave verifica que ciudades principales estén
// registradas en el departamento correcto (sin typos ni asignaciones erróneas).
func TestMunicipios_CiudadesClave(t *testing.T) {
	ciudades := []struct {
		ciudad      string
		departamento string
	}{
		{"Leticia", "Amazonas"},
		{"Medellín", "Antioquia"},
		{"Bogotá", "Bogotá D.C."},
		{"Barranquilla", "Atlántico"},
		{"Cartagena", "Bolívar"},
		{"Tunja", "Boyacá"},
		{"Manizales", "Caldas"},
		{"Florencia", "Caquetá"},
		{"Popayán", "Cauca"},
		{"Valledupar", "Cesar"},
		{"Quibdó", "Chocó"},
		{"Montería", "Córdoba"},
		{"Bogotá", "Bogotá D.C."},
		{"Inírida", "Guainía"},
		{"San José del Guaviare", "Guaviare"},
		{"Neiva", "Huila"},
		{"Riohacha", "La Guajira"},
		{"Santa Marta", "Magdalena"},
		{"Villavicencio", "Meta"},
		{"Pasto", "Nariño"},
		{"Cúcuta", "Norte de Santander"},
		{"Mocoa", "Putumayo"},
		{"Armenia", "Quindío"},
		{"Pereira", "Risaralda"},
		{"Bucaramanga", "Santander"},
		{"Sincelejo", "Sucre"},
		{"Ibagué", "Tolima"},
		{"Cali", "Valle del Cauca"},
	}

	for _, c := range ciudades {
		c := c
		t.Run(c.ciudad+"_en_"+c.departamento, func(t *testing.T) {
			munis := routes.GetMunicipios(c.departamento)
			encontrado := false
			for _, m := range munis {
				if m == c.ciudad {
					encontrado = true
					break
				}
			}
			if !encontrado {
				t.Errorf("Ciudad %q no encontrada en departamento %q", c.ciudad, c.departamento)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Comportamiento con entradas inválidas
// ─────────────────────────────────────────────────────────────────────────────

// TestMunicipios_DepartamentoInexistente verifica que GetMunicipios devuelva
// nil para un departamento que no existe en el mapa.
func TestMunicipios_DepartamentoInexistente(t *testing.T) {
	casos := []string{
		"NoExiste",
		"",
		"antioquia",           // minúsculas → no coincide
		"Bogotá",              // sin "D.C." → no coincide
		"Valle",               // nombre incompleto
	}
	for _, dep := range casos {
		dep := dep
		t.Run("no_existe_"+dep, func(t *testing.T) {
			munis := routes.GetMunicipios(dep)
			if munis != nil {
				t.Errorf("GetMunicipios(%q) debería ser nil, got %v", dep, munis)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Departamentos con más municipios que un mínimo esperado
// ─────────────────────────────────────────────────────────────────────────────

// TestMunicipios_DepartamentosGrandesConMasMunicipios verifica que los
// departamentos más grandes tengan un número razonable de municipios.
func TestMunicipios_DepartamentosGrandesConMasMunicipios(t *testing.T) {
	casos := []struct {
		departamento string
		minimo       int
	}{
		{"Antioquia", 50},      // tiene más de 100 municipios
		{"Boyacá", 50},         // 123 municipios
		{"Cundinamarca", 50},   // 116 municipios
		{"Nariño", 30},         // 64 municipios
		{"Santander", 30},      // 87 municipios
		{"Valle del Cauca", 20},// 42 municipios
	}

	for _, c := range casos {
		c := c
		t.Run(c.departamento, func(t *testing.T) {
			munis := routes.GetMunicipios(c.departamento)
			if len(munis) < c.minimo {
				t.Errorf("%s tiene %d municipios; esperado al menos %d",
					c.departamento, len(munis), c.minimo)
			}
		})
	}
}

// TestMunicipios_NoHayDuplicados verifica que dentro de un departamento
// no haya municipios duplicados (Antioquia como muestra representativa).
func TestMunicipios_NoHayDuplicados(t *testing.T) {
	deps := []string{"Antioquia", "Cundinamarca", "Santander", "Valle del Cauca"}
	for _, dep := range deps {
		dep := dep
		t.Run("sin_duplicados_en_"+dep, func(t *testing.T) {
			munis := routes.GetMunicipios(dep)
			vistos := make(map[string]bool)
			for _, m := range munis {
				if vistos[m] {
					t.Errorf("%s: municipio duplicado encontrado: %q", dep, m)
				}
				vistos[m] = true
			}
		})
	}
}

// TestMunicipios_StringsNoVacias verifica que ningún municipio sea una
// cadena vacía dentro de los slices.
func TestMunicipios_StringsNoVacias(t *testing.T) {
	deps := routes.GetDepartamentos()
	for _, dep := range deps {
		munis := routes.GetMunicipios(dep)
		for _, m := range munis {
			if m == "" {
				t.Errorf("Departamento %q tiene un municipio con nombre vacío", dep)
			}
		}
	}
}
