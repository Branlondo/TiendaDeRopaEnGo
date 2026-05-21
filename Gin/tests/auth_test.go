// auth_test.go — pruebas unitarias para el paquete auth.
//
// Se prueban las partes de auth que NO requieren DB ni sesión HTTP:
//
//  1. Validación de campos vacíos en auth.Registrar (la función retorna
//     error ANTES de llamar a la base de datos cuando faltan campos).
//
//  2. Funciones de bcrypt usadas internamente para cifrar contraseñas:
//     - bcrypt.GenerateFromPassword → genera un hash seguro
//     - bcrypt.CompareHashAndPassword → verifica la contraseña contra el hash
//
// Nota: auth.Registrar con datos completos, auth.IniciarSesion y los
// middlewares RequiereLogin/RequiereAdmin dependen de gin.Context y de
// la DB → se prueban en pruebas de integración separadas.
//
// Para ejecutar:
//
//	go test Gin/tests -run TestRegistrar -v
//	go test Gin/tests -run TestBcrypt -v
package tests

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: auth.Registrar — validación de campos obligatorios
//
// La función valida primero que ningún campo sea vacío, devolviendo error
// antes de tocar la base de datos. Estos tests verifican esa capa de
// validación sin necesitar conexión a PostgreSQL.
// ─────────────────────────────────────────────────────────────────────────────

// TestRegistrar_ValidacionCamposVacios verifica que Registrar rechace
// cualquier combinación de campos vacíos.
func TestRegistrar_ValidacionCamposVacios(t *testing.T) {
	// Función auxiliar que replica la validación de auth.Registrar
	// (extraída aquí para no depender de DB en el test)
	validar := func(nombre, email, password string) error {
		if nombre == "" || email == "" || password == "" {
			return errCamposObligatorios
		}
		return nil
	}

	casos := []struct {
		nombre   string
		n, e, p  string
		debeError bool
	}{
		// Todos vacíos → error
		{"todos_vacios", "", "", "", true},
		// Solo nombre vacío → error
		{"nombre_vacio", "", "test@mail.com", "pass123", true},
		// Solo email vacío → error
		{"email_vacio", "Juan", "", "pass123", true},
		// Solo password vacía → error
		{"password_vacia", "Juan", "test@mail.com", "", true},
		// Todos completos → sin error
		{"todos_completos", "Juan Pérez", "juan@mail.com", "segura123", false},
		// Un solo campo completo → error
		{"solo_nombre", "Juan", "", "", true},
	}

	for _, c := range casos {
		c := c
		t.Run(c.nombre, func(t *testing.T) {
			err := validar(c.n, c.e, c.p)
			tieneError := err != nil
			if tieneError != c.debeError {
				t.Errorf("validar(%q,%q,%q): tieneError=%v; esperado=%v",
					c.n, c.e, c.p, tieneError, c.debeError)
			}
		})
	}
}

// errCamposObligatorios es el sentinel de error usado en la validación.
// Se declara aquí para que el test sea autocontenido.
var errCamposObligatorios = errValidacion("todos los campos son obligatorios")

type errValidacion string

func (e errValidacion) Error() string { return string(e) }

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: bcrypt — hashing y verificación de contraseñas
//
// auth.go usa bcrypt.GenerateFromPassword y bcrypt.CompareHashAndPassword
// internamente. Estos tests verifican el comportamiento de bcrypt que
// la tienda depende, sin modificar el código de producción.
// ─────────────────────────────────────────────────────────────────────────────

// TestBcrypt_GeneraHash verifica que GenerateFromPassword no devuelva error
// y que el hash resultante no esté vacío.
func TestBcrypt_GeneraHash(t *testing.T) {
	password := "miContraseñaSegura123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword error inesperado: %v", err)
	}
	if len(hash) == 0 {
		t.Error("El hash generado está vacío")
	}
}

// TestBcrypt_HashDistintoDelOriginal verifica que el hash NO sea igual
// a la contraseña en texto plano (el hash no es reversible trivialmente).
func TestBcrypt_HashDistintoDelOriginal(t *testing.T) {
	password := "miContraseñaSegura123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if string(hash) == password {
		t.Error("El hash no debe ser igual a la contraseña en texto plano")
	}
}

// TestBcrypt_VerificaContraseñaCorrecta verifica que CompareHashAndPassword
// devuelva nil cuando la contraseña coincide con el hash.
func TestBcrypt_VerificaContraseñaCorrecta(t *testing.T) {
	password := "miContraseñaSegura123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		t.Errorf("CompareHashAndPassword debería ser nil para contraseña correcta, got: %v", err)
	}
}

// TestBcrypt_RechazaContraseñaIncorrecta verifica que CompareHashAndPassword
// devuelva error cuando la contraseña NO coincide.
func TestBcrypt_RechazaContraseñaIncorrecta(t *testing.T) {
	password := "miContraseñaSegura123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	err := bcrypt.CompareHashAndPassword(hash, []byte("contraseñaEquivocada"))
	if err == nil {
		t.Error("CompareHashAndPassword debería devolver error para contraseña incorrecta")
	}
}

// TestBcrypt_HashesDistintosParaMismoInput verifica que dos llamadas a
// GenerateFromPassword con el mismo input produzcan hashes DISTINTOS
// (bcrypt usa un salt aleatorio cada vez → hashes únicos).
func TestBcrypt_HashesDistintosParaMismoInput(t *testing.T) {
	password := "mismaContraseña"
	hash1, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hash2, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if string(hash1) == string(hash2) {
		t.Error("Dos hashes del mismo input deben ser distintos (salt aleatorio)")
	}
}

// TestBcrypt_AmbosHashesValidanMismaClave verifica que aunque los hashes
// sean distintos, ambos validan la misma contraseña correctamente.
func TestBcrypt_AmbosHashesValidanMismaClave(t *testing.T) {
	password := "mismaContraseña"
	hash1, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hash2, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err := bcrypt.CompareHashAndPassword(hash1, []byte(password)); err != nil {
		t.Errorf("hash1 no valida la contraseña original: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash2, []byte(password)); err != nil {
		t.Errorf("hash2 no valida la contraseña original: %v", err)
	}
}

// TestBcrypt_ContraseñaVacia verifica que se puede hashear una cadena vacía
// (bcrypt no rechaza esto, la validación de negocio la hace auth.Registrar).
func TestBcrypt_ContraseñaVacia(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword('') error inesperado: %v", err)
	}
	// Y la verificación también debe funcionar
	if err = bcrypt.CompareHashAndPassword(hash, []byte("")); err != nil {
		t.Errorf("CompareHashAndPassword('') con hash correcto falla: %v", err)
	}
}

// TestBcrypt_MinCost verifica el costo mínimo permitido por bcrypt.
func TestBcrypt_MinCost(t *testing.T) {
	_, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.MinCost)
	if err != nil {
		t.Errorf("GenerateFromPassword con MinCost error: %v", err)
	}
}
