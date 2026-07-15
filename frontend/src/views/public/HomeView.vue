<script setup>
// Content/copy ported verbatim from index.html (design §3 — no redesign).
</script>

<template>
  <section class="main">
    <img class="main__img" src="/assets/filas-main.png" alt="">
    <div class="main__textbox">
      <h2 class="main__title">¿Quiénes somos?</h2>
      <router-link to="/informacion" class="main__description">
        Fundación para la Integración Laboral y Autonomía Social de Personas con Necesidades Especiales. Propender a
        la integración socio laboral de las personas con necesidades especiales a fin de que participen en forma
        activa, contribuyendo a la vida económica de la comunidad a la cual pertenecen.
      </router-link>
    </div>
  </section>

  <section class="sections">
    <h2 class="sections__title">SECCIONES</h2>
    <div class="sections__container">
      <router-link to="/familias" class="sections__section sections__section--green">
        <img src="/assets/family.svg" alt="link a la sección familia" class="sections__img">
        <h3 class="sections__name">FAMILIAS (Centro de Día / Taller Protegido)</h3>
      </router-link>
      <router-link to="/organizaciones" class="sections__section">
        <img src="/assets/office.svg" alt="link a la sección organizaciones" class="sections__img">
        <h3 class="sections__name">Instituciones aliadas</h3>
      </router-link>
      <router-link to="/wilde-artesanal" class="sections__section sections__section--brown">
        <img src="/assets/cart.svg" alt="link a la sección Wilde Artesanal" class="sections__img">
        <h3 class="sections__name">PRODUCTOS Wilde Artesanal</h3>
      </router-link>
      <router-link to="/noticias" class="sections__section">
        <img src="/assets/bullhorn.svg" alt="link a la sección noticias" class="sections__img">
        <h3 class="sections__name">NOTICIAS</h3>
      </router-link>
    </div>
  </section>

  <section class="history">
    <h2 class="history__title">HISTORIA</h2>
    <p class="history__p">
      Fundación para la integración laboral y autonomía social de personas con necesidades especiales.
      <br><br>
      Nuestra fundación funciona desde el 98, fechas desde la cual se impulsan proyectos de integración laboral, se
      desarrolla un centro de día con un cupo aproximado de 15 asistentes, y se sostiene un taller protegido con más
      de 40 operarios trabajando.
      <br><br>
      Entre las misiones y funciones de nuestra institución se encuentra la de propender a la integración socio
      laboral de personas con discapacidad, a fin de que alcancen una autonomía social tal que puedan integrarse de
      forma activa en los espacios comunitarios de pertenencia.
    </p>
  </section>
</template>

<style scoped lang="scss">
.main {
  color: $color-background;
  width: 100%;
  height: 50vh;
  background: $color-primary;
  @include flex-row(center, center);
}

.main__img {
  height: 100%;
  margin-right: 80px;
}

.main__textbox {
  width: 50%;
  border-left: 2px solid $color-background;
  height: max-content;
  padding-left: 20px;
  padding-right: 30px;
}

.main__title {
  margin-bottom: 10px;
  font-size: 2rem;
}

.main__description {
  display: block;
  font-size: clamp(0.9rem, 1.55vw, 1.125rem);

  &:hover {
    background-color: rgba(0, 0, 0, 0.2);
  }
}

@include below-tablet {
  .main {
    @include flex-column(center, center);
    height: auto;
  }

  .main__img {
    width: 60%;
    margin-right: 0;
  }

  .main__textbox {
    width: 90%;
    margin-bottom: 40px;
    border: 2px solid $color-background;
    padding: 20px;
  }
}

@include below-mobile {
  .main__img {
    width: 80%;
    margin-right: 0;
  }
}

.sections {
  color: $color-background;
  height: 300px;
  border: 6px solid $color-primary;
  border-radius: 5px;
  margin: 30px;
  @include flex-column(center, center);
  background-color: $color-background;
}

.sections__title {
  color: $color-secondary;
  font-size: 2rem;
  margin-top: 14px;
}

.sections__container {
  display: flex;
  width: 100%;
  height: 100%;
  padding: 7px;
}

.sections__section {
  width: 25%;
  margin: 7px;
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: 0.8fr 1fr;
  justify-content: center;
  align-items: center;
  background-color: $color-primary;
  transition: 0.2s;

  &:hover {
    transform: scale(1.05);
    box-shadow: 0 0 4px black;
  }
}

.sections__section--green {
  background-color: $color-accent;
}

.sections__section--brown {
  background-color: $color-secondary;
}

.sections__img {
  height: 45px;
  margin: auto;
  filter: grayscale(100%) saturate(0%) sepia(100%);
}

.sections__name {
  font-size: clamp(1rem, 2vw, 1.5rem);
  padding-inline: 16px;
  padding-bottom: 5px;
  text-wrap: balance;
  text-transform: uppercase;
  text-align: center;
}

@media (max-width: 768px) {
  .sections {
    border-radius: 0;
    padding: 0;
    margin: 0;
    height: auto;
    padding-bottom: 20px;
  }

  .sections__container {
    flex-direction: column;
  }

  .sections__title {
    padding: 0;
    margin: 0;
    margin-top: 40px;
    margin-bottom: 20px;
  }

  .sections__section {
    width: auto;
    padding-block: 1rem;
    margin: 5px;

    &:hover {
      transform: scale(1);
      height: 120px;
    }
  }
}

.history {
  margin-block: 2rem;
  @include flex-column(center, center);
}

.history__title {
  text-align: center;
  font-size: 2rem;
  color: $color-secondary;
}

.history__p {
  max-width: 1600px;
  padding-inline: 2rem;
  text-align: center;
  text-wrap: balance;
  font-size: 1.125rem;
}
</style>
