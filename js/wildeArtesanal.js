(
() => {
    fetch('http://localhost/filasserver/api/products/')
        .then((response) => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then((data) => {
            // Handle the received data (list of products)
            // You can process and display the data here
            displayProducts(data);
        })
        .catch((error) => {
            console.error('There was a problem with the fetch operation:', error);
        });
}
)()

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
    productImage.src = product.Image != ''? product.Image : "assets/default-img.png";
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
            cart.push({ name: product.Name, quantity: 1 });
        }
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

}

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
