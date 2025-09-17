CREATE TABLE IF NOT EXISTS users(
    user_id serial PRIMARY KEY,
    login VARCHAR (50) UNIQUE NOT NULL,
    password VARCHAR (50) NOT NULL
);

CREATE TABLE IF NOT EXISTS orders(
    order_number VARCHAR (50)  PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR (50) NOT NULL,
    accrual INTEGER,
    upload_data VARCHAR (50) NOT NULL
);

CREATE TABLE IF NOT EXISTS withdraws(
    order_number VARCHAR (50)  PRIMARY KEY,
    user_id INTEGER NOT NULL,
    sum INTEGER NOT NULL,
    upload_data VARCHAR (50) NOT NULL
);

CREATE TABLE IF NOT EXISTS balances(
    user_id INTEGER PRIMARY KEY,
    accrual INTEGER NOT NULL,
    withdraw INTEGER NOT NULL
);
