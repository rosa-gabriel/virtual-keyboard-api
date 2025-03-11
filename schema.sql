create table combinations (
    options int[][]
);

create table users (
    username varchar(50) primary key,
    email varchar(50),
    password varchar(50) unique
);

