import {API} from './API.mjs'

API.Gallery.GetAll().then((data)=>{
    const imageContainer = document.querySelector('#image-container')
    if (data.length > 0) {
        let index = 0;
        data.forEach((image) => {
            const imgElement = document.createElement('img')
            imgElement.classList.add('gallery__photo')
            imgElement.setAttribute('src', image.Image)
            imgElement.setAttribute('alt', 'Imagen ' + (index + 1));
            imgElement.setAttribute('loading', 'lazy');
            imgElement.setAttribute('draggable', 'false');
            imageContainer.appendChild(imgElement);
            index++;
        });
        addClickListener();
    } else {
        imageContainer.innerHTML = '<p>No se encontraron imágenes.</p>';
    }
})

const addClickListener = () =>{
    const galleryPhotos = document.querySelectorAll('.gallery__photo');
    const closeImg = document.querySelector('.viewimg__close');
    
    galleryPhotos.forEach(photo => {
        photo.addEventListener('click', () => {
            document.querySelector('.viewimg').style.display = 'flex';
            document.querySelector('.viewimg__photo').src = photo.src;
        });
    });
    
    closeImg.addEventListener('click', () => {
        document.querySelector('.viewimg').style.display = 'none';
    });
}

