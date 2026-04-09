// togglePass: alterna visibilidad de contraseña via data-toggle-pass
function togglePass(inputId) {
    var input = document.getElementById(inputId);
    var eye   = document.getElementById("eye_" + inputId);
    if (!input) return;
    if (input.type === "password") {
        input.type = "text";
        if (eye) { eye.classList.replace("bi-eye", "bi-eye-slash"); }
    } else {
        input.type = "password";
        if (eye) { eye.classList.replace("bi-eye-slash", "bi-eye"); }
    }
}

document.addEventListener("DOMContentLoaded", function () {

    // Registra los botones de ojo
    document.querySelectorAll("[data-toggle-pass]").forEach(function (btn) {
        btn.addEventListener("click", function () {
            togglePass(btn.getAttribute("data-toggle-pass"));
        });
    });

    // Cambia entre panel login y panel registro usando delegación en document
    // data-panel="registro" muestra el registro, data-panel="login" vuelve al login
    document.addEventListener("click", function (e) {
        var link = e.target.closest("[data-panel]");
        if (!link) return;
        e.preventDefault();
        e.stopPropagation();
        var destino = link.getAttribute("data-panel");
        var titulo  = document.getElementById("sidebarTitle");

        if (destino === "registro") {
            document.getElementById("panelLogin").style.display    = "none";
            document.getElementById("panelRegistro").style.display = "block";
            if (titulo) titulo.textContent = "CREAR CUENTA";
        } else {
            document.getElementById("panelRegistro").style.display = "none";
            document.getElementById("panelLogin").style.display    = "block";
            if (titulo) titulo.textContent = "INICIAR SESIÓN";
        }
    });

    // Resetea el sidebar a login cada vez que se cierra.
    // Usamos null guards porque cuando el usuario está autenticado
    // el sidebar muestra el perfil y los paneles login/registro no existen en el DOM.
    var sidebar = document.getElementById("loginSidebar");
    if (sidebar) {
        sidebar.addEventListener("hidden.bs.offcanvas", function () {
            var pLogin    = document.getElementById("panelLogin");
            var pRegistro = document.getElementById("panelRegistro");
            var titulo    = document.getElementById("sidebarTitle");
            if (pLogin)    pLogin.style.display    = "block";
            if (pRegistro) pRegistro.style.display = "none";
            if (titulo)    titulo.textContent       = "INICIAR SESIÓN";
        });
    }

});