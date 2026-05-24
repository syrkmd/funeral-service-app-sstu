ALTER TABLE orders
    ADD COLUMN service_date DATE,
    ADD COLUMN service_time VARCHAR(50),
    ADD COLUMN service_address VARCHAR(255),
    ADD COLUMN cemetery VARCHAR(255),
    ADD COLUMN cemetery_notes VARCHAR(1000);
