<?php
switch ($request_method) {
    case 'GET':
        // Get all products or a specific product by ID
        if ($parts[4] !== "") {
            $product_id = intval($parts[4]);
            getProduct($product_id);
        } else {
            getProducts();
        }
        break;
    case 'POST':
        // Create a new product
        $data = json_decode(file_get_contents("php://input"));
        createProduct($data);
        break;
    case 'PUT':
        // Update a product by ID
        $data = json_decode(file_get_contents("php://input"));
        $product_id = intval($_GET['id']);
        updateProduct($product_id, $data);
        break;
    case 'DELETE':
        // Delete a product by ID
        $product_id = intval($_GET['id']);
        deleteProduct($product_id);
        break;
    default:
        http_response_code(405);
        echo json_encode(array("message" => "Method not allowed"));
        break;
}

function getProducts()
{
    global $conn;
    $sql = "SELECT * FROM Products";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $products = array();
        while ($row = $result->fetch_assoc()) {
            $products[] = $row;
        }
        echo json_encode($products);
    } else {
        echo json_encode(array());
    }
}

function getProduct($product_id)
{
    global $conn;
    $sql = "SELECT * FROM Products WHERE ID = $product_id";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        echo json_encode($row);
    } else {
        http_response_code(404);
        echo json_encode(array("message" => "Product not found"));
    }
}

function createProduct($data)
{
    global $conn;
    // Assuming $data contains the necessary fields for creating a product
    $name = $conn->real_escape_string($data->Name);
    $price = floatval($data->Price);
    $stock = intval($data->Stock);
    $image = $conn->real_escape_string($data->Image);
    $description = $conn->real_escape_string($data->Description);

    $sql = "INSERT INTO Products (Name, Price, Stock, Image, Description) 
            VALUES ('$name', $price, $stock, '$image', '$description')";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "Product created successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error creating product: " . $conn->error));
    }
}

function updateProduct($product_id, $data)
{
    global $conn;
    $id = intval($product_id);
    $name = $conn->real_escape_string($data->Name);
    $price = floatval($data->Price);
    $stock = intval($data->Stock);
    $image = $conn->real_escape_string($data->Image);
    $description = $conn->real_escape_string($data->Description);

    $sql = "INSERT INTO Products (Name, Price, Stock, Image, Description) 
            VALUES ('$name', $price, $stock, '$image', '$description') WHERE ID = $id";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "Product updated successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error creating product: " . $conn->error));
    }
}

function deleteProduct($product_id)
{
    global $conn;
    $id = intval($product_id);
    $sql = "DELETE FROM Products WHERE ID = $id";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "Product deleted successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error deleting product: " . $conn->error));
    }
}


?>
