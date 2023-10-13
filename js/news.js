(
    () => {
        fetch('http://localhost/filasserver/api/news/')
            .then((response) => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then((data) => {
                displayNews(data);
            })
            .catch((error) => {
                console.error('There was a problem with the fetch operation:', error);
            });
    }
)()

function displayNews(news){
    const newsSection = document.querySelector(".news");
    const newsContainer = document.createElement('div')
    newsContainer.classList.add('news__container');
    newsSection.appendChild(newsContainer);

    let index = 0;
    news.forEach(article => {
        const newsItem = document.createElement('div');
        newsItem.classList.add('news__new');

        if(index === 0){
            newsItem.classList.add('news__new--big');
            newsItem.innerHTML = `
            <img src="${article.Image}" alt="${article.Title}" class="news__img news__img--big">
            <div class="news__data new__data--big">
                <h2 class="news__title news__title--big">${article.Title}</h2>
                <a href="noticia.html?ID=${article.ID}" class="news__read news__read--big">Leer mas</a>
            </div>
        `;
        newsSection.appendChild(newsItem);
        }
        else{
            newsItem.innerHTML = `
                <img src="${article.Image}" alt="${article.Title}" class="news__img">
                <div class="news__data">
                    <h2 class="news__title">${article.Title}</h2>
                    <a href="noticia.html?ID=${article.ID}" class="news__read">Leer mas</a>
                </div>
            `;  
            newsContainer.appendChild(newsItem);
        }
        
        index++;
    });
}