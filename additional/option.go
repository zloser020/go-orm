package additional

type MyStructOption func(*MyStruct)
type MyStruct struct {
	// 必传
	id   string
	name string
	// 可选
	address string
}

func WithMyStructAddress(address string) MyStructOption {
	return func(ms *MyStruct) {
		ms.address = address
	}
}

func NewMyStruct(id string, name string, opts ...MyStructOption) *MyStruct {
	res := &MyStruct{
		id:   id,
		name: name,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}
