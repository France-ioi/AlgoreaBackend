-- +migrate Up
ALTER DATABASE CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

-- +migrate Down
ALTER DATABASE CHARACTER SET utf8mb3 COLLATE utf8_general_ci;
