-- phpMyAdmin SQL Dump
-- version 5.2.0
-- https://www.phpmyadmin.net/
--
-- Servidor: 127.0.0.1
-- Tiempo de generación: 30-10-2023 a las 06:55:16
-- Versión del servidor: 10.4.24-MariaDB
-- Versión de PHP: 8.1.6

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
-- Estructura de tabla para la tabla `mails`
--
CREATE TABLE `mails` (
  `ID` int(11) NOT NULL,
  `Mail` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
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
-- AUTO_INCREMENT de la tabla `family`
--
ALTER TABLE `family`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

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
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;
--
-- AUTO_INCREMENT de la tabla `products`
--
ALTER TABLE `products`
  MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;
COMMIT;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;