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

//Construct and implemend mgodb.Model
func NewUser() *User {
	user := new(User)
	user.Id = bson.NewObjectId()
	user.SetDoc(user)
	return user
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
