window.addEventListener("hashchange", loadContent);

loadContent();

async function loadContent() {

  if (!document.cookie.includes('token')) {
    window.location.href = "login.html";
  }

  const mainContent = document.querySelector(".main");
  const route = window.location.hash.slice(1);

  const Routes = {
    "":["js/admin/defaultController.html"],
    "familia": ["js/admin/familiaController.html", "js/admin/familiaController.mjs"],
    "organizaciones": ["js/admin/organizacionesController.html", "js/admin/organizacionesController.mjs"],
    "productos": ["js/admin/productosController.html", "js/admin/productosController.mjs"],
    "noticias": ["js/admin/noticiasController.html", "js/admin/noticiasController.mjs"],
    "galeria": ["js/admin/galeriaController.html", "js/admin/galeriaController.mjs"],
    "ordenes": ["js/admin/ordenesController.html", "js/admin/ordenesController.mjs"],
  };

  try {
    if (Routes[route][0]) {
      const response = await fetch(Routes[route][0]);
  
      if (response.ok) {
        const htmlContent = await response.text();
        mainContent.innerHTML = htmlContent;
        await document.querySelector(".section-script")?.remove();
        if(route !== ""){
          const script = document.createElement("script");
          await script.setAttribute("src", Routes[route][1]);
          await script.setAttribute("type", "module");
          await script.classList.add("section-script");
          await document.body.appendChild(script);
        }
      } else {
        throw new Error(`Failed to load ${route} content. HTTP status: ${response.status}`);
      }
    } else {
      mainContent.innerHTML = "404 - Page Not Found";
    }
  } catch (error) {
    console.error("Error loading content:", error);
    mainContent.innerHTML = "Error loading content.";
  }
}