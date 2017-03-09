package main

import (
	"errors"
	"github.com/dancannon/gorethink"
)

var (
	ErrNotExist = errors.New("not exist")
)

type DBTable struct {
	session *gorethink.Session
	db      gorethink.Term
	table   gorethink.Term
}

func (dt *DBTable) Init(tableName string, indexNames ...string) {
	dt.table = createTable(dt.session, dt.db, tableName, indexNames...)
}

func (dt *DBTable) Session() *gorethink.Session {
	return dt.session
}

func (dt *DBTable) BaseQuery() gorethink.Term {
	return dt.table
}

func (dt *DBTable) Insert(value interface{}) (string, error) {
	res, err := dt.table.
		Insert(value).
		RunWrite(dt.session, gorethink.RunOpts{MaxBatchRows: 200})
	if err != nil {
		return "", err
	}

	if res.Inserted == 0 {
		return "", errors.New("not inserted")
	}

	if len(res.GeneratedKeys) == 0 {
		//id is specifieid
		return "", nil
	} else {
		return res.GeneratedKeys[0], nil
	}
}

func (dt *DBTable) InsertBatch(value interface{}) error {
	_, err := dt.table.
		Insert(value).
		RunWrite(dt.session, gorethink.RunOpts{MaxBatchRows: 200})
	if err != nil {
		return err
	}
	return nil
}

func (dt *DBTable) Update(Id interface{}, value interface{}) error {
	//update data & time
	res, err := dt.table.
		Get(Id).
		Update(value).
		RunWrite(dt.session)
	if err != nil {
		return err
	}

	if res.Unchanged == 0 && res.Updated == 0 && res.Replaced == 0 {
		return ErrNotExist
	}

	return nil
}

func (dt *DBTable) UpdateById(Ids []string, hash interface{}) error {
	//update data
	_, err := dt.table.
		Filter(func(row gorethink.Term) gorethink.Term { return gorethink.Expr(Ids).Contains(row.Field("id")) }).
		Update(hash).
		RunWrite(dt.session)
	if err != nil {
		return err
	}

	return nil
}

func (dt *DBTable) UpdateByIndex(IndexName string, Index interface{}, hash map[string]interface{}) error {
	//update data
	_, err := dt.table.
		GetAllByIndex(IndexName, Index).
		Update(hash).
		RunWrite(dt.session)
	if err != nil {
		return err
	}

	return nil
}

func (dt *DBTable) Delete(Id interface{}) error {
	//delete data
	res, err := dt.table.
		Get(Id).
		Delete().
		RunWrite(dt.session)
	if err != nil {
		return err
	}

	if res.Deleted == 0 {
		return ErrNotExist
	}

	return nil
}

func (dt *DBTable) DeleteById(IndexName string, Index interface{}, Ids []string) error {
	//delete data
	_, err := dt.table.
		GetAllByIndex(IndexName, Index).
		Filter(func(row gorethink.Term) gorethink.Term { return gorethink.Expr(Ids).Contains(row.Field("id")) }).
		Delete().
		RunWrite(dt.session)
	if err != nil {
		return err
	}

	return nil
}

func (dt *DBTable) DeleteByIndex(IndexName string, Index interface{}) error {
	//delete data
	_, err := dt.table.
		GetAllByIndex(IndexName, Index).
		Delete().
		RunWrite(dt.session)
	if err != nil {
		return err
	}

	return nil
}
