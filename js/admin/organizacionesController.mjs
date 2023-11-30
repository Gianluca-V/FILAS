import { API, removeAllEventListeners } from "../API.mjs";
// Function to fetch and display Organization data
async function PopulateTable() {
    const OrganizationsData = await API.Organizations.GetAll().catch((e) => console.error(e));

    const OrganizationsTableBody = document.querySelector(".table__body");
    // Clear existing table rows
    OrganizationsTableBody.innerHTML = "";

    // Loop through the Organization data and create table rows
    OrganizationsData.forEach((Organization) => {
        const row = OrganizationsTableBody.insertRow();

        // Add data cells
        row.innerHTML = `
      <td class="table__cell table__cell--Organization" data-cell="#">${Organization.ID}</td>
      <td class="table__cell table__cell--Organization" data-cell="Titulo">${Organization.Title}</td>
      <td class="table__cell table__cell--Organization" data-cell="Cuerpo">${Organization.Description}</td>
      <td class="table__cell table__cell--Organization" data-cell="Imagen">
          <img class="table__image" src="${Organization.Image}" alt="${Organization.Title}">
      </td>
      <td class="table__cell table__cell--actions">
          <button class="table__button table__button--edit" data-id="${Organization.ID}"><svg class="table__svg table__svg--edit" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 512 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M471.6 21.7c-21.9-21.9-57.3-21.9-79.2 0L362.3 51.7l97.9 97.9 30.1-30.1c21.9-21.9 21.9-57.3 0-79.2L471.6 21.7zm-299.2 220c-6.1 6.1-10.8 13.6-13.5 21.9l-29.6 88.8c-2.9 8.6-.6 18.1 5.8 24.6s15.9 8.7 24.6 5.8l88.8-29.6c8.2-2.7 15.7-7.4 21.9-13.5L437.7 172.3 339.7 74.3 172.4 241.7zM96 64C43 64 0 107 0 160V416c0 53 43 96 96 96H352c53 0 96-43 96-96V320c0-17.7-14.3-32-32-32s-32 14.3-32 32v96c0 17.7-14.3 32-32 32H96c-17.7 0-32-14.3-32-32V160c0-17.7 14.3-32 32-32h96c17.7 0 32-14.3 32-32s-14.3-32-32-32H96z"/></svg></button>
          <button class="table__button table__button--delete" data-id="${Organization.ID}"><svg class="table__svg table__svg--delete" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M135.2 17.7L128 32H32C14.3 32 0 46.3 0 64S14.3 96 32 96H416c17.7 0 32-14.3 32-32s-14.3-32-32-32H320l-7.2-14.3C307.4 6.8 296.3 0 284.2 0H163.8c-12.1 0-23.2 6.8-28.6 17.7zM416 128H32L53.2 467c1.6 25.3 22.6 45 47.9 45H346.9c25.3 0 46.3-19.7 47.9-45L416 128z"/></svg></button>
      </td>
  `;
    });

    const addButton = document.querySelector(".add-button");
  const editButtons = document.querySelectorAll(".table__button--edit");
  const deleteButtons = document.querySelectorAll(".table__button--delete");
  const closeButton = document.querySelector(".form__close");
  const formContainer = document.querySelector(".form-container");

  addButton.addEventListener("click", function () {
    document.querySelector(".form__action").textContent = "Agregar";
    formContainer.showModal();

    document.querySelector(".form__input#titulo").value = ""
    document.querySelector(".form__input#cuerpo").value = ""
    document.querySelector(".form__input#imagen").value = ""

    removeAllEventListeners(document.querySelector(".form__input--submit"));
    const submitButton = document.querySelector(".form__input--submit");
    submitButton.addEventListener("click", async (e) => {
      e.preventDefault();

      const object = {
        Title: document.querySelector(".form__input#titulo").value,
        Description: document.querySelector(".form__input#cuerpo").value,
        Image: document.querySelector(".form__input#imagen").value,
      }

      formContainer.close();
      await API.Organizations.Post(object);
      PopulateTable();
    });
  });

  editButtons.forEach((editButton)=>{
    editButton.addEventListener("click", function (e) {
      const id = e.target.closest("[data-id]").getAttribute("data-id");
      document.querySelector(".form__action").textContent = "Editar";
      formContainer.showModal();

      const parentNode = Array.from(document.querySelectorAll('.table__cell[data-cell="#"]')).find(x => x.textContent == id).parentElement;
      const name = parentNode.querySelector("[data-cell=Titulo]").textContent;
      const body = parentNode.querySelector("[data-cell=Cuerpo]").textContent;
      const image = parentNode.querySelector("[data-cell=Imagen]").firstElementChild.getAttribute("src");

      document.querySelector(".form__input#titulo").value = name
      document.querySelector(".form__input#cuerpo").value = body
      document.querySelector(".form__input#imagen").value = image

      removeAllEventListeners(document.querySelector(".form__input--submit"));
      const submitButton = document.querySelector(".form__input--submit");
      console.log(submitButton);
      submitButton.addEventListener("click",async (e)=>{
        e.preventDefault();

        const object = {
          Title:  document.querySelector(".form__input#titulo").value,
          Description: document.querySelector(".form__input#cuerpo").value,
          Image: document.querySelector(".form__input#imagen").value,
        }

        formContainer.close();
        await API.Organizations.Put(id, object);
        PopulateTable();
      })
    });
  })

  deleteButtons.forEach((deleteButton) => {
    deleteButton.addEventListener("click", async function (e) {
      const id = e.target.closest("[data-id]").getAttribute("data-id");
      if(!confirm(`Estas seguro? esto eliminara el elemento ${id} para siempre`)) return;
        await API.Organizations.Delete(id);
        PopulateTable();
    });
  })

  closeButton.addEventListener("click", function (e) {
    e.preventDefault();
    const hash = window.location.hash.slice(1);
    formContainer.close();
    window.location.hash = hash;
  });
}
PopulateTable();
window.addEventListener("hashchange", () => {
    if (window.location.hash.slice(1) === "organizaciones") PopulateTable();
});

