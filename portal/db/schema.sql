CREATE TABLE teams (
    id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
    name VARCHAR(191) NOT NULL,
    password VARCHAR(191) NOT NULL,
    app_host VARCHAR(191) NOT NULL,
    color VARCHAR(191) NOT NULL,
    updated_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
) Engine=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE scores (
    id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
    pass BOOLEAN,
    score INT UNSIGNED,
    message TEXT,
    team_id INT,
    updated_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    KEY (team_id, created_at)
) Engine=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE jobs (
    id INT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
    team_id INT,
    status ENUM('waiting', 'running', 'done') NOT NULL,
    enqueued_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    KEY (team_id, created_at)
) Engine=InnoDB DEFAULT CHARSET=utf8mb4;

