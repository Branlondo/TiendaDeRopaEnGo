// auth.go — lógica de autenticación: registro, login, logout y middlewares.
// Este paquete es el único que toca la tabla "usuarios" de la base de datos.
package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"Gin/db"
	"Gin/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ─── Claves de sesión ────────────────────────────────────────────────────────
// Usar constantes evita errores tipográficos al leer/escribir la sesión.
// Todos los valores se guardan como string para evitar problemas de
// serialización gob con el cookie store.
const (
	sesUID      = "uid"        // ID del usuario como string ("0" = no autenticado)
	sesNombre   = "unombre"    // nombre visible en el navbar
	sesEmail    = "uemail"     // email del usuario
	sesRol      = "urol"       // "cliente" o "admin"
	sesFlashTip = "flash_tipo" // tipo de alerta Bootstrap: success, danger, warning
	sesFlashMsg = "flash_msg"  // texto del mensaje flash
)

// ─── Flash messages ──────────────────────────────────────────────────────────

// SetFlash guarda un mensaje de un solo uso en la sesión.
// El template lo lee la próxima vez que se carga la página y luego se borra.
// tipo acepta: "success", "danger", "warning", "info"
func SetFlash(c *gin.Context, tipo, msg string) {
	s := sessions.Default(c)
	s.Set(sesFlashTip, tipo)
	s.Set(sesFlashMsg, msg)
	s.Save()
}

// GetFlash lee el mensaje flash de la sesión y lo borra (one-shot).
// Devuelve ("", "") si no hay ningún flash pendiente.
func GetFlash(c *gin.Context) (tipo, msg string) {
	s := sessions.Default(c)
	tipo, _ = s.Get(sesFlashTip).(string)
	msg, _ = s.Get(sesFlashMsg).(string)
	if tipo != "" || msg != "" {
		s.Delete(sesFlashTip)
		s.Delete(sesFlashMsg)
		s.Save()
	}
	return
}

// ─── Sesión de usuario ───────────────────────────────────────────────────────

// UsuarioActual lee la sesión y devuelve los datos del usuario autenticado.
// Retorna nil si no hay sesión activa (usuario no ha iniciado sesión).
// Los handlers y templates usan esto para mostrar contenido condicional.
func UsuarioActual(c *gin.Context) *models.Usuario {
	s := sessions.Default(c)
	idStr, ok := s.Get(sesUID).(string)
	if !ok || idStr == "" {
		return nil
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		return nil
	}
	nombre, _ := s.Get(sesNombre).(string)
	email, _ := s.Get(sesEmail).(string)
	rol, _ := s.Get(sesRol).(string)
	return &models.Usuario{
		ID_Usuario: id,
		Nombre:     nombre,
		Email:      email,
		Rol:        rol,
	}
}

// guardarSesion escribe los datos del usuario en la cookie firmada.
// Se llama internamente después de un registro o login exitoso.
func guardarSesion(c *gin.Context, u *models.Usuario) {
	s := sessions.Default(c)
	s.Set(sesUID, strconv.Itoa(u.ID_Usuario))
	s.Set(sesNombre, u.Nombre)
	s.Set(sesEmail, u.Email)
	s.Set(sesRol, u.Rol)
	s.Save()
}

// ─── Registro ────────────────────────────────────────────────────────────────

// Registrar inserta un nuevo usuario con rol "cliente" en la base de datos.
// La contraseña se cifra con bcrypt antes de guardarse (costo 10).
// Devuelve error si el email ya existe (violación UNIQUE) o si faltan campos.
// Nota: PostgreSQL no soporta LastInsertId(), se usa RETURNING para obtener el ID.
func Registrar(nombre, email, password string) (*models.Usuario, error) {
	if nombre == "" || email == "" || password == "" {
		return nil, errors.New("todos los campos son obligatorios")
	}
	// bcrypt.DefaultCost = 10 iteraciones de hashing. Seguro y razonablemente rápido.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error cifrando contraseña: %w", err)
	}
	var id int
	err = db.DB.QueryRow(
		`INSERT INTO usuarios (Nombre, Email, Password, Rol) VALUES ($1, $2, $3, 'cliente') RETURNING ID_Usuario`,
		nombre, email, string(hash),
	).Scan(&id)
	if err != nil {
		// El error más común es la restricción UNIQUE del email
		return nil, errors.New("ese correo ya está registrado")
	}
	return &models.Usuario{
		ID_Usuario: id,
		Nombre:     nombre,
		Email:      email,
		Rol:        "cliente",
	}, nil
}

// ─── Login ───────────────────────────────────────────────────────────────────

// IniciarSesion busca el usuario por email, verifica la contraseña con bcrypt
// y guarda los datos en la sesión si todo es correcto.
// Devuelve error si el email no existe o la contraseña no coincide.
// El mensaje de error es genérico a propósito para no revelar si el email existe.
func IniciarSesion(c *gin.Context, email, password string) (*models.Usuario, error) {
	var u models.Usuario
	var hash string

	err := db.DB.QueryRow(
		`SELECT ID_Usuario, Nombre, Email, Password, Rol FROM usuarios WHERE Email = $1`,
		email,
	).Scan(&u.ID_Usuario, &u.Nombre, &u.Email, &hash, &u.Rol)

	if err == sql.ErrNoRows {
		return nil, errors.New("correo o contraseña incorrectos")
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando usuario: %w", err)
	}
	// CompareHashAndPassword devuelve error si la contraseña no coincide con el hash
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, errors.New("correo o contraseña incorrectos")
	}
	guardarSesion(c, &u)
	return &u, nil
}

// ─── Logout ──────────────────────────────────────────────────────────────────

// CerrarSesion borra todos los datos de la sesión del usuario.
// Después de esto UsuarioActual() devolverá nil.
func CerrarSesion(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Save()
}

// ─── Middlewares ─────────────────────────────────────────────────────────────

// RequiereLogin bloquea rutas que necesitan autenticación.
// Si el usuario no tiene sesión activa, lo redirige a "/" con un flash de aviso.
// Uso: r.GET("/mi-perfil", auth.RequiereLogin(), handler)
func RequiereLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if UsuarioActual(c) == nil {
			SetFlash(c, "warning", "Debes iniciar sesión para acceder a esa página.")
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequiereAdmin solo permite acceso a usuarios con rol "admin".
// Si el usuario es cliente o no está autenticado, redirige a "/" con flash de error.
// Uso: r.GET("/admin/dashboard", auth.RequiereAdmin(), handler)
func RequiereAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := UsuarioActual(c)
		if u == nil || u.Rol != "admin" {
			SetFlash(c, "danger", "No tienes permiso para acceder a esa página.")
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
