<?php
switch ($request_method) {
    case 'GET':
        // Handle GET request for news
        if ($parts[4] !== "") {
            $news_id = intval($parts[4]);
            $sql = "SELECT * FROM News WHERE ID = $news_id";
        } else {
            $sql = "SELECT * FROM News";
        }

        $result = $conn->query($sql);

        if ($result->num_rows > 0) {
            $news = array();
            while ($row = $result->fetch_assoc()) {
                $news[] = $row;
            }
            echo json_encode($news);
        } else {
            http_response_code(404);
            echo json_encode(array("message" => "No news found"));
        }
        break;

    case 'POST':
        // Handle POST request to add a new news item to the "News" table
        $data = json_decode(file_get_contents("php://input"));

        if (isset($data->Title) && isset($data->Body) && isset($data->Image)) {
            $title = $conn->real_escape_string($data->Title);
            $body = $conn->real_escape_string($data->Body);
            $image = $conn->real_escape_string($data->Image);

            $sql = "INSERT INTO News (Title, Body, Image) VALUES ('$title', '$body', '$image')";

            if ($conn->query($sql) === TRUE) {
                http_response_code(201);
                echo json_encode(array("message" => "News added successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error adding news: " . $conn->error));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing Title, Body, or Image parameter"));
        }
        break;

    case 'PUT':
        // Handle PUT request to update a news item by ID in the "News" table
        if (isset($_GET['id'])) {
            $news_id = intval($_GET['id']);
            $data = json_decode(file_get_contents("php://input"));

            if (isset($data->Title) && isset($data->Body) && isset($data->Image)) {
                $title = $conn->real_escape_string($data->Title);
                $body = $conn->real_escape_string($data->Body);
                $image = $conn->real_escape_string($data->Image);

                $sql = "UPDATE News SET Title = '$title', Body = '$body', Image = '$image' WHERE ID = $news_id";

                if ($conn->query($sql) === TRUE) {
                    http_response_code(200);
                    echo json_encode(array("message" => "News updated successfully"));
                } else {
                    http_response_code(500);
                    echo json_encode(array("message" => "Error updating news: " . $conn->error));
                }
            } else {
                http_response_code(400);
                echo json_encode(array("message" => "Missing Title, Body, or Image parameter"));
            }
        } else {
            http_response_code(400);
            echo json_encode(array("message" => "Missing ID parameter"));
        }
        break;

    case 'DELETE':
        // Handle DELETE request to delete a news item by ID from the "News" table
        if (isset($_GET['id'])) {
            $news_id = intval($_GET['id']);
            $sql = "DELETE FROM News WHERE ID = $news_id";

            if ($conn->query($sql) === TRUE) {
                http_response_code(200);
                echo json_encode(array("message" => "News deleted successfully"));
            } else {
                http_response_code(500);
                echo json_encode(array("message" => "Error deleting news: " . $conn->error));
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
