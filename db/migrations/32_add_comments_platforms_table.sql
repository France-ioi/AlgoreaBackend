-- +migrate Up
ALTER TABLE `platforms`
  COMMENT 'Platforms that host content',
  MODIFY COLUMN `sPublicKey` varchar(512) NOT NULL DEFAULT '' COMMENT 'Public key of this platform',
  MODIFY COLUMN `bUsesTokens` tinyint(1) NOT NULL COMMENT 'Whether this platform supports tokens. If true, data such as the score sent to the platform will be sent as JWT, and the data sent by the platform must be signed with the key as well.',
  MODIFY COLUMN `sRegexp` text COMMENT 'Regexp matching the urls, to automatically detect content from this platform. It is the only way to specify which items are from which platform.',
  MODIFY COLUMN `iPriority` int(11) NOT NULL DEFAULT '0' COMMENT 'Priority of the regexp compared to others (higher value is tried first).';

-- +migrate Down
ALTER TABLE `platforms`
  COMMENT '',
  MODIFY COLUMN `sPublicKey` varchar(512) NOT NULL DEFAULT '' COMMENT '',
  MODIFY COLUMN `bUsesTokens` tinyint(1) NOT NULL COMMENT '',
  MODIFY COLUMN `sRegexp` text COMMENT '',
  MODIFY COLUMN `iPriority` int(11) NOT NULL DEFAULT '0' COMMENT '';
