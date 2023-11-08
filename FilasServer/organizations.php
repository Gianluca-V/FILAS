<?php
switch ($request_method) {
    case 'GET':
        // Handle GET request for organizations
        if ($parts[4] !== "") {
            $organizations_id = intval($parts[4]);
            $sql = "SELECT * FROM Organizations WHERE ID = $organizations_id";
        } else {
            $sql = "SELECT * FROM Organizations";
        }

        $result = $conn->query($sql);

        if ($result->num_rows > 0) {
            $organizations = array();
            while ($row = $result->fetch_assoc()) {
                $organizations[] = $row;
            }
            echo json_encode($organizations);
        } else {
            http_response_code(404);
            echo json_encode(array("message" => "No organizations found"));
        }
        break;

    case 'POST':
        // Handle POST request to add a new organizations item to the "organizations" table
        $data = json_decode(file_get_contents("php://input"));
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

        if (isset($data->Title) && isset($data->Image)) {
            $title = $conn->real_escape_string($data->Title);
            $description = $conn->real_escape_string($data->Description);
            $image = $conn->real_escape_string($data->Image);

            $sql = "INSERT INTO Organizations (Title, Description, Image) VALUES ('$title', '$description', '$image')";

            if ($conn->query($sql) === TRUE) {
                http_response_code(201);
                echo json_encode(array("message" => "Organizations added successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error adding organizations: " . $conn->error));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing Title, or Image parameter"));
        }
        break;

    case 'PUT':
        // Handle PUT request to update a organizations item by ID in the "organizations" table
        if (intval($parts[4]) !== "") {
            $organizations_id = intval($parts[4]);
            $data = json_decode(file_get_contents("php://input"));
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

            if (isset($data->Title) && isset($data->Image)) {
                $title = $conn->real_escape_string($data->Title);
                $description = $conn->real_escape_string($data->Description);
                $image = $conn->real_escape_string($data->Image);

                $sql = "UPDATE Organizations SET Title = '$title', Description = '$description', Image = '$image' WHERE ID = $organizations_id";

                if ($conn->query($sql) === TRUE) {
                    http_response_code(200);
                    echo json_encode(array("message" => "Organizations updated successfully"));
                } else {
                    http_response_code(500);
                    echo json_encode(array("message" => "Error updating organizations: " . $conn->error));
                }
            } else {
                http_response_code(400);
                echo json_encode(array("message" => "Missing Title, or Image parameter"));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing ID parameter"));
        }
        break;

    case 'DELETE':
        // Handle DELETE request to delete a organizations item by ID from the "organizations" table
        if (intval($parts[4]) !== "") {
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
            $organizations_id = intval($parts[4]);
            $sql = "DELETE FROM Organizations WHERE ID = $organizations_id";

            if ($conn->query($sql) === TRUE) {
                http_response_code(200);
                echo json_encode(array("message" => "organizations deleted successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error deleting organizations: " . $conn->error));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing ID parameter"));
        }
        break;

    default:
        http_response_code(405);
        echo json_encode(array("message" => "Method not allowed"));
        break;
}

?>
