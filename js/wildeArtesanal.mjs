import {API} from './API.mjs';

API.Products.GetAll().then((data)=>{
    console.log(data)
    displayProducts(data);
})

let cart = []
// Function to generate a product card
function generateProductCard(product) {
    const productCard = document.createElement("div");
    productCard.classList.add("product__card");

    // Create elements for product details (e.g., name, price, image, description)
    const productName = document.createElement("h2");
    productName.textContent = product.Name;
    productName.classList.add("product__name");

    const productPrice = document.createElement("p");
    productPrice.textContent = `$${product.Price}`;
    productPrice.classList.add("product__price");

    const productImage = document.createElement("img");
    productImage.src = product.Image != '' ? product.Image : "assets/default-img.png";
    productImage.alt = product.Name;
    productImage.classList.add("product__img");

    const productDescription = document.createElement("p");
    productDescription.textContent = product.Description;
    productDescription.classList.add("product__description");

    const productBuyBtn = document.createElement("button");
    productBuyBtn.textContent = "Añadir al carrito";
    productBuyBtn.classList.add("product__buy");

    productBuyBtn.addEventListener("click", () => {
        const existingProduct = cart.find(item => item.name === product.Name);
        if (existingProduct) {
            existingProduct.quantity++;
        } else {
            cart.push({ name: product.Name, quantity: 1, price: product.Price });
        }
        updateCart();
    })

    const productData = document.createElement("div");
    productData.classList.add("product__data");
    productData.appendChild(productName);
    productData.appendChild(productPrice);
    productData.appendChild(productBuyBtn);

    // Append product details to the product card
    productCard.appendChild(productImage);
    productCard.appendChild(productData);
    //productCard.appendChild(productDescription);

    return productCard;
}

function updateCart() {
    const cartProductsContainer = document.querySelector(".cart__products-container")
    cartProductsContainer.innerHTML = ""

    let totalCost = 0;
    const cartTotalElement = document.querySelector(".cart__total");

    console.log(cart)
    if (cart.length <= 0) {
        cartProductsContainer.textContent = "Agregue productos al carrito";
        cartTotalElement.textContent = totalCost;
        document.querySelector(".cart__buy").setAttribute("disabled",true)
        return;
    }
    document.querySelector(".cart__buy").removeAttribute("disabled")

    cart.forEach(product => {
        const cartProduct = document.createElement("div");
        cartProduct.classList.add("cart__product");

        const productName = document.createElement("div")
        productName.classList.add("cart__product-name")
        productName.textContent = product.name;

        const productQuantityContainer = document.createElement("div")
        productQuantityContainer.classList.add("cart__product-quantity-container");
        productQuantityContainer.innerHTML = `
        X <input type="number" name="product-quantity" id="product-quantity" class="cart__product-quantity" min="1" value="${product.quantity}" placeholder="1" required>
        `

        const productQuantityInput = productQuantityContainer.querySelector("input");

        productQuantityInput.addEventListener("input", (e) => {
            product.quantity = e.target.value;
            updateCart();
            return;
        })

        const productDelete = document.createElement("button");
        productDelete.classList.add("cart__product-delete")
        productDelete.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" height="1em" viewBox="0 0 448 512"><!--! Font Awesome Free 6.4.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2023 Fonticons, Inc. --><path d="M135.2 17.7L128 32H32C14.3 32 0 46.3 0 64S14.3 96 32 96H416c17.7 0 32-14.3 32-32s-14.3-32-32-32H320l-7.2-14.3C307.4 6.8 296.3 0 284.2 0H163.8c-12.1 0-23.2 6.8-28.6 17.7zM416 128H32L53.2 467c1.6 25.3 22.6 45 47.9 45H346.9c25.3 0 46.3-19.7 47.9-45L416 128z"/></svg>';
        productDelete.addEventListener("click", () => {
            cart = cart.filter(item => item.name !== product.name);
            updateCart();
            return;
        });

        totalCost += parseFloat(product.price) * parseInt(product.quantity);
        cartTotalElement.textContent = totalCost;

        cartProduct.appendChild(productName);
        cartProduct.appendChild(productQuantityContainer);
        cartProduct.appendChild(productDelete);
        cartProductsContainer.appendChild(cartProduct);
    });

}

document.querySelector(".cart__buy").addEventListener("click",()=>{
    
})

// Function to display products in the product grid
function displayProducts(products) {
    const productGrid = document.querySelector(".products-section");

    // Clear any existing products in the grid
    productGrid.innerHTML = "";

    // Generate and append product cards to the grid
    products.forEach((product) => {
        const productCard = generateProductCard(product);
        productGrid.appendChild(productCard);
    });
}
