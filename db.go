package mgodb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
	"time"
)

var zeroVal reflect.Value
var zeroArgs []reflect.Value

var DbmInstance *Dbm

type Dbm struct {
	Database *mgo.Database
}

func (self *Dbm) GetInstance() *Dbm {
	return DbmInstance
}

func (self *Dbm) Init(connectUrl string, dbName string, timeout time.Duration) error {
	var err error
	var session *mgo.Session
	DbmInstance = &Dbm{}
	session, err = mgo.DialWithTimeout(connectUrl, timeout*time.Second)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not connect to %s: %s.", connectUrl, err.Error()))
	}
	session.SetMode(mgo.Monotonic, true)
	DbmInstance.Database = session.DB(dbName)
	return nil
}

func (self *Dbm) Find(collectionName string, query interface{}) *mgo.Query {
	return self.getCollection(collectionName).Find(query)
}

func (self *Dbm) Insert(collectionName string, doc interface{}) error {
	var err error
	if err = callToDoc("BeforeInsert", doc); err != nil {
		return err
	}
	if err = self.getCollection(collectionName).Insert(doc); err != nil {
		return err
	}
	if err = callToDoc("AfterInsert", doc); err != nil {
		return err
	}
	return nil
}

func (self *Dbm) Update(collectionName, id string, doc interface{}) error {
	var err error
	if err = callToDoc("BeforeUpdate", doc); err != nil {
		return err
	}
	if err = self.getCollection(collectionName).UpdateId(bson.ObjectIdHex(id), doc); err != nil {
		return err
	}
	if err = callToDoc("AfterUpdate", doc); err != nil {
		return err
	}
	return nil
}

func (self *Dbm) Delete(collectionName, id string, doc interface{}) error {
	var err error
	if err = callToDoc("BeforeDelete", doc); err != nil {
		return err
	}
	if err = self.getCollection(collectionName).RemoveId(bson.ObjectIdHex(id)); err != nil {
		return err
	}
	if err = callToDoc("AfterDelete", doc); err != nil {
		return err
	}
	return nil
}

func (self *Dbm) InsertAll(collectionName string, docs ...interface{}) error {
	return self.getCollection(collectionName).Insert(docs)
}

func (self *Dbm) UpdateAll(collectionName string, selector interface{}, change interface{}) (*mgo.ChangeInfo, error) {
	info, err := self.getCollection(collectionName).UpdateAll(selector, change)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (self *Dbm) DeleteAll(collectionName string, selector interface{}) (*mgo.ChangeInfo, error) {
	info, err := self.getCollection(collectionName).RemoveAll(selector)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (self *Dbm) getCollection(collectionName string) *mgo.Collection {
	return self.Database.C(collectionName)
}

func callToDoc(method string, doc interface{}) error {
	docV := reflect.ValueOf(doc)
	if docV.Kind() != reflect.Ptr {
		e := fmt.Sprintf("mgodb.Dbm: Passed non-pointer: %v (kind=%v), method:%s", doc, docV.Kind(), method)
		return errors.New(e)
	}
	fn := docV.Elem().Addr().MethodByName(method)
	if fn != zeroVal {
		ret := fn.Call(zeroArgs)
		if len(ret) > 0 && !ret[0].IsNil() {
			return ret[0].Interface().(error)
		}
	}
	return nil
}
