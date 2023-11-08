window.addEventListener("hashchange", loadContent);

loadContent();

async function loadContent() {

  if (!document.cookie.includes('token')) {
    window.location.href = "login.html";
  }

  const mainContent = document.querySelector(".main");
  const route = window.location.hash.slice(1);

  const Routes = {
    "familia": ["js/admin/familiaController.html", "js/admin/familiaController.mjs"],
    "organizaciones": ["js/admin/organizacionesController.html", "js/admin/organizacionesController.mjs"],
    "productos": ["js/admin/productosController.html", "js/admin/productosController.mjs"],
    "noticias": ["js/admin/noticiasController.html", "js/admin/noticiasController.mjs"],
    "galeria": ["js/admin/galeriaController.html", "js/admin/galeriaController.mjs"],
  };

  try {
    if (Routes[route][0]) {
      const response = await fetch(Routes[route][0]);
      if (response.ok) {
        const htmlContent = await response.text();
        mainContent.innerHTML = htmlContent;
        document.querySelector(".section-script")?.remove();
        const script = document.createElement("script");
        script.setAttribute("src", Routes[route][1]);
        script.setAttribute("type", "module");
        script.classList.add("section-script");
        document.body.appendChild(script);
      } else {
        throw new Error(`Failed to load ${route} content.`);
      }
    } else {
      mainContent.innerHTML = "404 - Page Not Found";
    }
  } catch (error) {
    console.error(error);
    mainContent.innerHTML = "Error loading content.";
  }
}