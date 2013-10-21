Mgodb
=====

Orm for mongodb, wrapper from http://labix.org/v2/mgo
Example usage in http://robfig.github.io/revel/ framework see example/init.revel.go and example/model.revel.go, support hooks concept from https://github.com/coopernurse/gorp

### CRUD ###

Example usage default CRUD:
```go
var err error
user := models.NewUser()
user.Username = "Bar"
user.Password = "ssdf"
err = user.Save() //Insert object
if err != nil {
    panic(err)
}
fmt.Println(user)
user.Username = "Foo"
err = user.Save() //Update object
if err != nil {
    panic(err)
}
err = user.Delete() //Delete object
if err != nil {
    panic(err)
}
fmt.Println(user)
```

### Hooks ###

Use hooks to before/after saving/delete to the db.
```go
//Full list of hooks that you can implement:
func (self *SomeModelName) BeforeInsert() error {}
func (self *SomeModelName) AfterInsert() error {}

func (self *SomeModelName) BeforeUpdate() error {}
func (self *SomeModelName) AfterUpdate() error {}

func (self *SomeModelName) BeforeDelete() error {}
func (self *SomeModelName) AfterDelete() error {}
```