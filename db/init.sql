CREATE TABLE Users(
    user_id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    phone_nr VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    img bytea
);

CREATE TABLE User_Followers(
    user_follower_id SERIAL PRIMARY KEY,
    fk_user_id INT REFERENCES Users(user_id) NOT NULL,
    fk_follower_id INT REFERENCES Users(user_id) NOT NULL
);

CREATE TABLE Product (
    product_id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    service BOOLEAN NOT NULL,
    price INT NOT NULL,
    upload_date DATE NOT NULL,
    description VARCHAR,
    fk_user_id INT REFERENCES Users(user_id) NOT NULL
);

CREATE TABLE Review (
    review_id SERIAL PRIMARY KEY,
    rating INT NOT NULL,
    content VARCHAR,
    fk_reviwer_id INT REFERENCES Users(user_id) NOT NULL,
    fk_product_id INT REFERENCES Product(product_id) NOT NULL
);


CREATE TABLE Community (
    community_id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL
);

CREATE TABLE User_Community (
    user_community_id SERIAL PRIMARY KEY,
    fk_user_id INT REFERENCES Users(user_id) NOT NULL,
    fk_community_id INT REFERENCES Community(community_id) NOT NULL
);


INSERT INTO Product (name,service,price,upload_date,description, fk_user_id ) VALUES ('Soffa',1,1,'2022-04-07','Hej',1),('Soffa',1,2,'2022-04-07','Hej',1);

INSERT INTO Review (rating,content, fk_reviwer_id, fk_product_id) VALUES (2,'SÄMST',1,1),(3,'SÄMRE',2,2);
INSERT INTO Users (name, phone_nr, password) VALUES ('Gustav', '+1 202-918-2132', 'lorem ipsum'), ('Victor', '+1 202-918-2131', 'lorem ipsum'), ('Rohat', 123, 'lorem ipsum');

INSERT INTO Users (name, phone_nr, password, img) VALUES ('Victor Kill', 123, 'lorem ipsum', pg_read_binary_file('/docker-entrypoint-initdb.d/victorkill.jpeg')::bytea);

INSERT INTO Community (name) VALUES ('Clothes'), ('Politics'), ('Memes');

INSERT INTO Product (name, service, price, upload_date, description, fk_user_id) VALUES ('Rosa soffa', False, 200, '2022-02-02', 
'A nice couch', 1);

INSERT INTO User_Community(fk_user_id, fk_community_id) VALUES (
    (SELECT user_id from Users where name='Gustav'),
    (SELECT community_id from Community where name='Memes'));

INSERT INTO User_Followers(fk_user_id, fk_follower_id) VALUES (1, 3);
