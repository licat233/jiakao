DROP DATABASE driving_exam;
CREATE DATABASE driving_exam charset=utf8;
use driving_exam;
CREATE TABLE score(
    id int PRIMARY KEY AUTO_INCREMENT,
    name varchar(10) not NULL,
    score INT
);