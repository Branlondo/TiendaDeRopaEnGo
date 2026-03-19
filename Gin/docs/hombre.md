<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
        <meta name="description" content="" />
        <meta name="author" content="" />
        <title>Ropa Fash Fashion - RFF</title>
        <!-- Favicon-->
        <link rel="icon" type="image/x-icon" href="/static/assets/Unilogo.ico" />
        <!-- Bootstrap icons-->
        <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.5.0/font/bootstrap-icons.css" rel="stylesheet" />
        <!-- Core theme CSS (includes Bootstrap)-->
        <link href="/static/css/styles.css" rel="stylesheet" />
    </head>
    <body>
        <!-- Navigation-->
        <nav class="navbar navbar-expand-lg navbar-light bg-light">
            <div class="container px-4 px-lg-5">
                <!-- Botón con ícono + texto -->
                <a class="btn btn-light d-flex align-items-center" href="index.html">
                    <img src="/static/assets/Unilogo.ico" alt="Logo" 
                        style="height:24px; width:auto; margin-right:8px;">
                    RFF
                </a>
                <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation"><span class="navbar-toggler-icon"></span></button>
                <div class="collapse navbar-collapse" id="navbarSupportedContent">
                    <ul class="navbar-nav me-auto mb-2 mb-lg-0 ms-lg-4">
                        <li class="nav-item"><a class="nav-link" href="/mujer">MUJER</a></li>
                        <li class="nav-item"><a class="nav-link" href="/hombre">HOMBRE</a></li>
                        <li class="nav-item dropdown">
                            <a class="nav-link dropdown-toggle" id="navbarDropdown" href="#" role="button" data-bs-toggle="dropdown" aria-expanded="false">Categorias</a>
                            <ul class="dropdown-menu" aria-labelledby="navbarDropdown">
                                <li><a class="dropdown-item" href="#!">All Products</a></li>
                                <li><hr class="dropdown-divider" /></li>
                                <li><a class="dropdown-item" href="#!">Camisas</a></li>
                                <li><a class="dropdown-item" href="#!">Jeans</a></li>
                                <li><a class="dropdown-item" href="#!">Pantalones</a></li>
                            </ul>
                        </li>
                    </ul>
                    <form class="d-flex gap me-2">
                        <button class="btn btn-outline-dark" type="button" onclick="window.location.href='/Carrito.html'">
                            <i class="bi-cart-fill me-1"></i>
                            Cart
                            <span class="badge bg-dark text-white ms-1 rounded-pill">0</span>
                        </button>
                    </form>
                    <form class="d-flex gap-2">
                    <button class="btn btn-outline-dark" type="button" 
                            data-bs-toggle="offcanvas" 
                            data-bs-target="#loginSidebar">
                        <i class="bi-person me-1"></i>
                        Iniciar sesión
                    </button>
                    </form>
                </div>
            </div>
        </nav>
        <!-- Header-->
        <header class="bg-dark py-5">
            <div class="container px-4 px-lg-5 my-5">
                <div class="text-center text-white">
                    <h1 class="display-4 fw-bolder">Ropa Flash Fashion</h1>
                    <p class="lead fw-normal text-white-50 mb-0">Ropa exclusivamente para hombres</p>
                </div>
            </div>
        </header>
        <!-- Section-->
        <section class="py-5">
            <div class="container px-4 px-lg-5 mt-5">
                <div class="row gx-4 gx-lg-5 row-cols-2 row-cols-md-3 row-cols-xl-4 justify-content-center">
                    {{ range .productos }}
                    <div class="col mb-5">
                        <div class="card h-100">
                            <!-- imagen del producto -->
                            <img class="card-img-top" src="{{ .Imagen }}" alt="{{ .Nombre }}" />
                            <!-- detalle -->
                            <div class="card-body p-4">
                                <div class="text-center">
                                    <h5 class="fw-bolder">{{ .Nombre }}</h5>
                                    ${{ .Precio }}
                                </div>
                            </div>
                            <!-- acciones -->
                            <div class="card-footer p-4 pt-0 border-top-0 bg-transparent">
                                <div class="text-center">
                                    <a class="btn btn-outline-dark mt-auto" href="/producto/{{ .ID }}">
                                        Ver detalle
                                    </a>
                                </div>
                            </div>
                        </div>
                    </div>
                    {{ end }}
                </div>
            </div>
        <!-- Sidebar login -->
        <div class="offcanvas offcanvas-end" tabindex="-1" id="loginSidebar">
            <div class="offcanvas-header border-bottom">
                <h5 class="offcanvas-title fw-bold">INICIAR SESIÓN</h5>
                <button type="button" class="btn-close" data-bs-dismiss="offcanvas"></button>
            </div>
        </div>
        </section>
        <!-- Footer-->
        <footer class="py-5 bg-dark">
            <div class="container"><p class="m-0 text-center text-white">Copyright &copy; Your Website 2023</p></div>
        </footer>
        
        <!-- ── Sidebar de login ──
             IMPORTANTE: debe estar fuera de cualquier section/div de contenido
             y antes del cierre de </body> para que Bootstrap lo gestione bien. -->
        <div class="offcanvas offcanvas-end" tabindex="-1" id="loginSidebar">
            <div class="offcanvas-header border-bottom">
                <h5 class="offcanvas-title fw-bold">INICIAR SESIÓN</h5>
                <button type="button" class="btn-close" data-bs-dismiss="offcanvas"></button>
            </div>
            <div class="offcanvas-body d-flex flex-column gap-3 pt-4">
                <input type="email" class="form-control" placeholder="E-MAIL">
                <!-- El botón del ojo llama a togglePass con el id del input.
                     El ícono tiene id "eye_passLogin" para que togglePass pueda cambiarlo. -->
                <div class="input-group">
                    <input type="password" id="passLogin" class="form-control" placeholder="CONTRASEÑA">
                    <button class="btn btn-outline-secondary" type="button"
                            onclick="togglePass('passLogin')">
                        <i class="bi-eye" id="eye_passLogin"></i>
                    </button>
                </div>
                <button class="btn btn-dark w-100 py-2 fw-bold" type="button">
                    INICIAR SESIÓN
                </button>
                <hr>
                <p class="text-center mb-0">
                    ¿No tienes cuenta?
                    <a href="/registro" class="fw-bold text-dark">REGÍSTRATE</a>
                </p>
            </div>
        </div>
        <!-- Bootstrap core JS-->
        <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/js/bootstrap.bundle.min.js"></script>
        <!-- Core theme JS-->
        <script src="/static/js/scripts.js"></script>
    </body>
</html>
