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

type User struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Username     string        `bson:"u"`
	Password     string        `bson:"p,omitempty"`
	PasswordHash []byte        `bson:"ph,omitempty"`
	Email        string        `bson:"e,omitempty"`
	Adm          bool          `bson:"a,omitempty"`
	Created      int64         `bson:"c,omitempty"`
}

func (self *User) Collection(_ mgodb.DbExecutor) string {
	return "users"
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

func (self *User) PreInsert(_ mgodb.DbExecutor) error {
	var err error
	self.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(self.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	self.Password = ""
	self.Created = time.Now().Unix()
	return nil
}

func (self *User) PreUpdate(_ mgodb.DbExecutor) error {
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
