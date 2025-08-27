CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    balance NUMERIC NOT NULL,
    date_update TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    from_wallet UUID REFERENCES wallets(id),
    to_wallet UUID REFERENCES wallets(id),
    amount NUMERIC NOT NULL,
    transfer_date TIMESTAMP NOT NULL
);


