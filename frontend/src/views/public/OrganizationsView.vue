<script setup>
import { onMounted, ref } from 'vue';

import { getOrganizations } from '../../api/resources.js';
import OrganizationSlider from '../../components/OrganizationSlider.vue';
import { isNotFound } from '../../utils/httpErrors.js';

const status = ref('loading');
const organizations = ref([]);

onMounted(async () => {
  try {
    organizations.value = await getOrganizations();
    status.value = 'ready';
  } catch (error) {
    if (isNotFound(error)) {
      status.value = 'empty';
    } else {
      status.value = 'error';
    }
  }
});
</script>

<template>
  <section id="team_box">
    <h2 id="team_box-team-title">Instituciones Aliadas</h2>

    <p v-if="status === 'loading'" class="team-status">Cargando instituciones aliadas...</p>
    <p v-else-if="status === 'error'" class="team-status">
      No pudimos cargar las instituciones aliadas. Intentá nuevamente más tarde.
    </p>
    <p v-else-if="status === 'empty'" class="team-status">Todavía no hay instituciones aliadas cargadas.</p>
    <OrganizationSlider v-else :organizations="organizations" />
  </section>

  <section id="contact_box">
    <h2 id="contact_box-contact-title">Contacto</h2>
    <section id="contact_box-main">
      <div id="contact_box-info">
        <div id="contact_box-info-num">
          <svg xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 512 512">
            <path
              d="M164.9 24.6c-7.7-18.6-28-28.5-47.4-23.2l-88 24C12.1 30.2 0 46 0 64C0 311.4 200.6 512 448 512c18 0 33.8-12.1 38.6-29.5l24-88c5.3-19.4-4.6-39.7-23.2-47.4l-96-40c-16.3-6.8-35.2-2.1-46.3 11.6L304.7 368C234.3 334.7 177.3 277.7 144 207.3L193.3 167c13.7-11.2 18.4-30 11.6-46.3l-40-96z" />
          </svg>
          <a href="tel:+541143532255" target="_blank" rel="noopener">
            <h2>11 4353-2255</h2>
          </a>
        </div>
        <div id="contact_box-info-mail">
          <svg xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 512 512">
            <path
              d="M48 64C21.5 64 0 85.5 0 112c0 15.1 7.1 29.3 19.2 38.4L236.8 313.6c11.4 8.5 27 8.5 38.4 0L492.8 150.4c12.1-9.1 19.2-23.3 19.2-38.4c0-26.5-21.5-48-48-48H48zM0 176V384c0 35.3 28.7 64 64 64H448c35.3 0 64-28.7 64-64V176L294.4 339.2c-22.8 17.1-54 17.1-76.8 0L0 176z" />
          </svg>
          <a href="mailto:fundacionfilas@gmail.com" target="_blank" rel="noopener">
            <h2>fundacionfilas@gmail.com</h2>
          </a>
        </div>
      </div>
      <form
        target="_blank"
        action="https://formsubmit.co/fundacionfilas@gmail.com"
        method="post"
        id="contact_box-form"
      >
        <h2 id="form-t">Enviar mensaje</h2>

        <label for="email" id="form-correo">E-mail:</label>
        <input
          type="text"
          id="email"
          name="email"
          class="input input--email"
          required
          placeholder="ingrese su mail"
        >

        <label for="message" id="form-mensaje">Mensaje:</label>
        <textarea id="message" name="message" class="input input--text" required placeholder="Mensaje..." />

        <input id="form-boton" type="submit" value="Enviar mensaje">
      </form>
    </section>
  </section>
</template>

<style scoped lang="scss">
$organizations-primary: #4a4a9e;
$organizations-accent: #367591;

#team_box-team-title {
  padding: 30px 0;
  font-size: clamp(2.25rem, 6vw, 5rem);
  text-align: center;
  color: $color-background;
}

#team_box {
  width: 100%;
  padding-bottom: 30px;
  background-color: $organizations-primary;
}

.team-status {
  text-align: center;
  color: $color-background;
  padding-block: 2rem;
}

#contact_box {
  width: 100%;
  margin: auto;
}

#contact_box-contact-title {
  color: black;
  font-size: 3.5rem;
  text-align: center;
  line-height: 70px;
}

#contact_box-main {
  background-color: $organizations-primary;
  border-radius: 10px;
  height: 500px;
  width: 75%;
  margin: 40px auto;
  margin-top: 20px;
  display: flex;
  align-items: center;
}

#contact_box-info {
  padding: 20px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  width: 50%;

  a {
    text-decoration: none;

    &:hover {
      text-decoration: underline;
      padding: 10px;
      border-radius: 10px;
    }
  }

  div {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  svg {
    fill: $color-background;
    opacity: 0.7;
  }

  h2 {
    font-size: 1.2rem;
    font-weight: 400;
    color: $color-background;
    opacity: 0.7;
  }
}

#contact_box-form {
  width: 50%;
  padding: 20px 30px;
  display: flex;
  flex-direction: column;

  h2 {
    color: $color-background;
    font-size: 1.7rem;
    margin-bottom: 15px;
  }

  label {
    color: $color-background;
    font-weight: 400;
    font-size: 0.9rem;
    margin-bottom: 10px;
  }

  input,
  textarea {
    margin-bottom: 10px;
    padding: 5px;
    background-color: transparent;
    border: 2px solid $organizations-accent;
    border-radius: 10px;
    color: $color-background;
    font-size: 1rem;
  }

  input {
    height: 35px;
  }

  textarea {
    resize: none;
    height: 60px;
  }

  input[type='submit'] {
    margin-top: 10px;
    height: 35px;
    width: 50%;
    cursor: pointer;
    color: $color-background;
    background-color: $organizations-accent;
  }
}

@media (max-width: 950px) {
  #contact_box-contact-title {
    padding-bottom: 10px;
    color: $color-background;
  }

  #contact_box {
    background: $organizations-primary;
    width: 100%;
    padding: 30px;
    height: auto;
  }

  #contact_box-main {
    width: 100%;
    flex-direction: column;
    background: none;
    height: auto;
  }

  #contact_box-info,
  #contact_box-form {
    width: 100%;
  }
}
</style>
