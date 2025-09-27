CREATE TABLE IF NOT EXISTS users(
    user_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    login VARCHAR (50) UNIQUE NOT NULL,
    password VARCHAR (64) NOT NULL
);

CREATE TABLE IF NOT EXISTS orders(
    order_number VARCHAR (50)  PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR (50) NOT NULL,
    accrual REAL,
    upload_data VARCHAR (50) NOT NULL
);

CREATE TABLE IF NOT EXISTS withdraws(
    order_number VARCHAR (50)  PRIMARY KEY,
    user_id INTEGER NOT NULL,
    sum REAL NOT NULL,
    upload_data VARCHAR (50) NOT NULL
);

CREATE TABLE IF NOT EXISTS balances(
    user_id INTEGER PRIMARY KEY,
    accrual REAL NOT NULL,
    withdraw REAL NOT NULL
);
