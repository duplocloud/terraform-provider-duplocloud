package duplocloud

import (
	"encoding/json"
	"reflect"
	"sort"
)

type PatchOperations []PatchOperation

func (po PatchOperations) MarshalJSON() ([]byte, error) {
	var v []PatchOperation = po
	return json.Marshal(v)
}

func (po PatchOperations) Equal(ops []PatchOperation) bool {
	var v []PatchOperation = po

	sort.Slice(v, sortByPathAsc(v))
	sort.Slice(ops, sortByPathAsc(ops))

	return reflect.DeepEqual(v, ops)
}

func sortByPathAsc(ops []PatchOperation) func(i, j int) bool {
	return func(i, j int) bool {
		return ops[i].GetPath() < ops[j].GetPath()
	}
}

type PatchOperation interface {
	MarshalJSON() ([]byte, error)
	GetPath() string
}

type ReplaceOperation struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Op    string      `json:"op"`
}

func (o *ReplaceOperation) GetPath() string {
	return o.Path
}

func (o *ReplaceOperation) MarshalJSON() ([]byte, error) {
	o.Op = "replace"
	return json.Marshal(*o)
}

func (o *ReplaceOperation) String() string {
	b, _ := o.MarshalJSON()
	return string(b)
}

type AddOperation struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Op    string      `json:"op"`
}

func (o *AddOperation) GetPath() string {
	return o.Path
}

func (o *AddOperation) MarshalJSON() ([]byte, error) {
	o.Op = "add"
	return json.Marshal(*o)
}

func (o *AddOperation) String() string {
	b, _ := o.MarshalJSON()
	return string(b)
}

type RemoveOperation struct {
	Path string `json:"path"`
	Op   string `json:"op"`
}

func (o *RemoveOperation) GetPath() string {
	return o.Path
}

func (o *RemoveOperation) MarshalJSON() ([]byte, error) {
	o.Op = "remove"
	return json.Marshal(*o)
}

func (o *RemoveOperation) String() string {
	b, _ := o.MarshalJSON()
	return string(b)
}
