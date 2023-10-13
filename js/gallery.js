$(document).ready(function() {
    $.ajax({
        url: 'http://localhost/FilasServer/api/gallery/',
        type: 'GET',
        dataType: 'json',
        success: function(data) {
            if (data.length > 0) {
                var imageContainer = $('#image-container');
                $.each(data, function(index, image) {
                    var imgElement = $('<img>').addClass('gallery__photo').attr('src', image.Image).attr('alt', 'Imagen ' + (index + 1));
                    imageContainer.append(imgElement);
                });
                addClickListener();
            } else {
                $('#image-container').html('<p>No se encontraron imágenes.</p>');
            }
        },
        error: function() {
            $('#image-container').html('<p>Error al cargar las imágenes.</p>');
        }
    });
});

const addClickListener = () =>{
    function toggleImage(){
        document.querySelector('.viewimg').style.display = 'none';
    }
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

