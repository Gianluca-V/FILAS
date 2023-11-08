class APIClass {
    SERVER_URL = 'http://localhost/filasServer/api/';
    table = "";
    apiURL = "";

    constructor(Table) {
        this.table = Table;
        this.apiURL = `${this.SERVER_URL}${this.table}/`;
    }

    async GetAll() {
        const response = await fetch(this.apiURL);
        return response.json();
    }
    async GetOne(id) {
        const response = await fetch(this.apiURL + id);
        return response.json();
    }
    async Post(data) {
        const token = getCookie("token");
        const response = await fetch(this.apiURL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify(data),
        });
        return response.json();
    }
    async Put(id, data){
        const token = getCookie("token");
        const response = await fetch(this.apiURL + id, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify(data),
        });
        return response.json();
    }
    async Delete(id){
        const token = getCookie("token");
        await fetch(this.apiURL + id, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
        });
        return true;
    }
}

function getCookie(name){
    const cookies = document.cookie.split("; ");
    let token = cookies
      .filter(cookie => cookie.split("=")[0] === name)
      .map(cookie => cookie.split("=")[1])

    return token;
}

const Gallery = Object.freeze(new APIClass("gallery"));
const Mails = Object.freeze(new APIClass("mails"));
const News = Object.freeze(new APIClass("news"));
const Products = Object.freeze(new APIClass("products"));
const Family = Object.freeze(new APIClass("family"));
const Admins = Object.freeze(new APIClass("admins"));
const Organizations = Object.freeze(new APIClass("organizations"));

const API = Object.freeze({
    Gallery:Gallery,
    Mails:Mails,
    News:News,
    Products:Products,
    Family:Family,
    Admins:Admins,
    Organizations:Organizations
})

export {API};