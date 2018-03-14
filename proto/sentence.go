package proto

import (
	"fmt"
	"reflect"
)

// Sentence is a line read from a RouterOS device.
type Sentence struct {
	// Word that begins with !
	Word string
	Tag  string
	List []Pair
	Map  map[string]string
}

type Pair struct {
	Key, Value string
}

func NewSentence() *Sentence {
	return &Sentence{
		Map: make(map[string]string),
	}
}

func (sen *Sentence) String() string {
	return fmt.Sprintf("%s @%s %#q", sen.Word, sen.Tag, sen.List)
}

//Unmarshal map data to struct
func (r *Sentence) Unmarshal(out interface{}) error {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("Out variable is not a pointer. Type: %v", reflect.TypeOf(out).Kind())
	}

	value := reflect.ValueOf(out).Elem()
	switch value.Kind() {
	case reflect.Struct:
		typ := value.Type()
		for i := 0; i < value.NumField(); i++ {

			tag := typ.Field(i).Tag.Get("routeros")

			if tag == "" {
				continue
			}

			if data, ok := r.Map[tag]; ok {
				value.FieldByName(typ.Field(i).Name).SetString(data)
			}
		}
	default:
		return fmt.Errorf("Invalid input data.")
	}
	return nil
}
