BEGIN;

UPDATE movie
SET
    title = CASE picture_file_key
        WHEN 'img/1.jpg' THEN '65'
        WHEN 'img/2.jpeg' THEN 'Джокер'
        WHEN 'img/3.jpg' THEN 'Тёмный рыцарь'
        WHEN 'img/4.jpg' THEN 'Гарри Поттер и философский камень'
        WHEN 'img/5.jpg' THEN 'Челюсти'
        ELSE title
    END,
    picture_file_key = CASE picture_file_key
        WHEN 'img/1.jpg' THEN 'img/65.jpg'
        WHEN 'img/2.jpeg' THEN 'img/joker.jpeg'
        WHEN 'img/3.jpg' THEN 'img/dark_knight.jpg'
        WHEN 'img/4.jpg' THEN 'img/garry_potter_sorcerers_stone.jpg'
        WHEN 'img/5.jpg' THEN 'img/jaws.jpg'
        ELSE picture_file_key
    END
WHERE picture_file_key IN ('img/1.jpg', 'img/2.jpeg', 'img/3.jpg', 'img/4.jpg', 'img/5.jpg');

COMMIT;
