-- +goose Up
WITH users_update AS(
    SELECT id, ROW_NUMBER() OVER (PARTITION BY username ORDER BY id) AS copy_no 
    FROM users
)
UPDATE users
SET username = users.username || '_' || c.id
FROM users_update AS c
WHERE users.id = c.id AND c.copy_no > 1;


-- +goose Down
-- операция необратима, старые имена нигде отдельно не хранятся
-- Объяснение того, что тут происходит: 
-- WITH users_update AS (...) - задаёт именованный временный результат, который сущесвует
-- только внутри одного запроса
-- ROW_NUMBER() OVER (...) это оконная функция: она нумерует строки, не схлопывая их.
-- Этим она отличается от GROUP BY, который превратил бы три bob в одну строку.
-- PARTITION BY username: разбить строки на группы по имени. Нумерация в каждой группе начинается с 1 заново.
-- ORDER BY id: внутри группы нумеровать по возрастанию id. Самый старый получает номер 1.
-- AS copy_no: имя новой колонки с номером копии.