-- Curăță datele existente (CASCADE șterge și asignările dependente)
TRUNCATE TABLE assignments, operators, machines RESTART IDENTITY CASCADE;

-- Seed: 10 mașini agricole
INSERT INTO machines (name, type, status, description) VALUES
    ('John Deere 8R 410',     'Tractor',      'available',  'Tractor de mare putere, 410 CP, 4WD'),
    ('Case IH Axial-Flow 250','Combină',       'available',  'Combină de recoltat cereale, 502 CP'),
    ('Fendt 724 Vario',       'Tractor',      'in_use',     'Tractor cu transmisie continuă variabilă, 240 CP'),
    ('New Holland T7.315',    'Tractor',      'available',  'Tractor cu Blue Power, 315 CP'),
    ('Claas Lexion 8900',     'Combină',      'in_service', 'Cea mai mare combină din lume, 790 CP'),
    ('Amazone ZA-TS 4200',    'Distribuitor', 'available',  'Distribuitor de îngrășăminte, 4200 L'),
    ('Horsch Joker 12 RT',    'Cultivator',   'available',  'Cultivator disc, 12 m lățime de lucru'),
    ('Kuhn Merge Maxx 902',   'Greblă',       'inactive',   'Greblă rotativă cu 9 rotoare, 9 m'),
    ('Väderstad Tempo V 16',  'Semănătoare',  'available',  'Semănătoare de precizie, 16 rânduri'),
    ('Krone BiG X 1180',      'Tocat furaje', 'in_use',     'Tocător autopropulsat, 1156 CP')
ON CONFLICT DO NOTHING;

-- Seed: 15 operatori
INSERT INTO operators (name, status) VALUES
    ('Alexandru Ionescu', 'active'),
    ('Mihai Popescu',     'active'),
    ('Gheorghe Dănilă',   'active'),
    ('Ion Constantin',    'active'),
    ('Vasile Marin',      'active'),
    ('Dumitru Florescu',  'active'),
    ('Nicolae Stancu',    'active'),
    ('Florin Gheorghiu',  'active'),
    ('Octavian Rus',      'active'),
    ('Petru Moldovan',    'active'),
    ('Andrei Popa',       'inactive'),
    ('Cristian Luca',     'inactive'),
    ('Bogdan Stoica',     'active'),
    ('Radu Nistor',       'active'),
    ('Sorin Enache',      'active')
ON CONFLICT DO NOTHING;

-- Seed: 15 asignări — referință după nume, nu ID hardcodat
INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '30 days', NOW() - INTERVAL '25 days', 'closed'
FROM machines m, operators o WHERE m.name = 'John Deere 8R 410'     AND o.name = 'Alexandru Ionescu'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '20 days', NOW() - INTERVAL '15 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Case IH Axial-Flow 250' AND o.name = 'Mihai Popescu'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '10 days', NULL, 'active'
FROM machines m, operators o WHERE m.name = 'Fendt 724 Vario'        AND o.name = 'Gheorghe Dănilă'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '5 days', NULL, 'active'
FROM machines m, operators o WHERE m.name = 'New Holland T7.315'     AND o.name = 'Ion Constantin'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '60 days', NOW() - INTERVAL '50 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Claas Lexion 8900'      AND o.name = 'Vasile Marin'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '45 days', NOW() - INTERVAL '40 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Amazone ZA-TS 4200'     AND o.name = 'Dumitru Florescu'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '3 days', NULL, 'active'
FROM machines m, operators o WHERE m.name = 'Horsch Joker 12 RT'     AND o.name = 'Nicolae Stancu'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '90 days', NOW() - INTERVAL '80 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Kuhn Merge Maxx 902'    AND o.name = 'Florin Gheorghiu'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '7 days', NULL, 'active'
FROM machines m, operators o WHERE m.name = 'Väderstad Tempo V 16'   AND o.name = 'Octavian Rus'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '15 days', NOW() - INTERVAL '10 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Krone BiG X 1180'       AND o.name = 'Petru Moldovan'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '120 days', NOW() - INTERVAL '100 days', 'closed'
FROM machines m, operators o WHERE m.name = 'John Deere 8R 410'      AND o.name = 'Bogdan Stoica'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '2 days', NULL, 'active'
FROM machines m, operators o WHERE m.name = 'Fendt 724 Vario'        AND o.name = 'Radu Nistor'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '50 days', NOW() - INTERVAL '45 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Claas Lexion 8900'      AND o.name = 'Sorin Enache'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '1 days', NULL, 'active'
FROM machines m, operators o WHERE m.name = 'Horsch Joker 12 RT'     AND o.name = 'Nicolae Stancu'
ON CONFLICT DO NOTHING;

INSERT INTO assignments (machine_id, operator_id, start_date, end_date, status)
SELECT m.id, o.id, NOW() - INTERVAL '8 days', NOW() - INTERVAL '3 days', 'closed'
FROM machines m, operators o WHERE m.name = 'Väderstad Tempo V 16'   AND o.name = 'Petru Moldovan'
ON CONFLICT DO NOTHING;
