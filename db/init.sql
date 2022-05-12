CREATE TABLE Users(
    user_id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    phone_number VARCHAR NOT NULL UNIQUE,
    password VARCHAR NOT NULL,
    picture bytea,
    rating float4
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
    upload_date DATE NOT NULL DEFAULT CURRENT_DATE,
    description VARCHAR,
    picture bytea,
    fk_user_id INT REFERENCES Users(user_id) NOT NULL
);

CREATE TABLE Review (
    review_id SERIAL PRIMARY KEY,
    rating INT NOT NULL,
    content VARCHAR,
    fk_reviewer_id INT REFERENCES Users(user_id) NOT NULL,
    fk_owner_id INT REFERENCES Users(user_id) NOT NULL

);

CREATE TABLE Pinned_Product (
    fk_product_id INT REFERENCES Product(product_id) ON DELETE CASCADE NOT NULL,
    fk_user_id INT REFERENCES Users(user_id) ON DELETE CASCADE NOT NULL,
    PRIMARY KEY(fk_product_id, fk_user_id)
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


/* test users user_id = 1 & 2 */
INSERT INTO Users (name, phone_number, password, picture, rating) VALUES ('Gustav', '+12029182132', 'lorem ipsum', encode(pg_read_binary_file('/docker-entrypoint-initdb.d/victorkill.jpeg'), 'base64')::bytea, 3);

INSERT INTO USERS (name, phone_number, password, rating) VALUES ('Victor', '+12027455483', 'lorem ipsum', 4);

/* test products product_id = 1 */
INSERT INTO Product (name,service,price,description, fk_user_id ) VALUES ('Soffa','true',1,'Hej',1);

/* test review review_id = 1 */
INSERT INTO Review (rating,content, fk_reviewer_id, fk_owner_id) VALUES (2,'SÄMST',1,2);

/* test communities community_id = 1 & 2 & 3 */
INSERT INTO Community (name) VALUES ('Clothes'), ('Politics'), ('Memes');

/* test pinned_product pinnedproduct_id = 1 */
INSERT INTO Pinned_Product (fk_product_id, fk_user_id) VALUES (1,1);

/* test user_community pinnedproduct_id = 1 */
INSERT INTO User_Community(fk_user_id, fk_community_id) VALUES (1,2);

/* test user_followers user_follower_id = 1 */
INSERT INTO User_Followers(fk_user_id, fk_follower_id) VALUES (2, 1);
