<?php
switch ($request_method) {
    case 'GET':
        // Get all mails or a specific mail by ID
        if ($parts[4] !== "") {
            $mail_id = intval($parts[4]);
            getMail($mail_id);
        } else {
            getMails();
        }
        break;
    case 'POST':
        // Create a new mail
        $data = json_decode(file_get_contents("php://input"));
        $headers = getallheaders();
        if(isset($headers['Authorization'])) {
            $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
            TokenValidationResponse($token);
        }
        createMail($data);
        break;
    case 'DELETE':
        // Delete a mail by ID
        $headers = getallheaders();
        if(!isset($headers['Authorization'])) {
            return http_response_code(400);
            break;
        }
        
        $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
        if(!TokenValidationResponse($token)){
            return http_response_code(401);
            break;
        }
        $mail_id = intval($_GET['id']);
        deleteMail($mail_id);
        break;
    // ... (OPTIONS handling remains unchanged)
    default:
        http_response_code(405);
        echo json_encode(array("message" => "Method not allowed"));
        break;
}

function getMails()
{
    global $conn;
    $sql = "SELECT * FROM Mails";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $mails = array();
        while ($row = $result->fetch_assoc()) {
            $mails[] = $row;
        }
        echo json_encode($mails);
    } else {
        echo json_encode(array());
    }
}

function getMail($mail_id)
{
    global $conn;
    $sql = "SELECT * FROM Mails WHERE ID = $mail_id";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        echo json_encode($row);
    } else {
        http_response_code(404);
        echo json_encode(array("message" => "Mail not found"));
    }
}

function createMail($data)
{
    global $conn;
    // Assuming $data contains the necessary fields for creating a mail
    $mail = $conn->real_escape_string($data->Mail);

    $sql = "INSERT INTO Mails (Mail) VALUES ('$mail')";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "Mail created successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error creating mail: " . $conn->error));
    }
}

function deleteMail($mail_id)
{
    global $conn;
    $sql = "DELETE FROM Mails WHERE ID = $mail_id";

    if ($conn->query($sql) === TRUE) {
        echo json_encode(array("message" => "Mail deleted successfully"));
    } else {
        http_response_code(500);
        echo json_encode(array("message" => "Error deleting mail: " . $conn->error));
    }
}

?>