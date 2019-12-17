DROP DATABASE IF EXISTS litterbox;
CREATE DATABASE litterbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE litterbox;

CREATE TABLE users
(
    id    VARCHAR(250),
    email VARCHAR(1000),
    name  VARCHAR(1000),

    PRIMARY KEY (id)
);
