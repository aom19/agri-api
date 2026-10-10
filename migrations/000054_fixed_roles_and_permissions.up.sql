-- T18: rolurile sunt fixe și definite doar aici. Înainte, permisiunile și rolurile veneau din 13
-- migrații plus seeds/rbac_seed.sql, iar ordinea conta: pe o bază nouă, migrațiile rulau
-- înaintea seed-ului, deci managerul nu primea operațiunile pe teren sau notificările.
--
-- Permisiunile sunt `modul:read` și `modul:write`; write include crearea, editarea, ștergerea,
-- dezactivarea și reactivarea. Singura excepție e field_operations:execute: operatorul pornește
-- și finalizează lucrările lui, dar nu le planifică. Rolurile nu se mai editează din aplicație,
-- deci rămâne doar roles:read.
--
-- Migrația aduce baza la starea de mai jos indiferent de starea de dinainte, deci poate fi
-- rulată din nou. O permisiune nouă cere o migrație nouă, care o adaugă și o dă rolurilor.

CREATE TEMP TABLE t18_permissions (name VARCHAR(100) PRIMARY KEY, description TEXT NOT NULL);
INSERT INTO t18_permissions (name, description) VALUES
    ('dashboard:read',           'Vizualizare dashboard'),
    ('reports:read',             'Vizualizare rapoarte și trimiterea lor pe e-mail'),
    ('notifications:read',       'Vizualizare notificări proprii'),
    ('audit:read',               'Vizualizare jurnal de audit'),
    ('roles:read',               'Vizualizare roluri și permisiuni'),
    ('users:read',               'Vizualizare utilizatori'),
    ('users:write',              'Creare, editare, dezactivare și schimbarea rolului utilizatorilor'),
    ('fields:read',              'Vizualizare terenuri'),
    ('fields:write',             'Creare, editare și ștergere terenuri'),
    ('crops:read',               'Vizualizare sezoane, culturi și culturi pe terenuri'),
    ('crops:write',              'Administrare sezoane, culturi, culturi pe terenuri și recoltă'),
    ('machines:read',            'Vizualizare mașini'),
    ('machines:write',           'Creare, editare și ștergere mașini'),
    ('implements:read',          'Vizualizare echipamente agricole'),
    ('implements:write',         'Creare, editare, ștergere și activare echipamente agricole'),
    ('operators:read',           'Vizualizare operatori'),
    ('operators:write',          'Creare, editare, ștergere și dezactivare operatori'),
    ('resources:read',           'Vizualizare resurse și stocuri'),
    ('resources:write',          'Creare, editare și ștergere resurse, mișcări de stoc'),
    ('operations:read',          'Vizualizare template-uri de operațiuni'),
    ('operations:write',         'Creare, editare și ștergere template-uri de operațiuni'),
    ('field_operations:read',    'Vizualizare operațiuni pe teren'),
    ('field_operations:write',   'Planificare, editare și ștergere operațiuni pe teren'),
    ('field_operations:execute', 'Pornire și finalizare operațiuni pe teren (timpi reali, consumuri)');

INSERT INTO permissions (name, description)
SELECT name, description FROM t18_permissions
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

-- formatele vechi (resource.*, resource_type.*, work_resource.*), permissions:read, :delete,
-- :disable, :enable, :start și :complete; role_permissions se curăță prin ON DELETE CASCADE
DELETE FROM permissions WHERE name NOT IN (SELECT name FROM t18_permissions);

-- ─── Roluri ──────────────────────────────────────────────────────────────────

-- un rol creat din aplicație cu numele unui rol fix ar bloca inserarea (roles.name e unic)
UPDATE roles SET name = name || ' (#' || id || ')'
WHERE code NOT IN ('admin', 'manager', 'operator', 'viewer')
  AND name IN ('Administrator', 'Manager', 'Operator', 'Vizualizator');

INSERT INTO roles (code, name, description) VALUES
    ('admin',    'Administrator', 'Acces total, inclusiv utilizatori și jurnalul de audit'),
    ('manager',  'Manager',       'Administrează ferma: terenuri, flotă, operatori, resurse și lucrări'),
    ('operator', 'Operator',      'Operator de teren: vede, pornește și finalizează lucrările proprii'),
    ('viewer',   'Vizualizator',  'Doar citire pe datele fermei')
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;

-- Rolurile create din aplicație (de exemplu TestC) dispar. Conturile lor primesc viewer, rolul
-- oricărui cont nou; un admin le poate da apoi alt rol.
UPDATE users
SET role_id = (SELECT id FROM roles WHERE code = 'viewer')
WHERE role_id IN (SELECT id FROM roles WHERE code NOT IN ('admin', 'manager', 'operator', 'viewer'));

DELETE FROM roles WHERE code NOT IN ('admin', 'manager', 'operator', 'viewer');

-- ─── Permisiunile fiecărui rol ───────────────────────────────────────────────

DELETE FROM role_permissions;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON
    -- admin: tot
    r.code = 'admin'
    -- manager: tot, mai puțin utilizatorii, rolurile și jurnalul de audit
    OR (r.code = 'manager' AND p.name NOT IN ('users:read', 'users:write', 'roles:read', 'audit:read'))
    -- operator: lucrările proprii (filtrate în backend după cont) și culturile lor
    OR (r.code = 'operator' AND p.name IN (
        'dashboard:read', 'notifications:read', 'crops:read',
        'field_operations:read', 'field_operations:execute'))
    -- viewer: citire pe datele fermei, fără utilizatori, roluri și audit
    OR (r.code = 'viewer' AND p.name LIKE '%:read'
        AND p.name NOT IN ('users:read', 'roles:read', 'audit:read'));

DROP TABLE t18_permissions;
