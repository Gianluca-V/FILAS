//Nav code

const toggle_open = document.getElementById('toggle_open');
toggle_open.addEventListener('click', togglemenu);

function togglemenu(){
    document.querySelector(".nav__ul").classList.toggle('active')
}
