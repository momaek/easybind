# easyreq
Bind req arguments easily in Golang.
Support Tag `pos`, specified that where we can get this value, only support one
- path: from url path, don't support nested struct
- query: from url query, don't support nested struct
- body: from request's body, default use json, support nested struct
- form: from request form
- required: this value is not null
pathQueryier get variables from path, GET /api/v1/users/:id , get id

```go
type Example struct {
	ID   string `json:"id"   pos:"path:id"`             // path value default is required
	Name string `json:"name" pos:"query:name,required"` // query specified that get
}
```

### Get Started

```
go get github.com/momaek/easybind
```

### Example

please check [bind\_test.go](bind_test.go)
