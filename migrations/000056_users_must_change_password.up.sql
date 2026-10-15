-- T23: contul de administrator creat la pornire (ADMIN_EMAIL / ADMIN_PASSWORD) primește o parolă
-- provizorie, care trebuie schimbată la prima autentificare. Până atunci API-ul acceptă doar
-- schimbarea parolei, profilul și deconectarea.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE;
