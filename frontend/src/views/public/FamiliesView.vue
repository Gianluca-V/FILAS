<script setup>
import { computed } from 'vue';

import { getFamily } from '../../api/resources.js';
import { useAsyncResource } from '../../composables/useAsyncResource.js';
import { resolveAssetPath } from '../../utils/assets.js';

const { data: family, status } = useAsyncResource(getFamily, { notFoundIsEmpty: true });

// Category enum from the spec: "Centro de dia" / "Taller protegido".
const dayCenterItems = computed(() => family.value.filter((item) => item.Category === 'Centro de dia'));
const workshopItems = computed(() => family.value.filter((item) => item.Category === 'Taller protegido'));
</script>

<template>
  <p v-if="status === 'loading'" class="family-status">Cargando información de familias...</p>
  <p v-else-if="status === 'error'" class="family-status">
    No pudimos cargar la información. Intentá nuevamente más tarde.
  </p>

  <section v-else class="family">
    <article class="family__article family__article--CD">
      <h2 class="family__title-container">Centro de Dia</h2>
      <p class="family__description-container">
        Nuestro Centro de Día es un dispositivo que trabaja en el fortalecimiento de las redes
        comunitarias/familiares, favoreciendo su autonomía y participación activa. Ofrecen actividades reflexivas,
        recreativas, culturales, corporales y cognitivas
      </p>
      <p class="family__description-container">
        Nuestro Centro de Día funciona de lunes a viernes, de 8 a 16hs, proponiendo pensión alimentaria a sus
        asistentes para desayuno, almuerzo y merienda. A lo largo de la jornada se llevan a cabo talleres de
        formación orientados a fortalecer habilidades sociales, estimular y desarrollar la creatividad, la expresión
        corporal y aptitudes relacionadas con el arte. En este espacio trabajan para contener a los concurrentes una
        planta compuesta por trabajadoras sociales, psicólogas, docentes especiales, cocineras, supervisores/as,
        talleristas, administrativos, entre otros.
        <ul class="family__nav-container">
          <h3 class="family__li">Ejes de trabajo:</h3>
          <li class="family__li">Área 1: Higiene y Salud personal.</li>
          <li class="family__li">Área 2: Cuidados personales.</li>
          <li class="family__li">Área 3: Habilidades domésticas.</li>
          <li class="family__li">Área 4: Prevención de riesgos en el hogar.</li>
          <li class="family__li">Área 5: Utilización de monedas y billetes y planificación de gastos e ingresos.</li>
          <li class="family__li">Área 6: Autonomía en la alimentación.</li>
          <li class="family__li">Área 7: Actividades comunitarias.</li>
          <li class="family__li">Área 8: Conocimiento y utilización del tiempo libre y de ocio.</li>
        </ul>
      </p>

      <p v-if="status === 'empty' || dayCenterItems.length === 0" class="family__empty">
        No hay contenido cargado para esta categoría todavía.
      </p>
      <div v-else class="family__cards-container family__cards-container--CD">
        <div v-for="item in dayCenterItems" :key="item.ID" class="family__card">
          <div v-if="item.Image" class="family__img">
            <img :src="resolveAssetPath(item.Image)" alt="Imagen de taller" loading="lazy" class="family__img">
          </div>
          <div class="family__body">{{ item.Body }}</div>
        </div>
      </div>
    </article>

    <article class="family__article family__article--TP">
      <h2 class="family__title-container">Talleres</h2>
      <p class="family__description-container">
        Nuestro T.P.P. es un espacio dedicado a brindar un servicio a la comunidad, creando espacios laborales
        incluir socio laboralmente a jóvenes y adultos entre 21 y 45 años de edad con discapacidad intelectual.
      </p>
      <p class="family__description-container">
        Nuestro Taller Protegido de Producción es un espacio dedicado a brindar un servicio a la comunidad, creando
        espacios laborales incluir socio laboralmente a personas entre 21 y 45 años de edad con discapacidad
        intelectual. La planta de operarios está integrada por jóvenes y adultos con discapacidad intelectual que
        presentan dificultades para insertarse en el sistema laboral formal. La jornada laboral de 8 horas de lunes a
        viernes está acompañada constantemente por supervisores y profesionales que velan por el bienestar de los
        operarios y el cumplimiento de sus metas en el ámbito laboral, acompañando sus necesidades y los desafíos
        que se presentan día a día. Para completar la jornada laboral, el taller comparte una relación estrecha con
        nuestro Centro de Día, compartiendo espacios de formación y recreación. Con esta propuesta diaria, el taller
        al momento mantiene 3 líneas de producción, ensamble, elaboración de dulces y encurtidos, y panadería,
        alternando en sus funciones a todos los operarios. Esta política de trabajo está orientada a evitar la
        mecanización de las conductas laborales y estimular mayor diversidad de habilidades motoras, cognoscitivas y
        sociales.
      </p>

      <p v-if="status === 'empty' || workshopItems.length === 0" class="family__empty">
        No hay contenido cargado para esta categoría todavía.
      </p>
      <div v-else class="family__cards-container family__cards-container--TP">
        <div v-for="item in workshopItems" :key="item.ID" class="family__card family__card--TP">
          <div v-if="item.Image" class="family__img-container family__img-container--TP">
            <img :src="resolveAssetPath(item.Image)" alt="Imagen de taller" loading="lazy" class="family__img--TP">
          </div>
        </div>
      </div>
    </article>
  </section>
</template>

<style scoped lang="scss">
.family-status {
  text-align: center;
  padding-block: 3rem;
  font-size: 1.125rem;
}

.family {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
  padding: 20px;
  margin: 0 auto;
}

.family__li {
  margin-bottom: 10px;
}

.family__nav-container {
  padding: 20px;
}

.family__description-container {
  text-align: center;
  margin: 20px 0;
}

.family__title-container {
  text-align: center;
  font-size: clamp(2rem, 4vw, 3.5rem);
  color: $color-background;
  background-color: $color-primary;
}

.family__empty {
  text-align: center;
  padding-block: 1rem;
  color: rgba(0, 0, 0, 0.6);
}

.family__cards-container {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.family__card {
  display: flex;
  flex-direction: column;
  background-color: $color-secondary;
  text-align: center;
  padding: 10px;
}

.family__card--TP {
  max-width: 400px;
  height: 200px;
  margin: auto;
  width: 100%;
}

.family__body {
  padding-inline: 16px;
  color: $color-background;
  font-weight: bold;
  max-height: 200px;
  overflow-y: auto;
}

.family__img {
  max-width: 300px;
  width: 100%;
  padding: 8px;
  margin-inline: auto;
}

.family__img-container--TP {
  height: 100%;
  width: 100%;
}

.family__img--TP {
  max-width: 100%;
  width: 100%;
  height: 100%;
}

@media (width < 800px) {
  .family__cards-container {
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
}

@media (width < 500px) {
  .family__cards-container {
    grid-template-columns: repeat(1, 1fr);
    gap: 20px;
  }
}
</style>
