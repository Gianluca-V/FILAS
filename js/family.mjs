import {API} from './API.mjs'

API.Family.GetAll().then((data)=>{
    displayFamily(data);
});

function displayFamily(family){
    const familySection = document.querySelector(".family");

    family.forEach(element => {
        familySection.innerHTML +=`<div class="family__events-container">
    <img src="${ element.Image != '' ? element.Image : "assets/default-img.png"}" alt="${element.Image}" loading="lazy">
    <p class="section--family__text">${element.Body}</p>
    </div>`;
    });

    
}