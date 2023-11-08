<?php

use Firebase\JWT\JWT;
use Firebase\JWT\Key;

function GetJWTKey(){
    $key = 'Fundación_para_la_Integración_Laboral_y_Autonomía_Social_de_Personas_con_Necesidades_Especiales';
    return $key;
}

function isTokenValid($token)
{
    try {
        $decoded = JWT::decode($token, new Key(GetJWTKey(), 'HS256'));
        // Token is valid
        return true;
    } catch (Exception $e) {
        // Token is invalid
        return false;
    }
}

function CreateToken($id)
{
    $payload = [
        'exp' => strtotime("now") + 3600 * 3, // 3 hours before expiration,
        'data' => $id
    ];

    $jwt = JWT::encode($payload, GetJWTKey(), 'HS256');

    return $jwt;
}

function TokenValidationResponse($token){
    if (!isTokenValid($token)) {
        // Token is invalid
        echo json_encode(array("message" => "Token is invalid"));
        return false;
    }
    //echo json_encode(array("message" => "Token is valid"));
    return true;
}