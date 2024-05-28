CREATE TABLE IF NOT EXISTS songs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    composer VARCHAR(255)
);

INSERT INTO songs (id, name, composer) VALUES
('1', 'Song 1', 'Composer 1'),
('2', 'Song 2', 'Composer 2'),
('3', 'Song 3', 'Composer 3');
