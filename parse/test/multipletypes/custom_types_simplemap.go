// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package multipletypes

type MyType1MyOtherTypeMap map[*MyType1]*MyOtherType

func (m MyType1MyOtherTypeMap) Has(key *MyType1) bool {
	_, ok := m[key]
	return ok
}

func (m MyType1MyOtherTypeMap) Get(key *MyType1) *MyOtherType {
	return m[key]
}

func (m MyType1MyOtherTypeMap) Set(key *MyType1, value *MyOtherType) MyType1MyOtherTypeMap {
	m[key] = value
	return m
}
