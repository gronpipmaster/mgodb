package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"github.com/gronpipmaster/mgodb"
	"github.com/robfig/revel"
	"labix.org/v2/mgo/bson"
	"regexp"
	"time"
)

//System methods all required

//Get collection name
func (self *User) GetCName() string {
	return "users"
}

//Get document id
func (self *User) GetId() string {
	return self.Id.Hex()
}

//get count
func CountUsers(findUsers *User) (n int, err error) {
	model := new(User)
	model.SetDoc(model)
	if n, err = model.Count(findUsers); err != nil {
		return 0, err
	}
	return n, nil
}

//load and construct models by any fields
func FindUsers(query mgodb.Query) ([]*User, error) {
	model := new(User)
	model.SetDoc(model)
	var users []*User
	if err := model.FindAll(query, &users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	for key, item := range users {
		item.ReloadDoc(item)
		users[key] = item
	}
	return users, nil
}

//load and construct model by any fields
func FindUserBy(findUser *User) (*User, error) {
	model := new(User)
	model.SetDoc(model)
	if err := model.FindOne(findUser, &model); err != nil {
		return nil, err
	}
	model.ReloadDoc(model)
	return model, nil
}

//load and construct model by id
func FindUser(id string) (*User, error) {
	model := new(User)
	model.SetDoc(model)
	if err := model.FindByPk(id, &model); err != nil {
		return nil, err
	}
	model.ReloadDoc(model)
	return model, nil
}

//Construct and implemend mgodb.Model
func NewUser() *User {
	model := new(User)
	model.Id = bson.NewObjectId()
	model.SetDoc(model)
	return model
}

//End system methods

type User struct {
	//Required
	mgodb.Model `,inline`
	//Required
	Id           bson.ObjectId `bson:"_id,omitempty" 	json:"Id,omitempty"`
	Username     string        `bson:"u,omitempty"		json:"Username,omitempty"`
	Password     string        `bson:"p,omitempty" 		json:"-,omitempty"`
	PasswordHash []byte        `bson:"ph,omitempty" 	json:"-,omitempty"`
	Email        string        `bson:"e,omitempty" 		json:"Email,omitempty"`
	Adm          bool          `bson:"a,omitempty" 		json:"Adm,omitempty"`
	Created      int64         `bson:"c,omitempty" 		json:"Created,omitempty"`
}

func (self *User) String() string {
	return fmt.Sprintf("User(%s)", self.Username)
}

func (self *User) Validate(v *revel.Validation) {
	userRegex := regexp.MustCompile("^\\w*$")
	v.Check(self.Username,
		revel.Required{},
		revel.MaxSize{15},
		revel.MinSize{3},
		revel.Match{userRegex},
	)
	v.Check(self.Email,
		revel.Required{},
		revel.Email{},
	)
	ValidatePassword(v, self.Password).
		Key("user.Password")
}

func ValidatePassword(v *revel.Validation, password string) *revel.ValidationResult {
	return v.Check(password,
		revel.Required{},
		revel.MaxSize{15},
		revel.MinSize{3},
	)
}

func (self *User) BeforeInsert() error {
	var err error
	self.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(self.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	self.Password = ""
	self.Created = time.Now().Unix()
	return nil
}

func (self *User) BeforeUpdate() error {
	var err error
	if self.Password == "" {
		return nil
	}
	self.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(self.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	self.Password = ""
	return nil
}
