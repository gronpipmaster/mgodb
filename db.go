package system

import (
	"errors"
	"fmt"
	"github.com/robfig/revel"
	"labix.org/v2/mgo"
	"net/url"
	"reflect"
	"time"
)

var zeroVal reflect.Value

type DbExecutor interface {
	Find(collectionName string, query interface{}) *mgo.Query
	Insert(collectionName string, docs ...interface{}) error
	Update(collectionName string, change interface{}, docs ...interface{}) error
	Delete(collectionName string, docs ...interface{}) error
}

type MongoDb struct {
	Database *mgo.Database
}

func (self *MongoDb) Init() *MongoDb {
	self.connect()
	return self
}

func (self *MongoDb) Find(collectionName string, query interface{}) *mgo.Query {
	return self.collection(collectionName).Find(query)
}

func (self *MongoDb) Insert(collectionName string, docs ...interface{}) error {
	var err error
	var elem reflect.Value
	hookarg := hookArg(DbExecutor(self))
	for _, doc := range docs {
		elem, err = self.getPointer(doc)
		if err != nil {
			return err
		}
		eptr := elem.Addr()
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

func (self *MongoDb) Update(collectionName string, change interface{}, docs ...interface{}) error {
	var err error
	var elem reflect.Value
	hookarg := hookArg(DbExecutor(self))
	for _, doc := range docs {
		elem, err = self.getPointer(doc)
		if err != nil {
			return err
		}
		eptr := elem.Addr()
		err = runHook("PreUpdate", eptr, hookarg)
		if err != nil {
			return err
		}
		_, err = self.collection(collectionName).UpdateAll(doc, change)
		if err != nil {
			return err
		}
		err = runHook("PostUpdate", eptr, hookarg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *MongoDb) Delete(collectionName string, docs ...interface{}) error {
	var err error
	var elem reflect.Value
	hookarg := hookArg(DbExecutor(self))
	for _, doc := range docs {
		elem, err = self.getPointer(doc)
		if err != nil {
			return err
		}
		eptr := elem.Addr()
		err = runHook("PreDelete", eptr, hookarg)
		if err != nil {
			return err
		}
		_, err = self.collection(collectionName).RemoveAll(doc)
		if err != nil {
			return err
		}
		err = runHook("PostDelete", eptr, hookarg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *MongoDb) collection(collectionName string) *mgo.Collection {
	return self.Database.C(collectionName)
}

func (self *MongoDb) connect() {
	var err error
	var found bool
	var host, port, user, pass, dbName, timeoutStr string
	var timeout time.Duration
	var session *mgo.Session

	connURL := &url.URL{Scheme: "mongodb"}

	if host, found = revel.Config.String("mongo.host"); !found {
		revel.ERROR.Fatal("No mongo.host found.")
	}

	if port, found = revel.Config.String("mongo.port"); !found {
		revel.ERROR.Fatal("No mongo.port found.")
	}

	if user, found = revel.Config.String("mongo.user"); !found {
		revel.ERROR.Fatal("No mongo.user found.")
	}

	if pass, found = revel.Config.String("mongo.pass"); !found {
		revel.ERROR.Fatal("No mongo.pass found.")
	}

	if dbName, found = revel.Config.String("mongo.db"); !found {
		revel.ERROR.Fatal("No mongo.db found.")
	}

	if timeoutStr, found = revel.Config.String("mongo.timeout"); !found {
		timeoutStr = "5s"
	}

	timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		revel.ERROR.Fatal("mongo.timeout found error, ", err.Error())
	}

	connURL.Host = fmt.Sprintf("%s:%s", host, port)
	connURL.User = url.UserPassword(user, pass)
	connURL.Path = "/" + dbName

	session, err = mgo.DialWithTimeout(connURL.String(), timeout*time.Second)
	session.SetMode(mgo.Monotonic, true)

	if err != nil {
		revel.ERROR.Fatalf("Could not connect to %s: %s.", host, err.Error())
	}
	self.Database = session.DB(dbName)
}

func (self *MongoDb) getPointer(doc interface{}) (reflect.Value, error) {
	docV := reflect.ValueOf(doc)
	if docV.Kind() != reflect.Ptr {
		e := fmt.Sprintf("bv.db: passed non-pointer: %v (kind=%v)", doc,
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
