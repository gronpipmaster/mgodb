package mgodb

import "labix.org/v2/mgo/bson"

type ModelInterface interface {
	CollectionName() string
	GetId() bson.ObjectId
	Save() error
	Delete() error
}

type Model struct {
}

func (self *Model) Save(docInface ModelInterface, doc interface{}) error {
	if docInface.GetId() == "" {
		return DbmInstance.Insert(doc)
	} else {
		return DbmInstance.Update(docInface.GetId(), doc)
	}
}

func (self *Model) Delete(docInface ModelInterface, doc interface{}) error {
	return DbmInstance.Delete(docInface.GetId(), doc)
}
