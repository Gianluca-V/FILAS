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
        createOrder($data);
        break;
    case 'PATCH':
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
        patchOrder($order_id, $data);
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
    default:
        http_response_code(405);
        echo json_encode(array("message" => "Method not allowed"));
        break;
}

function getOrders()
{
    global $conn;
    $sql = <<<SQL
        SELECT
            o.ID AS orderID,
            o.total AS orderTotal,
            o.state AS orderState,
            o.startDate AS orderStartDate,
            o.finishDate AS orderFinishDate,
            o.name AS orderName,
            o.phone AS orderPhone,
            CONCAT(
                '[',
                GROUP_CONCAT(
                    CONCAT(
                        '{"productName": "', p.name, '","productPrice": ', p.price, ',"productQuantity": ', op.productQuantity, '}'
                    ) SEPARATOR ','
                ),
                ']'
            ) AS products
        FROM
            orders o
        JOIN
            orderProduct op ON o.ID = op.orderID
        JOIN
            products p ON op.productID = p.ID
        GROUP BY
            o.ID
        ORDER BY
            o.ID;
    SQL;

    $result = $conn->query($sql);

    if ($result === false) {
        http_response_code(404);
        echo json_encode(array("message" => "No orders found"));
    } else {
        $orders = array();
        while ($row = $result->fetch_assoc()) {
            $row['products'] = json_decode(stripslashes($row['products']), true);
            $orders[] = $row;
        }
        echo json_encode($orders);
    }
}


function getOrder($order_id)
{
    global $conn;
    
    $sql = <<<SQL
        SELECT
            o.ID AS orderID,
            o.total AS orderTotal,
            o.state AS orderState,
            o.startDate AS orderStartDate,
            o.finishDate AS orderFinishDate,
            o.name AS orderName,
            o.phone AS orderPhone,
            CONCAT(
               '[',
                GROUP_CONCAT(
                    CONCAT(
                       '{"productName": "', p.name, '","productPrice": ', p.price, ',"productQuantity": ', op.productQuantity, '}'
                    ) SEPARATOR ','
                ),
               ']'
            ) AS products
        FROM
            orders o
        JOIN
            orderProduct op ON o.ID = op.orderID
        JOIN
            products p ON op.productID = p.ID
        WHERE
            o.ID = $order_id
        GROUP BY
            o.ID
        ORDER BY
            o.ID;
    SQL;
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $order = array();
        while ($row = $result->fetch_assoc()) {
            $row['products'] = json_decode(stripslashes($row['products']), true);
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
    if(!isset($data->orderProducts) || !isset($data->name) || !isset($data->phone)){
        http_response_code(400);
        echo json_encode(array("message" => "Parameters missing (OrderProducts, name or phone)"));
        return;
    }
    $orderProducts = $data->orderProducts;
    $orderName = $data->name;
    $phone = $data->phone;

    // Insert the order
    $orderInsertQuery = "INSERT INTO orders (name, phone, total) VALUES ('$orderName', '$phone', null)";
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
    $orderProducts = $data->orderProducts;

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

function patchOrder($order_id, $data)
{
    global $conn;
    $state = mysqli_real_escape_string($conn, $data->state);
    if($state != "finished" && $state != "canceled"){
        http_response_code(400);
        echo json_encode(array("message" => "Error patching order: invalid state"));
    }
    else{
        $patchOrderProductsQuery = <<<SQL
             UPDATE orders SET state = "$state" WHERE ID = $order_id 
        SQL;
        mysqli_query($conn, $patchOrderProductsQuery);

        if (mysqli_affected_rows($conn) > 0) {
            echo json_encode(array("message" => "Order patched successfully"));
        } else {
            http_response_code(500);
            echo json_encode(array("message" => "Error patching order: " . mysqli_error($conn) . " La consulta fue: ".$patchOrderProductsQuery));
        }
    }
}