CREATE TABLE memo (
    id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
    body VARCHAR(191),
    latlng GEOMETRY NOT NULL,
    updated_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
) Engine=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE memo_emoji (
    id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
    memo_id BIGINT UNSIGNED,
    emoji VARCHAR(191),
    created_at DATETIME NOT NULL
) Engine=InnoDB DEFAULT CHARSET=utf8mb4;
