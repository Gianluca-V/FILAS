-- FILAS schema, seeded from filas.sql for local Docker Compose development.
-- Row/column data below is copied verbatim from the source dump; only the
-- following are intentionally changed:
--
--  1. Trigger status: all 5 legacy MySQL triggers from filas.sql are
--     OMITTED here. Their logic (orderPrice, orders.total, finishDate
--     stamping) is ported to the Go `usecase/order.go` layer per ADR-3
--     (sdd/migrate-go-vue/design). Per the gate review (obs #28), there
--     are FIVE triggers in the original dump, not four:
--       - after_orderProduct_insert   (orderproduct)
--       - after_orderProduct_update   (orderproduct)
--       - before_orderProduct_insert  (orderproduct)
--       - before_orderProduct_update  (orderproduct)
--       - before_update_orders        (orders)
--     They are dropped from this seed ONLY after the ported usecase logic
--     has green parity tests (Phase 6 gate); until then the triggers live
--     only in a scratch DB used for Phase 5/6 characterization testing.
--
--  2. `admins` table: the source dump has NO primary key at all (ID is a
--     nullable int with zero constraints). Patched here with an explicit
--     PRIMARY KEY + AUTO_INCREMENT (task 2.1 will confirm this choice
--     against the Go AdminRepository).

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

-- --------------------------------------------------------

--
-- Table `admins` (PRIMARY KEY + AUTO_INCREMENT patched in, see note 2 above)
--

CREATE TABLE `admins` (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `username` text DEFAULT NULL,
  `password` text DEFAULT NULL,
  `salt` text DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `admins` (`ID`, `username`, `password`, `salt`) VALUES
(1543, 'FilasAdmin', '8b4abda03a7fb4eee79d8fb7c07f88497f28b0728779c84fab84a07b3f824cef', 'bcf76de2e59270b37b7e76d22b53917b');

ALTER TABLE `admins` AUTO_INCREMENT = 1544;

-- --------------------------------------------------------

--
-- Table `family`
--

CREATE TABLE `family` (
  `ID` int(11) NOT NULL,
  `Image` text DEFAULT NULL,
  `Body` text NOT NULL,
  `Category` varchar(20) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `family` (`ID`, `Image`, `Body`, `Category`) VALUES
(3, 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY=', 'aqui va la descripcion 3', 'Taller protegido'),
(5, 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg', '', 'Taller protegido'),
(6, '', 'Taller de música: Ritmo – canto- uso y confección de instrumentos', 'Centro de dia'),
(7, '', 'Taller de educación física: Actividades y juegos con distintos materiales y elementos', 'Centro de dia'),
(8, '', 'Taller de cocina: Lavado de manos, Planificación de receta, Compras, Uso de delantal, Reconocimiento de utensilios, Acciones, cortar, mezclar, amasar, batir; Reconocer y utilizar correctamente los dispositivos, usos y cuidados.', 'Centro de dia'),
(9, '', 'Taller de estimulación cognitiva: Se pretende activar, estimular y entrenar determinadas capacidades cognitivas y los componentes que la integran, de forma adecuada y sistemática, para transformarlas en una habilidad, hábito o destreza. Todo ello parece desligarse de otras dimensiones tales como la emocional y/o la conductual, pero no es así dado que también se trabajan para poder ser transferidas al entorno cotidiano general.\nPropuestas lúdicas.', 'Centro de dia'),
(10, '', 'Taller de manualidades: Diversas artesanías; Utilización de técnicas plásticas y grafo plásticas; Diversos materiales y texturas.', 'Centro de dia'),
(11, '', 'Taller de expresión corporal: La expresión corporal es un medio de comunicación a través del cual las personas comunican ideas, sensaciones, emociones, sentimientos y pensamientos, esta es la unidad del lenguaje gestual. Las personas comunican sus representaciones mentales sobre el ambiente a través del cuerpo. En consecuencia, las diferentes formas de expresión corporal son aprovechadas para construir significados de acuerdo con los contextos donde se manifiestan.', 'Centro de dia'),
(12, '', 'Expresión facial: Es el tipo de expresión corporal que se distingue por utilizar principalmente las diferentes partes del rostro para expresar sentimientos. Muchas veces ponemos en práctica esta forma de expresión si darnos cuenta, ya que es un elemento de uso común en el lenguaje humano.\nDiferentes propuestas para estimular la expresión corporal.', 'Centro de dia'),
(13, '', 'Expresión teatral: Como mencionamos antes, en el teatro la expresión corporal es fundamental, porque aunque podamos hablar, en este arte los movimientos de nuestro cuerpo dicen más que las palabras. Quienes logran manejar esta forma de expresión corporal logran transmitir muchas emociones.\nPor lo general, comunicarse a través de la expresión teatral significa utilizar todo el cuerpo..', 'Centro de dia'),
(14, '', 'Taller de relajación corporal: Las técnicas de relajación implican centrar la atención en algo que calme y aumente la conciencia del propio  cuerpo, favoreciendo la estimulación de las funciones ejecutivas. Ejercicios de relajación brindan  muchos beneficios, algunos son: Disminuir la frecuencia cardíaca; Disminuir la presión arterial; Disminuir la frecuencia respiratoria; Mejorar la digestión; Controlar los niveles de glucosa en la sangre; Reducir la actividad de las hormonas del estrés; Incrementar el flujo sanguíneo hacia los músculos más grandes; Reducir la tensión muscular y el dolor crónico; Mejorar la atención y el estado de ánimo; Mejorar la calidad del sueño; Disminuir la fatiga; Reducir la ira y la frustración; Desarrollar la confianza para resolver problemas;', 'Centro de dia');

ALTER TABLE `family` ADD PRIMARY KEY (`ID`);
ALTER TABLE `family` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=15;

-- --------------------------------------------------------

--
-- Table `gallery`
--

CREATE TABLE `gallery` (
  `ID` int(11) NOT NULL,
  `Image` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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

ALTER TABLE `gallery` ADD PRIMARY KEY (`ID`);
ALTER TABLE `gallery` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=15;

-- --------------------------------------------------------

--
-- Table `news`
--

CREATE TABLE `news` (
  `ID` int(11) NOT NULL,
  `Title` varchar(99) NOT NULL,
  `Body` text DEFAULT NULL,
  `Image` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `news` (`ID`, `Title`, `Body`, `Image`) VALUES
(1, 'Noticia de prueba 1', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(2, 'Noticia de prueba 2', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(3, 'Noticia de prueba 3', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(4, 'Noticia de prueba 4', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg'),
(5, 'Noticia de prueba 5', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam at nisl sed justo cursus finibus. Vivamus quis suscipit turpis. Fusce eu urna magna. Nulla sit amet nunc eget ligula luctus tristique non ornare risus. Quisque nec nulla at massa imperdiet accumsan vitae nec augue. Nunc et nisl justo. Nam venenatis odio sed sapien laoreet, nec lobortis tellus volutpat.', 'https://www.refugeesrespond.org/dadaabwikimedia/images/archive/a/a9/20201124034818%21Example.jpg');

ALTER TABLE `news` ADD PRIMARY KEY (`ID`);
ALTER TABLE `news` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

-- --------------------------------------------------------

--
-- Table `orderproduct` (5 legacy triggers intentionally omitted -- see
-- header note 1)
--

CREATE TABLE `orderproduct` (
  `ID` int(11) NOT NULL,
  `orderID` int(11) NOT NULL,
  `productID` int(11) NOT NULL,
  `productQuantity` int(11) NOT NULL,
  `orderPrice` double NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
(59, 27, 18, 3, 2100),
(60, 28, 1, 5, 3000),
(61, 29, 1, 5, 3000),
(62, 30, 1, 200, 120000),
(63, 31, 23, -100000, -85000000),
(64, 32, 22, 3, 2550);

-- --------------------------------------------------------

--
-- Table `orders` (1 legacy trigger intentionally omitted -- see header note 1)
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

INSERT INTO `orders` (`ID`, `total`, `startDate`, `finishDate`, `state`, `name`, `phone`) VALUES
(10, 3600, '2023-11-18 20:09:58', '2023-11-18 20:41:43', 'canceled', '', ''),
(11, 5000, '2023-11-18 20:09:58', '2023-11-18 20:41:43', 'pending', '', ''),
(14, 3250, '2023-11-18 20:09:58', '2023-11-18 20:42:33', 'finished', '', ''),
(21, 2000, '2023-11-20 00:22:57', NULL, 'pending', '', ''),
(22, 14650, '2023-11-20 00:25:08', NULL, 'pending', '', ''),
(23, 1600, '2023-11-20 00:27:01', '2023-11-20 00:41:45', 'canceled', '', ''),
(26, 77350, '2023-11-20 18:28:39', NULL, 'canceled', 'Gianluca Vespe', '32133423342421'),
(27, 13100, '2023-11-20 18:30:02', NULL, 'pending', 'Turco agustin', '32142432546'),
(28, 3000, '2023-11-20 21:58:04', NULL, 'canceled', 'Gianluca Vespe', '352435234234'),
(29, 3000, '2023-11-20 22:01:20', NULL, 'canceled', 'Gianluca Vespe', '352435234234'),
(30, 120000, '2023-11-20 22:21:06', NULL, 'canceled', 'gianluca', 'd323213'),
(31, -85000000, '2023-11-20 22:24:24', NULL, 'pending', 'gianluca', '352435234234'),
(32, 2550, '2023-11-20 22:28:45', NULL, 'pending', 'HoneyCorp', '1432546753');

ALTER TABLE `orders` ADD PRIMARY KEY (`ID`);
ALTER TABLE `orders` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=33;

-- --------------------------------------------------------

--
-- Table `organizations`
--

CREATE TABLE `organizations` (
  `ID` int(11) NOT NULL,
  `Title` text NOT NULL,
  `Description` text DEFAULT NULL,
  `Image` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `organizations` (`ID`, `Title`, `Description`, `Image`) VALUES
(1, 'organizacions de prueba 1', 'Descripcion de prueba 1', 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY='),
(2, 'organizacions de prueba 2', 'Descripcion de prueba 2', 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY='),
(3, 'organizacions de prueba 3', 'Descripcion de prueba 3', 'https://media.istockphoto.com/id/600072788/es/foto/contactos-de-delegados-en-la-recepci%C3%B3n-de-bebidas-de-la-conferencia.jpg?s=612x612&w=0&k=20&c=fxN0g917vwO_oUq62yO1Ouw9QkiZT5By68sq3v1gvVY=');

ALTER TABLE `organizations` ADD PRIMARY KEY (`ID`);
ALTER TABLE `organizations` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

-- --------------------------------------------------------

--
-- Table `products`
--

CREATE TABLE `products` (
  `ID` int(11) NOT NULL,
  `Name` varchar(99) NOT NULL,
  `Price` double NOT NULL,
  `Stock` int(11) NOT NULL,
  `Image` text DEFAULT NULL,
  `Description` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `products` (`ID`, `Name`, `Price`, `Stock`, `Image`, `Description`) VALUES
(1, 'Mermelada de pera', 600, 112, 'assets/default-img.png', ''),
(2, 'Mermelada de naranja inglesa', 600, 2, 'assets\\mermnaranja.jpg', ''),
(3, 'Mermelada de tomate', 600, 4, '', ''),
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
(21, 'Pan dulce', 800, 0, 'assets\\pandulce.jpeg', NULL),
(22, 'Fugazzeta', 850, -2, 'assets\\fugazzeta.jpg', NULL),
(23, 'Galletas (6 u.)', 850, 100092, 'assets\\galletitas.jpeg', NULL);

ALTER TABLE `products` ADD PRIMARY KEY (`ID`);
ALTER TABLE `products` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=25;

-- --------------------------------------------------------

--
-- Indexes and foreign keys for `orderproduct` (declared after all
-- referenced tables exist)
--

ALTER TABLE `orderproduct`
  ADD PRIMARY KEY (`ID`),
  ADD KEY `orderID` (`orderID`),
  ADD KEY `productID` (`productID`);

ALTER TABLE `orderproduct` MODIFY `ID` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=66;

ALTER TABLE `orderproduct`
  ADD CONSTRAINT `orderproduct_ibfk_1` FOREIGN KEY (`orderID`) REFERENCES `orders` (`ID`),
  ADD CONSTRAINT `orderproduct_ibfk_2` FOREIGN KEY (`productID`) REFERENCES `products` (`ID`);

COMMIT;
