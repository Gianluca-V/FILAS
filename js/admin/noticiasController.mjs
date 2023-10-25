import { API } from "../API.mjs";
// Function to fetch and display news data
async function PopulateTable() {
  const newsTableBody = document.querySelector(".news-table__body");

  const newsData = await API.News.GetAll().catch((e) => console.error(e));

  // Clear existing table rows
  newsTableBody.innerHTML = "";

  // Loop through the news data and create table rows
  newsData.forEach((news) => {
    const row = newsTableBody.insertRow();

    // Add data cells
    row.innerHTML = `
      <td class="news-table__cell">${news.ID}</td>
      <td class="news-table__cell">${news.Title}</td>
      <td class="news-table__cell">${news.Body}</td>
      <td class="news-table__cell">
          <img class="news-table__image" src="${news.Image}" alt="${news.Title}">
      </td>
      <td class="news-table__cell">
          <button class="news-table__button news-table__button--edit" data-id="${news.id}">Edit</button>
          <button class="news-table__button news-table__button--delete" data-id="${news.id}">Delete</button>
      </td>
  `;
  });
}
PopulateTable();