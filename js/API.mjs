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
        const response = await fetch(this.apiURL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });
        return response.json();
    }
    async Put(id, data){
        const response = await fetch(this.apiURL + id, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });
        return response.json();
    }
    async Delete(id){
        await fetch(this.apiURL + id, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        return true;
    }
}

const Gallery = Object.freeze(new APIClass("gallery"));
const Mails = Object.freeze(new APIClass("mails"));
const News = Object.freeze(new APIClass("news"));
const Products = Object.freeze(new APIClass("products"));

const API = Object.freeze({
    Gallery:Gallery,
    Mails:Mails,
    News:News,
    Products:Products
})

export {API};