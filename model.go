package mgodb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
)

type Query struct {
	QueryDoc interface{}
	Limit    int
	Skip     int
}

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

func (self *Model) FindAll(query Query, docs interface{}) (err error) {
	var mgoQuery *mgo.Query
	if mgoQuery, err = self.getQueryByFields(query.QueryDoc); err != nil {
		return err
	}
	if query.Skip > 0 {
		mgoQuery.Skip(query.Skip)
	}
	if query.Limit > 0 {
		mgoQuery.Limit(query.Limit)
	}
	return mgoQuery.All(docs)
}

func (self *Model) FindOne(queryDoc interface{}, doc interface{}) (err error) {
	var query *mgo.Query
	if query, err = self.getQueryByFields(queryDoc); err != nil {
		return err
	}
	return query.One(doc)
}

func (self *Model) FindByPk(id string, doc interface{}) (err error) {
	return self.FindOne(bson.M{"_id": bson.ObjectIdHex(id)}, doc)
}

func (self *Model) Find(query interface{}) (*mgo.Query, error) {
	if err := self.setValues(); err != nil {
		return nil, err
	}
	return DbmInstance.Find(self.collectionName, query), nil
}

func (self *Model) Count(queryDoc interface{}) (n int, err error) {
	var query *mgo.Query
	if query, err = self.getQueryByFields(queryDoc); err != nil {
		return 0, err
	}
	return query.Count()
}

func (self *Model) Save() (err error) {
	if err := self.setValues(); err != nil {
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

func (self *Model) Delete() (err error) {
	if err := self.setValues(); err != nil {
		return err
	}
	return DbmInstance.Delete(self.collectionName, self.docId, self.doc)
}

func (self *Model) getQueryByFields(queryDoc interface{}) (*mgo.Query, error) {
	if err := self.setValues(); err != nil {
		return nil, err
	}
	var query bson.M
	var err error
	if query, err = self.makeQuery(queryDoc); err != nil {
		return nil, err
	}
	return DbmInstance.Find(self.collectionName, query), nil
}

func (self *Model) makeQuery(doc interface{}) (bson.M, error) {
	var bsonData bson.M
	var tmpBlob []byte
	var err error
	if bsonDoc, ok := doc.(bson.M); ok {
		return bsonDoc, nil
	}
	if tmpBlob, err = bson.Marshal(doc); err != nil {
		return bsonData, err
	}
	if err = bson.Unmarshal(tmpBlob, &bsonData); err != nil {
		return bsonData, err
	}
	return bsonData, nil
}

func (self *Model) mergeResult(result interface{}, doc interface{}) (err error) {
	var tmpResult []byte
	if tmpResult, err = bson.Marshal(result); err != nil {
		return err
	}
	if err = bson.Unmarshal(tmpResult, doc); err != nil {
		return err
	}
	return nil
}

func (self *Model) setValues() (err error) {
	if self.collectionName == "" {
		if self.collectionName, err = self.getFromMtdName("GetCName"); err != nil {
			return err
		}
		if self.collectionName == "" {
			return errors.New("mdodb.Model: Collection name is empty.")
		}
	}
	if self.docId, err = self.getFromMtdName("GetId"); err != nil {
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
