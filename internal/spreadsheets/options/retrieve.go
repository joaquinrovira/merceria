package options

type RetrieveOptionsData struct {
	MaxRows int64
	Raw     bool
}

func Retrieve(opts []RetrieveOptions) RetrieveOptionsData {
	var data RetrieveOptionsData
	for _, opts := range opts {
		data = opts(data)
	}
	return data
}

type RetrieveOptions = func(v RetrieveOptionsData) RetrieveOptionsData

func UpTo(n int64) func(v RetrieveOptionsData) RetrieveOptionsData {
	return func(o RetrieveOptionsData) RetrieveOptionsData {
		if o.MaxRows >= 0 {
			o.MaxRows = n
		}
		return o
	}
}

func Raw(v bool) func(v RetrieveOptionsData) RetrieveOptionsData {
	return func(o RetrieveOptionsData) RetrieveOptionsData {
		o.Raw = v
		return o
	}
}
