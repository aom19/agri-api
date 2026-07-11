-- Expand operators table with phone, email, notes columns
ALTER TABLE operators
    ADD COLUMN IF NOT EXISTS phone TEXT,
    ADD COLUMN IF NOT EXISTS email TEXT,
    ADD COLUMN IF NOT EXISTS notes TEXT;

-- Add operators:disable permission
INSERT INTO permissions (name, description)
VALUES
    ('operators:disable', 'Dezactivare și reactivare operatori')
ON CONFLICT (name) DO NOTHING;

-- Grant operators:disable to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'operators:disable'
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;
