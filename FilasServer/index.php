<?php
// Database configuration
$db_host = 'localhost';
$db_name = 'filas';
$db_user = 'root';
$db_pass = '';

// Server config to avoid CORS
header("Access-Control-Allow-Origin: *");
header("Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS");
header("Access-Control-Allow-Headers: Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With");

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(204);
    exit();
}

require __DIR__ . '/vendor/autoload.php';
use Firebase\JWT\JWT;
use Firebase\JWT\Key;
include('JWT.php');

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
if ($parts[2] !== "api") {
    http_response_code(404);
    echo json_encode(array("message" => "Not Found","request"=>$request_uri,"parts" => print_r($parts)));
    exit();
}

// Determine the table based on the URI
$table = $parts[3];

$request_method = $_SERVER['REQUEST_METHOD'];

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
    case 'family':
        include('family.php'); // Include a separate script for the Family table
        break;
    case 'admins':
        include('admins.php'); // Include a separate script for the Admins table
        break;
    case 'organizations':
        include('organizations.php'); // Include a separate script for the Organizations table
        break;
    case 'orders':
        include('orders.php'); // Include a separate script for the Orders table
        break;
    default:
        http_response_code(404);
        echo json_encode(array("message" => "Table not found"));
        break;
}

// Close the database connection
$conn->close();
?>