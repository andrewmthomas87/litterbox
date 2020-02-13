DROP DATABASE IF EXISTS litterbox;
CREATE DATABASE litterbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE litterbox;

CREATE TABLE users
(
    id             VARCHAR(250),
    email          VARCHAR(1000),
    name           VARCHAR(1000),
    stage          TINYINT,

    onCampus       BOOLEAN,
    building       TINYINT,
    address        VARCHAR(500),
    onCampusFuture BOOLEAN,

    PRIMARY KEY (id)
);

CREATE TABLE pickupTimeSlots
(
    id        INT AUTO_INCREMENT,
    date      DATE,
    startTime TIME,
    endTime   TIME,

    PRIMARY KEY (id)
);
