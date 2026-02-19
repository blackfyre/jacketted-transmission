ALTER TABLE tracker
ADD COLUMN transmission_tx_result TEXT;

ALTER TABLE tracker
ADD COLUMN transmission_ratio REAL;

ALTER TABLE tracker
ADD COLUMN transmission_seed_time INTEGER;

ALTER TABLE tracker
ADD COLUMN transmission_status TEXT;

ALTER TABLE tracker
ADD COLUMN guid TEXT;

CREATE UNIQUE INDEX tracker_guid_index ON tracker (guid);