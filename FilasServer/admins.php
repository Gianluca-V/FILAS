<?php
switch ($request_method) {
    case 'GET':
        // Get all Admins or a specific User by ID
        if ($parts[4] !== "") {
            $Admins_id = intval($parts[4]);
            getUser($Admins_id);
        } else {
            getAdmins();
        }
        break;
    case 'POST':
        // Create a new User or Check User for login
        $data = json_decode(file_get_contents("php://input"));
        if (isset($data->login) && $data->login === true) {
            // Login check
            checkUserForLogin($data);
        } else {
            // Create a new User
            createUser($data);
        }
        break;
    case 'PUT':
        // Update a User by ID
        $data = json_decode(file_get_contents("php://input"));
        $Admin_id = intval($parts[4]);
        updateUser($Admin_id, $data);
        break;
    case 'DELETE':
        // Delete a User by ID
        $Admin_id = intval($parts[4]);
        deleteUser($Admin_id);
        break;
    default:
        http_response_code(405);
        echo json_encode(array("message" => "Method not allowed"));
        break;
}

function getAdmins()
{
    global $conn;
    $sql = "SELECT * FROM Admins";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $Admins = array();
        while ($row = $result->fetch_assoc()) {
            $Admins[] = $row;
        }
        echo json_encode($Admins);
    } else {
        echo json_encode(array());
    }
}

function getUser($Admin_id)
{
    global $conn;
    $sql = "SELECT * FROM Admins WHERE ID = $Admin_id";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        echo json_encode($row);
    } else {
        http_response_code(404);
        echo json_encode(array("message" => "User not found"));
    }
}

function createUser($data)
{
    global $conn;
    // Assuming $data contains the necessary fields for creating a User
    $username = $conn->real_escape_string($data->username);
    $rawPassword = $data->password;
    
    // Generate a random salt for each user
    $salt = bin2hex(random_bytes(16));
    
    // Combine the raw password and salt, and then hash it with SHA-256
    $hashedPassword = hash('sha256', $salt . $rawPassword);

    $sql = "INSERT INTO Admins (username, password, Salt) 
            VALUES ('$username', '$hashedPassword', '$salt')";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "User created successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error creating User: " . $conn->error));
    }
}

function updateUser($Admin_id, $data)
{
    global $conn;
    $id = intval($Admin_id);
    $username = $conn->real_escape_string($data->username);
    $password = floatval($data->password);

    $sql = "UPDATE Admins SET username='$username', password='$password' WHERE ID = $id";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "User updated successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error updating User: " . $conn->error));
    }
}

function deleteUser($Admin_id)
{
    global $conn;
    $id = intval($Admin_id);
    $sql = "DELETE FROM Admins WHERE ID = $id";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "User deleted successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error deleting User: " . $conn->error));
    }
}

function checkUserForLogin($data)
{
    global $conn;
    $username = $conn->real_escape_string($data->username);
    $rawPassword = $data->password;

    $sql = "SELECT * FROM Admins WHERE username = '$username'";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        $hashedPassword = $row['password'];
        $salt = $row['Salt'];
        
        // Combine the entered password with the stored salt and hash it
        $enteredPasswordHash = hash('sha256', $salt . $rawPassword);
        
        // Compare the entered password hash with the stored hash
        if ($enteredPasswordHash === $hashedPassword) {
            echo json_encode(array("message" => "Login successful"));
        } else {
            http_response_code(401);
            echo json_encode(array("message" => "Login failed"));
        }
    } else {
        http_response_code(401);
        echo json_encode(array("message" => "Login failed"));
    }
}
?>