ALTER TABLE field_operations
    ADD COLUMN IF NOT EXISTS check_machine_status BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS check_implement_status BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS check_field_area BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS check_notes_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS checklist_updated_at TIMESTAMPTZ;

INSERT INTO permissions (name, description) VALUES
  ('field_operations:checklist', 'Completare checklist plecare pentru operațiuni pe teren'),
  ('field_operations:start', 'Pornire operațiuni pe teren')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code IN ('admin', 'manager', 'operator')
  AND p.name IN ('field_operations:checklist', 'field_operations:start')
ON CONFLICT DO NOTHING;
