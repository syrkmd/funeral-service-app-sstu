INSERT INTO funeral_services (title, description, price, category_id, active, sort_order)
SELECT 'Церемония на кладбище',
       'Организация и сопровождение церемонии на месте захоронения',
       120000,
       category.id,
       true,
       6
FROM service_categories category
WHERE category.name = 'Основные услуги'
  AND NOT EXISTS (
    SELECT 1 FROM funeral_services service WHERE service.title = 'Церемония на кладбище'
);

INSERT INTO funeral_services (title, description, price, category_id, active, sort_order)
SELECT 'Прощание',
       'Организация церемонии прощания',
       80000,
       category.id,
       true,
       7
FROM service_categories category
WHERE category.name = 'Дополнительные услуги'
  AND NOT EXISTS (
    SELECT 1 FROM funeral_services service WHERE service.title = 'Прощание'
);
