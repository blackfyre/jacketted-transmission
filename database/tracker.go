package database

type Tracker struct {
	Hash                 string  `db:"hash"`
	CreatedAt            string  `db:"created_at"`
	TransmissionTxResult string  `db:"transmission_tx_result"`
	TransmissionRatio    float64 `db:"transmission_ratio"`
	TransmissionSeedTime int     `db:"transmission_seed_time"`
	TransmissionStatus   *string `db:"transmission_status"`
	Guid                 string  `db:"guid"`
}

func (db *DB) GetTracker(guid string) (Tracker, error) {
	var t Tracker
	err := db.Get(&t, "SELECT * FROM tracker WHERE guid = $1", guid)
	return t, err
}

func (db *DB) GetTrackerByHash(hash string) (Tracker, error) {
	var t Tracker
	err := db.Get(&t, "SELECT * FROM tracker WHERE hash = $1", hash)
	return t, err
}

func (db *DB) CreateTracker(guid string) error {
	_, err := db.Exec("INSERT INTO tracker (guid) VALUES ($1)", guid)
	return err
}

func (db *DB) UpdateTracker(guid string, txResult string, ratio float64, seedTime int, hash string) error {
	_, err := db.Exec("UPDATE tracker SET transmission_tx_result = $1, transmission_ratio = $2, transmission_seed_time = $3, hash = $4 WHERE guid = $5", txResult, ratio, seedTime, hash, guid)
	return err
}

func (db *DB) DeleteTrackersOlderThan(created_at string) error {
	_, err := db.Exec("DELETE FROM tracker WHERE created_at < $1", created_at)
	return err
}
