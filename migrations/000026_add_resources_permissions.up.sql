-- Add resource permissions and grant them to the admin role only.

INSERT INTO permissions (name, description)
VALUES
    ('resource_type.view',   'Vizualizare tipuri resurse'),
    ('resource_type.create', 'Creare tipuri resurse'),
    ('resource_type.update', 'Editare tipuri resurse'),
    ('resource_type.delete', 'Ștergere tipuri resurse'),
    ('resource.view',        'Vizualizare resurse'),
    ('resource.create',      'Creare resurse'),
    ('resource.update',      'Editare resurse'),
    ('resource.delete',      'Ștergere resurse'),
    ('stock.view',           'Vizualizare stocuri'),
    ('stock.create',         'Creare stocuri'),
    ('stock.update',         'Editare stocuri'),
    ('stock.delete',         'Ștergere stocuri'),
    ('work_resource.view',   'Vizualizare consum resurse'),
    ('work_resource.create', 'Creare consum resurse'),
    ('work_resource.update', 'Editare consum resurse'),
    ('work_resource.delete', 'Ștergere consum resurse')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'resource_type.view', 'resource_type.create', 'resource_type.update', 'resource_type.delete',
    'resource.view',      'resource.create',      'resource.update',      'resource.delete',
    'stock.view',         'stock.create',         'stock.update',         'stock.delete',
    'work_resource.view', 'work_resource.create', 'work_resource.update', 'work_resource.delete'
)
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;
