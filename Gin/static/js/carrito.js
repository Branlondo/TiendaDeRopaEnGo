// ============================================================
// carrito.js — gestión del carrito con localStorage
// ============================================================

const CART_KEY = "rff_carrito";

// ── Helpers de persistencia ──────────────────────────────────
function getCarrito() {
    try {
        return JSON.parse(localStorage.getItem(CART_KEY)) || [];
    } catch (e) {
        return [];
    }
}

function saveCarrito(carrito) {
    localStorage.setItem(CART_KEY, JSON.stringify(carrito));
}

// Clave única por producto + talla
function itemKey(id, talla) {
    return id + "_" + talla;
}

// ── Agregar o incrementar cantidad ──────────────────────────
function agregarAlCarrito(producto) {
    var carrito = getCarrito();
    var key = itemKey(producto.id, producto.talla);
    var existente = carrito.find(function (i) { return itemKey(i.id, i.talla) === key; });
    if (existente) {
        existente.cantidad += producto.cantidad;
    } else {
        carrito.push(producto);
    }
    saveCarrito(carrito);
    actualizarBadge();
}

// ── Eliminar item ────────────────────────────────────────────
function eliminarDelCarrito(id, talla) {
    var carrito = getCarrito().filter(function (i) {
        return itemKey(i.id, i.talla) !== itemKey(id, talla);
    });
    saveCarrito(carrito);
    renderCarrito();
    actualizarBadge();
}

// ── Cambiar cantidad desde el carrito ───────────────────────
function cambiarCantidad(id, talla, nuevaCantidad) {
    if (nuevaCantidad < 1) return;
    var carrito = getCarrito();
    var item = carrito.find(function (i) { return itemKey(i.id, i.talla) === itemKey(id, talla); });
    if (item) {
        item.cantidad = nuevaCantidad;
        saveCarrito(carrito);
        renderCarrito();
        actualizarBadge();
    }
}

// ── Badge del navbar ─────────────────────────────────────────
function actualizarBadge() {
    var carrito = getCarrito();
    var total = carrito.reduce(function (acc, i) { return acc + i.cantidad; }, 0);
    document.querySelectorAll("#cartCount").forEach(function (el) {
        el.textContent = total;
    });
}

// ── Renderizar carrito (página /carrito) ─────────────────────
function renderCarrito() {
    var carritoItems = document.getElementById("carritoItems");
    var carritoVacio = document.getElementById("carritoVacio");
    var carritoResumen = document.getElementById("carritoResumen");
    if (!carritoItems) return; // No estamos en la página del carrito

    var carrito = getCarrito();

    if (carrito.length === 0) {
        carritoVacio.style.display = "block";
        carritoItems.innerHTML = "";
        carritoResumen.style.display = "none";
        return;
    }

    carritoVacio.style.display = "none";
    carritoResumen.style.display = "block";

    var html = "";
    carrito.forEach(function (item) {
        var subtotalItem = (item.precio * item.cantidad).toFixed(2);
        html += '<div class="card mb-3">' +
            '<div class="card-body">' +
            '<div class="row align-items-center gx-4">' +

            // Imagen
            '<div class="col-md-2 text-center mb-3 mb-md-0">' +
            '<img src="' + item.imagen + '" class="img-fluid rounded" style="max-height:100px;object-fit:cover;" alt="' + item.nombre + '" />' +
            '</div>' +

            // Info
            '<div class="col-md-4 mb-3 mb-md-0">' +
            '<h5 class="fw-bolder mb-1">' + item.nombre + '</h5>' +
            '<span class="badge bg-secondary me-2">Talla: ' + item.talla + '</span>' +
            '<p class="text-muted small mb-1 mt-1">' + item.descripcion + '</p>' +
            '<span class="fw-bold text-dark">$' + item.precio.toFixed(2) + ' c/u</span>' +
            '</div>' +

            // Cantidad
            '<div class="col-md-3 d-flex align-items-center justify-content-center mb-3 mb-md-0">' +
            '<button class="btn btn-outline-dark btn-sm px-2" onclick="cambiarCantidad(' + item.id + ', \'' + item.talla + '\', ' + (item.cantidad - 1) + ')">' +
            '<i class="bi bi-dash"></i></button>' +
            '<input class="form-control text-center mx-2" type="number" value="' + item.cantidad + '" min="1" style="max-width:3.5rem;" ' +
            'onchange="cambiarCantidad(' + item.id + ', \'' + item.talla + '\', parseInt(this.value))" />' +
            '<button class="btn btn-outline-dark btn-sm px-2" onclick="cambiarCantidad(' + item.id + ', \'' + item.talla + '\', ' + (item.cantidad + 1) + ')">' +
            '<i class="bi bi-plus"></i></button>' +
            '</div>' +

            // Subtotal + eliminar
            '<div class="col-md-3 text-end">' +
            '<p class="fw-bolder fs-5 mb-1">$' + subtotalItem + '</p>' +
            '<button class="btn btn-outline-danger btn-sm" onclick="eliminarDelCarrito(' + item.id + ', \'' + item.talla + '\')">' +
            '<i class="bi-trash me-1"></i>Eliminar</button>' +
            '</div>' +

            '</div></div></div>';
    });

    carritoItems.innerHTML = html;

    // Calcular total
    var total = carrito.reduce(function (acc, i) { return acc + i.precio * i.cantidad; }, 0);
    document.getElementById("subtotalValor").textContent = "$" + total.toFixed(2);
    document.getElementById("totalValor").textContent = "$" + total.toFixed(2);
}

// ── Inicialización ───────────────────────────────────────────
document.addEventListener("DOMContentLoaded", function () {

    // Actualizar badge en cualquier página
    actualizarBadge();

    // Renderizar carrito si estamos en /carrito
    renderCarrito();

    // ── Lógica de la página de producto ──
    var tallaBtns = document.querySelectorAll(".talla-btn");
    var tallaSeleccionada = null;

    tallaBtns.forEach(function (btn) {
        btn.addEventListener("click", function () {
            tallaBtns.forEach(function (b) {
                b.classList.remove("btn-dark");
                b.classList.add("btn-outline-dark");
            });
            btn.classList.remove("btn-outline-dark");
            btn.classList.add("btn-dark");
            tallaSeleccionada = btn.getAttribute("data-talla");
            var err = document.getElementById("tallaError");
            if (err) err.style.display = "none";
        });
    });

    var btnAgregar = document.getElementById("btnAgregarCarrito");
    if (btnAgregar) {
        // Si solo hay una talla, seleccionarla automáticamente
        if (tallaBtns.length === 1) {
            tallaBtns[0].click();
        }

        btnAgregar.addEventListener("click", function () {
            var err = document.getElementById("tallaError");

            if (!tallaSeleccionada) {
                if (err) err.style.display = "block";
                return;
            }

            var cantidad = parseInt(document.getElementById("inputQuantity").value) || 1;
            if (cantidad < 1) cantidad = 1;

            var producto = {
                id:          parseInt(btnAgregar.getAttribute("data-id")),
                nombre:      btnAgregar.getAttribute("data-nombre"),
                precio:      parseFloat(btnAgregar.getAttribute("data-precio")),
                descripcion: btnAgregar.getAttribute("data-descripcion"),
                imagen:      btnAgregar.getAttribute("data-imagen"),
                talla:       tallaSeleccionada,
                cantidad:    cantidad
            };

            agregarAlCarrito(producto);

            // Mostrar confirmación
            var toast = document.getElementById("toastCarrito");
            if (toast) {
                toast.style.display = "block";
                setTimeout(function () { toast.style.display = "none"; }, 3000);
            }
        });
    }
});
