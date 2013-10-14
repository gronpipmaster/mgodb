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

//docs interface Dbm
type DbExecutor interface {
	//Wrapper from mgo.Find
	Find(collectionName string, query interface{}) *mgo.Query
	Insert(docs ...interface{}) error
	Save(doc interface{}) error
	Delete(doc interface{}) error
	//Hooks not affected
	UpdateAll(collectionName string, selector interface{}, change interface{}) (*mgo.ChangeInfo, error)
	//Hooks not affected
	DeleteAll(collectionName string, selector interface{}) (*mgo.ChangeInfo, error)
}

type Dbm struct {
	Database *mgo.Database
}

func (self *Dbm) Init(connectUrl string, dbName string, timeout time.Duration) (*Dbm, error) {
	var err error
	var session *mgo.Session
	session, err = mgo.DialWithTimeout(connectUrl, timeout*time.Second)
	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not connect to %s: %s.", connectUrl, err.Error()))
	}
	self.Database = session.DB(dbName)

	return self, nil
}

func (self *Dbm) Find(collectionName string, query interface{}) *mgo.Query {
	return self.collection(collectionName).Find(query)
}

func (self *Dbm) Insert(docs ...interface{}) error {
	var err error
	var elem reflect.Value
	var collectionName string
	hookarg := hookArg(DbExecutor(self))
	for _, doc := range docs {
		elem, err = self.getPointer(doc)
		if err != nil {
			return err
		}
		eptr := elem.Addr()
		collectionName, err = getCollectionName(eptr, hookarg)
		if err != nil {
			return err
		}
		err = runHook("PreInsert", eptr, hookarg)
		if err != nil {
			return err
		}
		err = self.collection(collectionName).Insert(doc)
		if err != nil {
			return err
		}
		err = runHook("PostInsert", eptr, hookarg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Dbm) Save(doc interface{}) error {
	var err error
	var elem reflect.Value
	var collectionName string
	hookarg := hookArg(DbExecutor(self))
	elem, err = self.getPointer(doc)
	if err != nil {
		return err
	}
	eptr := elem.Addr()
	err = runHook("PreUpdate", eptr, hookarg)
	if err != nil {
		return err
	}
	collectionName, err = getCollectionName(eptr, hookarg)
	if err != nil {
		return err
	}
	//TODO auto change all fieds
	change := bson.M{}
	err = self.collection(collectionName).Update(doc, change)
	if err != nil {
		return err
	}
	err = runHook("PostUpdate", eptr, hookarg)
	if err != nil {
		return err
	}
	return nil
}

func (self *Dbm) Delete(doc interface{}) error {
	var err error
	var elem reflect.Value
	var collectionName string
	hookarg := hookArg(DbExecutor(self))
	elem, err = self.getPointer(doc)
	if err != nil {
		return err
	}
	eptr := elem.Addr()
	err = runHook("PreDelete", eptr, hookarg)
	if err != nil {
		return err
	}
	collectionName, err = getCollectionName(eptr, hookarg)
	if err != nil {
		return err
	}
	//TODO getId doc
	err = self.collection(collectionName).Remove(doc)
	if err != nil {
		return err
	}
	err = runHook("PostDelete", eptr, hookarg)
	if err != nil {
		return err
	}
	return nil
}

func (self *Dbm) UpdateAll(collectionName string, selector interface{}, change interface{}) (*mgo.ChangeInfo, error) {
	info, err := self.collection(collectionName).UpdateAll(selector, change)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (self *Dbm) DeleteAll(collectionName string, selector interface{}) (*mgo.ChangeInfo, error) {
	info, err := self.collection(collectionName).RemoveAll(selector)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (self *Dbm) collection(collectionName string) *mgo.Collection {
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

func hookArg(exec DbExecutor) []reflect.Value {
	execval := reflect.ValueOf(exec)
	return []reflect.Value{execval}
}

func runHook(name string, eptr reflect.Value, arg []reflect.Value) error {
	hook := eptr.MethodByName(name)
	if hook != zeroVal {
		ret := hook.Call(arg)
		if len(ret) > 0 && !ret[0].IsNil() {
			return ret[0].Interface().(error)
		}
	}
	return nil
}

func getCollectionName(eptr reflect.Value, arg []reflect.Value) (string, error) {
	fn := eptr.MethodByName("Collection")
	if fn != zeroVal {
		ret := fn.Call(arg)
		if len(ret) > 0 && ret[0].String() != "" {
			return ret[0].String(), nil
		}
	}
	return "", errors.New(fmt.Sprintf("get Collection:%s err.", eptr.Type().String()))
}
