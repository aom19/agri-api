-- Un sezon poate fi „în curs”: data de sfârșit nu se cunoaște la creare și se completează când
-- sezonul se încheie. Sezonul e doar o etichetă de raportare (T24), deci nicio lucrare nu depinde
-- de această dată. CHECK (end_date >= start_date) rămâne valabil: e ignorat când end_date e NULL.
ALTER TABLE seasons ALTER COLUMN end_date DROP NOT NULL;
