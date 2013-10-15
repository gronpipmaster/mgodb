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
	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not connect to %s: %s.", connectUrl, err.Error()))
	}
	DbmInstance.Database = session.DB(dbName)
	return nil
}

func (self *Dbm) Find(collectionName string, query interface{}) *mgo.Query {
	return self.getCollection(collectionName).Find(query)
}

func (self *Dbm) Insert(docs ...interface{}) error {
	var err error
	var elem reflect.Value
	var collectionName string
	for _, doc := range docs {
		elem, err = self.getPointer(doc)
		if err != nil {
			return err
		}
		eptr := elem.Addr()
		collectionName, err = getCollectionName(eptr)
		if err != nil {
			return err
		}
		err = runHook("PreInsert", eptr)
		if err != nil {
			return err
		}
		err = self.getCollection(collectionName).Insert(doc)
		if err != nil {
			return err
		}
		err = runHook("PostInsert", eptr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Dbm) Update(Id bson.ObjectId, doc interface{}) error {
	var err error
	var elem reflect.Value
	var collectionName string
	elem, err = self.getPointer(doc)
	if err != nil {
		return err
	}
	eptr := elem.Addr()
	err = runHook("PreUpdate", eptr)
	if err != nil {
		return err
	}
	collectionName, err = getCollectionName(eptr)
	if err != nil {
		return err
	}
	//TODO auto change all fieds
	change := bson.M{}
	err = self.getCollection(collectionName).UpdateId(bson.M{"_id": Id}, change)
	if err != nil {
		return err
	}
	err = runHook("PostUpdate", eptr)
	if err != nil {
		return err
	}
	return nil
}

func (self *Dbm) Delete(Id bson.ObjectId, doc interface{}) error {
	var err error
	var elem reflect.Value
	var collectionName string
	elem, err = self.getPointer(doc)
	if err != nil {
		return err
	}
	eptr := elem.Addr()
	err = runHook("PreDelete", eptr)
	if err != nil {
		return err
	}
	collectionName, err = getCollectionName(eptr)
	if err != nil {
		return err
	}
	err = self.getCollection(collectionName).RemoveId(bson.M{"_id": Id})
	if err != nil {
		return err
	}
	err = runHook("PostDelete", eptr)
	if err != nil {
		return err
	}
	return nil
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

func (self *Dbm) getPointer(doc interface{}) (reflect.Value, error) {
	docV := reflect.ValueOf(doc)
	if docV.Kind() != reflect.Ptr {
		e := fmt.Sprintf("mgodb.Dbm: passed non-pointer: %v (kind=%v)", doc,
			docV.Kind())
		return reflect.Value{}, errors.New(e)
	}
	elem := docV.Elem()
	return elem, nil
}

func runHook(name string, eptr reflect.Value) error {
	hook := eptr.MethodByName(name)
	if hook != zeroVal {
		ret := hook.Call(zeroArgs)
		if len(ret) > 0 && !ret[0].IsNil() {
			return ret[0].Interface().(error)
		}
	}
	return nil
}

func getCollectionName(eptr reflect.Value) (string, error) {
	fn := eptr.MethodByName("CollectionName")
	if fn != zeroVal {
		ret := fn.Call(zeroArgs)
		if len(ret) > 0 && ret[0].String() != "" {
			return ret[0].String(), nil
		}
	}
	return "", errors.New(fmt.Sprintf("get Collection:%s err.", eptr.Type().String()))
}
