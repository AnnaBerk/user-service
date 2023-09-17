
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       name TEXT NOT NULL,
                       surname TEXT NOT NULL,
                       patronymic TEXT,
                       age INT,
                       gender TEXT,
                       nationality TEXT
);

-- down.sql

DROP TABLE users;