-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.users ADD CONSTRAINT users_unique UNIQUE (guid);

CREATE TABLE customer (
    user_id UUID NOT NULL PRIMARY KEY,
    stripe_id TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users (guid)
);

CREATE TRIGGER update_customer_updated_at
BEFORE UPDATE ON customer
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();


CREATE TABLE subscription (
    stripe_id TEXT NOT NULL PRIMARY KEY,
    customer_id UUID NOT NULL,
    stripe_price_id TEXT NOT NULL,
    stripe_product_name TEXT NOT NULL,
    status TEXT NOT NULL,
    quantity INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (customer_id) REFERENCES customer (user_id)
);
CREATE TRIGGER update_subscription_updated_at
BEFORE UPDATE ON subscription
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE subscription;
DROP TABLE customer;
ALTER TABLE public.users DROP CONSTRAINT users_unique;
-- +goose StatementEnd
