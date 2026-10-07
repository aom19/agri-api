-- T11: regula de compatibilitate are un singur nivel, template → tipuri de mașini și echipamente
-- (template_machine_types, template_implement_types), verificat de API la crearea și editarea
-- operațiunii pe teren. Celelalte două niveluri erau doar filtre în dropdown-uri.
DROP TABLE IF EXISTS implement_compatibilities;

ALTER TABLE user_profiles DROP COLUMN IF EXISTS allowed_machine_types;
