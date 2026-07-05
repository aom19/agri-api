-- Curăță datele existente (CASCADE șterge și asignările dependente)
TRUNCATE TABLE assignments, operators, machines RESTART IDENTITY CASCADE;

-- Seed: mașini agricole (format nume: <TIP>-<BRAND>-<MODEL>-<NR>)
INSERT INTO machines (
    name,
    code,
    type,
    brand,
    model,
    year,
    registration_number,
    fuel_type,
    status,
    notes
) VALUES
    ('TR-JOHNDEERE-6R185-01', 'MCH-0001', 'tractor', 'John Deere', '6R 185', 2018, 'B-AG-0001', 'diesel', 'active', 'Tractor principal lucrari camp'),
    ('TR-FENDT-724VARIO-02', 'MCH-0002', 'tractor', 'Fendt', '724 Vario', 2019, 'B-AG-0002', 'diesel', 'active', 'Tractor pentru transport'),
    ('TR-NEWHOLLAND-T7270-03', 'MCH-0003', 'tractor', 'New Holland', 'T7.270', 2017, 'B-AG-0003', 'diesel', 'maintenance', 'Revizie motor programata'),
    ('TR-CASEIH-PUMA240-04', 'MCH-0004', 'tractor', 'Case IH', 'Puma 240', 2020, 'B-AG-0004', 'diesel', 'active', 'Tractor pentru arat'),
    ('TR-MASSEY-7720-05', 'MCH-0005', 'tractor', 'Massey Ferguson', '7720', 2016, 'B-AG-0005', 'diesel', 'inactive', 'Rezerva sezon'),
    ('TR-VALTRA-T254-06', 'MCH-0006', 'tractor', 'Valtra', 'T254', 2022, 'B-AG-0006', 'diesel', 'active', 'Utilaj nou'),
    ('TR-SAME-EXPLORER120-07', 'MCH-0007', 'tractor', 'Same', 'Explorer 120', 2015, 'B-AG-0007', 'diesel', 'active', 'Lucrari usoare'),
    ('TR-ZETOR-FORTERRA140-08', 'MCH-0008', 'tractor', 'Zetor', 'Forterra 140', 2016, 'B-AG-0008', 'diesel', 'maintenance', 'Verificare transmisie'),
    ('DR-DJI-AGRAST30-01', 'MCH-0009', 'drone', 'DJI', 'Agras T30', 2022, 'B-AG-0009', 'electric', 'active', 'Monitorizare culturi'),
    ('DR-DJI-AGRAST40-02', 'MCH-0010', 'drone', 'DJI', 'Agras T40', 2023, 'B-AG-0010', 'electric', 'active', 'Stropiri localizate'),
    ('CB-CLAAS-LEXION780-01', 'MCH-0011', 'combine', 'Claas', 'Lexion 780', 2019, 'B-AG-0011', 'diesel', 'active', 'Recoltare grau'),
    ('CB-JOHNDEERE-S780-02', 'MCH-0012', 'combine', 'John Deere', 'S780', 2021, 'B-AG-0012', 'diesel', 'active', 'Recoltare porumb'),
    ('CB-FENDT-IDEAL9T-03', 'MCH-0013', 'combine', 'Fendt', 'Ideal 9T', 2021, 'B-AG-0013', 'diesel', 'maintenance', 'Service pre-campanie'),
    ('CAR-DACIA-DUSTER-01', 'MCH-0014', 'car', 'Dacia', 'Duster', 2020, 'B-AG-0014', 'gasoline', 'active', 'Masina deplasari teren'),
    ('CAR-DACIA-DUSTER-02', 'MCH-0015', 'car', 'Dacia', 'Duster', 2021, 'B-AG-0015', 'gasoline', 'active', 'Masina echipa tehnica'),
    ('CAR-DACIA-DUSTER-03', 'MCH-0016', 'car', 'Dacia', 'Duster', 2022, 'B-AG-0016', 'hybrid', 'inactive', 'Back-up administrativ'),
    ('ST-MERCEDES-SPRINTER-01', 'MCH-0017', 'small_truck', 'Mercedes', 'Sprinter', 2019, 'B-AG-0017', 'diesel', 'active', 'Transport piese si echipamente'),
    ('ST-MERCEDES-SPRINTER-02', 'MCH-0018', 'small_truck', 'Mercedes', 'Sprinter', 2021, 'B-AG-0018', 'diesel', 'maintenance', 'Revizie flota')
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
SELECT
    m.id,
    o.id,
    NOW() - (l.start_days || ' days')::interval,
    CASE WHEN l.status = 'closed' THEN NOW() - (l.end_days || ' days')::interval ELSE NULL END,
    l.status
FROM (
    VALUES
        ('MCH-0001', 'Alexandru Ionescu', 30, 25, 'closed'),
        ('MCH-0002', 'Mihai Popescu', 20, 15, 'closed'),
        ('MCH-0003', 'Gheorghe Dănilă', 10, NULL, 'active'),
        ('MCH-0004', 'Ion Constantin', 5, NULL, 'active'),
        ('MCH-0005', 'Vasile Marin', 60, 50, 'closed'),
        ('MCH-0006', 'Dumitru Florescu', 45, 40, 'closed'),
        ('MCH-0007', 'Nicolae Stancu', 3, NULL, 'active'),
        ('MCH-0008', 'Florin Gheorghiu', 90, 80, 'closed'),
        ('MCH-0009', 'Octavian Rus', 7, NULL, 'active'),
        ('MCH-0010', 'Petru Moldovan', 15, 10, 'closed'),
        ('MCH-0011', 'Bogdan Stoica', 120, 100, 'closed'),
        ('MCH-0012', 'Radu Nistor', 2, NULL, 'active'),
        ('MCH-0013', 'Sorin Enache', 50, 45, 'closed'),
        ('MCH-0014', 'Nicolae Stancu', 1, NULL, 'active'),
        ('MCH-0015', 'Petru Moldovan', 8, 3, 'closed')
) AS l(machine_code, operator_name, start_days, end_days, status)
JOIN machines m ON m.code = l.machine_code
JOIN operators o ON o.name = l.operator_name
ON CONFLICT DO NOTHING;
