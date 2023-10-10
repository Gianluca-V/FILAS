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
    let index = 0;
    news.forEach(article => {
        if(index === 0){
            const newsItem = document.createElement('div');
            newsItem.classList.add('news__new');
            newsItem.classList.add('news__new--big');
            newsItem.innerHTML = `
            <img src="${article.Image}" alt="${article.Title}" class="new__img new__img--big">
            <div class="new__data new__data--big">
                <h2 class="new__title new__title--big">${article.Title}</h2>
                <a href="noticia.html?ID=${article.ID}" class="new__read new__read--big">Leer mas</a>
            </div>
        `
        }
        else{
            const newsItem = document.createElement('div');
            newsItem.classList.add('news__new');
            newsItem.innerHTML = `
                <img src="${article.Image}" alt="${article.Title}" class="new__img">
                <div class="new__data">
                    <h2 class="new__title">${article.Title}</h2>
                    <a href="noticia.html?ID=${article.ID}" class="new__read">Leer mas</a>
                </div>
            `
        }
        index++;
    });
}