 window.addEventListener("hashchange", loadContent);

 loadContent();

 async function loadContent() {
   const mainContent = document.querySelector(".main");
   const route = window.location.hash.slice(1);

   const HTMLRoutes = {
     "familia": "js/admin/familiaController.html",
     "organizaciones": "js/admin/organizacionesController.html",
     "productos": "js/admin/productosController.html",
     "noticias": "js/admin/noticiasController.html",
     "galeria": "js/admin/galeriaController.html",
   };
   const JSRoutes = {
    "familia": "js/admin/familiaController.mjs",
    "organizaciones": "js/admin/organizacionesController.mjs",
    "productos": "js/admin/productosController.mjs",
    "noticias": "js/admin/noticiasController.mjs",
    "galeria": "js/admin/galeriaController.mjs",
  };

   try {
     if (HTMLRoutes[route]) {
       const response = await fetch(HTMLRoutes[route]);
       if (response.ok) {
         const htmlContent = await response.text();
         mainContent.innerHTML = htmlContent;
         const script = document.createElement("script");
         script.setAttribute("src",JSRoutes[route]);
         script.setAttribute("type","module");
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