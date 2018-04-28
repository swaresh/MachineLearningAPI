CREATE DATABASE machine_learning;
USE machine_learning;
CREATE TABLE accuracy (id INT NOT NULL PRIMARY KEY AUTO_INCREMENT, 
learning_rate FLOAT8,
layer INT,
steps INT,
accuracy FLOAT8);