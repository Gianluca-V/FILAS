-- phpMyAdmin SQL Dump
-- version 5.1.0
-- https://www.phpmyadmin.net/
--
-- Servidor: 127.0.0.1
-- Tiempo de generación: 21-11-2023 a las 01:19:36
-- Versión del servidor: 10.4.19-MariaDB
-- Versión de PHP: 8.0.6

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
-- Estructura de tabla para la tabla `admins`
--

CREATE TABLE `admins` (
  `ID` int(11) DEFAULT NULL,
  `username` text DEFAULT NULL,
  `password` text DEFAULT NULL,
  `salt` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `admins`
--

INSERT INTO `admins` (`ID`, `username`, `password`, `salt`) VALUES
(1543, 'FilasAdmin', '8b4abda03a7fb4eee79d8fb7c07f88497f28b0728779c84fab84a07b3f824cef', 'bcf76de2e59270b37b7e76d22b53917b');

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `family`
--

CREATE TABLE `family` (
  `ID` int(11) NOT NULL,
  `Image` text DEFAULT NULL,
  `Body` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `family`
--

INSERT INTO `family` (`ID`, `Image`, `Body`) VALUES
(1, 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY=', 'aqui va la descripcion'),
(2, 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY=', 'aqui va la descripcion 2'),
(3, 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY=', 'aqui va la descripcion 3');

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `gallery`
--

CREATE TABLE `gallery` (
  `ID` int(11) NOT NULL,
  `Image` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
-- Estructura de tabla para la tabla `news`
--

CREATE TABLE `news` (
  `ID` int(11) NOT NULL,
  `Title` varchar(99) NOT NULL,
  `Body` text DEFAULT NULL,
  `Image` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `news`
--

INSERT INTO `news` (`ID`, `Title`, `Body`, `Image`) VALUES
(1, 'Noticia de prueba 1', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat. Aliquam erat tortor, faucibus sit amet lectus sit amet, tristique mollis magna. Aliquam sollicitudin quis nibh sed dictum.Sed in tempus nisi. Nam in magna pharetra, tempus ante non, tristique nulla. Vivamus consequat lectus eu nulla tristique ultricies. In a sapien vitae nibh accumsan aliquet. Etiam pharetra nisi sit amet consectetur laoreet. Suspendisse id tempor nunc. Vivamus pretium posuere nunc id malesuada. Quisque auctor ut nibh id sollicitudin. In posuere nunc nec imperdiet malesuada. Ut dignissim purus id venenatis porta. In hendrerit purus massa, non tristique eros fermentum nec. Curabitur placerat id ligula sed pharetra. Quisque ac felis nisi. Maecenas gravida cursus placerat. In nisl urna, scelerisque et leo vel, dictum pharetra velit.Nam augue est, elementum eu mauris sit amet, euismod lobortis nunc. Sed tincidunt augue at dolor bibendum, id euismod eros ullamcorper. Morbi elementum libero fringilla lectus volutpat, et rutrum risus blandit. Integer vehicula mi non odio egestas rutrum. Donec iaculis ante ut tincidunt mattis. Praesent sagittis at enim sed auctor. Nulla facilisi. Duis vitae risus sit amet urna convallis cursus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Aenean vitae sodales ex, ut semper magna. Aenean rhoncus dignissim ligula, vitae tincidunt nunc vestibulum eget. Morbi gravida diam sem, congue ultrices ipsum interdum at. Aliquam pellentesque eros sed sapien rhoncus, sed varius ante cursus.Phasellus euismod bibendum lorem at rhoncus. In quis ipsum congue, facilisis purus tincidunt, efficitur felis. Morbi eu laoreet massa. Sed nec metus mauris. Ut dictum sed turpis vel lobortis. Integer facilisis odio vel interdum cursus. Pellentesque accumsan purus non lectus tempus semper. Duis libero sapien, sagittis at scelerisque eu, aliquet dapibus metus. Vestibulum sit amet augue vel nisi tincidunt dapibus. Sed nunc arcu, rutrum in dictum id, euismod in nibh. Curabitur sit amet tempus purus, quis faucibus justo. Mauris et lacus pellentesque libero viverra dignissim. Morbi mattis nisl eu enim placerat, ut consectetur sem sodales. Cras molestie, elit id maximus viverra, magna ex fermentum neque, vehicula pretium nulla erat vel nibh.Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Cras consequat nibh sit amet auctor bibendum. Nunc pharetra interdum metus, id pretium erat efficitur sit amet. Aliquam erat volutpat. Aliquam ac iaculis arcu. Suspendisse maximus lacinia mi, ac vehicula odio malesuada a. Aliquam vitae porttitor lacus. Nulla facilisi. Nulla facilisi. Mauris finibus tortor quam, nec accumsan augue rhoncus euismod.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(2, 'Noticia de prueba 2', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat. Aliquam erat tortor, faucibus sit amet lectus sit amet, tristique mollis magna. Aliquam sollicitudin quis nibh sed dictum.Sed in tempus nisi. Nam in magna pharetra, tempus ante non, tristique nulla. Vivamus consequat lectus eu nulla tristique ultricies. In a sapien vitae nibh accumsan aliquet. Etiam pharetra nisi sit amet consectetur laoreet. Suspendisse id tempor nunc. Vivamus pretium posuere nunc id malesuada. Quisque auctor ut nibh id sollicitudin. In posuere nunc nec imperdiet malesuada. Ut dignissim purus id venenatis porta. In hendrerit purus massa, non tristique eros fermentum nec. Curabitur placerat id ligula sed pharetra. Quisque ac felis nisi. Maecenas gravida cursus placerat. In nisl urna, scelerisque et leo vel, dictum pharetra velit.Nam augue est, elementum eu mauris sit amet, euismod lobortis nunc. Sed tincidunt augue at dolor bibendum, id euismod eros ullamcorper. Morbi elementum libero fringilla lectus volutpat, et rutrum risus blandit. Integer vehicula mi non odio egestas rutrum. Donec iaculis ante ut tincidunt mattis. Praesent sagittis at enim sed auctor. Nulla facilisi. Duis vitae risus sit amet urna convallis cursus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Aenean vitae sodales ex, ut semper magna. Aenean rhoncus dignissim ligula, vitae tincidunt nunc vestibulum eget. Morbi gravida diam sem, congue ultrices ipsum interdum at. Aliquam pellentesque eros sed sapien rhoncus, sed varius ante cursus.Phasellus euismod bibendum lorem at rhoncus. In quis ipsum congue, facilisis purus tincidunt, efficitur felis. Morbi eu laoreet massa. Sed nec metus mauris. Ut dictum sed turpis vel lobortis. Integer facilisis odio vel interdum cursus. Pellentesque accumsan purus non lectus tempus semper. Duis libero sapien, sagittis at scelerisque eu, aliquet dapibus metus. Vestibulum sit amet augue vel nisi tincidunt dapibus. Sed nunc arcu, rutrum in dictum id, euismod in nibh. Curabitur sit amet tempus purus, quis faucibus justo. Mauris et lacus pellentesque libero viverra dignissim. Morbi mattis nisl eu enim placerat, ut consectetur sem sodales. Cras molestie, elit id maximus viverra, magna ex fermentum neque, vehicula pretium nulla erat vel nibh.Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Cras consequat nibh sit amet auctor bibendum. Nunc pharetra interdum metus, id pretium erat efficitur sit amet. Aliquam erat volutpat. Aliquam ac iaculis arcu. Suspendisse maximus lacinia mi, ac vehicula odio malesuada a. Aliquam vitae porttitor lacus. Nulla facilisi. Nulla facilisi. Mauris finibus tortor quam, nec accumsan augue rhoncus euismod.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(3, 'Noticia de prueba 3', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat. Aliquam erat tortor, faucibus sit amet lectus sit amet, tristique mollis magna. Aliquam sollicitudin quis nibh sed dictum.Sed in tempus nisi. Nam in magna pharetra, tempus ante non, tristique nulla. Vivamus consequat lectus eu nulla tristique ultricies. In a sapien vitae nibh accumsan aliquet. Etiam pharetra nisi sit amet consectetur laoreet. Suspendisse id tempor nunc. Vivamus pretium posuere nunc id malesuada. Quisque auctor ut nibh id sollicitudin. In posuere nunc nec imperdiet malesuada. Ut dignissim purus id venenatis porta. In hendrerit purus massa, non tristique eros fermentum nec. Curabitur placerat id ligula sed pharetra. Quisque ac felis nisi. Maecenas gravida cursus placerat. In nisl urna, scelerisque et leo vel, dictum pharetra velit.Nam augue est, elementum eu mauris sit amet, euismod lobortis nunc. Sed tincidunt augue at dolor bibendum, id euismod eros ullamcorper. Morbi elementum libero fringilla lectus volutpat, et rutrum risus blandit. Integer vehicula mi non odio egestas rutrum. Donec iaculis ante ut tincidunt mattis. Praesent sagittis at enim sed auctor. Nulla facilisi. Duis vitae risus sit amet urna convallis cursus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Aenean vitae sodales ex, ut semper magna. Aenean rhoncus dignissim ligula, vitae tincidunt nunc vestibulum eget. Morbi gravida diam sem, congue ultrices ipsum interdum at. Aliquam pellentesque eros sed sapien rhoncus, sed varius ante cursus.Phasellus euismod bibendum lorem at rhoncus. In quis ipsum congue, facilisis purus tincidunt, efficitur felis. Morbi eu laoreet massa. Sed nec metus mauris. Ut dictum sed turpis vel lobortis. Integer facilisis odio vel interdum cursus. Pellentesque accumsan purus non lectus tempus semper. Duis libero sapien, sagittis at scelerisque eu, aliquet dapibus metus. Vestibulum sit amet augue vel nisi tincidunt dapibus. Sed nunc arcu, rutrum in dictum id, euismod in nibh. Curabitur sit amet tempus purus, quis faucibus justo. Mauris et lacus pellentesque libero viverra dignissim. Morbi mattis nisl eu enim placerat, ut consectetur sem sodales. Cras molestie, elit id maximus viverra, magna ex fermentum neque, vehicula pretium nulla erat vel nibh.Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Cras consequat nibh sit amet auctor bibendum. Nunc pharetra interdum metus, id pretium erat efficitur sit amet. Aliquam erat volutpat. Aliquam ac iaculis arcu. Suspendisse maximus lacinia mi, ac vehicula odio malesuada a. Aliquam vitae porttitor lacus. Nulla facilisi. Nulla facilisi. Mauris finibus tortor quam, nec accumsan augue rhoncus euismod.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(4, 'Noticia de prueba 4', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat. Aliquam erat tortor, faucibus sit amet lectus sit amet, tristique mollis magna. Aliquam sollicitudin quis nibh sed dictum.Sed in tempus nisi. Nam in magna pharetra, tempus ante non, tristique nulla. Vivamus consequat lectus eu nulla tristique ultricies. In a sapien vitae nibh accumsan aliquet. Etiam pharetra nisi sit amet consectetur laoreet. Suspendisse id tempor nunc. Vivamus pretium posuere nunc id malesuada. Quisque auctor ut nibh id sollicitudin. In posuere nunc nec imperdiet malesuada. Ut dignissim purus id venenatis porta. In hendrerit purus massa, non tristique eros fermentum nec. Curabitur placerat id ligula sed pharetra. Quisque ac felis nisi. Maecenas gravida cursus placerat. In nisl urna, scelerisque et leo vel, dictum pharetra velit.Nam augue est, elementum eu mauris sit amet, euismod lobortis nunc. Sed tincidunt augue at dolor bibendum, id euismod eros ullamcorper. Morbi elementum libero fringilla lectus volutpat, et rutrum risus blandit. Integer vehicula mi non odio egestas rutrum. Donec iaculis ante ut tincidunt mattis. Praesent sagittis at enim sed auctor. Nulla facilisi. Duis vitae risus sit amet urna convallis cursus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Aenean vitae sodales ex, ut semper magna. Aenean rhoncus dignissim ligula, vitae tincidunt nunc vestibulum eget. Morbi gravida diam sem, congue ultrices ipsum interdum at. Aliquam pellentesque eros sed sapien rhoncus, sed varius ante cursus.Phasellus euismod bibendum lorem at rhoncus. In quis ipsum congue, facilisis purus tincidunt, efficitur felis. Morbi eu laoreet massa. Sed nec metus mauris. Ut dictum sed turpis vel lobortis. Integer facilisis odio vel interdum cursus. Pellentesque accumsan purus non lectus tempus semper. Duis libero sapien, sagittis at scelerisque eu, aliquet dapibus metus. Vestibulum sit amet augue vel nisi tincidunt dapibus. Sed nunc arcu, rutrum in dictum id, euismod in nibh. Curabitur sit amet tempus purus, quis faucibus justo. Mauris et lacus pellentesque libero viverra dignissim. Morbi mattis nisl eu enim placerat, ut consectetur sem sodales. Cras molestie, elit id maximus viverra, magna ex fermentum neque, vehicula pretium nulla erat vel nibh.Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Cras consequat nibh sit amet auctor bibendum. Nunc pharetra interdum metus, id pretium erat efficitur sit amet. Aliquam erat volutpat. Aliquam ac iaculis arcu. Suspendisse maximus lacinia mi, ac vehicula odio malesuada a. Aliquam vitae porttitor lacus. Nulla facilisi. Nulla facilisi. Mauris finibus tortor quam, nec accumsan augue rhoncus euismod.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(5, 'Noticia de prueba 5', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat. Aliquam erat tortor, faucibus sit amet lectus sit amet, tristique mollis magna. Aliquam sollicitudin quis nibh sed dictum.Sed in tempus nisi. Nam in magna pharetra, tempus ante non, tristique nulla. Vivamus consequat lectus eu nulla tristique ultricies. In a sapien vitae nibh accumsan aliquet. Etiam pharetra nisi sit amet consectetur laoreet. Suspendisse id tempor nunc. Vivamus pretium posuere nunc id malesuada. Quisque auctor ut nibh id sollicitudin. In posuere nunc nec imperdiet malesuada. Ut dignissim purus id venenatis porta. In hendrerit purus massa, non tristique eros fermentum nec. Curabitur placerat id ligula sed pharetra. Quisque ac felis nisi. Maecenas gravida cursus placerat. In nisl urna, scelerisque et leo vel, dictum pharetra velit.Nam augue est, elementum eu mauris sit amet, euismod lobortis nunc. Sed tincidunt augue at dolor bibendum, id euismod eros ullamcorper. Morbi elementum libero fringilla lectus volutpat, et rutrum risus blandit. Integer vehicula mi non odio egestas rutrum. Donec iaculis ante ut tincidunt mattis. Praesent sagittis at enim sed auctor. Nulla facilisi. Duis vitae risus sit amet urna convallis cursus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Aenean vitae sodales ex, ut semper magna. Aenean rhoncus dignissim ligula, vitae tincidunt nunc vestibulum eget. Morbi gravida diam sem, congue ultrices ipsum interdum at. Aliquam pellentesque eros sed sapien rhoncus, sed varius ante cursus.Phasellus euismod bibendum lorem at rhoncus. In quis ipsum congue, facilisis purus tincidunt, efficitur felis. Morbi eu laoreet massa. Sed nec metus mauris. Ut dictum sed turpis vel lobortis. Integer facilisis odio vel interdum cursus. Pellentesque accumsan purus non lectus tempus semper. Duis libero sapien, sagittis at scelerisque eu, aliquet dapibus metus. Vestibulum sit amet augue vel nisi tincidunt dapibus. Sed nunc arcu, rutrum in dictum id, euismod in nibh. Curabitur sit amet tempus purus, quis faucibus justo. Mauris et lacus pellentesque libero viverra dignissim. Morbi mattis nisl eu enim placerat, ut consectetur sem sodales. Cras molestie, elit id maximus viverra, magna ex fermentum neque, vehicula pretium nulla erat vel nibh.Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Cras consequat nibh sit amet auctor bibendum. Nunc pharetra interdum metus, id pretium erat efficitur sit amet. Aliquam erat volutpat. Aliquam ac iaculis arcu. Suspendisse maximus lacinia mi, ac vehicula odio malesuada a. Aliquam vitae porttitor lacus. Nulla facilisi. Nulla facilisi. Mauris finibus tortor quam, nec accumsan augue rhoncus euismod.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg');

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `orderproduct`
--

CREATE TABLE `orderproduct` (
  `ID` int(11) NOT NULL,
  `orderID` int(11) NOT NULL,
  `productID` int(11) NOT NULL,
  `productQuantity` int(11) NOT NULL,
  `orderPrice` double NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `orderproduct`
--

INSERT INTO `orderproduct` (`ID`, `orderID`, `productID`, `productQuantity`, `orderPrice`) VALUES
(16, 10, 1, 2, 600),
(17, 10, 2, 1, 600),
(18, 10, 3, 3, 600),
(19, 11, 9, 2, 1400),
(20, 11, 3, 3, 1800),
(21, 11, 1, 3, 1800),
(25, 14, 1, 2, 1200),
(26, 14, 8, 1, 1000),
(27, 14, 13, 3, 1050),
(31, 21, 10, 2, 2000),
(32, 22, 16, 2, 1700),
(33, 22, 17, 2, 2400),
(34, 22, 18, 1, 700),
(35, 22, 23, 2, 1700),
(36, 22, 22, 3, 2550),
(37, 22, 20, 7, 5600),
(38, 23, 21, 2, 1600),
(50, 26, 23, 91, 77350),
(51, 27, 23, 1, 850),
(52, 27, 22, 1, 850),
(53, 27, 21, 1, 800),
(54, 27, 19, 1, 850),
(55, 27, 20, 1, 800),
(56, 27, 16, 1, 850),
(57, 27, 15, 2, 1200),
(58, 27, 17, 4, 4800),
(59, 27, 18, 3, 2100);

--
-- Disparadores `orderproduct`
--
DELIMITER $$
CREATE TRIGGER `after_orderProduct_insert` AFTER INSERT ON `orderproduct` FOR EACH ROW BEGIN
    UPDATE orders
    SET total = (
        SELECT SUM(orderPrice)
        FROM orderProduct
        WHERE orderID = NEW.orderID
    )
    WHERE ID = NEW.orderID;
END
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `after_orderProduct_update` AFTER UPDATE ON `orderproduct` FOR EACH ROW BEGIN
    UPDATE orders
    SET total = (
        SELECT SUM(orderPrice)
        FROM orderProduct
        WHERE orderID = NEW.orderID
    )
    WHERE ID = NEW.orderID;
END
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `before_orderProduct_insert` BEFORE INSERT ON `orderproduct` FOR EACH ROW BEGIN
    DECLARE productPrice DOUBLE;

    SELECT price * NEW.productQuantity INTO productPrice
    FROM products
    WHERE ID = NEW.productID;

    SET NEW.orderPrice = productPrice;
END
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `before_orderProduct_update` BEFORE UPDATE ON `orderproduct` FOR EACH ROW BEGIN
    DECLARE productPrice DOUBLE;

    SELECT price * NEW.productQuantity INTO productPrice
    FROM products
    WHERE ID = NEW.productID;

    SET NEW.OrderPrice = productPrice;
END
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `orders`
--

CREATE TABLE `orders` (
  `ID` int(11) NOT NULL,
  `total` double DEFAULT NULL,
  `startDate` datetime NOT NULL DEFAULT current_timestamp(),
  `finishDate` datetime DEFAULT NULL,
  `state` enum('pending','finished','canceled') NOT NULL DEFAULT 'pending',
  `name` varchar(50) NOT NULL,
  `phone` varchar(30) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `orders`
--

INSERT INTO `orders` (`ID`, `total`, `startDate`, `finishDate`, `state`, `name`, `phone`) VALUES
(10, 3600, '2023-11-18 20:09:58', '2023-11-18 20:41:43', 'pending', '', ''),
(11, 5000, '2023-11-18 20:09:58', '2023-11-18 20:41:43', 'pending', '', ''),
(14, 3250, '2023-11-18 20:09:58', '2023-11-18 20:42:33', 'finished', '', ''),
(21, 2000, '2023-11-20 00:22:57', NULL, 'pending', '', ''),
(22, 14650, '2023-11-20 00:25:08', NULL, 'pending', '', ''),
(23, 1600, '2023-11-20 00:27:01', '2023-11-20 00:41:45', 'canceled', '', ''),
(26, 77350, '2023-11-20 18:28:39', NULL, 'pending', 'Gianluca Vespe', '32133423342421'),
(27, 13100, '2023-11-20 18:30:02', NULL, 'pending', 'Turco agustin', '32142432546');

--
-- Disparadores `orders`
--
DELIMITER $$
CREATE TRIGGER `before_update_orders` BEFORE UPDATE ON `orders` FOR EACH ROW BEGIN
    IF NEW.state = 'finished' AND OLD.state != 'finished' THEN
        SET NEW.finishDate = CURRENT_TIMESTAMP;
    END IF;
END
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Estructura de tabla para la tabla `organizations`
--

CREATE TABLE `organizations` (
  `ID` int(11) NOT NULL,
  `Title` text NOT NULL,
  `Description` text DEFAULT NULL,
  `Image` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `organizations`
--

INSERT INTO `organizations` (`ID`, `Title`, `Description`, `Image`) VALUES
(1, 'organizacions de prueba 1', 'Descripcion de prueba 1', 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY='),
(2, 'organizacions de prueba 2', 'Descripcion de prueba 2', 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY='),
(3, 'organizacions de prueba 3', 'Descripcion de prueba 3', 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY=');

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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Volcado de datos para la tabla `products`
--

INSERT INTO `products` (`ID`, `Name`, `Price`, `Stock`, `Image`, `Description`) VALUES
(1, 'Mermelada de pera', 600, 1, '', ''),
(2, 'Mermelada de naranja inglesa', 600, 1, 'assets\\mermnaranja.jpg', ''),
(3, 'Mermelada de tomate', 600, 1, '', ''),
(4, 'Mermelada de zapallo', 600, 1, '', ''),
(5, 'Mermelada de ciruela', 800, 1, '', ''),
(6, 'Mermelada de higo', 800, 1, '', ''),
(7, 'Encurtido de berenjena', 700, 1, 'assets\\berenjena.jpg', ''),
(8, 'Encurtido de berenjena grande', 1000, 1, '', ''),
(9, 'Encurtido de champiñón', 700, 1, '', ''),
(10, 'Encurtido de champiñón grande', 1000, 1, '', ''),
(11, 'Bandeja de alfajores (6 u.)', 850, 1, 'assets\\alfajores.jpeg', ''),
(12, 'Pre pizza de tomate', 350, 1, '', ''),
(13, 'Pre pizza de cebolla', 350, 1, '', ''),
(14, 'Pasta frola', 900, 1, '', ''),
(15, 'Pan de figasa (1kg)', 600, 1, '', ''),
(16, 'Pan integral (800gr)', 850, 1, '', ''),
(17, 'Tarta de coco y dulce de leche', 1200, 1, '', ''),
(18, 'Encurtido de ajo', 700, 1, 'assets\\ajo.jpg', NULL),
(19, 'Bandeja de alfajores de chocolate (6 u.)', 850, 1, 'assets\\alfajores-chocolate.jpeg', NULL),
(20, 'Budin', 800, 1, 'assets\\budin.jpeg', NULL),
(21, 'Pan dulce', 800, 1, 'assets\\pandulce.jpeg', NULL),
(22, 'Fugazzeta', 850, 1, 'assets\\fugazzeta.jpg', NULL),
(23, 'Galletas (6 u.)', 850, 1, 'assets\\galletitas.jpeg', NULL);

--
-- Índices para tablas volcadas
--

--
-- Indices de la tabla `family`
--
ALTER TABLE `family`
  ADD PRIMARY KEY (`ID`);

--
-- Indices de la tabla `gallery`
--
ALTER TABLE `gallery`
  ADD PRIMARY KEY (`ID`);

--
-- Indices de la tabla `orderproduct`
--
ALTER TABLE `orderproduct`
  ADD PRIMARY KEY (`ID`),
  ADD KEY `orderID` (`orderID`),
  ADD KEY `productID` (`productID`);

--
-- Indices de la tabla `orders`
--
ALTER TABLE `orders`
  ADD PRIMARY KEY (`ID`);

--
-- Indices de la tabla `organizations`
--
ALTER TABLE `organizations`
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
-- AUTO_INCREMENT de la tabla `family`
--
ALTER TABLE `family`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT de la tabla `gallery`
--
ALTER TABLE `gallery`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=13;

--
-- AUTO_INCREMENT de la tabla `news`
--
ALTER TABLE `news`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

--
-- AUTO_INCREMENT de la tabla `orderproduct`
--
ALTER TABLE `orderproduct`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=60;

--
-- AUTO_INCREMENT de la tabla `orders`
--
ALTER TABLE `orders`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=28;

--
-- AUTO_INCREMENT de la tabla `organizations`
--
ALTER TABLE `organizations`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT de la tabla `products`
--
ALTER TABLE `products`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=25;

--
-- Restricciones para tablas volcadas
--

--
-- Filtros para la tabla `orderproduct`
--
ALTER TABLE `orderproduct`
  ADD CONSTRAINT `orderproduct_ibfk_1` FOREIGN KEY (`orderID`) REFERENCES `orders` (`ID`),
  ADD CONSTRAINT `orderproduct_ibfk_2` FOREIGN KEY (`productID`) REFERENCES `products` (`ID`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
