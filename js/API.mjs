class APIClass {
    SERVER_URL = 'http://localhost/filasserver/api/';
    table = "";
    apiURL = "";

    constructor(Table) {
        this.table = Table;
        this.apiURL = `http://localhost/filasserver/api/${this.table}/`;
    }

    GetAll() {
        return fetch(this.apiURL)
            .then((response) => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .catch((error) => {
                console.error('There was a problem with the fetch operation:', error);
                return false;
            });
    }
    GetOne(id) {
        return fetch(this.apiURL + id)
            .then((response) => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .catch((error) => {
                console.error('There was a problem with the fetch operation:', error);
                return false;
            });
    }
    Post(data) {
        return fetch(this.apiURL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .catch((error) => {
                console.error('There was a problem with the fetch operation:', error);
                return false;
            });
    }
    Put(id,data){
        return fetch(this.apiURL + id, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .catch((error) => {
                console.error('There was a problem with the fetch operation:', error);
                return false;
            });
    }
    Delete(id){
        return fetch(this.apiURL + id, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return true;
            })
            .catch((error) => {
                console.error('There was a problem with the fetch operation:', error);
                return false;
            });
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