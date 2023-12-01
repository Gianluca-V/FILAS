import {API} from './API.mjs'

API.Family.GetAll().then((data)=>{
    displayFamily(data);
});


function displayFamily(family){
    const CDcardsContainer = document.querySelector('.family__cards-container--CD')
    const TPcardsContainer = document.querySelector('.family__cards-container--TP')
    
    family.forEach(element => {

        console.log(CDcardsContainer)

        if(element.Category == "Centro de dia"){
            CDcardsContainer.innerHTML+=`
            <div class="family__card">
                <div class="family__img">${element.Image !== ''? `<img src="" alt="${element.Image}" loading="lazy" class="family__img">` : ''}</div>
                <div class="family__body">${element.Body}</div>
            </div>`
        }
        
        if(element.Category == 'Taller protegido'){
            TPcardsContainer.innerHTML+=`
            <div class="family__card">
                <div class="family__img">${element.Image !== ''? `<img src="" alt="${element.Image}" loading="lazy" class="family__img">` : ''}<div>
            </div>`
        }
    });
}