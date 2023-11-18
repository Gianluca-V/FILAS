<?php
switch ($request_method) {
    case 'GET':

        $headers = getallheaders();
        if (!isset($headers['Authorization'])) {
            return http_response_code(400);
            break;
        }

        $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
        if (!TokenValidationResponse($token)) {
            return http_response_code(401);
            break;
        }

        // Get all orders or a specific order by ID
        if ($parts[4] !== "") {
            $order_id = intval($parts[4]);
            getOrder($order_id);
        } else {
            getOrders();
        }
        break;
    case 'POST':
        // Create a new order
        $data = json_decode(file_get_contents("php://input"));
        $headers = getallheaders();
        if (!isset($headers['Authorization'])) {
            return http_response_code(400);
            break;
        }

        $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
        if (!TokenValidationResponse($token)) {
            return http_response_code(401);
            break;
        }
        createOrder($data);
        break;
    case 'PUT':
        // Update a order by ID
        $data = json_decode(file_get_contents("php://input"));
        $order_id = intval($parts[4]);
        $headers = getallheaders();
        if (!isset($headers['Authorization'])) {
            return http_response_code(400);
            break;
        }

        $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
        if (!TokenValidationResponse($token)) {
            return http_response_code(401);
            break;
        }
        updateOrder($order_id, $data);
        break;
    case 'DELETE':
        // Delete a order by ID
        $order_id = intval($parts[4]);
        $headers = getallheaders();
        if (!isset($headers['Authorization'])) {
            return http_response_code(400);
            break;
        }

        $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
        if (!TokenValidationResponse($token)) {
            return http_response_code(401);
            break;
        }
        deleteOrder($order_id);
        break;
    default:
        http_response_code(405);
        echo json_encode(array("message" => "Method not allowed"));
        break;
}

function getOrders()
{
    global $conn;
    $sql = "SELECT
                o.ID AS orderID,
                o.total AS orderTotal,
                op.productID,
                p.name AS productName,
                op.productQuantity,
                op.orderPrice AS orderProductPrice
            FROM
                orders o
            JOIN
                orderProduct op ON o.ID = op.orderID
            JOIN
                products p ON op.productID = p.ID
            ORDER BY
                o.ID, op.productID;";

    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $orders = array();
        while ($row = $result->fetch_assoc()) {
            $orders[] = $row;
        }
        echo json_encode($orders);
    } else {
        echo json_encode(array());
    }
}

function getOrder($order_id)
{
    global $conn;
    $sql = "SELECT
                o.ID AS orderID,
                o.total AS orderTotal,
                op.productID,
                p.name AS productName,
                op.productQuantity,
                op.orderPrice AS orderProductPrice
            FROM
                orders o
            JOIN
                orderProduct op ON o.ID = op.orderID
            JOIN
                products p ON op.productID = p.ID
            WHERE
                o.ID = $order_id
            ORDER BY
                o.ID, op.productID;";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $order = array();
        while ($row = $result->fetch_assoc()) {
            $order[] = $row;
        }
        echo json_encode($order);
    } else {
        http_response_code(404);
        echo json_encode(array("message" => "Order not found"));
    }
}

function createOrder($data)
{
    global $conn;
    // Assuming $data contains the necessary information for creating an order
    $orderProducts = $data->orderProducts;

    // Insert the order
    $orderInsertQuery = "INSERT INTO orders (total) VALUES (null)";
    mysqli_query($conn, $orderInsertQuery);

    // Get the ID of the newly inserted order
    $orderId = mysqli_insert_id($conn);

    // Insert order products
    foreach ($orderProducts as $product) {
        $productId = mysqli_real_escape_string($conn, $product->productID);
        $quantity = mysqli_real_escape_string($conn, $product->quantity);

        $orderProductInsertQuery = "INSERT INTO orderProduct (orderID, productID, productQuantity) 
                                    VALUES ($orderId, $productId, $quantity)";
        mysqli_query($conn, $orderProductInsertQuery);
    }

    if (mysqli_affected_rows($conn) > 0) {
        echo json_encode(array("message" => "Order created successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error creating order: " . mysqli_error($conn)));
    }
}

function updateOrder($order_id, $data)
{
    global $conn;

    // Assuming $data contains the necessary information for updating an order
    $orderProducts = $data->orderProducts; // assuming $orderProducts is an array

    // Delete existing order products for the given order ID
    $deleteOrderProductsQuery = "DELETE FROM orderProduct WHERE orderID = $order_id";
    mysqli_query($conn, $deleteOrderProductsQuery);

    // Insert updated order products
    foreach ($orderProducts as $product) {
        $productId = mysqli_real_escape_string($conn, $product->productID);
        $quantity = mysqli_real_escape_string($conn, $product->quantity);

        $orderProductInsertQuery = "INSERT INTO orderProduct (orderID, productID, productQuantity) 
                                    VALUES ($order_id, $productId, $quantity)";
        mysqli_query($conn, $orderProductInsertQuery);
    }

    if (mysqli_affected_rows($conn) > 0) {
        echo json_encode(array("message" => "Order updated successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error updating order: " . mysqli_error($conn)));
    }
}

function deleteOrder($order_id)
{
    global $conn;

    // Delete order products for the given order ID
    $deleteOrderProductsQuery = "DELETE FROM orderProduct WHERE orderID = $order_id";
    mysqli_query($conn, $deleteOrderProductsQuery);

    // Delete the order
    $deleteOrderQuery = "DELETE FROM orders WHERE ID = $order_id";
    mysqli_query($conn, $deleteOrderQuery);

    if (mysqli_affected_rows($conn) > 0) {
        echo json_encode(array("message" => "Order deleted successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error deleting order: " . mysqli_error($conn)));
    }
}
