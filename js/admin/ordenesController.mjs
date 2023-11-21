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
        <td class="table__cell table__cell--orders" data-cell="Estado"><button class="table__client-btn">Ver datos de cliente</button></td>
        <td class="table__cell table__cell--orders" data-cell="Fecha de Inicio">${Order.orderStartDate}</td>
        <td class="table__cell table__cell--orders" data-cell="Fecha de Finalización">${Order.orderFinishDate}</td>
        <td class="table__cell table__cell--actions">
            <button class="table__button table__button--finish" data-id="${Order.orderID}"><svg class="table__svg table__svg--finish" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M64 32C28.7 32 0 60.7 0 96V416c0 35.3 28.7 64 64 64H384c35.3 0 64-28.7 64-64V96c0-35.3-28.7-64-64-64H64zM337 209L209 337c-9.4 9.4-24.6 9.4-33.9 0l-64-64c-9.4-9.4-9.4-24.6 0-33.9s24.6-9.4 33.9 0l47 47L303 175c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9z"/></svg></button>
            <button class="table__button table__button--cancel" data-id="${Order.orderID}"><svg class="table__svg table__svg--cancel" xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M64 32C28.7 32 0 60.7 0 96V416c0 35.3 28.7 64 64 64H384c35.3 0 64-28.7 64-64V96c0-35.3-28.7-64-64-64H64zm79 143c9.4-9.4 24.6-9.4 33.9 0l47 47 47-47c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9l-47 47 47 47c9.4 9.4 9.4 24.6 0 33.9s-24.6 9.4-33.9 0l-47-47-47 47c-9.4 9.4-24.6 9.4-33.9 0s-9.4-24.6 0-33.9l47-47-47-47c-9.4-9.4-9.4-24.6 0-33.9z"/></svg></button>
        </td>
    `;

    const showProductsButton = document.querySelectorAll(".table__products-btn")[index];
    showProductsButton.addEventListener("click", () => {
      const productDialog = document.querySelector(".products-dialog");
      productDialog.showModal();
      document.querySelector(".products-dialog__container").innerHTML = products;
    });

    const showClientButton = document.querySelectorAll(".table__client-btn")[index];
    showClientButton.addEventListener("click", () => {
      const productDialog = document.querySelector(".client-dialog");
      productDialog.showModal();
      document.querySelector(".client-dialog__container").innerHTML = `
      <div class="table__client"> <span>Nombre: ${Order.orderName}</span> <span>Tel: ${Order.orderPhone}</span> </div>
      `;
    });

    index++;
  });

  const finishButtons = document.querySelectorAll(".table__button--finish");
  const cancelButtons = document.querySelectorAll(".table__button--cancel");

  cancelButtons.forEach((cancelButton) => {
    cancelButton.addEventListener("click", async function (e) {
      const id = e.target.closest("[data-id]").getAttribute("data-id");
      if (!confirm(`Estas seguro? esto cancelara la orden ${id} para siempre`)) return;
      const response = await API.Orders.cancel(id);
      if(!response.ok){
        alert("No se pudo modificar la orden, es posible que ya este finalizada/cancelada");
      }
      PopulateTable();
    });
  })

  finishButtons.forEach((finishButton) => {
    finishButton.addEventListener("click", async function (e) {
      const id = e.target.closest("[data-id]").getAttribute("data-id");
      if (!confirm(`Estas seguro? esto dará la orden ${id} por finalizada`)) return;
      const response = await API.Orders.finish(id);
      if(!response.ok){
        alert("No se pudo modificar la orden, es posible que ya este finalizada/cancelada");
      }
      PopulateTable();
    });
  })

  document.querySelector(".products-dialog__close").addEventListener("click", function (e) {
    document.querySelector(".products-dialog").close();
  });
  document.querySelector(".client-dialog__close").addEventListener("click", function (e) {
    document.querySelector(".client-dialog").close();
  });
}
PopulateTable();
window.addEventListener("hashchange", () => {
  if (window.location.hash.slice(1) === "ordenes") PopulateTable();
});

