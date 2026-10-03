DELETE FROM stock_movements WHERE notes = 'Sold inițial (reconciliere istoric)';

ALTER TABLE stock_movements
    ALTER COLUMN quantity_delta TYPE NUMERIC(14,3),
    ALTER COLUMN resulting_quantity TYPE NUMERIC(14,3);
