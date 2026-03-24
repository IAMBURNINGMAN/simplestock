CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(100) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    full_name   VARCHAR(200) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'storekeeper',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE categories (
    id    BIGSERIAL PRIMARY KEY,
    name  VARCHAR(200) UNIQUE NOT NULL
);

CREATE TABLE products (
    id             BIGSERIAL PRIMARY KEY,
    name           VARCHAR(300) NOT NULL,
    sku            VARCHAR(100) UNIQUE NOT NULL,
    category_id    BIGINT REFERENCES categories(id),
    unit           VARCHAR(50) NOT NULL DEFAULT 'шт',
    min_stock      INTEGER NOT NULL DEFAULT 0,
    quantity       INTEGER NOT NULL DEFAULT 0,
    purchase_price NUMERIC(12,2),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE documents (
    id              BIGSERIAL PRIMARY KEY,
    doc_type        VARCHAR(20) NOT NULL,
    doc_number      VARCHAR(50) UNIQUE NOT NULL,
    counterparty    VARCHAR(300),
    expense_type    VARCHAR(50),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    user_id         BIGINT REFERENCES users(id) NOT NULL,
    doc_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE document_items (
    id            BIGSERIAL PRIMARY KEY,
    document_id   BIGINT REFERENCES documents(id) ON DELETE CASCADE NOT NULL,
    product_id    BIGINT REFERENCES products(id) NOT NULL,
    quantity      INTEGER NOT NULL CHECK (quantity > 0),
    price         NUMERIC(12,2)
);

CREATE TABLE movements (
    id            BIGSERIAL PRIMARY KEY,
    product_id    BIGINT REFERENCES products(id) NOT NULL,
    document_id   BIGINT REFERENCES documents(id),
    inventory_id  BIGINT,
    movement_type VARCHAR(30) NOT NULL,
    quantity      INTEGER NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventories (
    id           BIGSERIAL PRIMARY KEY,
    inv_number   VARCHAR(50) UNIQUE NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'active',
    user_id      BIGINT REFERENCES users(id) NOT NULL,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE inventory_items (
    id                BIGSERIAL PRIMARY KEY,
    inventory_id      BIGINT REFERENCES inventories(id) ON DELETE CASCADE NOT NULL,
    product_id        BIGINT REFERENCES products(id) NOT NULL,
    expected_quantity INTEGER NOT NULL,
    actual_quantity   INTEGER NOT NULL,
    difference        INTEGER GENERATED ALWAYS AS (actual_quantity - expected_quantity) STORED
);

ALTER TABLE movements ADD CONSTRAINT fk_inventory
    FOREIGN KEY (inventory_id) REFERENCES inventories(id);

CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_low_stock ON products(quantity, min_stock) WHERE quantity <= min_stock;
CREATE INDEX idx_movements_product ON movements(product_id, created_at DESC);
CREATE INDEX idx_documents_type_date ON documents(doc_type, doc_date);
