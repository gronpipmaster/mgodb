package mgodb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
)

type Model struct {
	doc            interface{}
	collectionName string
	docId          string
	isNew          bool
}

func (self *Model) SetDoc(doc interface{}) {
	self.isNew = true
	self.doc = doc
}

func (self *Model) ReloadDoc(doc interface{}) {
	self.isNew = false
	self.doc = doc
}

func (self *Model) FindByPk(id string, doc interface{}) error {
	var err error
	if err := self.setValues(); err != nil {
		return err
	}
	var result interface{}
	err = DbmInstance.Find(self.collectionName, bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err != nil {
		return err
	}
	var tmpResult []byte
	tmpResult, err = bson.Marshal(result)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(tmpResult, doc) //merge result from doc
	if err != nil {
		return err
	}
	return nil
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
	if !self.isNew {
		return DbmInstance.Update(self.collectionName, self.docId, self.doc)
	} else {
		err = DbmInstance.Insert(self.collectionName, self.doc)
		if err != nil {
			return err
		}
		self.isNew = false
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
	if self.collectionName == "" {
		self.collectionName, err = self.getFromMtdName("GetCName")
		if err != nil {
			return err
		}
		if self.collectionName == "" {
			return errors.New("mdodb.Model: Collection name is empty.")
		}
	}
	self.docId, err = self.getFromMtdName("GetId")
	if err != nil {
		return err
	}
	return nil
}

func (self *Model) getFromMtdName(method string) (string, error) {
	docV := reflect.ValueOf(self.doc)
	if docV.Kind() != reflect.Ptr {
		e := fmt.Sprintf("mgodb.Model: Passed non-pointer: %v (kind=%v), method:%s", self.doc, docV.Kind(), method)
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
