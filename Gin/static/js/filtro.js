// ============================================================
// filtro.js — filtrado de productos por subcategoría
// Sin recargar la página; funciona con cualquier cantidad de
// subcategorías que el backend pase dinámicamente al template.
// ============================================================

document.addEventListener("DOMContentLoaded", function () {

    var filtroItems   = document.querySelectorAll(".filtro-item");
    var productoCards = document.querySelectorAll(".producto-card");
    var labelActivo   = document.getElementById("filtroActivo");
    var sinResultados = document.getElementById("sinResultados");

    if (!filtroItems.length) return; // No estamos en una página con filtros

    filtroItems.forEach(function (enlace) {
        enlace.addEventListener("click", function (e) {
            e.preventDefault();

            var filtro = enlace.getAttribute("data-filtro"); // "todas" o nombre de subcategoría

            // Marcar el item activo en el dropdown
            filtroItems.forEach(function (el) { el.classList.remove("active"); });
            enlace.classList.add("active");

            // Actualizar etiqueta visible
            if (labelActivo) {
                labelActivo.textContent = filtro === "todas" ? "Todos los productos" : filtro;
            }

            // Mostrar / ocultar tarjetas
            var visibles = 0;
            productoCards.forEach(function (card) {
                var sub = card.getAttribute("data-subcategoria");
                var mostrar = filtro === "todas" || sub === filtro;
                card.style.display = mostrar ? "" : "none";
                if (mostrar) visibles++;
            });

            // Mensaje si no hay resultados
            if (sinResultados) {
                sinResultados.style.display = visibles === 0 ? "block" : "none";
            }
        });
    });
});
