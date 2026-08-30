CREATE TABLE if not exists `feed_states` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `feed_id` bigint NOT NULL,
  `last_refreshed` datetime DEFAULT NULL,
  `last_success` datetime DEFAULT NULL,
  `item_count` int DEFAULT NULL,
  `error` text,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_feed_state_feed_id` (`feed_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO feed_states (feed_id, error)
SELECT e.feed_id, max(e.error) FROM feed_errors e
JOIN feeds f ON f.id = e.feed_id GROUP BY e.feed_id
ON DUPLICATE KEY UPDATE error = values(error);

INSERT INTO feed_states (feed_id, item_count)
SELECT z.feed_id, max(z.size) FROM feed_sizes z
JOIN feeds f ON f.id = z.feed_id GROUP BY z.feed_id
ON DUPLICATE KEY UPDATE item_count = values(item_count);

DROP TABLE IF EXISTS feed_errors;
DROP TABLE IF EXISTS feed_sizes;

update `settings` set val = '4' where `key` = 'db_version';
