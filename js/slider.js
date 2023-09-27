const slider = document.querySelector(".slider");
const botonNext = document.querySelector(".next-btn");
const botonPrev = document.querySelector(".prev-btn");
const slides = document.querySelectorAll(".slide");
const slideIconos = document.querySelectorAll(".slide-icon");
const cantidaDeSlides = slides.length;
var slideNumero = 0;

window.addEventListener("scroll", reveal);

function reveal() {
  //reveals es igual a la cantidad de divs con "reveal" en el nombre
    var reveals = document.querySelectorAll(".reveal");
    for (var i = 0; i < reveals.length; i++) {
      //variable que guarda la altura de la pantalla
      var windowHeight = window.innerHeight;
      //el metodo getBoundingClientRect() devuelve el tamaño de un elemento y su posición relativa a la ventana gráfica.
      var elementTop = reveals[i].getBoundingClientRect().top;
      var elementVisible = 150;
  
      // comprueba si los elementos estan en pantalla para mostrarlos o no segun corresponda
      if (elementTop < windowHeight - elementVisible) {
        reveals[i].classList.add("active");
      } else {
        reveals[i].classList.remove("active");
      }
    }
  }
  
   //Boton de siguiente imagen del slider
   botonNext.addEventListener("click", () => {
    slides.forEach((slide) => {
      slide.classList.remove("active");
    });
    slideIconos.forEach((slideIcon) => {
      slideIcon.classList.remove("active");
    });

    slideNumero++;

    if(slideNumero > (cantidaDeSlides - 1)){
      slideNumero = 0;
    }

    slides[slideNumero].classList.add("active");
    slideIconos[slideNumero].classList.add("active");
  });

  //Boton imagen previa del slider
  botonPrev.addEventListener("click", () => {
    slides.forEach((slide) => {
      slide.classList.remove("active");
    });
    slideIconos.forEach((slideIcon) => {
      slideIcon.classList.remove("active");
    });

    slideNumero--;

    if(slideNumero < 0){
      slideNumero = cantidaDeSlides - 1;
    }

    slides[slideNumero].classList.add("active");
    slideIconos[slideNumero].classList.add("active");
  });

  //Reprduccion automatica del slider
  var playSlider;

  var repetidor = () => {
    playSlider = setInterval(function(){
      slides.forEach((slide) => {
        slide.classList.remove("active");
      });
      slideIconos.forEach((slideIcon) => {
        slideIcon.classList.remove("active");
      });

      slideNumero++;

      if(slideNumero > (cantidaDeSlides - 1)){
        slideNumero = 0;
      }

      slides[slideNumero].classList.add("active");
      slideIconos[slideNumero].classList.add("active");
    }, 10000);
  }
  repetidor();

  //Desactiva la reproduccion automatica en mouseover
  slider.addEventListener("mouseover", () => {
    clearInterval(playSlider);
  });

  //Activa la reproduccion automatica en mouseout
  slider.addEventListener("mouseout", () => {
    repetidor();
  });
    
