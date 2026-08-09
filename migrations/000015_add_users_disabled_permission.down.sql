-- Revert 000015 by removing users:disable and users:enable permissions.

DELETE FROM permissions WHERE name IN ('users:enable', 'users:disable');
