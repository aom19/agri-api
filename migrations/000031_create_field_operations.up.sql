CREATE TABLE IF NOT EXISTS field_operations (
    id                    BIGSERIAL PRIMARY KEY,
    field_id              UUID NOT NULL REFERENCES fields(id),
    operation_type_id     BIGINT NOT NULL REFERENCES operation_types(id),
    operation_template_id BIGINT REFERENCES operation_templates(id),
    machine_id            BIGINT REFERENCES machines(id),
    implement_id          BIGINT REFERENCES implements(id),
    operator_id           BIGINT REFERENCES operators(id),
    planned_start_at      TIMESTAMPTZ,
    planned_end_at        TIMESTAMPTZ,
    area_planned_ha       DOUBLE PRECISION,
    notes                 TEXT NOT NULL DEFAULT '',
    status                VARCHAR(50) NOT NULL DEFAULT 'planned',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT field_operations_status_check
        CHECK (status IN ('planned', 'in_progress', 'completed', 'canceled')),
    CONSTRAINT field_operations_area_check
        CHECK (area_planned_ha IS NULL OR area_planned_ha >= 0),
    CONSTRAINT field_operations_dates_check
        CHECK (
            planned_start_at IS NULL
            OR planned_end_at IS NULL
            OR planned_end_at >= planned_start_at
        )
);

CREATE INDEX IF NOT EXISTS idx_field_operations_field_id           ON field_operations(field_id);
CREATE INDEX IF NOT EXISTS idx_field_operations_operation_type_id  ON field_operations(operation_type_id);
CREATE INDEX IF NOT EXISTS idx_field_operations_template_id        ON field_operations(operation_template_id);
CREATE INDEX IF NOT EXISTS idx_field_operations_machine_id         ON field_operations(machine_id);
CREATE INDEX IF NOT EXISTS idx_field_operations_operator_id        ON field_operations(operator_id);
CREATE INDEX IF NOT EXISTS idx_field_operations_status             ON field_operations(status);
CREATE INDEX IF NOT EXISTS idx_field_operations_planned_start_at   ON field_operations(planned_start_at);

INSERT INTO permissions (name, description) VALUES
    ('field_operations:read',   'Vizualizare operațiuni pe teren'),
    ('field_operations:write',  'Creare și editare operațiuni pe teren'),
    ('field_operations:delete', 'Ștergere operațiuni pe teren')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
  AND p.name IN ('field_operations:read', 'field_operations:write', 'field_operations:delete')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'manager'
  AND p.name IN ('field_operations:read', 'field_operations:write')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'viewer'
  AND p.name = 'field_operations:read'
ON CONFLICT DO NOTHING;
