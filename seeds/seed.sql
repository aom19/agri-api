-- Curăță datele existente (CASCADE șterge și asignările dependente)
TRUNCATE TABLE assignments, operators, machines RESTART IDENTITY CASCADE;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'implement_compatibilities') THEN
        EXECUTE 'TRUNCATE TABLE implement_compatibilities RESTART IDENTITY';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'implements') THEN
        EXECUTE 'TRUNCATE TABLE implements RESTART IDENTITY';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'resource_types')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'resources')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stocks') THEN
        EXECUTE 'TRUNCATE TABLE stocks, resources, resource_types RESTART IDENTITY';
    END IF;
END $$;

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
    operating_hours,
    asset_status,
    notes
) VALUES
    ('TR-JOHNDEERE-6R185-01', 'MCH-0001', 'tractor', 'John Deere', '6R 185', 2018, 'B-AG-0001', 'diesel', 6420, 'active', 'Tractor principal lucrari camp'),
    ('TR-FENDT-724VARIO-02', 'MCH-0002', 'tractor', 'Fendt', '724 Vario', 2019, 'B-AG-0002', 'diesel', 5190, 'active', 'Tractor pentru transport'),
    ('TR-NEWHOLLAND-T7270-03', 'MCH-0003', 'tractor', 'New Holland', 'T7.270', 2017, 'B-AG-0003', 'diesel', 7340, 'maintenance', 'Revizie motor programata'),
    ('TR-CASEIH-PUMA240-04', 'MCH-0004', 'tractor', 'Case IH', 'Puma 240', 2020, 'B-AG-0004', 'diesel', 4810, 'active', 'Tractor pentru arat'),
    ('TR-MASSEY-7720-05', 'MCH-0005', 'tractor', 'Massey Ferguson', '7720', 2016, 'B-AG-0005', 'diesel', 8125, 'inactive', 'Rezerva sezon'),
    ('TR-VALTRA-T254-06', 'MCH-0006', 'tractor', 'Valtra', 'T254', 2022, 'B-AG-0006', 'diesel', 2130, 'active', 'Utilaj nou'),
    ('TR-SAME-EXPLORER120-07', 'MCH-0007', 'tractor', 'Same', 'Explorer 120', 2015, 'B-AG-0007', 'diesel', 9020, 'active', 'Lucrari usoare'),
    ('TR-ZETOR-FORTERRA140-08', 'MCH-0008', 'tractor', 'Zetor', 'Forterra 140', 2016, 'B-AG-0008', 'diesel', 6880, 'maintenance', 'Verificare transmisie'),
    ('DR-DJI-AGRAST30-01', 'MCH-0009', 'drone', 'DJI', 'Agras T30', 2022, 'B-AG-0009', 'electric', 940, 'active', 'Monitorizare culturi'),
    ('DR-DJI-AGRAST40-02', 'MCH-0010', 'drone', 'DJI', 'Agras T40', 2023, 'B-AG-0010', 'electric', 610, 'active', 'Stropiri localizate'),
    ('CB-CLAAS-LEXION780-01', 'MCH-0011', 'combine', 'Claas', 'Lexion 780', 2019, 'B-AG-0011', 'diesel', 3890, 'active', 'Recoltare grau'),
    ('CB-JOHNDEERE-S780-02', 'MCH-0012', 'combine', 'John Deere', 'S780', 2021, 'B-AG-0012', 'diesel', 2750, 'active', 'Recoltare porumb'),
    ('CB-FENDT-IDEAL9T-03', 'MCH-0013', 'combine', 'Fendt', 'Ideal 9T', 2021, 'B-AG-0013', 'diesel', 3180, 'maintenance', 'Service pre-campanie'),
    ('CAR-DACIA-DUSTER-01', 'MCH-0014', 'car', 'Dacia', 'Duster', 2020, 'B-AG-0014', 'gasoline', 2260, 'active', 'Masina deplasari teren'),
    ('CAR-DACIA-DUSTER-02', 'MCH-0015', 'car', 'Dacia', 'Duster', 2021, 'B-AG-0015', 'gasoline', 1980, 'active', 'Masina echipa tehnica'),
    ('CAR-DACIA-DUSTER-03', 'MCH-0016', 'car', 'Dacia', 'Duster', 2022, 'B-AG-0016', 'hybrid', 1290, 'inactive', 'Back-up administrativ'),
    ('ST-MERCEDES-SPRINTER-01', 'MCH-0017', 'small_truck', 'Mercedes', 'Sprinter', 2019, 'B-AG-0017', 'diesel', 4560, 'active', 'Transport piese si echipamente'),
    ('ST-MERCEDES-SPRINTER-02', 'MCH-0018', 'small_truck', 'Mercedes', 'Sprinter', 2021, 'B-AG-0018', 'diesel', 3010, 'maintenance', 'Revizie flota'),
    ('SP-HARDI-ALPHAEVO-01', 'MCH-0019', 'sprayer', 'Hardi', 'Alpha Evo', 2020, 'B-AG-0019', 'diesel', 2870, 'active', 'Pulverizator autopropulsat pentru tratamente')
ON CONFLICT DO NOTHING;

-- Seed: implementuri (echipamente atașabile, fără motor propriu)
INSERT INTO implements (
    name,
    code,
    type,
    brand,
    model,
    year,
    working_width,
    capacity,
    status,
    notes
) VALUES
    ('PL-LEMKEN-JUWEL8-01', 'IMP-0001', 'plow', 'Lemken', 'Juwel 8', 2020, 2.4, NULL, 'active', 'Plug reversibil pentru arat'),
    ('SE-HORSCH-PRONTO6DC-01', 'IMP-0002', 'seeder', 'Horsch', 'Pronto 6 DC', 2021, 6.0, NULL, 'active', 'Semanatoare cereale paioase'),
    ('FS-AMAZONE-ZA-TS4200-01', 'IMP-0003', 'fertilizer_spreader', 'Amazone', 'ZA-TS 4200', 2019, NULL, 4200, 'active', 'Distribuitor ingrasaminte'),
    ('SP-RAUCH-AERO32-01', 'IMP-0004', 'sprayer', 'Rauch', 'Aero 32', 2022, 32.0, 3200, 'maintenance', 'Pulverizator tractat'),
    ('TR-KRAMPE-HALFPIPE-01', 'IMP-0005', 'trailer', 'Krampe', 'Halfpipe', 2018, NULL, 18000, 'active', 'Remorca transport cereale'),
    ('HD-CLAAS-VARIO1080-01', 'IMP-0006', 'header', 'Claas', 'Vario 1080', 2020, 10.8, NULL, 'active', 'Header pentru combine'),
    ('DH-KVERNELAND-QUALIDISC-01', 'IMP-0007', 'disc_harrow', 'Kverneland', 'Qualidisc', 2019, 4.0, NULL, 'active', 'Grapa cu discuri pentru pregatirea patului germinativ'),
    ('CV-KONGSKILDE-VIBROFLEX-01', 'IMP-0008', 'cultivator', 'Kongskilde', 'Vibro Flex', 2017, 4.5, NULL, 'active', 'Cultivator pentru lucrari superficiale'),
    ('OT-UNIVERSAL-PLATFORM-01', 'IMP-0009', 'other', 'Universal', 'Platform', 2015, NULL, 2500, 'inactive', 'Implement generic pentru utilizari diverse')
ON CONFLICT DO NOTHING;

-- Seed: compatibilități între tipuri de mașini și implementuri
INSERT INTO implement_compatibilities (machine_type, implement_type) VALUES
    ('tractor', 'plow'),
    ('tractor', 'seeder'),
    ('tractor', 'fertilizer_spreader'),
    ('tractor', 'sprayer'),
    ('tractor', 'trailer'),
    ('combine', 'header'),
    ('combine', 'trailer'),
    ('drone', 'sprayer')
ON CONFLICT DO NOTHING;

-- Seed: 15 operatori
INSERT INTO operators (name, phone, email, status, allowed_machine_types) VALUES
    ('Alexandru Ionescu', '+37369100001', 'alexandru.ionescu@agri.ro', 'active',   '{tractor,combine}'),
    ('Mihai Popescu',     '+37369100002', 'mihai.popescu@agri.ro',     'active',   '{tractor}'),
    ('Gheorghe Dănilă',   '+37369100003', NULL,                        'active',   '{tractor,sprayer}'),
    ('Ion Constantin',    '+37369100004', NULL,                        'active',   '{tractor,small_truck}'),
    ('Vasile Marin',      '+37369100005', 'vasile.marin@agri.ro',      'active',   '{combine}'),
    ('Dumitru Florescu',  '+37369100006', NULL,                        'active',   '{tractor}'),
    ('Nicolae Stancu',    '+37369100007', 'nicolae.stancu@agri.ro',    'active',   '{tractor,car}'),
    ('Florin Gheorghiu',  '+37369100008', NULL,                        'active',   '{tractor,small_truck}'),
    ('Octavian Rus',      '+37369100009', 'octavian.rus@agri.ro',      'active',   '{drone}'),
    ('Petru Moldovan',    '+37369100010', NULL,                        'active',   '{drone,sprayer}'),
    ('Andrei Popa',       '+37369100011', NULL,                        'inactive', '{tractor}'),
    ('Cristian Luca',     '+37369100012', NULL,                        'inactive', '{car,small_truck}'),
    ('Bogdan Stoica',     '+37369100013', 'bogdan.stoica@agri.ro',     'active',   '{combine,tractor}'),
    ('Radu Nistor',       '+37369100014', NULL,                        'active',   '{tractor}'),
    ('Sorin Enache',      '+37369100015', 'sorin.enache@agri.ro',      'active',   '{tractor,sprayer}')
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

-- Seed: tipuri de resurse agricole
INSERT INTO resource_types (name, category, default_unit) VALUES
    ('Combustibil', 'fuel', 'L'),
    ('Fertilizant', 'fertilizer', 'kg'),
    ('Seminte', 'seed', 'kg'),
    ('Pesticid', 'pesticide', 'L'),
    ('Apa', 'water', 'm3'),
    ('Alte resurse', 'other', 'buc')
ON CONFLICT DO NOTHING;

-- Seed: resurse (consumabile)
INSERT INTO resources (name, resource_type_id, price_per_unit, notes)
VALUES
    ('Motorina flota utilaje', 1, 28.95, 'Rezervor principal pentru tractoare si combine'),
    ('Benzina pentru autoturisme', 1, 31.20, 'Consum pentru vehicule usoare'),
    ('Uree pentru fertilizare faziala', 2, 7.85, 'Aplicare primavara pe grau'),
    ('NPK pentru pregatire teren', 2, 6.60, 'Fertilizare de baza inainte de semanat'),
    ('Samanta grau lot A', 3, 4.75, 'Lot certificat C1'),
    ('Samanta porumb lot B', 3, 690.00, 'Saci 80.000 boabe'),
    ('Erbicid camp est', 4, 43.90, 'Tratament post-recoltare'),
    ('Fungicid lot rapita', 4, 96.50, 'Control boli foliare'),
    ('Apa sistem pivot 1', 5, 1.35, 'Cost operational mediu'),
    ('Material absorbant atelier', 6, 18.00, 'Consumabil mentenanta')
ON CONFLICT DO NOTHING;

-- Seed: stocuri pentru resurse
INSERT INTO stocks (resource_id, quantity, minimum_quantity)
SELECT
    res.id,
    s.quantity,
    s.minimum_quantity
FROM (
    VALUES
        ('Motorina flota utilaje', 12450.0000, 3000.0000),
        ('Benzina pentru autoturisme', 1850.0000, 500.0000),
        ('Uree pentru fertilizare faziala', 9200.0000, 2500.0000),
        ('NPK pentru pregatire teren', 7800.0000, 2000.0000),
        ('Samanta grau lot A', 5600.0000, 1500.0000),
        ('Samanta porumb lot B', 140.0000, 40.0000),
        ('Erbicid camp est', 620.0000, 180.0000),
        ('Fungicid lot rapita', 240.0000, 80.0000),
        ('Apa sistem pivot 1', 48000.0000, 10000.0000),
        ('Material absorbant atelier', 85.0000, 20.0000)
) AS s(resource_name, quantity, minimum_quantity)
JOIN resources res ON res.name = s.resource_name
ON CONFLICT DO NOTHING;
