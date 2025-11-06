CREATE TABLE if not exists `settings`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `key` varchar(50) not null UNIQUE,
  `val` text DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO settings (`key`,val) VALUES ('db_version','1');

CREATE TABLE if not exists `folders`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(500) not null,
  `is_expanded` boolean DEFAULT false,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE if not exists `feeds` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `folder_id` bigint DEFAULT NULL,
  `title` varchar(200) DEFAULT NULL,
  `description` varchar(500) DEFAULT NULL,
  `link` varchar(300) DEFAULT NULL,
  `feed_link` varchar(300) NOT NULL UNIQUE,
  `icon` blob,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `items` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `guid` varchar(200) DEFAULT NULL,
  `feed_id` bigint NOT NULL,
  `title` varchar(200) DEFAULT NULL,
  `link` varchar(300) DEFAULT NULL,
  `description` varchar(500) DEFAULT NULL,
  `content` varchar(500) DEFAULT NULL,
  `author` varchar(50) DEFAULT NULL,
  `date` datetime DEFAULT NULL,
  `date_updated` datetime DEFAULT NULL,
  `date_arrived` datetime DEFAULT NULL,
  `status` int DEFAULT NULL,
  `media_links` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_item_guid` (`feed_id`,`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `http_states` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `feed_id` bigint NOT NULL,
  `last_refreshed` datetime DEFAULT NULL,
  `last_modified` datetime DEFAULT NULL,
  `etag` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_feed_id` (`feed_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `feed_sizes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `feed_id` bigint NOT NULL,
  `size` int DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `feed_errors` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `feed_id` bigint NOT NULL,
  `error` text,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


