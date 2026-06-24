-- ============================================================
-- RBAC seed: roles, permissions, role_permissions
-- Rulează DUPĂ migrate-up (migration 010)
-- ============================================================

-- ─── Permissions ────────────────────────────────────────────
INSERT INTO permissions (name, description) VALUES
    ('machines:read',     'Vizualizare mașini'),
    ('machines:write',    'Creare și editare mașini'),
    ('machines:delete',   'Ștergere mașini'),
  ('fields:read',       'Vizualizare terenuri'),
  ('fields:write',      'Creare și editare terenuri'),
  ('fields:delete',     'Ștergere terenuri'),
    ('operators:read',    'Vizualizare operatori'),
    ('operators:write',   'Creare și editare operatori'),
    ('operators:delete',  'Ștergere operatori'),
    ('assignments:read',  'Vizualizare asignări'),
    ('assignments:write', 'Creare și editare asignări'),
    ('assignments:delete','Ștergere asignări'),
    ('roles:read',        'Vizualizare roluri și permisiuni'),
    ('roles:write',       'Creare și editare roluri și permisiuni'),
    ('roles:delete',      'Ștergere roluri'),
    ('users:read',        'Vizualizare utilizatori'),
    ('users:write',       'Modificare rol utilizator')
ON CONFLICT (name) DO NOTHING;

-- ─── Roles ──────────────────────────────────────────────────
INSERT INTO roles (code, name, description) VALUES
  ('admin',   'Administrator', 'Administrator complet — acces total'),
  ('manager', 'Manager',       'Manager — administrare resurse, fără gestiunea rolurilor'),
  ('viewer',  'Vizualizator',  'Vizualizator — acces doar citire')
ON CONFLICT (code) DO NOTHING;

-- ─── Role → Permissions ──────────────────────────────────────

-- admin: toate permisiunile
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

-- manager: read+write pe machines/operators/assignments + roles:read + users:read
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'manager'
  AND p.name IN (
      'machines:read',    'machines:write',
  'fields:read',      'fields:write',
      'operators:read',   'operators:write',
      'assignments:read', 'assignments:write',
      'roles:read',
      'users:read'
  )
ON CONFLICT DO NOTHING;

-- viewer: doar :read pe toate resursele
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'viewer'
  AND p.name IN (
      'machines:read',
  'fields:read',
      'operators:read',
      'assignments:read',
      'roles:read',
      'users:read'
  )
ON CONFLICT DO NOTHING;

-- ─── Atribuie rolul admin primului user existent (dacă există) ────────────────
UPDATE users
SET role_id = (SELECT id FROM roles WHERE code = 'admin')
WHERE id = (SELECT id FROM users ORDER BY id LIMIT 1)
  AND role_id IS NULL;
