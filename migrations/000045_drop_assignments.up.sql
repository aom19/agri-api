-- Alocările sunt operațiunile pe teren planificate sau în lucru; tabelul assignments nu era
-- folosit de aplicație, iar rapoartele citeau din el cifre care nu reflectau nimic real.
-- role_permissions se curăță prin ON DELETE CASCADE.
DELETE FROM permissions WHERE name IN ('assignments:read', 'assignments:write', 'assignments:delete');

DROP TABLE IF EXISTS assignments;
