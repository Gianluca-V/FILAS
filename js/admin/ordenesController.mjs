import { API } from "../API.mjs";
// Function to fetch and display Order data
async function PopulateTable() {
  const OrdersData = await API.Orders.GetAll().catch((e) => console.error(e));

  const OrdersTableBody = document.querySelector(".table__body");
  // Clear existing table rows
  OrdersTableBody.innerHTML = "";

  // Loop through the Order data and create table rows
  let index = 0;
  OrdersData.forEach((Order) => {
    const row = OrdersTableBody.insertRow();

    let products = "";
    Order.products.forEach((product) => {
      products += `<div class="table__ticket"> <span>${product.productName}</span>  <span>X${product.productQuantity}</span> <span>Subtotal: ${product.productPrice * product.productQuantity}</span> </div>`
    });

    // Add data cells
    row.innerHTML = `
        <td class="table__cell table__cell--orders" data-cell="#">${Order.orderID}</td>

        <td class="table__cell table__cell--orders" data-cell="Productos"><button class="table__products-btn">Ver productos</button></td>

        <td class="table__cell table__cell--orders" data-cell="Total">${Order.orderTotal}</td>
        <td class="table__cell table__cell--orders" data-cell="Estado">${Order.orderState}</td>
        <td class="table__cell table__cell--orders" data-cell="Fecha de Inicio">${Order.orderStartDate}</td>
        <td class="table__cell table__cell--orders" data-cell="Fecha de Finalización">${Order.orderFinishDate}</td>
        <td class="table__cell table__cell--actions">
            <button class="table__button table__button--finish" data-id="${Order.orderID}"><svg class="table__svg table__svg--edit" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 512 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M471.6 21.7c-21.9-21.9-57.3-21.9-79.2 0L362.3 51.7l97.9 97.9 30.1-30.1c21.9-21.9 21.9-57.3 0-79.2L471.6 21.7zm-299.2 220c-6.1 6.1-10.8 13.6-13.5 21.9l-29.6 88.8c-2.9 8.6-.6 18.1 5.8 24.6s15.9 8.7 24.6 5.8l88.8-29.6c8.2-2.7 15.7-7.4 21.9-13.5L437.7 172.3 339.7 74.3 172.4 241.7zM96 64C43 64 0 107 0 160V416c0 53 43 96 96 96H352c53 0 96-43 96-96V320c0-17.7-14.3-32-32-32s-32 14.3-32 32v96c0 17.7-14.3 32-32 32H96c-17.7 0-32-14.3-32-32V160c0-17.7 14.3-32 32-32h96c17.7 0 32-14.3 32-32s-14.3-32-32-32H96z"/></svg></button>
            <button class="table__button table__button--cancel" data-id="${Order.orderID}"><svg class="table__svg table__svg--delete" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M135.2 17.7L128 32H32C14.3 32 0 46.3 0 64S14.3 96 32 96H416c17.7 0 32-14.3 32-32s-14.3-32-32-32H320l-7.2-14.3C307.4 6.8 296.3 0 284.2 0H163.8c-12.1 0-23.2 6.8-28.6 17.7zM416 128H32L53.2 467c1.6 25.3 22.6 45 47.9 45H346.9c25.3 0 46.3-19.7 47.9-45L416 128z"/></svg></button>
        </td>
    `;

    const showProductsButton = document.querySelectorAll(".table__products-btn")[index];
    showProductsButton.addEventListener("click", () => {
      const productDialog = document.querySelector(".products-dialog");
      productDialog.showModal();
      document.querySelector(".products-dialog__container").innerHTML = products;
    });
    index++;
  });

  const finishButtons = document.querySelectorAll(".table__button--finish");
  const cancelButtons = document.querySelectorAll(".table__button--cancel");

  cancelButtons.forEach((cancelButton) => {
    cancelButton.addEventListener("click", async function (e) {
      const id = e.target.closest("[data-id]").getAttribute("data-id");
      if (!confirm(`Estas seguro? esto cancelara la orden ${id} para siempre`)) return;
      await API.Orders.cancel(id);
      PopulateTable();
    });
  })

  finishButtons.forEach((finishButton) => {
    finishButton.addEventListener("click", async function (e) {
      const id = e.target.closest("[data-id]").getAttribute("data-id");
      if (!confirm(`Estas seguro? esto dará la orden ${id} por finalizada`)) return;
      await API.Orders.finish(id);
      PopulateTable();
    });
  })

  document.querySelector(".products-dialog__close").addEventListener("click", function (e) {
    document.querySelector(".products-dialog").close();
  });
}
PopulateTable();
window.addEventListener("hashchange", () => {
  if (window.location.hash.slice(1) === "ordenes") PopulateTable();
});

