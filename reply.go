package routeros

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/pawelsocha/routeros/proto"
)

// Reply has all the sentences from a reply.
type Reply struct {
	Re   []*proto.Sentence
	Done *proto.Sentence
}

func (r *Reply) String() string {
	b := &bytes.Buffer{}
	for _, re := range r.Re {
		fmt.Fprintf(b, "%s\n", re)
	}
	fmt.Fprintf(b, "%s", r.Done)
	return b.String()
}

// readReply reads one reply synchronously. It returns the reply.
func (c *Client) readReply() (*Reply, error) {
	r := &Reply{}
	for {
		sen, err := c.r.ReadSentence()
		if err != nil {
			return nil, err
		}
		done, err := r.processSentence(sen)
		if err != nil {
			return nil, err
		}
		if done {
			return r, nil
		}
	}
}

func (r *Reply) processSentence(sen *proto.Sentence) (bool, error) {
	switch sen.Word {
	case "!re":
		r.Re = append(r.Re, sen)
	case "!done":
		r.Done = sen
		return true, nil
	case "!trap", "!fatal":
		return true, &DeviceError{sen}
	case "":
		// API docs say that empty sentences should be ignored
	default:
		return true, &UnknownReplyError{sen}
	}
	return false, nil
}

func (r *Reply) fillRow(record *proto.Sentence, value reflect.Value) reflect.Value {

	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {

		tag := typ.Field(i).Tag.Get("routeros")

		if tag == "" {
			continue
		}

		if data, ok := record.Map[tag]; ok {
			value.FieldByName(typ.Field(i).Name).SetString(data)
		}
	}
	return value
}

func Generator(r *Reply) chan *proto.Sentence {
	ret := make(chan *proto.Sentence, len(r.Re))
	defer close(ret)

	for _, data := range r.Re {
		ret <- data
	}

	return ret
}

//Fetch map data to struct
func (r *Reply) Fetch(out interface{}) error {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("Out variable is not a pointer. Type: %v", reflect.TypeOf(out).Kind())
	}

	if len(r.Re) < 1 {
		return fmt.Errorf("Empty data returned from routeros")
	}

	value := reflect.ValueOf(out).Elem()
	switch value.Kind() {
	case reflect.Struct:
		if len(r.Re) > 1 {
			return fmt.Errorf("Too many records returned from routeros")
		}

		value.Set(r.fillRow(r.Re[0], value))
	case reflect.Slice:
		newobj := reflect.New(value.Type())
		for _, data := range r.Re {
			value.Set(reflect.Append(value, r.fillRow(data, newobj)))
		}
	}
	return nil
}
