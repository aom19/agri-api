#!/usr/bin/env python3
"""
Generează seed-urile pe sezoane: sezon_2023_2024.sql, sezon_2024_2025.sql, sezon_2025_2026.sql.

Fiecare fișier e de sine stătător și se rulează DUPĂ seeds/seed.sql (are nevoie de terenuri, mașini,
echipamente, operatori, resurse și șabloane). Conține:
  - preludiu idempotent: resursele, șabloanele și resursele de recoltă care lipsesc;
  - curățenia propriului sezon (poate fi rulat de mai multe ori);
  - sezonul, culturile pe terenuri, lucrările (de la arat la recoltat și balotat) și toate
    mișcările de stoc (aprovizionări, consumuri la finalizare, recolte, paie), cu invariantul
    „stocul = suma mișcărilor” păstrat de funcția pg_temp.seed_move.

Determinist: același fișier la fiecare rulare. Rulează: python3 seeds/seasons/generate.py
"""
import math
import random
from datetime import date, datetime, timedelta
from pathlib import Path

OUT_DIR = Path(__file__).resolve().parent
TODAY = date(2026, 10, 15)  # „azi” pentru sezonul în curs; fix, ca fișierele să fie deterministe

# ─── Sezoane ────────────────────────────────────────────────────────────────
SEASONS = [
    # (an de început, nume, sfârșit, activ, factor de an pentru randamente, notă)
    (2023, "Sezon 2023/2024", "2024-10-31", False, 0.88, "An secetos: randamente sub așteptări la porumb și floarea-soarelui."),
    (2024, "Sezon 2024/2025", "2025-10-31", False, 1.06, "An bun, cu precipitații normale."),
    (2025, "Sezon 2025/2026", None, True, 0.97, "În curs. Grâul de pe Tartaul - Tarla Mica a fost compromis de îngheț și reînsămânțat cu porumb."),
]

# ─── Terenuri și rotația culturilor (2023/24 → 2024/25 → 2025/26) ───────────
# „Vie” nu primește culturi de câmp; Cirpesti - Lot Sud rămâne liber în 2025/26.
ROTATION = {
    "Plopi - Tarla Mare":          ("Grâu", "Floarea-soarelui", "Porumb"),
    "Gotesti - Tarla Mare":        ("Porumb", "Grâu", "Rapiță"),
    "Gotesti - Tarla Est":         ("Rapiță", "Grâu", "Floarea-soarelui"),
    "Tartaul - Tarla Mare":        ("Floarea-soarelui", "Porumb", "Grâu"),
    "Gotesti - Tarla Vest":        ("Lucernă", "Lucernă", "Lucernă"),
    "Tartaul - Tarla Centrala":    ("Orz", "Soia", "Grâu"),
    "Plopi - Tarla Nord":          ("Grâu", "Porumb", "Sfeclă de zahăr"),
    "Plopi - Tarla Sud":           ("Soia", "Grâu", "Porumb"),
    "Tartaul - Tarla Mica":        ("Grâu", "Rapiță", "Grâu"),   # 2025/26: compromis, apoi porumb
    "Lingura - Teren agricol Nord": ("Porumb", "Grâu", "Orz"),
    "Cirpesti - Lot Nord":         ("Grâu", "Porumb", "Grâu"),
    "Cirpesti - Lot Central":      ("Porumb", "Grâu", "Soia"),
    "Cirpesti - Lot Sud":          ("Floarea-soarelui", "Grâu", None),
}
AREAS = {
    "Plopi - Tarla Mare": 210.25, "Gotesti - Tarla Mare": 130.03, "Gotesti - Tarla Est": 92.34,
    "Tartaul - Tarla Mare": 84.02, "Gotesti - Tarla Vest": 68.47, "Tartaul - Tarla Centrala": 53.5,
    "Plopi - Tarla Nord": 52.41, "Plopi - Tarla Sud": 48.28, "Tartaul - Tarla Mica": 9.84,
    "Lingura - Teren agricol Nord": 7.98, "Cirpesti - Lot Nord": 2.38, "Cirpesti - Lot Central": 1.91,
    "Cirpesti - Lot Sud": 1.56,
}
IRRIGATED = {"Plopi - Tarla Sud", "Plopi - Tarla Nord"}  # au pivot

# ─── Resurse noi (idempotent), pe lângă cele din seed.sql ─────────────────────
# (nume, tip, categorie, unitate, preț, stoc inițial, minim)
EXTRA_RESOURCES = [
    ("Samanta rapita hibrid",          "Seminte",  "seed",    "kg", 95.00, 0, 20),
    ("Samanta floarea-soarelui hibrid", "Seminte", "seed",    "kg", 320.00, 0, 15),
    ("Samanta orz",                    "Seminte",  "seed",    "kg", 4.20, 0, 500),
    ("Samanta soia",                   "Seminte",  "seed",    "kg", 6.50, 0, 400),
    ("Samanta sfecla de zahar",        "Seminte",  "seed",    "kg", 900.00, 0, 5),
    ("Samanta lucerna",                "Seminte",  "seed",    "kg", 38.00, 0, 50),
    ("Paie baloti",                    "Recoltă",  "harvest", "t",  180.00, 0, 0),
]
HARVEST_PRICES = {  # lei/t, pentru valoarea stocului de recoltă
    "Grâu": 900, "Porumb": 850, "Orz": 800, "Floarea-soarelui": 1900, "Rapiță": 2100,
    "Soia": 1950, "Sfeclă de zahăr": 280, "Lucernă": 650,
}

# ─── Șabloane (nume → tip, norme pe hectar). Cele din seed.sql sunt refolosite. ──
TEMPLATES = {
    # existente în seed.sql (normele de aici trebuie să fie egale cu cele din bază)
    "Arat adanc standard":        ("soil_preparation", {"Motorina flota utilaje": 18}, None),
    "Pregatire pat germinativ":   ("soil_preparation", {"Motorina flota utilaje": 10, "NPK pentru pregatire teren": 180}, None),
    "Semanat grau standard":      ("seeding",  {"Motorina flota utilaje": 9, "Samanta grau lot A": 220}, "Grâu"),
    "Semanat porumb standard":    ("seeding",  {"Motorina flota utilaje": 8.5, "Samanta porumb lot B": 0.85}, "Porumb"),
    "Fertilizare faziala grau":   ("fertilization", {"Motorina flota utilaje": 6, "Uree pentru fertilizare faziala": 160}, "Grâu"),
    "Tratament fungicid rapita":  ("spraying", {"Motorina flota utilaje": 4.8, "Fungicid lot rapita": 1, "Apa sistem pivot 1": 0.22}, "Rapiță"),
    "Erbicidare camp est":        ("spraying", {"Motorina flota utilaje": 4.5, "Erbicid camp est": 2.2, "Apa sistem pivot 1": 0.25}, None),
    "Recoltare grau":             ("harvesting", {"Motorina flota utilaje": 14}, "Grâu"),
    "Irigare pivot":              ("irrigation", {"Apa sistem pivot 1": 350}, None),
    # noi (create idempotent de fiecare seed)
    "Semanat rapita standard":    ("seeding",  {"Motorina flota utilaje": 7, "Samanta rapita hibrid": 4}, "Rapiță"),
    "Semanat floarea-soarelui standard": ("seeding", {"Motorina flota utilaje": 8, "Samanta floarea-soarelui hibrid": 3.5}, "Floarea-soarelui"),
    "Semanat orz standard":       ("seeding",  {"Motorina flota utilaje": 9, "Samanta orz": 180}, "Orz"),
    "Semanat soia standard":      ("seeding",  {"Motorina flota utilaje": 8.5, "Samanta soia": 100}, "Soia"),
    "Semanat sfecla standard":    ("seeding",  {"Motorina flota utilaje": 9, "Samanta sfecla de zahar": 2.5}, "Sfeclă de zahăr"),
    "Semanat lucerna standard":   ("seeding",  {"Motorina flota utilaje": 8, "Samanta lucerna": 22}, "Lucernă"),
    "Fertilizare lucerna":        ("fertilization", {"Motorina flota utilaje": 6, "NPK pentru pregatire teren": 150}, "Lucernă"),
    "Tratament fungicid grau":    ("spraying", {"Motorina flota utilaje": 4.8, "Fungicid lot rapita": 0.8, "Apa sistem pivot 1": 0.22}, "Grâu"),
    "Recoltare porumb":           ("harvesting", {"Motorina flota utilaje": 16}, "Porumb"),
    "Recoltare rapita":           ("harvesting", {"Motorina flota utilaje": 13}, "Rapiță"),
    "Recoltare floarea-soarelui": ("harvesting", {"Motorina flota utilaje": 12}, "Floarea-soarelui"),
    "Recoltare orz":              ("harvesting", {"Motorina flota utilaje": 13}, "Orz"),
    "Recoltare soia":             ("harvesting", {"Motorina flota utilaje": 12}, "Soia"),
    "Recoltare sfecla":           ("harvesting", {"Motorina flota utilaje": 22}, "Sfeclă de zahăr"),
    "Cosit si balotat lucerna":   ("harvesting", {"Motorina flota utilaje": 8}, "Lucernă"),
    "Balotat paie":               ("harvesting", {"Motorina flota utilaje": 3}, None),
}
EXISTING_TEMPLATES = {"Arat adanc standard", "Pregatire pat germinativ", "Semanat grau standard", "Semanat porumb standard",
                      "Fertilizare faziala grau", "Tratament fungicid rapita", "Erbicidare camp est", "Recoltare grau", "Irigare pivot"}

# ha/oră pe tip de lucrare → durata planificată; irigarea e pe oră fixă
RATES = {"soil_preparation": 1.6, "seeding": 3.2, "fertilization": 7.0, "spraying": 9.0, "harvesting": 2.4}
# echipajul: (tip mașină, tip echipament) pe tip de lucrare; șabloanele speciale suprascriu
CREW = {
    "soil_preparation": ("tractor", "plow"), "seeding": ("tractor", "seeder"), "fertilization": ("tractor", "fertilizer_spreader"),
    "spraying": ("sprayer", None), "harvesting": ("combine", "header"), "irrigation": (None, None),
}
CREW_OVERRIDE = {"Pregatire pat germinativ": ("tractor", "disc_harrow"), "Balotat paie": ("tractor", "trailer"),
                 "Cosit si balotat lucerna": ("tractor", "trailer")}
MACHINES = {  # mașini active din seed.sql, pe tip (rotite pe terenuri)
    "tractor": ["TR-JOHNDEERE-6R185-01", "TR-FENDT-724VARIO-02", "TR-CASEIH-PUMA240-04", "TR-VALTRA-T254-06", "TR-SAME-EXPLORER120-07"],
    "combine": ["CB-CLAAS-LEXION780-01", "CB-JOHNDEERE-S780-02"],
    "sprayer": ["SP-HARDI-ALPHAEVO-01"],
}
IMPLEMENTS = {"plow": "PL-LEMKEN-JUWEL8-01", "seeder": "SE-HORSCH-PRONTO6DC-01", "fertilizer_spreader": "FS-AMAZONE-ZA-TS4200-01",
              "header": "HD-CLAAS-VARIO1080-01", "disc_harrow": "DH-KVERNELAND-QUALIDISC-01", "trailer": "TR-KRAMPE-HALFPIPE-01"}
OPERATORS = 13  # conturi active cu rolul operator, luate prin OFFSET după id

# ─── Planul agronomic pe cultură: (șablon, lună, zi, an relativ: 0 = anul de început) ─────
EXPECTED = {"Grâu": 5.8, "Orz": 5.2, "Rapiță": 3.2, "Porumb": 9.0, "Floarea-soarelui": 2.6, "Soia": 2.8, "Sfeclă de zahăr": 60.0, "Lucernă": 9.5}
STRAW = {"Grâu": 2.4, "Orz": 2.0}  # t/ha paie balotate

def plan(crop, field, lucerna_year):
    y0, y1 = 0, 1
    if crop == "Grâu":
        return [("Arat adanc standard", 8, 25, y0), ("Pregatire pat germinativ", 9, 28, y0), ("Semanat grau standard", 10, 8, y0),
                ("Fertilizare faziala grau", 3, 12, y1), ("Erbicidare camp est", 4, 18, y1), ("Tratament fungicid grau", 5, 10, y1),
                ("Recoltare grau", 7, 5, y1), ("Balotat paie", 7, 12, y1)]
    if crop == "Orz":
        return [("Arat adanc standard", 8, 22, y0), ("Pregatire pat germinativ", 9, 20, y0), ("Semanat orz standard", 9, 30, y0),
                ("Fertilizare faziala grau", 3, 6, y1), ("Erbicidare camp est", 4, 12, y1),
                ("Recoltare orz", 6, 28, y1), ("Balotat paie", 7, 3, y1)]
    if crop == "Rapiță":
        return [("Arat adanc standard", 8, 5, y0), ("Pregatire pat germinativ", 8, 20, y0), ("Semanat rapita standard", 8, 28, y0),
                ("Fertilizare faziala grau", 3, 5, y1), ("Tratament fungicid rapita", 4, 15, y1), ("Recoltare rapita", 6, 28, y1)]
    if crop == "Porumb":
        ops = [("Arat adanc standard", 10, 25, y0), ("Pregatire pat germinativ", 4, 10, y1), ("Semanat porumb standard", 4, 22, y1),
               ("Erbicidare camp est", 5, 20, y1)]
        if field in IRRIGATED:
            ops += [("Irigare pivot", 6, 25, y1), ("Irigare pivot", 7, 20, y1)]
        return ops + [("Recoltare porumb", 10, 2, y1)]
    if crop == "Floarea-soarelui":
        return [("Arat adanc standard", 10, 28, y0), ("Pregatire pat germinativ", 4, 5, y1), ("Semanat floarea-soarelui standard", 4, 14, y1),
                ("Erbicidare camp est", 5, 12, y1), ("Recoltare floarea-soarelui", 9, 8, y1)]
    if crop == "Soia":
        return [("Arat adanc standard", 11, 2, y0), ("Pregatire pat germinativ", 4, 18, y1), ("Semanat soia standard", 4, 28, y1),
                ("Erbicidare camp est", 5, 28, y1), ("Recoltare soia", 9, 24, y1)]
    if crop == "Sfeclă de zahăr":
        return [("Arat adanc standard", 10, 30, y0), ("Pregatire pat germinativ", 3, 22, y1), ("Semanat sfecla standard", 3, 30, y1),
                ("Erbicidare camp est", 5, 5, y1), ("Erbicidare camp est", 6, 2, y1), ("Irigare pivot", 7, 10, y1), ("Recoltare sfecla", 10, 28, y1)]
    if crop == "Lucernă":
        if lucerna_year == 1:
            return [("Arat adanc standard", 8, 20, y0), ("Pregatire pat germinativ", 3, 12, y1), ("Semanat lucerna standard", 3, 22, y1),
                    ("Cosit si balotat lucerna", 5, 28, y1), ("Cosit si balotat lucerna", 7, 8, y1), ("Cosit si balotat lucerna", 8, 22, y1), ("Cosit si balotat lucerna", 9, 28, y1)]
        return [("Fertilizare lucerna", 3, 8, y1), ("Cosit si balotat lucerna", 5, 15, y1), ("Cosit si balotat lucerna", 6, 28, y1),
                ("Cosit si balotat lucerna", 8, 10, y1), ("Cosit si balotat lucerna", 9, 20, y1)]
    raise ValueError(crop)

LUCERNA_CUTS = {1: [1.8, 2.6, 2.2, 1.6], 2: [2.8, 3.2, 2.6, 1.9], 3: [2.6, 3.0, 2.4, 1.7]}  # t/ha pe coasă

# ─── Utilitare SQL ───────────────────────────────────────────────────────────
def q(s):
    return "'" + str(s).replace("'", "''") + "'"

def ts(d):
    return q(d.strftime("%Y-%m-%d %H:%M:%S+00"))

def r2(x): return round(x + 1e-9, 2)
def r3(x): return round(x + 1e-9, 3)

class Season:
    def __init__(self, year, name, end, active, factor, note, index):
        self.year, self.name, self.end, self.active, self.factor, self.note = year, name, end, active, factor, note
        self.start = date(year, 8, 1)
        self.rng = random.Random(1000 + year)
        self.tag = f"[{name}]"
        self.lines = []           # SQL în DO block
        self.needs = {}           # resursă → cantitate consumată (pentru aprovizionări)
        self.first_use = {}       # resursă → prima dată de folosire
        self.ops = 0
        self.index = index
        self.op_counter = 0
        self.sales = []           # (resursă, cantitate, dată, notă): ieșiri după recoltă

    def emit(self, s=""):
        self.lines.append(s)

    def need(self, resource, qty, when):
        self.needs[resource] = self.needs.get(resource, 0) + qty
        if resource not in self.first_use or when < self.first_use[resource]:
            self.first_use[resource] = when

    # ── o lucrare completă, cu consum ──
    def op(self, field, crop_var, template, start: datetime, area, status="completed", note="", crew_offset=0,
           late=None, irrigate_hours=20, cancel_reason=None, in_progress=False, overdue=False, no_crop=False):
        tcode, norms, _ = TEMPLATES[template]
        mtype, itype = CREW_OVERRIDE.get(template, CREW[tcode])
        machine = MACHINES[mtype][(crew_offset + self.op_counter) % len(MACHINES[mtype])] if mtype else None
        implement = IMPLEMENTS[itype] if itype else None
        operator_offset = (crew_offset + self.op_counter) % OPERATORS
        self.op_counter += 1
        hours = irrigate_hours if tcode == "irrigation" else max(1.5, area / RATES[tcode])
        planned_end = start + timedelta(hours=hours)
        rng = self.rng
        actual_start = actual_end = None
        machine_hours = area_done = None
        completion_notes = ""
        if status == "completed":
            actual_start = start + timedelta(minutes=rng.randint(-20, 60))
            if late is None:
                late = rng.random() < 0.18
            shift = rng.uniform(1, 6) if late else rng.uniform(-1.5, 0.5)
            actual_end = planned_end + timedelta(hours=shift)
            if actual_end <= actual_start:
                actual_end = actual_start + timedelta(hours=1)
            machine_hours = r2(hours * rng.uniform(0.92, 1.1)) if machine else None
            area_done = r2(area * (1 if rng.random() < 0.8 else rng.uniform(0.95, 1.0)))
            completion_notes = "Întârziere: " + rng.choice(["ploaie", "defecțiune minoră", "teren umed"]) if late else ""
        elif status == "in_progress":
            actual_start = start + timedelta(minutes=rng.randint(0, 45))
        elif status == "canceled":
            completion_notes = cancel_reason or ""
        fc_expr = "NULL" if no_crop else "v_fc"
        notes = note
        if no_crop:
            notes = f"{self.tag} {note}".strip()
        # tipul lucrării vine din șablon (T14): operation_type rămâne NULL când există șablon
        self.emit(f"  INSERT INTO field_operations (field_id, operation_type, operation_template_id, machine_id, implement_id, operator_id,")
        self.emit(f"      planned_start_at, planned_end_at, area_planned_ha, notes, status, field_crop_id,")
        self.emit(f"      actual_start_at, actual_end_at, area_completed_ha, machine_hours, completion_notes,")
        self.emit(f"      overdue_notified_at, created_at, updated_at)")
        self.emit(f"  VALUES (v_field, NULL, (SELECT id FROM operation_templates WHERE name = {q(template)} AND deleted_at IS NULL),")
        self.emit(f"      {('(SELECT id FROM machines WHERE name = ' + q(machine) + ')') if machine else 'NULL'}, {('(SELECT id FROM implements WHERE name = ' + q(implement) + ')') if implement else 'NULL'},")
        self.emit(f"      (SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'operator' AND u.deleted_at IS NULL ORDER BY u.id OFFSET {operator_offset} LIMIT 1),")
        self.emit(f"      {ts(start)}, {ts(planned_end)}, {r2(area)}, {q(notes)}, {q(status)}, {fc_expr},")
        self.emit(f"      {ts(actual_start) if actual_start else 'NULL'}, {ts(actual_end) if actual_end else 'NULL'}, {area_done if area_done is not None else 'NULL'}, {machine_hours if machine_hours is not None else 'NULL'}, {q(completion_notes)},")
        self.emit(f"      {ts(planned_end + timedelta(minutes=1)) if overdue else 'NULL'}, {ts(start - timedelta(days=7))}, {ts(actual_end or start)})")
        self.emit(f"  RETURNING id INTO v_op;")
        self.ops += 1
        # consumul din stoc, la finalizare
        if status == "completed":
            when = actual_end
            for resource, norm in norms.items():
                factor = rng.uniform(1.0, 1.12) if resource.startswith("Motorina") else (1.0 if tcode == "irrigation" else rng.uniform(0.97, 1.04))
                qty = r3(norm * area_done * factor)
                if qty <= 0:
                    continue
                label = "Combustibil raportat la finalizare" if resource.startswith("Motorina") else "Consum la finalizarea operațiunii"
                self.emit(f"  PERFORM pg_temp.seed_move({q(resource)}, 'out', {qty}, {ts(when)}, v_op, NULL, {q(label)});")
                self.need(resource, qty, when)
        return actual_end or planned_end, area_done or area

    def sell(self, resource, qty, when, label):
        """Vânzarea unei părți din recoltă: ieșire din stoc, fără lucrare, etichetată pentru curățenie."""
        when = when + timedelta(days=self.rng.randint(30, 60), hours=3)
        if when.date() <= TODAY and qty > 0:
            self.sales.append((resource, r2(qty), when, f"{self.tag} {label}"))

    def harvest_in(self, crop, field, qty, when, cumulative, first):
        notes = f"Recoltă {crop} - {field} ({self.name})"
        self.emit(f"  PERFORM pg_temp.seed_move({q('Recoltă ' + crop)}, 'in', {qty}, {ts(when)}, NULL, v_fc, {q(notes)});")
        self.sell("Recoltă " + crop, qty * (0.8 if crop == "Lucernă" else 0.85), when, f"Vânzare {crop} — contract, {field}")
        self.emit(f"  UPDATE field_crops SET production_total = {cumulative}, harvest_recorded_quantity = {cumulative}, harvest_recorded_at = {ts(when)}, harvested_at = {q(when.date().isoformat())}, updated_at = NOW() WHERE id = v_fc;")

    def straw_in(self, crop, field, qty, when):
        notes = f"Paie {crop} - {field}, balotat"
        self.emit(f"  PERFORM pg_temp.seed_move('Paie baloti', 'in', {qty}, {ts(when)}, NULL, v_fc, {q(notes)});")
        self.sell("Paie baloti", qty * 0.6, when, f"Vânzare paie — {field}")


def field_block(season: Season, field, crop, lucerna_year, field_index):
    """Emite DO-block-ul pentru un teren: cultura pe teren + lucrările + mișcările."""
    rng = season.rng
    area = AREAS[field]
    y = season.year
    current = season.end is None
    exp = EXPECTED[crop] * (1.15 if (crop == "Porumb" and field in IRRIGATED) else 1.0)
    jitter = rng.uniform(0.94, 1.06)
    season.emit(f"-- ── {field} ({area} ha): {crop}" + (f", anul {lucerna_year}" if crop == "Lucernă" else "") + " ──")
    season.emit("DO $seed$")
    season.emit("DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;")
    season.emit("BEGIN")
    season.emit(f"  SELECT id INTO v_season FROM seasons WHERE name = {q(season.name)};")
    season.emit(f"  SELECT id INTO v_field FROM fields WHERE name = {q(field)} AND deleted_at IS NULL;")
    season.emit(f"  IF v_field IS NULL THEN RAISE NOTICE 'terenul {field} lipsește; sărit'; RETURN; END IF;")

    ops = plan(crop, field, lucerna_year)
    sow = next((o for o in ops if o[0].startswith("Semanat")), None)
    planted = date(y + sow[3], sow[1], sow[2]) if sow else date(y + 1, 3, 1)
    if crop == "Lucernă" and lucerna_year > 1:
        planted = None  # semănată în anul 1
    # starea și datele de închidere se completează după ce știm recolta
    # Sezoanele încheiate se inserează direct „închise” (closed_at provizoriu, corectat după recoltă):
    # indexul unic acceptă o singură cultură deschisă pe teren, iar fișierele pot fi rulate în orice ordine.
    initial_status = "open" if current else "closed"
    initial_closed = "NULL" if current else q(date(y + 1, 7, 15).isoformat())
    season.emit(f"  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, closed_at, created_at, updated_at)")
    season.emit(f"  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = {q(crop)}), {area}, {q(planted.isoformat()) if planted else 'NULL'}, {r2(exp)}, {q('Lucernă anul ' + str(lucerna_year) if crop == 'Lucernă' else '')}, {q(initial_status)}, {initial_closed}, {ts(datetime(y, 8, 1, 8))}, NOW())")
    season.emit(f"  RETURNING id INTO v_fc;")

    cumulative = 0.0
    cut_index = 0
    last_end = None
    harvest_done = False
    for (template, month, day, rel) in ops:
        start = datetime(y + rel, month, day, 7, 0) + timedelta(days=rng.randint(-2, 3))
        if current and start.date() > TODAY:
            # lucrare în viitor: planificată
            season.op(field, "v_fc", template, start, area, status="planned", crew_offset=field_index)
            continue
        # un fungicid pe sezon e anulat (ploaie), ca să existe și lucrări anulate
        if template.startswith("Tratament fungicid") and field_index % 5 == 2:
            season.op(field, "v_fc", template, start, area, status="canceled", crew_offset=field_index,
                      cancel_reason="Anulată: ploi persistente în fereastra de tratament")
            continue
        end, done = season.op(field, "v_fc", template, start, area, crew_offset=field_index)
        last_end = end
        if template.startswith("Recoltare"):
            production = r2(done * exp * season.factor * jitter)
            cumulative = production
            season.harvest_in(crop, field, production, end + timedelta(hours=20), cumulative, True)
            harvest_done = True
        elif template == "Cosit si balotat lucerna":
            per_ha = LUCERNA_CUTS[lucerna_year][cut_index] * season.factor * jitter
            cut_index += 1
            qty = r2(done * per_ha)
            cumulative = r2(cumulative + qty)
            season.harvest_in(crop, field, qty, end + timedelta(hours=30), cumulative, cut_index == 1)
            harvest_done = True
        elif template == "Balotat paie":
            qty = r2(done * STRAW[crop] * rng.uniform(0.9, 1.1))
            season.straw_in(crop, field, qty, end + timedelta(hours=6))
    # închiderea culturii
    if not current or (harvest_done and crop != "Lucernă" and crop != "Sfeclă de zahăr"):
        close_at = (last_end + timedelta(days=5)).date() if last_end else date(y + 1, 7, 15)
        season.emit(f"  UPDATE field_crops SET status = 'closed', closed_at = {q(close_at.isoformat())}, harvested_at = COALESCE(harvested_at, {q(close_at.isoformat())}), updated_at = NOW() WHERE id = v_fc;")
    season.emit("END $seed$;")
    season.emit("")


def compromised_then_corn(season: Season, field, field_index):
    """2025/26, Tartaul - Tarla Mica: grâu compromis de îngheț în februarie, apoi porumb în același sezon."""
    rng = season.rng
    area = AREAS[field]
    y = season.year
    season.emit(f"-- ── {field} ({area} ha): Grâu compromis (îngheț), apoi Porumb în același sezon ──")
    season.emit("DO $seed$")
    season.emit("DECLARE v_season bigint; v_field uuid; v_fc bigint; v_op bigint;")
    season.emit("BEGIN")
    season.emit(f"  SELECT id INTO v_season FROM seasons WHERE name = {q(season.name)};")
    season.emit(f"  SELECT id INTO v_field FROM fields WHERE name = {q(field)} AND deleted_at IS NULL;")
    season.emit(f"  IF v_field IS NULL THEN RETURN; END IF;")
    season.emit(f"  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, production_total, notes, status, closed_at, created_at, updated_at)")
    season.emit(f"  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Grâu'), {area}, '2025-10-09', 5.8, 0, 'Compromisă: îngheț 12–18.02.2026, plante distruse pe 90% din suprafață. Reînsămânțat cu porumb.', 'compromised', '2026-02-28', {ts(datetime(y, 8, 1, 8))}, NOW())")
    season.emit(f"  RETURNING id INTO v_fc;")
    for template, month, day, rel in [("Arat adanc standard", 8, 24, 0), ("Pregatire pat germinativ", 9, 29, 0), ("Semanat grau standard", 10, 9, 0)]:
        season.op(field, "v_fc", template, datetime(y + rel, month, day, 7), area, crew_offset=field_index)
    season.emit("  -- porumbul, pe același teren, în același sezon")
    season.emit(f"  INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, expected_yield_per_ha, notes, status, created_at, updated_at)")
    season.emit(f"  VALUES (v_field, v_season, (SELECT id FROM crops WHERE name = 'Porumb'), {area}, '2026-04-20', 8.5, 'Reînsămânțare după grâul compromis.', 'open', {ts(datetime(y + 1, 3, 1, 8))}, NOW())")
    season.emit(f"  RETURNING id INTO v_fc;")
    cumulative = 0
    for template, month, day in [("Pregatire pat germinativ", 4, 8), ("Semanat porumb standard", 4, 20), ("Erbicidare camp est", 5, 18), ("Recoltare porumb", 10, 5)]:
        end, done = season.op(field, "v_fc", template, datetime(y + 1, month, day, 7), area, crew_offset=field_index)
        if template.startswith("Recoltare"):
            production = r2(done * 8.5 * season.factor * rng.uniform(0.95, 1.03))
            season.harvest_in("Porumb", field, production, end + timedelta(hours=20), production, True)
            season.emit(f"  UPDATE field_crops SET status = 'closed', closed_at = {q((end + timedelta(days=4)).date().isoformat())}, updated_at = NOW() WHERE id = v_fc;")
    season.emit("END $seed$;")
    season.emit("")


def autumn_without_crop(season: Season):
    """Toamna 2026: lucrări pentru culturile sezonului următor, încă nedeschise → fără cultură (field_crop_id NULL).
    Dau în dashboard: una în lucru, una întârziată (în lucru, cu sfârșitul planificat depășit), una planificată."""
    season.emit("-- ── Toamna 2026: lucrări fără cultură (sezonul următor nu e deschis încă) ──")
    season.emit("DO $seed$")
    season.emit("DECLARE v_field uuid; v_fc bigint; v_op bigint;")
    season.emit("BEGIN")
    for field, template, start, status, overdue, note in [
        ("Gotesti - Tarla Est", "Arat adanc standard", datetime(2026, 10, 14, 7), "in_progress", False, "Arat de toamnă pentru grâu 2026/2027"),
        ("Tartaul - Tarla Mare", "Pregatire pat germinativ", datetime(2026, 10, 8, 7), "in_progress", True, "Pregătire pentru rapiță; a depășit termenul"),
        ("Plopi - Tarla Sud", "Arat adanc standard", datetime(2026, 10, 21, 7), "planned", False, "Arat de toamnă"),
        ("Cirpesti - Lot Sud", "Arat adanc standard", datetime(2026, 10, 27, 7), "planned", False, "Arat de toamnă pe lotul liber"),
    ]:
        season.emit(f"  SELECT id INTO v_field FROM fields WHERE name = {q(field)} AND deleted_at IS NULL;")
        season.emit(f"  IF v_field IS NOT NULL THEN")
        season.op(field, None, template, start, AREAS[field], status=status, note=note, overdue=overdue, no_crop=True, crew_offset=3)
        season.emit(f"  END IF;")
    season.emit("END $seed$;")
    season.emit("")


def prelude(season: Season):
    out = []
    out.append(f"-- Seed sezon {season.name} — generat de seeds/seasons/generate.py; nu edita manual, modifică generatorul.")
    out.append("-- Rulează după seeds/seed.sql. Poate fi rulat de mai multe ori: își șterge întâi propriile date.")
    out.append("BEGIN;")
    out.append("")
    out.append("-- ─── Funcție temporară: o mișcare de stoc care păstrează invariantul stoc = Σ mișcări ───")
    out.append("""CREATE OR REPLACE FUNCTION pg_temp.seed_move(p_resource text, p_type text, p_qty numeric, p_at timestamptz, p_op bigint, p_fc bigint, p_notes text)
RETURNS void LANGUAGE plpgsql AS $f$
DECLARE v_res bigint; v_price numeric; v_qty numeric; v_delta numeric; v_new numeric;
BEGIN
    SELECT r.id, r.price_per_unit, r.quantity INTO v_res, v_price, v_qty
    FROM resources r WHERE r.name = p_resource FOR UPDATE;
    IF v_res IS NULL THEN RAISE EXCEPTION 'resursa % nu există', p_resource; END IF;
    v_delta := CASE WHEN p_type = 'in' THEN ROUND(p_qty, 4) ELSE -ROUND(p_qty, 4) END;
    v_new := v_qty + v_delta;
    IF v_new < 0 THEN RAISE EXCEPTION 'stoc negativ pentru % (% %)', p_resource, v_qty, v_delta; END IF;
    INSERT INTO stock_movements (resource_id, field_operation_id, movement_type, quantity_delta, resulting_quantity, unit_cost, total_cost, notes, field_crop_id, created_at)
    VALUES (v_res, p_op, p_type, v_delta, v_new, v_price, ABS(v_delta) * v_price, p_notes, p_fc, p_at);
    UPDATE resources SET quantity = v_new, updated_at = NOW() WHERE id = v_res;
END $f$;""")
    out.append("")
    out.append(f"-- ─── Curățenie: datele acestui sezon, dacă a mai fost rulat ───")
    out.append(f"""DO $clean$
DECLARE v_season bigint;
BEGIN
    SELECT id INTO v_season FROM seasons WHERE name = {q(season.name)};
    IF v_season IS NOT NULL THEN
        DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season));
        DELETE FROM stock_movements WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_operations WHERE field_crop_id IN (SELECT id FROM field_crops WHERE season_id = v_season);
        DELETE FROM field_crops WHERE season_id = v_season;
        DELETE FROM seasons WHERE id = v_season;
    END IF;
    -- aprovizionările și lucrările fără cultură ale sezonului sunt etichetate în notes
    DELETE FROM stock_movements WHERE field_operation_id IN (SELECT id FROM field_operations WHERE notes LIKE {q(season.tag + '%')});
    DELETE FROM field_operations WHERE notes LIKE {q(season.tag + '%')};
    DELETE FROM stock_movements WHERE notes LIKE {q(season.tag + '%')};
    UPDATE resources r SET quantity = COALESCE((SELECT SUM(sm.quantity_delta) FROM stock_movements sm WHERE sm.resource_id = r.id), 0);
END $clean$;""")
    out.append("")
    out.append("-- ─── Resurse, resurse de recoltă și șabloane care lipsesc (idempotent) ───")
    for name, tname, cat, unit, price, qty, minimum in EXTRA_RESOURCES:
        out.append(f"INSERT INTO resource_types (name, category, default_unit) SELECT {q(tname)}, {q(cat)}, {q(unit)} WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = {q(cat)} AND default_unit = {q(unit)});")
        out.append(f"INSERT INTO resources (name, resource_type_id, price_per_unit, minimum_quantity, notes) SELECT {q(name)}, (SELECT id FROM resource_types WHERE category = {q(cat)} AND default_unit = {q(unit)} ORDER BY id LIMIT 1), {price}, {minimum}, 'Creată de seed-ul pe sezoane' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = {q(name)});")
    out.append("INSERT INTO resource_types (name, category, default_unit) SELECT 'Recoltă', 'harvest', 't' WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE category = 'harvest' AND default_unit = 't');")
    for crop, price in HARVEST_PRICES.items():
        rn = "Recoltă " + crop
        out.append(f"INSERT INTO resources (name, resource_type_id, price_per_unit, notes) SELECT {q(rn)}, (SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = 't' ORDER BY id LIMIT 1), {price}, 'Resursa de recoltă a culturii (seed pe sezoane)' WHERE NOT EXISTS (SELECT 1 FROM resources WHERE name = {q(rn)});")
        out.append(f"UPDATE resources SET price_per_unit = {price} WHERE name = {q(rn)} AND price_per_unit = 0;")
        out.append(f"UPDATE crops SET harvest_resource_id = (SELECT id FROM resources WHERE name = {q(rn)}) WHERE name = {q(crop)} AND harvest_resource_id IS NULL;")
    for tname, (tcode, norms, crop) in TEMPLATES.items():
        if tname in EXISTING_TEMPLATES:
            continue
        crop_expr = f"(SELECT id FROM crops WHERE name = {q(crop)})" if crop else "NULL"
        out.append(f"INSERT INTO operation_templates (operation_type, name, description, unit, crop_id) SELECT {q(tcode)}, {q(tname)}, 'Șablon creat de seed-ul pe sezoane', 'ha', {crop_expr} WHERE NOT EXISTS (SELECT 1 FROM operation_templates WHERE name = {q(tname)} AND deleted_at IS NULL);")
        for res, norm in norms.items():
            out.append(f"INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) SELECT (SELECT id FROM operation_templates WHERE name = {q(tname)} AND deleted_at IS NULL), (SELECT id FROM resources WHERE name = {q(res)}), {norm} WHERE NOT EXISTS (SELECT 1 FROM template_resources tr WHERE tr.template_id = (SELECT id FROM operation_templates WHERE name = {q(tname)} AND deleted_at IS NULL) AND tr.resource_id = (SELECT id FROM resources WHERE name = {q(res)}));")
    out.append("")
    out.append("-- ─── Sezonul ───")
    end_expr = q(season.end) if season.end else "NULL"
    if season.active:
        out.append("UPDATE seasons SET is_active = FALSE WHERE is_active;")
    out.append(f"INSERT INTO seasons (name, start_date, end_date, is_active, notes) VALUES ({q(season.name)}, {q(season.start.isoformat())}, {end_expr}, {str(season.active).lower()}, {q(season.note)});")
    out.append("")
    return out


def supplies(season: Season):
    """Aprovizionări: suficiente pentru consumul sezonului, înaintea primei folosiri; motorina în două tranșe."""
    out = ["-- ─── Aprovizionări (dimensionate după consumul sezonului, +5%) ───", "DO $sup$ BEGIN"]
    for resource, total in sorted(season.needs.items()):
        first = season.first_use[resource] - timedelta(days=6)
        step = 500 if resource.startswith("Motorina") or resource.startswith("Apa") else (50 if total > 500 else 5)
        qty = math.ceil(total * 1.05 / step) * step
        label = "Alimentare bazin pivot" if resource.startswith("Apa") else "Recepție furnizor"
        if resource.startswith("Motorina"):
            half = math.ceil(qty / 2 / step) * step
            out.append(f"  PERFORM pg_temp.seed_move({q(resource)}, 'in', {half}, {ts(first)}, NULL, NULL, {q(season.tag + ' ' + label + ' — tranșa 1')});")
            out.append(f"  PERFORM pg_temp.seed_move({q(resource)}, 'in', {half}, {ts(first + timedelta(days=150))}, NULL, NULL, {q(season.tag + ' ' + label + ' — tranșa 2')});")
        else:
            out.append(f"  PERFORM pg_temp.seed_move({q(resource)}, 'in', {qty}, {ts(first)}, NULL, NULL, {q(season.tag + ' ' + label)});")
    out.append("END $sup$;")
    out.append("")
    return out


def sales_block(season: Season):
    if not season.sales:
        return []
    out = ["-- ─── Vânzări din stoc, după recoltă (ieșiri fără lucrare) ───", "DO $sale$ BEGIN"]
    for resource, qty, when, notes in sorted(season.sales, key=lambda x: x[2]):
        out.append(f"  PERFORM pg_temp.seed_move({q(resource)}, 'out', {qty}, {ts(when)}, NULL, NULL, {q(notes)});")
    out.append("END $sale$;")
    out.append("")
    return out


def epilogue():
    return [
        "-- ─── Istoricul stocului în ordine cronologică (resulting_quantity ca sold curent la acea dată) ───",
        "UPDATE stock_movements sm SET resulting_quantity = t.run FROM (SELECT id, SUM(quantity_delta) OVER (PARTITION BY resource_id ORDER BY created_at, id) AS run FROM stock_movements) t WHERE t.id = sm.id;",
        "",
        "COMMIT;",
        "",
    ]


def generate(season: Season):
    fields = list(ROTATION.items())
    for i, (field, crops) in enumerate(fields):
        crop = crops[season.index]
        if crop is None:
            season.emit(f"-- ── {field}: fără cultură în acest sezon (teren liber) ──")
            season.emit("")
            continue
        if season.index == 2 and field == "Tartaul - Tarla Mica":
            compromised_then_corn(season, field, i)
            continue
        lucerna_year = season.index + 1 if crop == "Lucernă" else 0
        field_block(season, field, crop, lucerna_year, i)
    if season.end is None:
        autumn_without_crop(season)
    body = season.lines
    # aprovizionările trebuie să fie ÎNAINTE de consumuri (seed_move refuză stoc negativ)
    sql = prelude(season) + supplies(season) + body + sales_block(season) + epilogue()
    path = OUT_DIR / f"sezon_{season.year}_{season.year + 1}.sql"
    path.write_text("\n".join(sql), encoding="utf-8")
    return path, season.ops


if __name__ == "__main__":
    for i, (year, name, end, active, factor, note) in enumerate(SEASONS):
        s = Season(year, name, end, active, factor, note, i)
        path, ops = generate(s)
        print(f"{path.name}: {ops} lucrări, {len(s.needs)} resurse aprovizionate, {len(s.sales)} vânzări")
