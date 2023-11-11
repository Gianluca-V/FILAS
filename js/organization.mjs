import { API } from './API.mjs';

const slider = document.querySelector(".slider");
let slides = document.querySelectorAll(".slide");
let cantidaDeSlides = slides.length;
const slideIconos = document.querySelectorAll(".slide-icon");
var slideNumero = 0;

window.addEventListener("scroll", reveal);

API.Organizations.GetAll().then((data) => {
  populateSlider(data);
});

const populateSlider = async (data) => {
  console.log(data)
  const sliderContainer = document.querySelector(".slider");
  await data.forEach((slide) => {
    sliderContainer.innerHTML += `
    <div class="slide">
      <div class="img-abogado" style="background-image: url('${slide.Image}');">

      </div>
      <div class="info">
        <h2>${slide.Title}</h2>
        <p id="p-equipo5">
          ${slide.Description}
        </p>
      </div>

    </div>
    `;
  });


  sliderContainer.innerHTML += `
  <div class="navigation">
    <b class="fas fa-chevron-left prev-btn"><</b>
        <b class="fas fa-chevron-right next-btn">></b>
  </div>
  `;
  slides = document.querySelectorAll(".slide");
  cantidaDeSlides = slides.length;
  slides.item(0).classList.add("active");

  const botonNext = document.querySelector(".next-btn");
  const botonPrev = document.querySelector(".prev-btn");

  //Boton de siguiente imagen del slider
  botonNext.addEventListener("click", () => {
    slides.forEach((slide) => {
      slide.classList.remove("active");
    });
    slideIconos.forEach((slideIcon) => {
      slideIcon.classList.remove("active");
    });

    slideNumero++;

    if (slideNumero > (cantidaDeSlides - 1)) {
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

    if (slideNumero < 0) {
      slideNumero = cantidaDeSlides - 1;
    }

    slides[slideNumero].classList.add("active");
    slideIconos[slideNumero].classList.add("active");
  });

  await repetidor();

  //Desactiva la reproduccion automatica en mouseover
  slider.addEventListener("mouseover", () => {
    clearInterval(playSlider);
  });

  //Activa la reproduccion automatica en mouseout
  slider.addEventListener("mouseout", () => {
    repetidor();
  });
}

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

//Reprduccion automatica del slider
var playSlider;

var repetidor = () => {
  playSlider = setInterval(function () {
    slides.forEach((slide) => {
      slide.classList.remove("active");
    });
    slideIconos.forEach((slideIcon) => {
      slideIcon.classList.remove("active");
    });

    slideNumero++;

    if (slideNumero > (cantidaDeSlides - 1)) {
      slideNumero = 0;
    }

    slides[slideNumero].classList.add("active");
    //slideIconos[slideNumero].classList.add("active");
  }, 10000);
}



