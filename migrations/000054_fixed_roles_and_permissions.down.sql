-- Readuce permisiunile și rolurile de dinainte de T18, cum le lăsau migrațiile plus
-- seeds/rbac_seed.sql. Rolurile create din aplicație și rolul vechi al conturilor mutate pe
-- viewer nu se recuperează.

CREATE TEMP TABLE t18_permissions (name VARCHAR(100) PRIMARY KEY, description TEXT NOT NULL);
INSERT INTO t18_permissions (name, description) VALUES
    ('audit:read',                'Vizualizare jurnal de audit'),
    ('crops:read',                'Vizualizare sezoane, culturi și culturi pe terenuri'),
    ('crops:write',               'Administrare sezoane, culturi și culturi pe terenuri'),
    ('dashboard:read',            'Vizualizare carduri dashboard'),
    ('field_operations:complete', 'Finalizare operațiuni pe teren (timpi reali, consumuri)'),
    ('field_operations:delete',   'Ștergere operațiuni pe teren'),
    ('field_operations:read',     'Vizualizare operațiuni pe teren'),
    ('field_operations:start',    'Pornire operațiuni pe teren'),
    ('field_operations:write',    'Creare și editare operațiuni pe teren'),
    ('fields:delete',             'Stergere terenuri'),
    ('fields:read',               'Vizualizare terenuri'),
    ('fields:write',              'Creare si editare terenuri'),
    ('implements:delete',         'Ștergere implementuri'),
    ('implements:read',           'Vizualizare implementuri'),
    ('implements:write',          'Creare și editare implementuri'),
    ('machines:delete',           'Ștergere mașini'),
    ('machines:read',             'Vizualizare mașini'),
    ('machines:write',            'Creare și editare mașini'),
    ('notifications:read',        'Vizualizare notificări proprii'),
    ('operations:delete',         'Ștergere template-uri de operațiuni'),
    ('operations:read',           'Vizualizare template-uri de operațiuni'),
    ('operations:write',          'Creare și editare template-uri de operațiuni'),
    ('operators:delete',          'Ștergere operatori'),
    ('operators:disable',         'Dezactivare și reactivare operatori'),
    ('operators:read',            'Vizualizare operatori'),
    ('operators:write',           'Creare și editare operatori'),
    ('permissions:read',          'Vizualizare permisiuni'),
    ('reports:read',              'Vizualizare rapoarte'),
    ('resource.create',           'Creare resurse'),
    ('resource.delete',           'Ștergere resurse'),
    ('resources:delete',          'Ștergere resurse'),
    ('resources:read',            'Vizualizare resurse și stocuri'),
    ('resources:write',           'Creare și editare resurse, mișcări de stoc'),
    ('resource_type.create',      'Creare tipuri resurse'),
    ('resource_type.delete',      'Ștergere tipuri resurse'),
    ('resource_type.update',      'Editare tipuri resurse'),
    ('resource_type.view',        'Vizualizare tipuri resurse'),
    ('resource.update',           'Editare resurse'),
    ('resource.view',             'Vizualizare resurse'),
    ('roles:delete',              'Ștergere roluri'),
    ('roles:read',                'Vizualizare roluri și permisiuni'),
    ('roles:write',               'Creare și editare roluri și permisiuni'),
    ('users:disable',             'Dezactivare utilizatori'),
    ('users:enable',              'Reactivare utilizatori'),
    ('users:read',                'Vizualizare utilizatori'),
    ('users:write',               'Modificare rol utilizator'),
    ('work_resource.create',      'Creare consum resurse'),
    ('work_resource.delete',      'Ștergere consum resurse'),
    ('work_resource.update',      'Editare consum resurse'),
    ('work_resource.view',        'Vizualizare consum resurse');

INSERT INTO permissions (name, description)
SELECT name, description FROM t18_permissions
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

DELETE FROM permissions WHERE name NOT IN (SELECT name FROM t18_permissions);

UPDATE roles SET description = CASE code
    WHEN 'admin'    THEN 'Administrator complet — acces total'
    WHEN 'manager'  THEN 'Manager — administrare resurse, fără gestiunea rolurilor'
    WHEN 'operator' THEN 'Operator de teren — acces la lucrările și alocările proprii'
    WHEN 'viewer'   THEN 'Fallback role for migrated users'
END
WHERE code IN ('admin', 'manager', 'operator', 'viewer');

DELETE FROM role_permissions;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON
    r.code = 'admin'
    OR (r.code = 'manager' AND p.name IN (
        'crops:read', 'crops:write', 'dashboard:read', 'field_operations:complete',
        'fields:read', 'fields:write', 'implements:delete', 'implements:read', 'implements:write',
        'machines:read', 'machines:write', 'operations:read', 'operations:write',
        'operators:disable', 'operators:read', 'operators:write', 'reports:read',
        'resources:delete', 'resources:read', 'resources:write'))
    OR (r.code = 'operator' AND p.name IN (
        'crops:read', 'dashboard:read', 'field_operations:complete', 'field_operations:read',
        'field_operations:start', 'notifications:read'))
    OR (r.code = 'viewer' AND p.name IN (
        'crops:read', 'dashboard:read', 'field_operations:read', 'fields:read', 'machines:read',
        'operations:read', 'operators:read', 'reports:read', 'resources:read'));

DROP TABLE t18_permissions;
