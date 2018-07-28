CREATE DATABASE IF NOT EXISTS shortlink CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE shortlink

CREATE TABLE `t_entry` (
  `id` int(11) NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `_key` varchar(64) NOT NULL,
  `_value` varchar(2048) NOT NULL,
  `_duration` int(11),
  `_password` varchar(16),
  `_dt` bigint(20),
  INDEX `key_index` (`_key`)
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `t_access_record` (
  `id` int(11) NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `_key` varchar(64) NOT NULL,
  `_ua` varchar(2048),
  `_ip` varchar(64),
  `_status` tinyint(4) NOT NULL,
  `_dt` bigint(20),
  INDEX `key_index` (`_key`)
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci;
