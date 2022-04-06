CREATE TABLE Users(
    user_id SERIAL PRIMARY KEY,
    name VARCHAR,
    phone_nr INT,
    address VARCHAR

);

CREATE TABLE Product (
    product_id SERIAL PRIMARY KEY,
    name VARCHAR,
    service INT,
    price INT,
    upload_date DATE,
    description VARCHAR,
    fk_user_id INT REFERENCES Users(user_id)
);

CREATE TABLE Review (
    review_id SERIAL PRIMARY KEY,
    rating INT,
    content VARCHAR,
    fk_reviwer_id INT REFERENCES Users(user_id),
    fk_product_id INT REFERENCES Product(product_id)
);


CREATE TABLE Community (
    community_id SERIAL PRIMARY KEY,
    name VARCHAR 
);

CREATE TABLE User_Community (
    -- user_community_link_id SERIAL PRIMARY KEY,
    fk_user_id INT REFERENCES Users(user_id),
    fk_community_id INT REFERENCES Community(community_id) 
);

INSERT INTO Users (name) VALUES ('Gustav'), ('Victor'),('Kimiya'),('Aishe');

INSERT INTO Product (name,service,price,upload_date,description, fk_user_id ) VALUES ('Soffa',1,1,'2022-04-07','Hej',1),('Soffa',1,2,'2022-04-07','Hej',1);

INSERT INTO Community (name) VALUES ('Clothes'), ('Politics'), ('Memes');

INSERT INTO Review (rating,content, fk_reviwer_id, fk_product_id) VALUES (2,'SÄMST',1,1),(3,'SÄMRE',2,2);

INSERT INTO User_Community(fk_user_id, fk_community_id) VALUES (
    (SELECT user_id from Users where name='Gustav'),
    (SELECT community_id from Community where name='Memes'));
