// carrito.js — MIGRADO A GO
//
// Toda la lógica del carrito (agregar, eliminar, cambiar cantidad,
// calcular total, mostrar badge) ahora vive en el servidor Go:
//
//   - Gin/carrito/carrito.go  → lógica de negocio (sesiones)
//   - Gin/routes/routes.go    → rutas POST: /carrito/agregar,
//                               /carrito/eliminar, /carrito/actualizar
//   - Gin/templates/Carrito.html → renderizado con Go templates
//
// Este archivo se conserva solo como referencia histórica.
// No contiene código activo.
