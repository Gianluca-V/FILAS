<?php
switch ($request_method) {
    case 'GET':
        // Handle GET request for gallery
        if (isset($_GET['id'])) {
            $image_id = intval($_GET['id']);
            $sql = "SELECT * FROM Gallery WHERE ID = $image_id";
        } else {
            $sql = "SELECT * FROM Gallery";
        }

        $result = $conn->query($sql);

        if ($result->num_rows > 0) {
            $images = array();
            while ($row = $result->fetch_assoc()) {
                $images[] = $row;
            }
            echo json_encode($images);
        } else {
            http_response_code(404);
            echo json_encode(array("message" => "No images found"));
        }
        break;

    case 'POST':
        // Handle POST request to add a new image to the "Gallery" table
        $data = json_decode(file_get_contents("php://input"));

        if (isset($data->Image)) {
            $image = $conn->real_escape_string($data->Image);
            $sql = "INSERT INTO Gallery (Image) VALUES ('$image')";

            if ($conn->query($sql) === TRUE) {
                http_response_code(201);
                echo json_encode(array("message" => "Image added successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error adding image: " . $conn->error));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing Image parameter"));
        }
        break;

    case 'PUT':
        // Handle PUT request to update an image by ID in the "Gallery" table
        if (isset($_GET['id'])) {
            $image_id = intval($_GET['id']);
            $data = json_decode(file_get_contents("php://input"));

            if (isset($data->Image)) {
                $image = $conn->real_escape_string($data->Image);
                $sql = "UPDATE Gallery SET Image = '$image' WHERE ID = $image_id";

                if ($conn->query($sql) === TRUE) {
                    http_response_code(200);
                    echo json_encode(array("message" => "Image updated successfully"));
                } else {
                    http_response_code(500);
                    echo json_encode(array("message" => "Error updating image: " . $conn->error));
                }
            } else {
                http_response_code(400);
                echo json_encode(array("message" => "Missing Image parameter"));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing ID parameter"));
        }
        break;

    case 'DELETE':
        // Handle DELETE request to delete an image by ID from the "Gallery" table
        if (isset($_GET['id'])) {
            $image_id = intval($_GET['id']);
            $sql = "DELETE FROM Gallery WHERE ID = $image_id";

            if ($conn->query($sql) === TRUE) {
                http_response_code(200);
                echo json_encode(array("message" => "Image deleted successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error deleting image: " . $conn->error));
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
