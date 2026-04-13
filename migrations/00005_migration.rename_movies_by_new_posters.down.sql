BEGIN;

UPDATE movie
SET
    title = CASE picture_file_key
        WHEN 'img/65.jpg' THEN 'Дюна: Часть Вторая'
        WHEN 'img/joker.jpeg' THEN 'Джокер'
        WHEN 'img/dark_knight.jpg' THEN 'Тёмный рыцарь'
        WHEN 'img/garry_potter_sorcerers_stone.jpg' THEN 'Престиж'
        WHEN 'img/jaws.jpg' THEN 'Ford против Ferrari'
        ELSE title
    END,
    picture_file_key = CASE picture_file_key
        WHEN 'img/65.jpg' THEN 'img/1.jpg'
        WHEN 'img/joker.jpeg' THEN 'img/2.jpeg'
        WHEN 'img/dark_knight.jpg' THEN 'img/3.jpg'
        WHEN 'img/garry_potter_sorcerers_stone.jpg' THEN 'img/4.jpg'
        WHEN 'img/jaws.jpg' THEN 'img/5.jpg'
        ELSE picture_file_key
    END
WHERE picture_file_key IN (
    'img/65.jpg',
    'img/joker.jpeg',
    'img/dark_knight.jpg',
    'img/garry_potter_sorcerers_stone.jpg',
    'img/jaws.jpg'
);

COMMIT;
