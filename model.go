package mgodb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"reflect"
)

type Model struct {
	doc            interface{}
	collectionName string
	docId          string
	insert         bool
}

func (self *Model) SetDoc(doc interface{}) {
	self.doc = doc
	self.insert = false
}

func (self *Model) Find(query interface{}) (*mgo.Query, error) {
	if err := self.setValues(); err != nil {
		return nil, err
	}
	return DbmInstance.Find(self.collectionName, query), nil
}

func (self *Model) Save() error {
	var err error
	if err = self.setValues(); err != nil {
		return err
	}
	if self.insert {
		return DbmInstance.Update(self.collectionName, self.docId, self.doc)
	} else {
		err = DbmInstance.Insert(self.collectionName, self.doc)
		if err != nil {
			return err
		}
		self.insert = true
		return nil
	}
}

func (self *Model) Delete() error {
	if err := self.setValues(); err != nil {
		return err
	}
	return DbmInstance.Delete(self.collectionName, self.docId, self.doc)
}

func (self *Model) setValues() error {
	var err error
	self.collectionName, err = self.getFromMtdName("GetCName")
	if err != nil {
		return err
	}
	if self.collectionName == "" {
		return errors.New("mdodb.Model: Collection name is empty.")
	}
	self.docId, err = self.getFromMtdName("GetId")
	if err != nil {
		return err
	}
	if self.docId == "" {
		return errors.New("mdodb.Model: Document id is empty.")
	}
	return nil
}

func (self *Model) getFromMtdName(method string) (string, error) {
	docV := reflect.ValueOf(self.doc)
	if docV.Kind() != reflect.Ptr {
		e := fmt.Sprintf("mgodb.Model: Passed non-pointer: %v (kind=%v)", self.doc,
			docV.Kind())
		return "", errors.New(e)
	}
	fn := docV.Elem().Addr().MethodByName(method)
	if fn != zeroVal {
		ret := fn.Call(zeroArgs)
		if len(ret) > 0 && ret[0].String() != "" {
			return ret[0].String(), nil
		}
	}
	return "", nil
}
