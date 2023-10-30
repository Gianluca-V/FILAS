<?php
switch ($request_method) {
    case 'GET':
        // Handle GET request for Workshops
        if ($parts[4] !== "") {
            $family_id = intval($parts[4]);
            $sql = "SELECT * FROM Family WHERE ID = $family_id";
        } else {
            $sql = "SELECT * FROM Family";
        }

        $result = $conn->query($sql);

        if ($result->num_rows > 0) {
            $Workshops = array();
            while ($row = $result->fetch_assoc()) {
                $Workshops[] = $row;
            }
            echo json_encode($Workshops);
        } else {
            http_response_code(404);
            echo json_encode(array("message" => "No workshops found"));
        }
        break;

    case 'POST':
        // Handle POST request to add a new Workshops item to the "Workshops" table
        $data = json_decode(file_get_contents("php://input"));

        if (isset($data->Body) && isset($data->Image)) {
            $body = $conn->real_escape_string($data->Body);
            $image = $conn->real_escape_string($data->Image);

            $sql = "INSERT INTO family (Body, Image) VALUES ('$body', '$image')";

            if ($conn->query($sql) === TRUE) {
                http_response_code(201);
                echo json_encode(array("message" => "Workshops added successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error adding Workshops: " . $conn->error));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing Body or Image parameter"));
        }
        break;

    case 'PUT':
        // Handle PUT request to update a Workshops item by ID in the "Workshops" table
        if (isset($_GET['id'])) {
            $Workshops_id = intval($_GET['id']);
            $data = json_decode(file_get_contents("php://input"));

            if (isset($data->Body) && isset($data->Image)) {
                $body = $conn->real_escape_string($data->Body);
                $image = $conn->real_escape_string($data->Image);

                $sql = "UPDATE Workshops SET Body = '$body', Image = '$image' WHERE ID = $Workshops_id";

                if ($conn->query($sql) === TRUE) {
                    http_response_code(200);
                    echo json_encode(array("message" => "Workshops updated successfully"));
                } else {
                    http_response_code(500);
                    echo json_encode(array("message" => "Error updating Workshops: " . $conn->error));
                }
            } else {
                http_response_code(400);
                echo json_encode(array("message" => "Missing Body or Image parameter"));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing ID parameter"));
        }
        break;

    case 'DELETE':
        // Handle DELETE request to delete a Workshops item by ID from the "Workshops" table
        if (isset($_GET['id'])) {
            $Workshops_id = intval($_GET['id']);
            $sql = "DELETE FROM Workshops WHERE ID = $Workshops_id";

            if ($conn->query($sql) === TRUE) {
                http_response_code(200);
                echo json_encode(array("message" => "Workshops deleted successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error deleting Workshops: " . $conn->error));
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
