CREATE TABLE IF NOT EXISTS subscription (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    servic_name TEXT NOT NULL CHECK(
        length(servic_name) >= 2 AND length(servic_name) <= 50
    ),
    price INT NOT NULL,
    client_id UUID NOT NULL,
    start_date TIMESTAMP NOT NULL,
    finish_date TIMESTAMP NULL
);