CREATE TABLE if not exists `feed_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `url` varchar(300) NOT NULL,
  `config` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

update `settings` set val = '2' where `key` = 'db_version';
