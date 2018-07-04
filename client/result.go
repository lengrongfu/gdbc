package client

type GdbcResult struct {
	affectedRows uint64 //affected rows
	lastInsertID uint64 //last insert-id
}

func (r GdbcResult) LastInsertId() (int64, error) {
	return int64(r.lastInsertID), nil
}

func (r GdbcResult) RowsAffected() (int64, error) {
	return int64(r.affectedRows), nil
}
