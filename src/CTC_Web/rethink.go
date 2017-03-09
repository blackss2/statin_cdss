package main

import (
	"github.com/blackss2/utility/convert"
	"github.com/dancannon/gorethink"
	"log"
	"strings"
)

func createDB(session *gorethink.Session, dbName string) gorethink.Term {
	if true {
		hasDB := false
		res, err := gorethink.DBList().Run(session)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer res.Close()

		var row interface{}
		for res.Next(&row) {
			if convert.String(row) == dbName {
				hasDB = true
			}
		}
		if !hasDB {
			gorethink.DBCreate(dbName).Run(session)
		}
	}
	return gorethink.DB(dbName)
}

func createTable(session *gorethink.Session, db gorethink.Term, tableName string, indexNames ...string) gorethink.Term {
	if true {
		thash := tableHash(session, db)
		if !thash[tableName] {
			_, err := db.TableCreate(tableName).Run(session)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
		if len(indexNames) > 0 {
			ihash := indexHash(session, db, tableName)
			for _, v := range indexNames {
				if ihash[v] {
					delete(ihash, v)
				} else {
					if strings.Index(v, ".") >= 0 {
						list := strings.Split(v, ".")
						fields := make([]interface{}, 0)
						for _, f := range list {
							fields = append(fields, gorethink.Row.Field(f))
						}

						err := db.Table(tableName).IndexCreateFunc(v, fields).Exec(session)
						if err != nil {
							log.Fatalln(err.Error())
						}
					} else if strings.Index(v, "->") >= 0 {
						list := strings.Split(v, "->")
						var fields gorethink.Term
						for i, f := range list {
							if i == 0 {
								fields = gorethink.Row.Field(f)
							} else {
								fields = fields.Field(f)
							}
						}

						err := db.Table(tableName).IndexCreateFunc(v, fields).Exec(session)
						if err != nil {
							log.Fatalln(err.Error())
						}
					} else {
						_, err := db.Table(tableName).IndexCreate(v).Run(session)
						if err != nil {
							log.Fatalln(err.Error())
						}
					}
				}
			}
			for k, _ := range ihash {
				_, err := db.Table(tableName).IndexDrop(k).Run(session)
				if err != nil {
					log.Fatalln(err.Error())
				}
			}
		}
	}
	return db.Table(tableName)
}

func tableHash(session *gorethink.Session, db gorethink.Term) map[string]bool {
	res, err := db.TableList().Run(session)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer res.Close()

	hash := make(map[string]bool)
	var row interface{}
	for res.Next(&row) {
		hash[convert.String(row)] = true
	}
	return hash
}

func indexHash(session *gorethink.Session, db gorethink.Term, tableName string) map[string]bool {
	res, err := db.Table(tableName).IndexList().Run(session)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer res.Close()

	hash := make(map[string]bool, 0)
	var row interface{}
	for res.Next(&row) {
		hash[convert.String(row)] = true
	}
	return hash
}
