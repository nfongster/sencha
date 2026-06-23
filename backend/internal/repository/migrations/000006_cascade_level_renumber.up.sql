ALTER TABLE vocabulary DROP CONSTRAINT vocabulary_level_number_fkey;
ALTER TABLE vocabulary ADD CONSTRAINT vocabulary_level_number_fkey
    FOREIGN KEY (level_number) REFERENCES levels(number) ON UPDATE CASCADE;

ALTER TABLE sentences DROP CONSTRAINT sentences_level_number_fkey;
ALTER TABLE sentences ADD CONSTRAINT sentences_level_number_fkey
    FOREIGN KEY (level_number) REFERENCES levels(number) ON UPDATE CASCADE;
