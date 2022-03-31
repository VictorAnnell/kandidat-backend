CREATE TABLE Users(
    user_id SERIAL PRIMARY KEY,
    name VARCHAR,
    phone_nr INT,
    address VARCHAR

);

CREATE TABLE Product (
    product_id SERIAL PRIMARY KEY,
    name VARCHAR,
    service BOOLEAN,
    price INT,
    upload_date DATE,
    description VARCHAR,
    fk_user_id INT REFERENCES Users(user_id)
);

CREATE TABLE Review (
    review_id SERIAL PRIMARY KEY,
    rating INT,
    content VARCHAR,
    fk_user_id INT REFERENCES Users(user_id),
    fk_product_id INT REFERENCES Product(product_id)
);


CREATE TABLE Community (
    community_id SERIAL PRIMARY KEY,
    name VARCHAR 
);

CREATE TABLE UserCommunityLink (
    user_community_link_id SERIAL PRIMARY KEY,
    fk_user_id INT REFERENCES Users(user_id),
    fk_community_id INT REFERENCES Community(community_id) 
);


    

INSERT INTO Users (name) VALUES ('Gustav'), ('Victor');

INSERT INTO Community (name) VALUES ('Clothes'), ('Politics'), ('Memes');

INSERT INTO UserCommunityLink(fk_user_id, fk_community_id) VALUES (
    (SELECT user_id from Users where name='Gustav'),
    (SELECT community_id from Community where name='Memes'));
