function toggleMap() {
    var mapFrame = document.getElementById('map_frame');
    if (mapFrame.style.display === 'none' || mapFrame.style.display === '') {
        mapFrame.style.display = 'block'; // Mostrar el mapa
    } else {
        mapFrame.style.display = 'none'; // Ocultar el mapa
    }
}
