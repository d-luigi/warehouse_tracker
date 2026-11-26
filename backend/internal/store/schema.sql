CREATE TABLE IF NOT EXISTS items (
    id            BIGSERIAL PRIMARY KEY,
    sku           TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    location      TEXT        NOT NULL DEFAULT '',
    quantity      INTEGER     NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reorder_level INTEGER     NOT NULL DEFAULT 0 CHECK (reorder_level >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS movements (
    id         BIGSERIAL PRIMARY KEY,
    item_id    BIGINT      NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    type       TEXT        NOT NULL CHECK (type IN ('in', 'out', 'adjust')),
    delta      INTEGER     NOT NULL,
    balance    INTEGER     NOT NULL CHECK (balance >= 0),
    note       TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS movements_item_idx ON movements (item_id, id DESC);
