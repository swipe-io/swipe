package option

type SliceStringValue struct {
	Value []string
}

type StringValue struct {
	Value *string
}

func (v StringValue) IsValid() bool {
	return v.Value != nil
}

func (v StringValue) Take() string {
	if v.Value == nil {
		return ""
	}
	return *v.Value
}

type IntValue struct {
	Value *int
}

func (v IntValue) IsValid() bool {
	return v.Value != nil
}

func (v IntValue) Take() int {
	if v.Value == nil {
		return 0
	}
	return *v.Value
}

type Int64Value struct {
	Value *int64
}

func (v Int64Value) IsValid() bool {
	return v.Value != nil
}

func (v Int64Value) Take() int64 {
	if v.Value == nil {
		return 0
	}
	return *v.Value
}

type BoolValue struct {
	Value *bool
}

func (v BoolValue) IsValid() bool {
	return v.Value != nil
}

func (v BoolValue) Take() bool {
	if v.Value == nil {
		return false
	}
	return *v.Value
}
