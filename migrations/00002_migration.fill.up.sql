BEGIN;

INSERT INTO country (title) VALUES
    ('США'),
    ('Испания'),
    ('Великобритания')
ON CONFLICT (title) DO NOTHING;

INSERT INTO language (title) VALUES
    ('Английский')
ON CONFLICT (title) DO NOTHING;

INSERT INTO genre (title) VALUES
    ('Фантастика'),
    ('Драма'),
    ('Приключения'),
    ('Триллер'),
    ('Боевик'),
    ('Криминал'),
    ('Детектив'),
    ('Биография'),
    ('Спорт')
ON CONFLICT (title) DO NOTHING;

INSERT INTO selection (title) VALUES
    ('Популярные'),
    ('Новинки'),
    ('Топ-10'),
    ('Фильмы'),
    ('Мультфильмы'),
    ('Сериалы'),
    ('Короткометражки'),
    ('Семейное')
ON CONFLICT (title) DO NOTHING;

INSERT INTO movie (
    title, description, director, content_type, release_year,
    duration_seconds, age_limit, original_language_id, country_id, picture_file_key, poster_file_key
) VALUES
    (
        'Интерстеллар',
        'Команда исследователей отправляется через червоточину в поисках нового дома для человечества.',
        'Кристофер Нолан',
        'film',
        2014,
        10140,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'interstellar.webp',
        'interstellar.webp'
    ),
    (
        'Начало',
        'Специалист по проникновению в сны получает задание внедрить идею в сознание цели.',
        'Кристофер Нолан',
        'film',
        2010,
        8880,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'inception.webp',
        'inception.webp'
    ),
    (
        'Матрица',
        'Хакер узнает, что привычный мир является симуляцией, и присоединяется к сопротивлению.',
        'Лана Вачовски, Лилли Вачовски',
        'film',
        1999,
        8160,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'matrix.webp',
        'matrix.webp'
    ),
    (
        'Гладиатор',
        'Римский генерал превращается в гладиатора и идет к мести.',
        'Ридли Скотт',
        'film',
        2000,
        9300,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'gladiator.webp',
        'gladiator.webp'
    ),
    (
        'Отступники',
        'Полицейский под прикрытием и агент мафии пытаются вычислить друг друга.',
        'Мартин Скорсезе',
        'film',
        2006,
        9060,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'the_departed.webp',
        'the_departed.webp'
    ),
    (
        'Одержимость',
        'Молодой барабанщик сталкивается с жестким преподавателем на пути к совершенству.',
        'Дэмьен Шазелл',
        'film',
        2014,
        6420,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'whiplash.webp',
        'whiplash.webp'
    ),
    (
        'Остров проклятых',
        'Маршал расследует исчезновение пациентки в клинике на удаленном острове.',
        'Мартин Скорсезе',
        'film',
        2010,
        8280,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'shutter_island.webp',
        'shutter_island.webp'
    ),
    (
        'Семь',
        'Два детектива преследуют серийного убийцу, который наказывает за смертные грехи.',
        'Дэвид Финчер',
        'film',
        1995,
        7620,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'se7en.webp',
        'se7en.webp'
    ),
    (
        'Пленницы',
        'После похищения двух девочек отец одной из них начинает собственное расследование.',
        'Дени Вильнёв',
        'film',
        2013,
        9180,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'prisoners.webp',
        'prisoners.webp'
    ),
    (
        'Социальная сеть',
        'История создания Facebook и конфликта вокруг стремительного роста компании.',
        'Дэвид Финчер',
        'film',
        2010,
        7200,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'the_social_network.webp',
        'the_social_network.webp'
    ),
    (
        'История игрушек',
        'Игрушки оживают, когда люди не видят, и переживают большое приключение.',
        'Джон Лассетер',
        'film',
        1995,
        4860,
        0,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'toy_story.webp',
        'toy_story.webp'
    ),
    (
        'Король Лев',
        'Молодой лев Симба проходит путь от изгнания к возвращению на трон.',
        'Роджер Аллерс, Роб Минкофф',
        'film',
        1994,
        5340,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'lion_king.webp',
        'lion_king.webp'
    ),
    (
        'ВАЛЛ-И',
        'Робот-уборщик на опустевшей Земле отправляется в космическое путешествие.',
        'Эндрю Стэнтон',
        'film',
        2008,
        5880,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'wall_e.webp',
        'wall_e.webp'
    ),
    (
        'Головоломка',
        'Эмоции в голове девочки пытаются помочь ей пережить большие перемены.',
        'Пит Доктер',
        'film',
        2015,
        5700,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'inside_out.webp',
        'inside_out.webp'
    ),
    (
        'Тайна Коко',
        'Мальчик попадает в мир мертвых и раскрывает семейную тайну.',
        'Ли Анкрич',
        'film',
        2017,
        6300,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'coco.webp',
        'coco.webp'
    ),
    (
        'Шрек',
        'Огр отправляется спасать принцессу и неожиданно находит друзей.',
        'Эндрю Адамсон, Вики Дженсон',
        'film',
        2001,
        5400,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'shrek.webp',
        'shrek.webp'
    ),
    (
        'Как приручить дракона',
        'Подросток-викинг дружится с драконом и меняет взгляд своего племени на мир.',
        'Крис Сандерс, Дин Деблуа',
        'film',
        2010,
        5880,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'how_to_train_your_dragon.webp',
        'how_to_train_your_dragon.webp'
    ),
    (
        'Зверополис',
        'Крольчиха-полицейский и хитрый лис раскрывают крупный заговор.',
        'Байрон Ховард, Рич Мур',
        'film',
        2016,
        6480,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'zootopia.webp',
        'zootopia.webp'
    ),
    (
        'Вверх',
        'Пожилой ворчун и юный следопыт улетают навстречу приключениям.',
        'Пит Доктер',
        'film',
        2009,
        5760,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'up.webp',
        'up.webp'
    ),
    (
        'Рататуй',
        'Крысенок мечтает стать шеф-поваром и находит путь на кухню лучшего ресторана.',
        'Брэд Бёрд',
        'film',
        2007,
        6660,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'ratatouille.webp',
        'ratatouille.webp'
    ),
    (
        'Во все тяжкие',
        'Учитель химии начинает производить метамфетамин, чтобы обеспечить семью.',
        'Винс Гиллиган',
        'series',
        2008,
        2820,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'breaking_bad.webp',
        'breaking_bad.webp'
    ),
    (
        'Очень странные дела',
        'Подростки сталкиваются с потусторонними силами в маленьком городке.',
        'Братья Даффер',
        'series',
        2016,
        3000,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'stranger_things.webp',
        'stranger_things.webp'
    ),
    (
        'Черное зеркало',
        'Антология о технологиях и их влиянии на общество и человека.',
        'Чарли Брукер',
        'series',
        2011,
        3600,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'black_mirror.webp',
        'black_mirror.webp'
    ),
    (
        'Игра престолов',
        'Дома Вестероса борются за власть на фоне древней угрозы.',
        'Дэвид Бениофф, Д. Б. Уайсс',
        'series',
        2011,
        3300,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'game_of_thrones.webp',
        'game_of_thrones.webp'
    ),
    (
        'Шерлок',
        'Современная версия приключений Шерлока Холмса и доктора Ватсона.',
        'Марк Гэтисс, Стивен Моффат',
        'series',
        2010,
        5400,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'sherlock.webp',
        'sherlock.webp'
    ),
    (
        'Настоящий детектив',
        'Детективы расследуют запутанные преступления, которые меняют их судьбы.',
        'Ник Пиццолатто',
        'series',
        2014,
        3300,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'true_detective.webp',
        'true_detective.webp'
    ),
    (
        'Чернобыль',
        'История аварии на Чернобыльской АЭС и ее последствий.',
        'Крэйг Мэйзин',
        'series',
        2019,
        3960,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'chernobyl.webp',
        'chernobyl.webp'
    ),
    (
        'Мандалорец',
        'Охотник за головами сопровождает загадочного ребенка по опасной галактике.',
        'Джон Фавро',
        'series',
        2019,
        2520,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'mandalorian.webp',
        'mandalorian.webp'
    ),
    (
        'Аркейн',
        'История двух сестер на фоне конфликта между Пилтовером и Зауном.',
        'Кристиан Линке, Алекс Йи',
        'series',
        2021,
        2460,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'arcane.webp',
        'arcane.webp'
    ),
    (
        'Дом дракона',
        'Хроника борьбы дома Таргариенов за Железный трон.',
        'Райан Кондал, Джордж Р. Р. Мартин',
        'series',
        2022,
        3600,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'house_of_the_dragon.webp',
        'house_of_the_dragon.webp'
    ),
    (
        'Долгое прощание',
        'Короткая музыкальная драма о страхе, идентичности и утрате дома.',
        'Анис Чаганти',
        'film',
        2020,
        720,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'the_long_goodbye.webp',
        'the_long_goodbye.webp'
    ),
    (
        'Два далеких незнакомца',
        'Мужчина снова и снова проживает один и тот же трагический день.',
        'Травон Фри, Мартин Десмонд Роу',
        'film',
        2020,
        1920,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'two_distant_strangers.webp',
        'two_distant_strangers.webp'
    ),
    (
        'Заикание',
        'Молодой человек пытается преодолеть страх близости и собственного голоса.',
        'Бенджамин Клири',
        'film',
        2015,
        780,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'stutterer.webp',
        'stutterer.webp'
    ),
    (
        'Тихий ребенок',
        'Сурдологопед помогает глухой девочке установить контакт с миром.',
        'Крис Овертон',
        'film',
        2017,
        1200,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'the_silent_child.webp',
        'the_silent_child.webp'
    ),
    (
        'Комендантский час',
        'Отчаявшийся мужчина получает шанс переосмыслить жизнь за одну ночь.',
        'Шон Кристенсен',
        'film',
        2012,
        1140,
        16,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'curfew.webp',
        'curfew.webp'
    ),
    (
        'Пой',
        'Короткая история о жесткой школьной дисциплине и силе творчества.',
        'Кристофф Дек',
        'film',
        2001,
        1500,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'sing_short.webp',
        'sing_short.webp'
    ),
    (
        'Кожа',
        'Неонацист сталкивается с шансом вырваться из среды ненависти.',
        'Гай Наттив',
        'film',
        2018,
        1200,
        18,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'skin_short.webp',
        'skin_short.webp'
    ),
    (
        'Окно по соседству',
        'Повседневная жизнь семьи меняется после одного взгляда в чужое окно.',
        'Маршалл Карри',
        'film',
        2019,
        1200,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'the_neighbors_window.webp',
        'the_neighbors_window.webp'
    ),
    (
        'Чудесная история Генри Шугара',
        'Игрок и миллионер пытается освоить невероятный дар.',
        'Уэс Андерсон',
        'film',
        2023,
        2340,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'henry_sugar.webp',
        'henry_sugar.webp'
    ),
    (
        'Ирландское прощание',
        'Два брата вынуждены выполнить список желаний покойной матери.',
        'Том Беркли, Росс Уайт',
        'film',
        2022,
        1380,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'an_irish_goodbye.webp',
        'an_irish_goodbye.webp'
    ),
    (
        'Гарри Поттер и философский камень',
        'Мальчик-волшебник впервые попадает в Хогвартс и узнает правду о себе.',
        'Крис Коламбус',
        'film',
        2001,
        9120,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'harry_potter_1.webp',
        'harry_potter_1.webp'
    ),
    (
        'Один дома',
        'Мальчик остается один дома и защищает его от двух грабителей.',
        'Крис Коламбус',
        'film',
        1990,
        6180,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'home_alone.webp',
        'home_alone.webp'
    ),
    (
        'Приключения Паддингтона 2',
        'Медвежонок пытается купить подарок тете и попадает в запутанную историю.',
        'Пол Кинг',
        'film',
        2017,
        6240,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'paddington_2.webp',
        'paddington_2.webp'
    ),
    (
        'Хроники Нарнии: Лев, колдунья и волшебный шкаф',
        'Четверо детей попадают в волшебный мир и вступают в битву за Нарнию.',
        'Эндрю Адамсон',
        'film',
        2005,
        8580,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'narnia.webp',
        'narnia.webp'
    ),
    (
        'Малефисента',
        'История сказочной волшебницы, чья судьба оказалась сложнее легенд.',
        'Роберт Стромберг',
        'film',
        2014,
        5820,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'maleficent.webp',
        'maleficent.webp'
    ),
    (
        'Мэри Поппинс возвращается',
        'Няня-волшебница снова приходит на помощь семье Бэнкс.',
        'Роб Маршалл',
        'film',
        2018,
        7800,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'Великобритания'),
        'mary_poppins_returns.webp',
        'mary_poppins_returns.webp'
    ),
    (
        'Джуманджи: Зов джунглей',
        'Подростки попадают в игру и оказываются в телах своих игровых аватаров.',
        'Джейк Кэздан',
        'film',
        2017,
        7140,
        12,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'jumanji_welcome_to_the_jungle.webp',
        'jumanji_welcome_to_the_jungle.webp'
    ),
    (
        'Душа',
        'Музыкант оказывается между мирами и заново учится ценить жизнь.',
        'Пит Доктер',
        'film',
        2020,
        6000,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'soul.webp',
        'soul.webp'
    ),
    (
        'Лука',
        'Два морских чудовища проводят лето в итальянском городке под видом мальчиков.',
        'Энрико Касароса',
        'film',
        2021,
        5700,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'luca.webp',
        'luca.webp'
    ),
    (
        'Энканто',
        'Девочка из волшебной семьи пытается спасти чудо, на котором держится ее дом.',
        'Джаред Буш, Байрон Ховард',
        'film',
        2021,
        6120,
        6,
        (SELECT id FROM language WHERE title = 'Английский'),
        (SELECT id FROM country WHERE title = 'США'),
        'encanto.webp',
        'encanto.webp'
    )
ON CONFLICT (picture_file_key) DO NOTHING;

WITH genre_map(movie_title, genre_title) AS (
    VALUES
        ('Интерстеллар', 'Фантастика'),
        ('Интерстеллар', 'Драма'),
        ('Интерстеллар', 'Приключения'),
        ('Начало', 'Фантастика'),
        ('Начало', 'Триллер'),
        ('Начало', 'Боевик'),
        ('Матрица', 'Фантастика'),
        ('Матрица', 'Боевик'),
        ('Гладиатор', 'Драма'),
        ('Гладиатор', 'Боевик'),
        ('Гладиатор', 'Приключения'),
        ('Отступники', 'Драма'),
        ('Отступники', 'Криминал'),
        ('Отступники', 'Триллер'),
        ('Одержимость', 'Драма'),
        ('Остров проклятых', 'Триллер'),
        ('Остров проклятых', 'Детектив'),
        ('Остров проклятых', 'Драма'),
        ('Семь', 'Триллер'),
        ('Семь', 'Детектив'),
        ('Семь', 'Криминал'),
        ('Пленницы', 'Триллер'),
        ('Пленницы', 'Детектив'),
        ('Пленницы', 'Драма'),
        ('Социальная сеть', 'Драма'),
        ('Социальная сеть', 'Биография'),
        ('История игрушек', 'Приключения'),
        ('Король Лев', 'Приключения'),
        ('ВАЛЛ-И', 'Фантастика'),
        ('Головоломка', 'Драма'),
        ('Тайна Коко', 'Приключения'),
        ('Шрек', 'Приключения'),
        ('Как приручить дракона', 'Приключения'),
        ('Зверополис', 'Детектив'),
        ('Вверх', 'Приключения'),
        ('Рататуй', 'Драма'),
        ('Во все тяжкие', 'Драма'),
        ('Во все тяжкие', 'Криминал'),
        ('Очень странные дела', 'Фантастика'),
        ('Очень странные дела', 'Драма'),
        ('Черное зеркало', 'Фантастика'),
        ('Черное зеркало', 'Триллер'),
        ('Игра престолов', 'Драма'),
        ('Игра престолов', 'Приключения'),
        ('Шерлок', 'Детектив'),
        ('Шерлок', 'Драма'),
        ('Настоящий детектив', 'Детектив'),
        ('Настоящий детектив', 'Драма'),
        ('Чернобыль', 'Драма'),
        ('Мандалорец', 'Фантастика'),
        ('Мандалорец', 'Приключения'),
        ('Аркейн', 'Фантастика'),
        ('Аркейн', 'Драма'),
        ('Дом дракона', 'Драма'),
        ('Дом дракона', 'Приключения'),
        ('Долгое прощание', 'Драма'),
        ('Два далеких незнакомца', 'Драма'),
        ('Заикание', 'Драма'),
        ('Тихий ребенок', 'Драма'),
        ('Комендантский час', 'Драма'),
        ('Пой', 'Драма'),
        ('Кожа', 'Драма'),
        ('Окно по соседству', 'Драма'),
        ('Чудесная история Генри Шугара', 'Приключения'),
        ('Ирландское прощание', 'Драма'),
        ('Гарри Поттер и философский камень', 'Приключения'),
        ('Один дома', 'Приключения'),
        ('Приключения Паддингтона 2', 'Приключения'),
        ('Хроники Нарнии: Лев, колдунья и волшебный шкаф', 'Приключения'),
        ('Малефисента', 'Приключения'),
        ('Мэри Поппинс возвращается', 'Приключения'),
        ('Джуманджи: Зов джунглей', 'Приключения'),
        ('Душа', 'Драма'),
        ('Лука', 'Приключения'),
        ('Энканто', 'Приключения')
)
INSERT INTO genre_to_movie (genre_id, movie_id)
SELECT g.id, m.id
FROM genre_map gm
JOIN movie m ON m.title = gm.movie_title
JOIN genre g ON g.title = gm.genre_title
ON CONFLICT DO NOTHING;

WITH actor_seed(full_name, birthdate, biography, country_title, picture_file_key) AS (
    VALUES
        (
            'Мэттью Макконахи',
            '1969-11-04'::date,
            'Американский актёр драматических и приключенческих фильмов.',
            'США',
            'matthew_mcconaughey.webp'
        ),
        (
            'Энн Хэтэуэй',
            '1982-11-12'::date,
            'Американская актриса, лауреат премии «Оскар».',
            'США',
            'anne_hathaway.webp'
        ),
        (
            'Леонардо ДиКаприо',
            '1974-11-11'::date,
            'Американский актёр драматического кино.',
            'США',
            'leonardo_dicaprio.webp'
        ),
        (
            'Джозеф Гордон-Левитт',
            '1981-02-17'::date,
            'Американский актёр.',
            'США',
            'joseph_gordon_levitt.webp'
        ),
        (
            'Киану Ривз',
            '1964-09-02'::date,
            'Американский и канадский актёр.',
            'США',
            'keanu_reeves.webp'
        ),
        (
            'Кэрри-Энн Мосс',
            '1967-08-21'::date,
            'Канадская актриса.',
            'США',
            'carrie_anne_moss.webp'
        ),
        (
            'Рассел Кроу',
            '1964-04-07'::date,
            'Актёр и продюсер.',
            'Великобритания',
            'russell_crowe.webp'
        ),
        (
            'Хоакин Феникс',
            '1974-10-28'::date,
            'Американский актёр.',
            'США',
            'phoenix.webp'
        ),
        (
            'Мэтт Деймон',
            '1970-10-08'::date,
            'Американский актёр и сценарист.',
            'США',
            'matt_damon.webp'
        ),
        (
            'Джек Николсон',
            '1937-04-22'::date,
            'Американский актёр.',
            'США',
            'jack_nicholson.webp'
        ),
        (
            'Майлз Теллер',
            '1987-02-20'::date,
            'Американский актёр.',
            'США',
            'miles_teller.webp'
        ),
        (
            'Дж. К. Симмонс',
            '1955-01-09'::date,
            'Американский актёр.',
            'США',
            'jk_simmons.webp'
        ),
        (
            'Марк Руффало',
            '1967-11-22'::date,
            'Американский актёр.',
            'США',
            'mark_ruffalo.webp'
        ),
        (
            'Брэд Питт',
            '1963-12-18'::date,
            'Американский актёр и продюсер.',
            'США',
            'brad_pitt.webp'
        ),
        (
            'Морган Фримен',
            '1937-06-01'::date,
            'Американский актёр.',
            'США',
            'morgan_freeman.webp'
        ),
        (
            'Хью Джекман',
            '1968-10-12'::date,
            'Актёр театра и кино.',
            'Великобритания',
            'hugh_jackman.webp'
        ),
        (
            'Джейк Джилленхол',
            '1980-12-19'::date,
            'Американский актёр.',
            'США',
            'jake_gyllenhaal.webp'
        ),
        (
            'Джесси Айзенберг',
            '1983-10-05'::date,
            'Американский актёр.',
            'США',
            'jesse_eisenberg.webp'
        ),
        (
            'Эндрю Гарфилд',
            '1983-08-20'::date,
            'Британский и американский актёр.',
            'Великобритания',
            'andrew_garfield.webp'
        ),
        (
            'Том Хэнкс',
            '1956-07-09'::date,
            'Американский актёр.',
            'США',
            'tom_hanks.webp'
        ),
        (
            'Тим Аллен',
            '1953-06-13'::date,
            'Американский актёр.',
            'США',
            'tim_allen.webp'
        ),
        (
            'Мэттью Бродерик',
            '1962-03-21'::date,
            'Американский актёр.',
            'США',
            'matthew_broderick.webp'
        ),
        (
            'Джеймс Эрл Джонс',
            '1931-01-17'::date,
            'Американский актёр.',
            'США',
            'james_earl_jones.webp'
        ),
        (
            'Бен Бертт',
            '1948-07-12'::date,
            'Американский актёр озвучания и звукорежиссёр.',
            'США',
            'ben_burtt.webp'
        ),
        (
            'Элиза Найт',
            '1983-02-15'::date,
            'Американская актриса и сценаристка.',
            'США',
            'elissa_knight.webp'
        ),
        (
            'Эми Полер',
            '1971-09-16'::date,
            'Американская актриса.',
            'США',
            'amy_poehler.webp'
        ),
        (
            'Филлис Смит',
            '1949-07-10'::date,
            'Американская актриса.',
            'США',
            'phyllis_smith.webp'
        ),
        (
            'Энтони Гонсалес',
            '2004-09-23'::date,
            'Американский актёр.',
            'США',
            'anthony_gonzalez.webp'
        ),
        (
            'Гаэль Гарсиа Берналь',
            '1978-11-30'::date,
            'Актёр и продюсер.',
            'Испания',
            'gael_garcia_bernal.webp'
        ),
        (
            'Майк Майерс',
            '1963-05-25'::date,
            'Канадский актёр и комик.',
            'США',
            'mike_myers.webp'
        ),
        (
            'Эдди Мерфи',
            '1961-04-03'::date,
            'Американский актёр и комик.',
            'США',
            'eddie_murphy.webp'
        ),
        (
            'Джей Барушель',
            '1982-04-09'::date,
            'Канадский актёр.',
            'США',
            'jay_baruchel.webp'
        ),
        (
            'Америка Феррера',
            '1984-04-18'::date,
            'Американская актриса.',
            'США',
            'america_ferrera.webp'
        ),
        (
            'Джейсон Бейтман',
            '1969-01-14'::date,
            'Американский актёр.',
            'США',
            'jason_bateman.webp'
        ),
        (
            'Джиннифер Гудвин',
            '1978-05-22'::date,
            'Американская актриса.',
            'США',
            'ginnifer_goodwin.webp'
        ),
        (
            'Эдвард Эснер',
            '1929-11-15'::date,
            'Американский актёр.',
            'США',
            'ed_asner.webp'
        ),
        (
            'Джордан Нагаи',
            '2000-02-05'::date,
            'Американский актёр.',
            'США',
            'jordan_nagai.webp'
        ),
        (
            'Пэттон Освальт',
            '1969-01-27'::date,
            'Американский актёр и комик.',
            'США',
            'patton_oswalt.webp'
        ),
        (
            'Лу Романо',
            '1972-04-15'::date,
            'Американский актёр и художник-постановщик.',
            'США',
            'lou_romano.webp'
        ),
        (
            'Брайан Крэнстон',
            '1956-03-07'::date,
            'Американский актёр.',
            'США',
            'bryan_cranston.webp'
        ),
        (
            'Аарон Пол',
            '1979-08-27'::date,
            'Американский актёр.',
            'США',
            'aaron_paul.webp'
        ),
        (
            'Милли Бобби Браун',
            '2004-02-19'::date,
            'Британская актриса.',
            'Великобритания',
            'millie_bobby_brown.webp'
        ),
        (
            'Финн Вулфард',
            '2002-12-23'::date,
            'Канадский актёр.',
            'США',
            'finn_wolfhard.webp'
        ),
        (
            'Брайс Даллас Ховард',
            '1981-03-02'::date,
            'Американская актриса.',
            'США',
            'bryce_dallas_howard.webp'
        ),
        (
            'Дэниел Калуя',
            '1989-02-24'::date,
            'Британский актёр.',
            'Великобритания',
            'daniel_kaluuya.webp'
        ),
        (
            'Эмилия Кларк',
            '1986-10-23'::date,
            'Британская актриса.',
            'Великобритания',
            'emilia_clarke.webp'
        ),
        (
            'Кит Харингтон',
            '1986-12-26'::date,
            'Британский актёр.',
            'Великобритания',
            'kit_harington.webp'
        ),
        (
            'Бенедикт Камбербэтч',
            '1976-07-19'::date,
            'Британский актёр.',
            'Великобритания',
            'benedict_cumberbatch.webp'
        ),
        (
            'Мартин Фриман',
            '1971-09-08'::date,
            'Британский актёр.',
            'Великобритания',
            'martin_freeman.webp'
        ),
        (
            'Вуди Харрельсон',
            '1961-07-23'::date,
            'Американский актёр.',
            'США',
            'woody_harrelson.webp'
        ),
        (
            'Джаред Харрис',
            '1961-08-24'::date,
            'Британский актёр.',
            'Великобритания',
            'jared_harris.webp'
        ),
        (
            'Эмили Уотсон',
            '1967-01-14'::date,
            'Британская актриса.',
            'Великобритания',
            'emily_watson.webp'
        ),
        (
            'Педро Паскаль',
            '1975-04-02'::date,
            'Чилийско-американский актёр.',
            'США',
            'pedro_pascal.webp'
        ),
        (
            'Кэти Сакхофф',
            '1980-04-08'::date,
            'Американская актриса.',
            'США',
            'katee_sackhoff.webp'
        ),
        (
            'Элла Пернелл',
            '1996-09-17'::date,
            'Британская актриса.',
            'Великобритания',
            'ella_purnell.webp'
        ),
        (
            'Хейли Стайнфелд',
            '1996-12-11'::date,
            'Американская актриса.',
            'США',
            'hailee_steinfeld.webp'
        ),
        (
            'Мэтт Смит',
            '1982-10-28'::date,
            'Британский актёр.',
            'Великобритания',
            'matt_smith.webp'
        ),
        (
            'Эмма ДАрси',
            '1992-06-27'::date,
            'Британская актриса.',
            'Великобритания',
            'emma_darcy.webp'
        ),
        (
            'Риз Ахмед',
            '1982-12-01'::date,
            'Британский актёр.',
            'Великобритания',
            'riz_ahmed.webp'
        ),
        (
            'Джои Бэти',
            '1989-04-06'::date,
            'Британский актёр.',
            'Великобритания',
            'joey_batey.webp'
        ),
        (
            'Микайла Коул',
            '1987-10-01'::date,
            'Британская актриса.',
            'Великобритания',
            'michaela_coel.webp'
        ),
        (
            'Рэйчел Шентон',
            '1987-12-21'::date,
            'Британская актриса.',
            'Великобритания',
            'rachel_shenton.webp'
        ),
        (
            'Мэйси Слай',
            '2009-01-01'::date,
            'Британская актриса.',
            'Великобритания',
            'maisie_sly.webp'
        ),
        (
            'Мэттью Нидхэм',
            '1984-01-01'::date,
            'Британский актёр.',
            'Великобритания',
            'matthew_needham.webp'
        ),
        (
            'Хлоя Пирри',
            '1987-08-25'::date,
            'Британская актриса.',
            'Великобритания',
            'chloe_pirrie.webp'
        ),
        (
            'Фатима Боудженах',
            '1988-01-01'::date,
            'Актриса короткометражного кино.',
            'США',
            'fatima_bojenah.webp'
        ),
        (
            'Отигба Укот',
            '1980-01-01'::date,
            'Актёр короткометражного кино.',
            'США',
            'utigba_ukoh.webp'
        ),
        (
            'Шон Кристенсен',
            '1979-01-01'::date,
            'Американский актёр и режиссёр.',
            'США',
            'sean_christensen.webp'
        ),
        (
            'Фатима Птачек',
            '2000-08-20'::date,
            'Американская актриса.',
            'США',
            'fatima_ptacek.webp'
        ),
        (
            'Кеке Палмер',
            '1993-08-26'::date,
            'Американская актриса.',
            'США',
            'keke_palmer.webp'
        ),
        (
            'Брэдли Уитфорд',
            '1959-10-10'::date,
            'Американский актёр.',
            'США',
            'bradley_whitford.webp'
        ),
        (
            'Джонатан Такер',
            '1982-05-31'::date,
            'Американский актёр.',
            'США',
            'jonathan_tucker.webp'
        ),
        (
            'Даниэль Макдональд',
            '1991-05-19'::date,
            'Австралийская актриса.',
            'Великобритания',
            'danielle_macdonald.webp'
        ),
        (
            'Мария Диззия',
            '1974-12-29'::date,
            'Американская актриса.',
            'США',
            'maria_dizzia.webp'
        ),
        (
            'Грег Келлер',
            '1975-01-01'::date,
            'Американский актёр.',
            'США',
            'greg_keller.webp'
        ),
        (
            'Рэйф Файнс',
            '1962-12-22'::date,
            'Британский актёр.',
            'Великобритания',
            'ralph_fiennes.webp'
        ),
        (
            'Шеймус ОХара',
            '1985-01-01'::date,
            'Ирландский актёр.',
            'Великобритания',
            'seamus_ohara.webp'
        ),
        (
            'Джеймс Мартин',
            '1992-01-01'::date,
            'Ирландский актёр.',
            'Великобритания',
            'james_martin.webp'
        ),
        (
            'Дэниел Рэдклифф',
            '1989-07-23'::date,
            'Британский актёр.',
            'Великобритания',
            'daniel_radcliffe.webp'
        ),
        (
            'Эмма Уотсон',
            '1990-04-15'::date,
            'Британская актриса.',
            'Великобритания',
            'emma_watson.webp'
        ),
        (
            'Маколей Калкин',
            '1980-08-26'::date,
            'Американский актёр.',
            'США',
            'macaulay_culkin.webp'
        ),
        (
            'Джо Пеши',
            '1943-02-09'::date,
            'Американский актёр.',
            'США',
            'joe_pesci.webp'
        ),
        (
            'Бен Уишоу',
            '1980-10-14'::date,
            'Британский актёр.',
            'Великобритания',
            'ben_whishaw.webp'
        ),
        (
            'Хью Грант',
            '1960-09-09'::date,
            'Британский актёр.',
            'Великобритания',
            'hugh_grant.webp'
        ),
        (
            'Джорджи Хенли',
            '1995-07-09'::date,
            'Британская актриса.',
            'Великобритания',
            'georgie_henley.webp'
        ),
        (
            'Скандар Кейнс',
            '1991-09-05'::date,
            'Британский актёр.',
            'Великобритания',
            'skandar_keynes.webp'
        ),
        (
            'Анджелина Джоли',
            '1975-06-04'::date,
            'Американская актриса.',
            'США',
            'angelina_jolie.webp'
        ),
        (
            'Эль Фаннинг',
            '1998-04-09'::date,
            'Американская актриса.',
            'США',
            'elle_fanning.webp'
        ),
        (
            'Эмили Блант',
            '1983-02-23'::date,
            'Британская актриса.',
            'Великобритания',
            'emily_blunt.webp'
        ),
        (
            'Лин-Мануэль Миранда',
            '1980-01-16'::date,
            'Американский актёр и композитор.',
            'США',
            'lin_manuel_miranda.webp'
        ),
        (
            'Дуэйн Джонсон',
            '1972-05-02'::date,
            'Американский актёр.',
            'США',
            'dwayne_johnson.webp'
        ),
        (
            'Кевин Харт',
            '1979-07-06'::date,
            'Американский актёр и комик.',
            'США',
            'kevin_hart.webp'
        ),
        (
            'Джейми Фокс',
            '1967-12-13'::date,
            'Американский актёр и музыкант.',
            'США',
            'jamie_foxx.webp'
        ),
        (
            'Тина Фей',
            '1970-05-18'::date,
            'Американская актриса и сценаристка.',
            'США',
            'tina_fey.webp'
        ),
        (
            'Джейкоб Тремблей',
            '2006-10-05'::date,
            'Канадский актёр.',
            'США',
            'jacob_tremblay.webp'
        ),
        (
            'Майя Рудольф',
            '1972-07-27'::date,
            'Американская актриса.',
            'США',
            'maya_rudolph.webp'
        ),
        (
            'Стефани Беатрис',
            '1981-02-10'::date,
            'Американская актриса.',
            'США',
            'stephanie_beatriz.webp'
        ),
        (
            'Джон Легуизамо',
            '1960-07-22'::date,
            'Американский актёр.',
            'США',
            'john_leguizamo.webp'
        )
)
INSERT INTO actor (full_name, birthdate, biography, country_id, picture_file_key)
SELECT a.full_name, a.birthdate, a.biography, c.id, a.picture_file_key
FROM actor_seed a
JOIN country c ON c.title = a.country_title
ON CONFLICT (full_name, birthdate) DO NOTHING;

WITH cast_map(movie_title, actor_name) AS (
    VALUES
        ('Интерстеллар', 'Мэттью Макконахи'),
        ('Интерстеллар', 'Энн Хэтэуэй'),
        ('Начало', 'Леонардо ДиКаприо'),
        ('Начало', 'Джозеф Гордон-Левитт'),
        ('Матрица', 'Киану Ривз'),
        ('Матрица', 'Кэрри-Энн Мосс'),
        ('Гладиатор', 'Рассел Кроу'),
        ('Гладиатор', 'Хоакин Феникс'),
        ('Отступники', 'Леонардо ДиКаприо'),
        ('Отступники', 'Мэтт Деймон'),
        ('Одержимость', 'Майлз Теллер'),
        ('Одержимость', 'Дж. К. Симмонс'),
        ('Остров проклятых', 'Леонардо ДиКаприо'),
        ('Остров проклятых', 'Марк Руффало'),
        ('Семь', 'Брэд Питт'),
        ('Семь', 'Морган Фримен'),
        ('Пленницы', 'Хью Джекман'),
        ('Пленницы', 'Джейк Джилленхол'),
        ('Социальная сеть', 'Джесси Айзенберг'),
        ('Социальная сеть', 'Эндрю Гарфилд'),
        ('История игрушек', 'Том Хэнкс'),
        ('История игрушек', 'Тим Аллен'),
        ('Король Лев', 'Мэттью Бродерик'),
        ('Король Лев', 'Джеймс Эрл Джонс'),
        ('ВАЛЛ-И', 'Бен Бертт'),
        ('ВАЛЛ-И', 'Элиза Найт'),
        ('Головоломка', 'Эми Полер'),
        ('Головоломка', 'Филлис Смит'),
        ('Тайна Коко', 'Энтони Гонсалес'),
        ('Тайна Коко', 'Гаэль Гарсиа Берналь'),
        ('Шрек', 'Майк Майерс'),
        ('Шрек', 'Эдди Мерфи'),
        ('Как приручить дракона', 'Джей Барушель'),
        ('Как приручить дракона', 'Америка Феррера'),
        ('Зверополис', 'Джейсон Бейтман'),
        ('Зверополис', 'Джиннифер Гудвин'),
        ('Вверх', 'Эдвард Эснер'),
        ('Вверх', 'Джордан Нагаи'),
        ('Рататуй', 'Пэттон Освальт'),
        ('Рататуй', 'Лу Романо'),
        ('Во все тяжкие', 'Брайан Крэнстон'),
        ('Во все тяжкие', 'Аарон Пол'),
        ('Очень странные дела', 'Милли Бобби Браун'),
        ('Очень странные дела', 'Финн Вулфард'),
        ('Черное зеркало', 'Брайс Даллас Ховард'),
        ('Черное зеркало', 'Дэниел Калуя'),
        ('Игра престолов', 'Эмилия Кларк'),
        ('Игра престолов', 'Кит Харингтон'),
        ('Шерлок', 'Бенедикт Камбербэтч'),
        ('Шерлок', 'Мартин Фриман'),
        ('Настоящий детектив', 'Мэттью Макконахи'),
        ('Настоящий детектив', 'Вуди Харрельсон'),
        ('Чернобыль', 'Джаред Харрис'),
        ('Чернобыль', 'Эмили Уотсон'),
        ('Мандалорец', 'Педро Паскаль'),
        ('Мандалорец', 'Кэти Сакхофф'),
        ('Аркейн', 'Элла Пернелл'),
        ('Аркейн', 'Хейли Стайнфелд'),
        ('Дом дракона', 'Мэтт Смит'),
        ('Дом дракона', 'Эмма ДАрси'),
        ('Долгое прощание', 'Риз Ахмед'),
        ('Долгое прощание', 'Джои Бэти'),
        ('Два далеких незнакомца', 'Джои Бэти'),
        ('Два далеких незнакомца', 'Микайла Коул'),
        ('Заикание', 'Мэттью Нидхэм'),
        ('Заикание', 'Хлоя Пирри'),
        ('Тихий ребенок', 'Рэйчел Шентон'),
        ('Тихий ребенок', 'Мэйси Слай'),
        ('Комендантский час', 'Шон Кристенсен'),
        ('Комендантский час', 'Фатима Птачек'),
        ('Пой', 'Кеке Палмер'),
        ('Пой', 'Брэдли Уитфорд'),
        ('Кожа', 'Джонатан Такер'),
        ('Кожа', 'Даниэль Макдональд'),
        ('Окно по соседству', 'Мария Диззия'),
        ('Окно по соседству', 'Грег Келлер'),
        ('Чудесная история Генри Шугара', 'Бенедикт Камбербэтч'),
        ('Чудесная история Генри Шугара', 'Рэйф Файнс'),
        ('Ирландское прощание', 'Шеймус ОХара'),
        ('Ирландское прощание', 'Джеймс Мартин'),
        ('Гарри Поттер и философский камень', 'Дэниел Рэдклифф'),
        ('Гарри Поттер и философский камень', 'Эмма Уотсон'),
        ('Один дома', 'Маколей Калкин'),
        ('Один дома', 'Джо Пеши'),
        ('Приключения Паддингтона 2', 'Бен Уишоу'),
        ('Приключения Паддингтона 2', 'Хью Грант'),
        ('Хроники Нарнии: Лев, колдунья и волшебный шкаф', 'Джорджи Хенли'),
        ('Хроники Нарнии: Лев, колдунья и волшебный шкаф', 'Скандар Кейнс'),
        ('Малефисента', 'Анджелина Джоли'),
        ('Малефисента', 'Эль Фаннинг'),
        ('Мэри Поппинс возвращается', 'Эмили Блант'),
        ('Мэри Поппинс возвращается', 'Лин-Мануэль Миранда'),
        ('Джуманджи: Зов джунглей', 'Дуэйн Джонсон'),
        ('Джуманджи: Зов джунглей', 'Кевин Харт'),
        ('Душа', 'Джейми Фокс'),
        ('Душа', 'Тина Фей'),
        ('Лука', 'Джейкоб Тремблей'),
        ('Лука', 'Майя Рудольф'),
        ('Энканто', 'Стефани Беатрис'),
        ('Энканто', 'Джон Легуизамо')
)
INSERT INTO actor_to_movie (actor_id, movie_id)
SELECT a.id, m.id
FROM cast_map cm
JOIN movie m ON m.title = cm.movie_title
JOIN actor a ON a.full_name = cm.actor_name
ON CONFLICT DO NOTHING;

INSERT INTO episode (
    movie_id, description, season_number, episode_number, title,
    duration_seconds, picture_file_key, video_file_key
)
SELECT
    m.id,
    m.description,
    1,
    1,
    CASE m.title
        WHEN 'Во все тяжкие' THEN 'Пилот'
        WHEN 'Очень странные дела' THEN 'Исчезновение Уилла Байерса'
        WHEN 'Черное зеркало' THEN 'Национальный гимн'
        WHEN 'Игра престолов' THEN 'Зима близко'
        WHEN 'Шерлок' THEN 'Этюд в розовых тонах'
        WHEN 'Настоящий детектив' THEN 'Долгая яркая тьма'
        WHEN 'Чернобыль' THEN '1:23:45'
        WHEN 'Мандалорец' THEN 'Глава 1'
        WHEN 'Аркейн' THEN 'Добро пожаловать на игровую площадку'
        WHEN 'Дом дракона' THEN 'Наследники дракона'
    END,
    m.duration_seconds,
    '' ||
        CASE m.title
            WHEN 'Во все тяжкие' THEN 'breaking_bad_s01e01.webp'
            WHEN 'Очень странные дела' THEN 'stranger_things_s01e01.webp'
            WHEN 'Черное зеркало' THEN 'black_mirror_s01e01.webp'
            WHEN 'Игра престолов' THEN 'game_of_thrones_s01e01.webp'
            WHEN 'Шерлок' THEN 'sherlock_s01e01.webp'
            WHEN 'Настоящий детектив' THEN 'true_detective_s01e01.webp'
            WHEN 'Чернобыль' THEN 'chernobyl_s01e01.webp'
            WHEN 'Мандалорец' THEN 'mandalorian_s01e01.webp'
            WHEN 'Аркейн' THEN 'arcane_s01e01.webp'
            WHEN 'Дом дракона' THEN 'house_of_the_dragon_s01e01.webp'
        END,
    '' ||
        CASE m.title
            WHEN 'Начало' THEN '1.mp4'
            WHEN 'Во все тяжкие' THEN 'breaking_bad_s01e01.mp4'
            WHEN 'Очень странные дела' THEN 'stranger_things_s01e01.mp4'
            WHEN 'Черное зеркало' THEN 'black_mirror_s01e01.mp4'
            WHEN 'Игра престолов' THEN 'game_of_thrones_s01e01.mp4'
            WHEN 'Шерлок' THEN 'sherlock_s01e01.mp4'
            WHEN 'Настоящий детектив' THEN 'true_detective_s01e01.mp4'
            WHEN 'Чернобыль' THEN 'chernobyl_s01e01.mp4'
            WHEN 'Мандалорец' THEN 'mandalorian_s01e01.mp4'
            WHEN 'Аркейн' THEN 'arcane_s01e01.mp4'
            WHEN 'Дом дракона' THEN 'house_of_the_dragon_s01e01.mp4'
        END
FROM movie m
WHERE m.content_type = 'series'
ON CONFLICT (movie_id, season_number, episode_number) DO NOTHING;

WITH selection_map(selection_title, movie_title) AS (
    VALUES
        ('Популярные', 'Интерстеллар'),
        ('Популярные', 'Начало'),
        ('Популярные', 'Матрица'),
        ('Популярные', 'Гладиатор'),
        ('Популярные', 'Остров проклятых'),
        ('Популярные', 'История игрушек'),
        ('Популярные', 'Король Лев'),
        ('Популярные', 'Во все тяжкие'),
        ('Популярные', 'Игра престолов'),
        ('Популярные', 'Гарри Поттер и философский камень'),
        ('Топ-10', 'Интерстеллар'),
        ('Топ-10', 'Одержимость'),
        ('Топ-10', 'Семь'),
        ('Топ-10', 'Пленницы'),
        ('Топ-10', 'Матрица'),
        ('Топ-10', 'ВАЛЛ-И'),
        ('Топ-10', 'Тайна Коко'),
        ('Топ-10', 'Шерлок'),
        ('Топ-10', 'Чернобыль'),
        ('Топ-10', 'Душа'),
        ('Фильмы', 'Интерстеллар'),
        ('Фильмы', 'Начало'),
        ('Фильмы', 'Матрица'),
        ('Фильмы', 'Гладиатор'),
        ('Фильмы', 'Отступники'),
        ('Фильмы', 'Одержимость'),
        ('Фильмы', 'Остров проклятых'),
        ('Фильмы', 'Семь'),
        ('Фильмы', 'Пленницы'),
        ('Фильмы', 'Социальная сеть'),
        ('Мультфильмы', 'История игрушек'),
        ('Мультфильмы', 'Король Лев'),
        ('Мультфильмы', 'ВАЛЛ-И'),
        ('Мультфильмы', 'Головоломка'),
        ('Мультфильмы', 'Тайна Коко'),
        ('Мультфильмы', 'Шрек'),
        ('Мультфильмы', 'Как приручить дракона'),
        ('Мультфильмы', 'Зверополис'),
        ('Мультфильмы', 'Вверх'),
        ('Мультфильмы', 'Рататуй'),
        ('Сериалы', 'Во все тяжкие'),
        ('Сериалы', 'Очень странные дела'),
        ('Сериалы', 'Черное зеркало'),
        ('Сериалы', 'Игра престолов'),
        ('Сериалы', 'Шерлок'),
        ('Сериалы', 'Настоящий детектив'),
        ('Сериалы', 'Чернобыль'),
        ('Сериалы', 'Мандалорец'),
        ('Сериалы', 'Аркейн'),
        ('Сериалы', 'Дом дракона'),
        ('Короткометражки', 'Долгое прощание'),
        ('Короткометражки', 'Два далеких незнакомца'),
        ('Короткометражки', 'Заикание'),
        ('Короткометражки', 'Тихий ребенок'),
        ('Короткометражки', 'Комендантский час'),
        ('Короткометражки', 'Пой'),
        ('Короткометражки', 'Кожа'),
        ('Короткометражки', 'Окно по соседству'),
        ('Короткометражки', 'Чудесная история Генри Шугара'),
        ('Короткометражки', 'Ирландское прощание'),
        ('Семейное', 'Гарри Поттер и философский камень'),
        ('Семейное', 'Один дома'),
        ('Семейное', 'Приключения Паддингтона 2'),
        ('Семейное', 'Хроники Нарнии: Лев, колдунья и волшебный шкаф'),
        ('Семейное', 'Малефисента'),
        ('Семейное', 'Мэри Поппинс возвращается'),
        ('Семейное', 'Джуманджи: Зов джунглей'),
        ('Семейное', 'Душа'),
        ('Семейное', 'Лука'),
        ('Семейное', 'Энканто'),
        ('Новинки', 'Аркейн'),
        ('Новинки', 'Дом дракона'),
        ('Новинки', 'Лука'),
        ('Новинки', 'Энканто'),
        ('Новинки', 'Чудесная история Генри Шугара')
)
INSERT INTO movie_to_selection (movie_id, selection_id)
SELECT m.id, s.id
FROM selection_map sm
JOIN movie m ON m.title = sm.movie_title
JOIN selection s ON s.title = sm.selection_title
ON CONFLICT DO NOTHING;

COMMIT;
