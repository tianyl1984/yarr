CREATE TABLE `feed_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `url` varchar(300) NOT NULL,
  `config` text,
  `use_proxy` tinyint(1) NOT NULL DEFAULT '0',
  `use_browserless` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

update `settings` set val = '2' where `key` = 'db_version';
