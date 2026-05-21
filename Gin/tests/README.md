# Pruebas Unitarias — RFF (Ropa Flash Fashion)

Carpeta de pruebas unitarias del proyecto. Cubre **las funciones puras y de
lógica de negocio** que no dependen de base de datos ni de contexto HTTP.

---

## Cómo ejecutar

```bash
# Todas las pruebas de esta carpeta
go test ./tests/... -v

# Con reporte de cobertura
go test ./tests/... -cover

# Solo un archivo / grupo de tests
go test Gin/tests -run TestItemCarrito -v
go test Gin/tests -run TestTotal -v
go test Gin/tests -run TestFormatCOP -v
go test Gin/tests -run TestMunicipios -v

# Pruebas del paquete helpers directamente
go test Gin/helpers/... -v
```

---

## Estructura de archivos

```
tests/
├── models_test.go       → structs y método Total() de ItemCarrito
├── carrito_test.go      → funciones puras del carrito (Total, ContarItems)
├── helpers_test.go      → utilidades de templates (FormatCOP, Join, etc.)
├── auth_test.go         → validación de campos y bcrypt
├── municipios_test.go   → integridad del mapa de municipios de Colombia
└── README.md            → este archivo
```

---

## Descripción de cada archivo

### `models_test.go` — Paquete `models`

Prueba el struct `ItemCarrito` y su método `Total()`.

| Test | Qué verifica |
|------|-------------|
| `TestItemCarritoTotal_CasosBase` | 10 subtests: precio normal, precio cero, cantidad cero, cantidad 1, cantidad grande, precio decimal, precio millones, cantidad 10, precio y cantidad 1, precio mínimo |
| `TestItemCarritoTotal_Multiplicacion` | Que precio × cantidad sea algebraicamente correcto |
| `TestItemCarritoStruct_CamposVacios` | 5 subtests: que un struct vacío tenga valores cero por defecto de Go |
| `TestItemCarritoStruct_AsignacionCampos` | 4 subtests: que todos los campos se asignen correctamente al construir con valores explícitos |

**Total tests: 21**

---

### `carrito_test.go` — Paquete `carrito`

Prueba `carrito.Total()` y `carrito.ContarItems()` (sin DB ni sesión HTTP).

| Test | Qué verifica |
|------|-------------|
| `TestTotal_SliceVacio` | Carrito vacío → total 0.0 |
| `TestTotal_NilSlice` | nil se trata como vacío |
| `TestTotal_UnSoloItem` | Cálculo correcto con 1 ítem |
| `TestTotal_MultiplesItems` | Suma de 3 ítems con distintos precios/cantidades |
| `TestTotal_TablaVariantes` | 5 subtests: todos precio 0, cantidad 0, precio decimal, cantidad 100, dos ítems misma cantidad |
| `TestTotal_ConsistenciaConItemTotal` | carrito.Total() == suma manual de item.Total() |
| `TestTotal_OrdenNoImporta` | Total es el mismo sin importar el orden de ítems |
| `TestContarItems_SliceVacio` | ContarItems vacío → 0 |
| `TestContarItems_NilSlice` | nil devuelve 0 |
| `TestContarItems_TablaVariantes` | 7 subtests: cantidad 1, cantidad 5, dos ítems suma, tres ítems, cantidad 0, todos cantidad 1, cantidad grande |
| `TestContarItems_SumaManual` | ContarItems() coincide con suma manual de cantidades |
| `TestContarItems_IndependienteDelPrecio` | El precio no afecta el conteo de unidades |

**Total tests: 22**

---

### `helpers_test.go` — Paquete `helpers`

Prueba todas las funciones utilitarias extraídas de `routes.go` al paquete `helpers`.

| Test | Qué verifica |
|------|-------------|
| `TestFormatCOP` | 12 subtests: cero, 900, 1000, 10000, 100000, 1M, precio típico, precio premium, decimal trunca, costo envío, dos separadores, un peso |
| `TestFormatCOP_PrefijoDolar` | Siempre empieza con "$" |
| `TestAdd` | 5 casos: 0+0, positivos, con negativo, conmutativity |
| `TestAdd_Conmutativity` | Add(3,7) == Add(7,3) |
| `TestSub` | 5 subtests: resta normal, exactamente 1, a==b devuelve 1, negativo devuelve 1, mínimo carrito |
| `TestSub_NuncaMenorQueUno` | Sub nunca devuelve valor < 1 para 4 entradas extremas |
| `TestFSub` | 4 subtests: normal, resultado negativo, resta cero, ambos iguales |
| `TestJoin` | 6 subtests: vacío, un elemento, tallas típicas, separador coma, separador vacío, separador guion |
| `TestContiene` | 7 subtests: primer elemento, medio, último, no existe, string vacío, case-sensitive lower, case-sensitive mixed |
| `TestContiene_SliceVacio` | Slice vacío siempre false |
| `TestContiene_SliceNil` | nil trata como vacío |
| `TestContieneInt` | 6 subtests: primer, medio, último, no existe, cero, negativo |
| `TestContieneInt_SliceVacio` | Slice vacío siempre false |
| `TestStrSlice` | 4 subtests: sin args, un arg, múltiples args, preserva orden |
| `TestToLower` | 5 casos: mayúsculas, todas mayúsculas, mixto, vacío, ya minúsculas |
| `TestToUpper` | 4 casos: minúsculas, ya mayúsculas, mixto, vacío |
| `TestFormatFecha_FormatoDD_MM_YYYY` | Salida exacta "14/05/2026 11:30" |
| `TestFormatFecha_DiaCeroRelleno` | Día y hora < 10 llevan cero inicial |
| `TestFormatFecha_LongitudFija` | Resultado siempre tiene 16 caracteres |
| `TestFormatFecha_Separadores` | Posiciones correctas de '/', ' ' y ':' |

**Total tests: 56**

---

### `auth_test.go` — Paquete `auth` (lógica pura)

Prueba la validación de campos obligatorios y el hashing bcrypt.

| Test | Qué verifica |
|------|-------------|
| `TestRegistrar_ValidacionCamposVacios` | 6 subtests: todos vacíos, nombre vacío, email vacío, password vacía, todos completos, solo nombre |
| `TestBcrypt_GeneraHash` | GenerateFromPassword no da error y el hash no está vacío |
| `TestBcrypt_HashDistintoDelOriginal` | El hash no es igual al texto plano |
| `TestBcrypt_VerificaContraseñaCorrecta` | CompareHashAndPassword devuelve nil para contraseña correcta |
| `TestBcrypt_RechazaContraseñaIncorrecta` | CompareHashAndPassword devuelve error para contraseña incorrecta |
| `TestBcrypt_HashesDistintosParaMismoInput` | Dos hashes del mismo input son distintos (salt aleatorio) |
| `TestBcrypt_AmbosHashesValidanMismaClave` | Ambos hashes (con distinto salt) validan la misma contraseña |
| `TestBcrypt_ContraseñaVacia` | bcrypt puede hashear cadena vacía (la validación la hace Registrar) |
| `TestBcrypt_MinCost` | GenerateFromPassword funciona con MinCost |

**Total tests: 14**

---

### `municipios_test.go` — Paquete `routes`

Prueba la integridad del mapa `municipiosColombia` usando los accesores exportados
`GetMunicipios()` y `GetDepartamentos()`.

| Test | Qué verifica |
|------|-------------|
| `TestMunicipios_DepartamentosTotales` | El mapa tiene exactamente 33 entradas |
| `TestMunicipios_DepartamentosExisten` | 33 subtests: cada departamento existe con el nombre exacto |
| `TestMunicipios_TodosTienenAlMenosUno` | 33 subtests: ningún departamento tiene slice vacío |
| `TestMunicipios_NingunoNil` | GetMunicipios no devuelve nil para ningún departamento registrado |
| `TestMunicipios_CiudadesClave` | 28 subtests: capitales y ciudades principales en sus departamentos correctos |
| `TestMunicipios_DepartamentoInexistente` | 5 subtests: nil para nombre incorrecto, vacío, minúsculas, incompleto |
| `TestMunicipios_DepartamentosGrandesConMasMunicipios` | 6 subtests: Antioquia ≥50, Boyacá ≥50, etc. |
| `TestMunicipios_NoHayDuplicados` | Sin municipios repetidos en Antioquia, Cundinamarca, Santander, Valle del Cauca |
| `TestMunicipios_StringsNoVacias` | Ningún municipio tiene nombre vacío en todos los departamentos |

**Total tests: ≈ 80+ subtests solo en municipios**

---

## Resumen de cobertura

| Archivo | Función/Área | Tests |
|---------|-------------|-------|
| `models_test.go` | `ItemCarrito.Total()`, struct fields | 21 |
| `carrito_test.go` | `carrito.Total()`, `carrito.ContarItems()` | 22 |
| `helpers_test.go` | FormatCOP, Add, Sub, FSub, Join, Contiene, ContieneInt, StrSlice, ToLower, ToUpper, FormatFecha | 56 |
| `auth_test.go` | Validación campos, bcrypt hash/verify | 14 |
| `municipios_test.go` | Existencia, integridad, ciudades clave, entradas inválidas | 80+ |
| **TOTAL** | | **~193 subtests** |

---

## Qué NO cubre esta carpeta (pruebas de integración pendientes)

| Área | Motivo de exclusión |
|------|-------------------|
| `auth.IniciarSesion` | Requiere `gin.Context` + DB activa |
| `auth.Registrar` (ruta feliz) | Requiere INSERT a PostgreSQL |
| `carrito.Agregar/Eliminar/CambiarCantidad` | Requieren `gin.Context` con sesión activa |
| `producto.Listar/BuscarPorID` etc. | Requieren DB activa (PostgreSQL) |
| Rutas HTTP (`/mujer`, `/admin`, etc.) | Requieren `httptest` + DB seed |
| Middlewares `RequiereLogin/RequiereAdmin` | Requieren `gin.Context` con sesión |

Para pruebas de integración se recomienda usar `net/http/httptest` con una
base de datos PostgreSQL de prueba (variable de entorno `DATABASE_URL_TEST`).
