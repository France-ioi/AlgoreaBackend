-- +migrate Up
ALTER TABLE `items_strings`
  COMMENT 'Textual content associated with an item, in a given language.',
  MODIFY COLUMN `sTranslator` varchar(100) DEFAULT NULL COMMENT 'Name of the translator(s) of this content',
  MODIFY COLUMN `sTitle` varchar(200) DEFAULT NULL COMMENT 'Title of the item, in the specified language',
  MODIFY COLUMN `sImageUrl` text COMMENT 'Url of a small image associated with this item.',
  MODIFY COLUMN `sSubtitle` varchar(200) DEFAULT NULL COMMENT 'Subtitle of the item in the specified language',
  MODIFY COLUMN `sDescription` text COMMENT 'Description of the item in the specified language',
  MODIFY COLUMN `sEduComment` text COMMENT 'Information about what this item teaches, in the specified language.';

-- +migrate Down
ALTER TABLE `items_strings`
  COMMENT '',
  MODIFY COLUMN `sTranslator` varchar(100) DEFAULT NULL COMMENT '',
  MODIFY COLUMN `sTitle` varchar(200) DEFAULT NULL COMMENT '',
  MODIFY COLUMN `sImageUrl` text COMMENT '',
  MODIFY COLUMN `sSubtitle` varchar(200) DEFAULT NULL COMMENT '',
  MODIFY COLUMN `sDescription` text COMMENT '',
  MODIFY COLUMN `sEduComment` text COMMENT '';
