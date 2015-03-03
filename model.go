package mgodb

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

type Query struct {
	QueryDoc   interface{}
	Limit      int
	Skip       int
	sortFields []string
}

var (
	SortedDesc string = "-"
	SortedAsk  string = ""
)

// Example:
// 	query := mgodb.Query{QueryDoc: &models.User{}}
// 	query.SetSort(&models.User{Created: 0})
// 	users, err := models.FindUsers(query)
// http://godoc.org/labix.org/v2/mgo#Query.Sort
func (self *Query) SetSort(doc interface{}, sorted string) (err error) {
	bsonData, err := docToBson(doc)
	if err != nil {
		return
	}
	for key, _ := range bsonData {
		self.sortFields = append(self.sortFields, sorted+key)
	}
	return nil

}

type Model struct {
	doc            interface{}
	collectionName string
	docId          string
	isNew          bool
	//Crutch for success encoding gob
	Tmp *bool `bson:"-" json:"-" xml:"-"`
}

func (self *Model) SetDoc(doc interface{}) {
	self.isNew = true
	self.doc = doc
}

func (self *Model) ReloadDoc(doc interface{}) {
	self.isNew = false
	self.doc = doc
}

func (self *Model) MergeDoc(docOld interface{}, docNew interface{}) error {
	self.ReloadDoc(docOld)
	if err := self.setValues(); err != nil {
		return err
	}
	oldDocBson, err := docToBson(docOld)
	if err != nil {
		return err
	}
	newDocBson, err := docToBson(docNew)
	if err != nil {
		return err
	}
	for field, value := range newDocBson {
		if self.isEmpty(value) {
			continue
		}
		oldDocBson[field] = value
	}
	self.ReloadDoc(oldDocBson)
	return DbmInstance.GetCollection(self.collectionName).UpdateId(ObjectIdHex(self.docId), self.doc)
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
	if len(query.sortFields) > 0 {
		mgoQuery.Sort(query.sortFields...)
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
	return self.FindOne(bson.M{"_id": ObjectIdHex(id)}, doc)
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
	if reflect.ValueOf(queryDoc).IsNil() {
		queryDoc = self.doc
	}
	if query, err = docToBson(queryDoc); err != nil {
		return nil, err
	}
	return DbmInstance.Find(self.collectionName, query), nil
}

func docToBson(doc interface{}) (bsonData bson.M, err error) {
	if bsonData, ok := doc.(bson.M); ok {
		return bsonData, nil
	}
	var tmpBlob []byte
	if tmpBlob, err = bson.Marshal(doc); err != nil {
		return
	}
	if err = bson.Unmarshal(tmpBlob, &bsonData); err != nil {
		return
	}
	if Debug {
		fmt.Print("mgodb.Model debug:", fmt.Sprintf("%#v\n", bsonData))
	}
	return
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

func (self *Model) isEmpty(obj interface{}) bool {
	if obj == nil {
		return true
	}
	if i, ok := obj.(int); ok {
		return i == 0
	}
	return false
}
