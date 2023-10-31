-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Servidor: 127.0.0.1
-- Tiempo de generación: 31-10-2023 a las 15:07:52
-- Versión del servidor: 10.4.28-MariaDB
-- Versión de PHP: 8.2.4

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Base de datos: `filas`
--

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `gallery`
--

CREATE TABLE `gallery` (
  `ID` int(11) NOT NULL,
  `Image` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Volcado de datos para la tabla `gallery`
--

INSERT INTO `gallery` (`ID`, `Image`) VALUES
(1, 'assets/galeria-1.jpg'),
(2, 'assets/galeria-2.jpg'),
(3, 'assets/galeria-3.jpg'),
(4, 'assets/galeria-4.jpg'),
(5, 'assets/galeria-5.jpg'),
(6, 'assets/galeria-6.jpg'),
(7, 'assets/galeria-7.jpg'),
(8, 'assets/galeria-8.jpg'),
(9, 'assets/galeria-9.jpg');

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `mails`
--

CREATE TABLE `mails` (
  `ID` int(11) NOT NULL,
  `Mail` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `news`
--

CREATE TABLE `news` (
  `ID` int(11) NOT NULL,
  `Title` varchar(99) NOT NULL,
  `Body` text DEFAULT NULL,
  `Image` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `products`
--

CREATE TABLE `products` (
  `ID` int(11) NOT NULL,
  `Name` varchar(99) NOT NULL,
  `Price` double NOT NULL,
  `Stock` int(11) NOT NULL,
  `Image` text DEFAULT NULL,
  `Description` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Volcado de datos para la tabla `products`
--

INSERT INTO `products` (`ID`, `Name`, `Price`, `Stock`, `Image`, `Description`) VALUES
(1, 'Mermelada de pera', 600, 1, '', ''),
(2, 'Mermelada de naranja inglesa', 600, 1, '', ''),
(3, 'Mermelada de tomate', 600, 1, '', ''),
(4, 'Mermelada de zapallo', 600, 1, '', ''),
(5, 'Mermelada de ciruela', 800, 1, '', ''),
(6, 'Mermelada de higo', 800, 1, '', ''),
(7, 'Encurtido de berenjena', 700, 1, '', ''),
(8, 'Encurtido de berenjena grande', 1000, 1, '', ''),
(9, 'Encurtido de champiñón', 700, 1, '', ''),
(10, 'Encurtido de champiñón grande', 1000, 1, '', ''),
(11, 'Bandeja de alfajores (6 u.)', 850, 1, '', ''),
(12, 'Pre pizza de tomate', 350, 1, '', ''),
(13, 'Pre pizza de cebolla', 350, 1, '', ''),
(14, 'Pasta frola', 900, 1, '', ''),
(15, 'Pan de figasa (1kg)', 600, 1, '', ''),
(16, 'Pan integral (800gr)', 850, 1, '', ''),
(17, 'Tarta de coco y dulce de leche', 1200, 1, '', '');

--
-- Índices para tablas volcadas
--

--
-- Indices de la tabla `gallery`
--
ALTER TABLE `gallery`
  ADD PRIMARY KEY (`ID`);

--
-- Indices de la tabla `mails`
--
ALTER TABLE `mails`
  ADD PRIMARY KEY (`ID`);

--
-- Indices de la tabla `news`
--
ALTER TABLE `news`
  ADD PRIMARY KEY (`ID`);

--
-- Indices de la tabla `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`ID`);

--
-- AUTO_INCREMENT de las tablas volcadas
--

--
-- AUTO_INCREMENT de la tabla `gallery`
--
ALTER TABLE `gallery`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT de la tabla `mails`
--
ALTER TABLE `mails`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de la tabla `news`
--
ALTER TABLE `news`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de la tabla `products`
--
ALTER TABLE `products`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
