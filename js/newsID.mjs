import {API} from './API.mjs'
window.addEventListener("DOMContentLoaded",()=>{
    const params = new URLSearchParams(window.location.search);
    const articleId = params.get('ID');
    console.log(`http://localhost/FilasServer/api/news/${articleId}`)


    // Fetch the specific news article from your backend based on articleId.
    API.News.GetOne(articleId)
        .then(article => {
            // Populate the news template with data.
            const newsImage = document.querySelector('.article__image');
            const newsTitle = document.querySelector('.article__title');
            const newsContent = document.querySelector('.article__content');

            article = article[0];
            console.log(article)
            newsImage.src = article.Image;
            newsImage.alt = article.Title;
            newsTitle.textContent = article.Title;
            newsContent.textContent = article.Body;
        })
        .catch(error => {
            console.error('Error fetching data:', error);
        });
})