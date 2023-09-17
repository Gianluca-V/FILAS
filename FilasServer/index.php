<?php
// Database configuration
$db_host = 'localhost';
$db_name = 'filas';
$db_user = 'root';
$db_pass = '';

// Server config to avoid CORS
header("Access-Control-Allow-Origin: http://127.0.0.1:5501");
header("Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS");
header("Access-Control-Allow-Headers: Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With");

// Create a database connection
$conn = new mysqli($db_host, $db_user, $db_pass, $db_name);

// Check connection
if ($conn->connect_error) {
    die("Connection failed: " . $conn->connect_error);
}

// Define API routes
$request_uri = $_SERVER['REQUEST_URI'];
$parts = explode('/', $request_uri);

// Check for the base path, e.g., /api
if ($parts[1] !== 'api') {
    http_response_code(404);
    echo json_encode(array("message" => "Not Found"));
    exit();
}

// Determine the table based on the URI
$table = $parts[2];

switch ($table) {
    case 'products':
        include('products.php'); // Include a separate script for the Products table
        break;
    case 'mails':
        include('mails.php'); // Include a separate script for the Mails table
        break;
    case 'gallery':
        include('gallery.php'); // Include a separate script for the Gallery table
        break;
    case 'news':
        include('news.php'); // Include a separate script for the News table
        break;
    default:
        http_response_code(404);
        echo json_encode(array("message" => "Table not found"));
        break;
}

// Close the database connection
$conn->close();
?>