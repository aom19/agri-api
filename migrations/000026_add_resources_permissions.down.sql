DELETE FROM permissions
WHERE name IN (
    'resource_type.view', 'resource_type.create', 'resource_type.update', 'resource_type.delete',
    'resource.view',      'resource.create',      'resource.update',      'resource.delete',
    'stock.view',         'stock.create',         'stock.update',         'stock.delete',
    'work_resource.view', 'work_resource.create', 'work_resource.update', 'work_resource.delete'
);
