-- ============================================================
-- RBAC seed: roles, permissions, role_permissions
-- Rulează DUPĂ migrate-up (migration 010)
-- ============================================================

-- ─── Permissions ────────────────────────────────────────────
INSERT INTO permissions (name, description) VALUES
    ('machines:read',     'Vizualizare mașini'),
    ('machines:write',    'Creare și editare mașini'),
    ('machines:delete',   'Ștergere mașini'),
  ('implements:read',   'Vizualizare Echipamente agricole'),
  ('implements:write',  'Creare și editare Echipamente agricole'),
  ('implements:delete', 'Ștergere Echipamente agricole'),
    ('fields:read',       'Vizualizare terenuri'),
    ('fields:write',      'Creare și editare terenuri'),
    ('fields:delete',     'Ștergere terenuri'),
    ('operators:read',    'Vizualizare operatori'),
    ('operators:write',   'Creare și editare operatori'),
    ('operators:delete',  'Ștergere operatori'),
    ('operators:disable', 'Dezactivare și reactivare operatori'),
    ('resources:read',    'Vizualizare resurse și stocuri'),
    ('resources:write',   'Creare și editare resurse, mișcări de stoc'),
    ('resources:delete',  'Ștergere resurse'),
    ('roles:read',        'Vizualizare roluri și permisiuni'),
    ('roles:write',       'Creare și editare roluri și permisiuni'),
    ('roles:delete',      'Ștergere roluri'),
    ('users:read',        'Vizualizare utilizatori'),
    ('users:write',       'Modificare rol utilizator'),
    ('users:disable',     'Dezactivare utilizatori'),
    ('users:enable',      'Reactivare utilizatori'),
    ('permissions:read',  'Vizualizare permisiuni'),
    ('dashboard:read',    'Vizualizare carduri dashboard'),
    ('operations:read',   'Vizualizare template-uri de operațiuni'),
    ('operations:write',  'Creare și editare template-uri de operațiuni'),
    ('operations:delete', 'Ștergere template-uri de operațiuni'),
    ('reports:read',      'Vizualizare rapoarte'),
    ('field_operations:complete', 'Finalizare operațiuni pe teren (timpi reali, consumuri)'),
    ('crops:read',        'Vizualizare sezoane, culturi și culturi pe terenuri'),
    ('crops:write',       'Administrare sezoane, culturi și culturi pe terenuri')

ON CONFLICT (name) DO NOTHING;

-- ─── Roles ──────────────────────────────────────────────────
INSERT INTO roles (code, name, description) VALUES
  ('admin',   'Administrator', 'Administrator complet — acces total'),
  ('manager', 'Manager',       'Manager — administrare resurse, fără gestiunea rolurilor'),
  ('operator','Operator',      'Operator de teren — acces la lucrările și alocările proprii'),
  ('viewer',  'Vizualizator',  'Vizualizator — acces doar citire')
ON CONFLICT (code) DO NOTHING;

-- ─── Role → Permissions ──────────────────────────────────────

-- admin: toate permisiunile
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

-- admin: asigură explicit dreptul de reactivare utilizatori
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
  AND p.name = 'users:enable'
ON CONFLICT DO NOTHING;

-- manager: read+write pe machines/operators
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'manager'
  AND p.name IN (
      'machines:read',    'machines:write',
      'implements:read',  'implements:write', 'implements:delete',
      'fields:read',      'fields:write',
      'operators:read',   'operators:write',  'operators:disable',
      'resources:read',   'resources:write',  'resources:delete',
      'dashboard:read',   'reports:read',
      'operations:read',  'operations:write',
      'field_operations:complete',
      'crops:read',       'crops:write'
  )
ON CONFLICT DO NOTHING;

-- viewer: doar :read pe resursele operaționale
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'viewer'
  AND p.name IN (
      'machines:read',
      'resources:read',
      'fields:read',
      'operators:read',
      'dashboard:read',
      'reports:read',
      'crops:read',
      'operations:read'
  )
ON CONFLICT DO NOTHING;

-- operator: citire doar pentru dashboard și operațiunile pe teren asignate
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'operator'
  AND p.name IN (
      'dashboard:read',
      'field_operations:read',
      'field_operations:complete',
      'crops:read'
  )
ON CONFLICT DO NOTHING;

-- ─── Atribuie rolul admin primului user existent (dacă există) ────────────────
UPDATE users
SET role_id = (SELECT id FROM roles WHERE code = 'admin')
WHERE id = (SELECT id FROM users ORDER BY id LIMIT 1)
  AND role_id IS NULL;
