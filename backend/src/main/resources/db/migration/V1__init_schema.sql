CREATE TABLE admin_users (
                             id BIGSERIAL PRIMARY KEY,
                             login VARCHAR(255) NOT NULL UNIQUE,
                             password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE service_categories (
                                    id BIGSERIAL PRIMARY KEY,
                                    name VARCHAR(255) NOT NULL,
                                    sort_order INTEGER,
                                    active BOOLEAN NOT NULL
);

CREATE TABLE funeral_services (
                                  id BIGSERIAL PRIMARY KEY,
                                  title VARCHAR(255) NOT NULL,
                                  description VARCHAR(1000),
                                  price NUMERIC(12, 2) NOT NULL,
                                  category_id BIGINT REFERENCES service_categories(id),
                                  active BOOLEAN NOT NULL,
                                  sort_order INTEGER
);

CREATE TABLE product_categories (
                                    id BIGSERIAL PRIMARY KEY,
                                    name VARCHAR(255) NOT NULL,
                                    sort_order INTEGER,
                                    active BOOLEAN NOT NULL
);

CREATE TABLE catalog_products (
                                  id BIGSERIAL PRIMARY KEY,
                                  title VARCHAR(255) NOT NULL,
                                  description VARCHAR(1000),
                                  price NUMERIC(12, 2) NOT NULL,
                                  image_url VARCHAR(1000),
                                  category_id BIGINT REFERENCES product_categories(id),
                                  active BOOLEAN NOT NULL,
                                  sort_order INTEGER
);

CREATE TABLE orders (
                        id VARCHAR(32) PRIMARY KEY,
                        order_date DATE NOT NULL,
                        status VARCHAR(50) NOT NULL,
                        total_amount NUMERIC(12, 2) NOT NULL,
                        paid BOOLEAN NOT NULL,
                        client_name VARCHAR(255) NOT NULL,
                        client_phone VARCHAR(50) NOT NULL,
                        client_email VARCHAR(255),
                        deceased_name VARCHAR(255) NOT NULL,
                        deceased_date_of_birth DATE,
                        deceased_date_of_death DATE NOT NULL
);

CREATE TABLE order_services (
                                id BIGSERIAL PRIMARY KEY,
                                order_id VARCHAR(32) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                                name VARCHAR(255) NOT NULL,
                                price NUMERIC(12, 2) NOT NULL
);

CREATE TABLE order_products (
                                id BIGSERIAL PRIMARY KEY,
                                order_id VARCHAR(32) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                                name VARCHAR(255) NOT NULL,
                                price NUMERIC(12, 2) NOT NULL
);

CREATE TABLE order_documents (
                                 id BIGSERIAL PRIMARY KEY,
                                 order_id VARCHAR(32) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                                 name VARCHAR(255) NOT NULL,
                                 type VARCHAR(100) NOT NULL,
                                 document_date DATE NOT NULL,
                                 size VARCHAR(100)
);