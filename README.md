Mgodb
=====

Orm for mongodb, wrapper from http://labix.org/v2/mgo
Example usage in http://robfig.github.io/revel/ framework see example/init.revel.go and example/model.revel.go, support hooks,

    func (self *SomeModelName) PreInsert() error {}
    func (self *SomeModelName) PostInsert() error {}
    func (self *SomeModelName) PreUpdate() error {}
    func (self *SomeModelName) PostUpdate() error {}
    func (self *SomeModelName) PreDelete() error {}
    func (self *SomeModelName) PostDelete() error {}

Example usage default CRUD:

    var err error
    user := &models.User{}
    user.Username = "Bar"
    user.Password = "ssdf"
    err = user.Save() //Insert object
    if err != nil {
        fmt.Fatal(err.Error())
    }
    fmt.Println(user)
    user.Username = "Foo"
    err = user.Save() //Update object
    if err != nil {
        fmt.Fatal(err.Error())
    }
    err = user.Delete() //Delete object
    if err != nil {
        fmt.Fatal(err.Error())
    }
    fmt.Println(user)

