package routeros

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pawelsocha/routeros/proto"
)

type asyncReply struct {
	chanReply
	Reply
}

// Run simply calls RunArgs().
func (c *Client) Run(sentence ...string) (*Reply, error) {
	return c.RunArgs(sentence)
}

// RunArgs sends a sentence to the RouterOS device and waits for the reply.
func (c *Client) RunArgs(sentence []string) (*Reply, error) {
	c.w.BeginSentence()
	for _, word := range sentence {
		c.w.WriteWord(word)
	}
	if !c.async {
		return c.endCommandSync()
	}
	a, err := c.endCommandAsync()
	if err != nil {
		return nil, err
	}
	for range a.reC {
	}
	return &a.Reply, a.err
}

func (c *Client) endCommandSync() (*Reply, error) {
	err := c.w.EndSentence()

	if err != nil {
		return nil, err
	}
	return c.readReply()
}

func (c *Client) endCommandAsync() (*asyncReply, error) {
	c.nextTag++
	a := &asyncReply{}
	a.reC = make(chan *proto.Sentence)
	a.tag = fmt.Sprintf("r%d", c.nextTag)
	c.w.WriteWord(".tag=" + a.tag)

	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.w.EndSentence()
	if err != nil {
		return nil, err
	}
	if c.tags == nil {
		return nil, errAsyncLoopEnded
	}
	c.tags[a.tag] = a
	return a, nil
}

func (c *Client) proplist(obj interface{}) string {
	var proplist []string
	var typ reflect.Type

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		typ = reflect.ValueOf(obj).Elem().Type()
	} else {
		typ = reflect.TypeOf(obj)
	}

	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i).Tag.Get("routeros")
		if p == "" {
			continue
		}

		proplist = append(proplist, p)
	}

	return strings.Join(proplist, ",")
}

func (c *Client) valuelist(obj interface{}) []string {
	var values []string
	elem := reflect.ValueOf(obj)
	typ := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		p := elem.Field(i)
		switch p.Type().Name() {
		case "string":
			if p.Interface() != "" {
				values = append(
					values,
					fmt.Sprintf("=%s=%s", typ.Field(i).Tag.Get("routeros"), p.Interface()),
				)
			}
		}
	}
	return values
}

//Print print data from specific path
func (c *Client) Print(i Entity) error {
	sentence := []string{
		fmt.Sprintf("%s/print", i.Path()),
	}

	where := i.Where()
	if where != "" {
		sentence = append(sentence, where)
	}

	plist := c.proplist(i)
	if plist != "" {
		sentence = append(sentence, "=.proplist="+plist)
	}

	ret, err := c.RunArgs(sentence)
	if err != nil {
		return err
	}

	err = ret.Fetch(i)
	return err
}

//Add create new entity
func (c *Client) Add(i Entity) error {
	sentence := []string{
		fmt.Sprintf("%s/add", i.Path()),
	}
	sentence = append(sentence, c.valuelist(i)...)

	ret, err := c.RunArgs(sentence)
	return err
}

//Remove remove object from routeros
func (c *Client) Remove(i Entity) error {

	id := i.GetId()
	if id == "" {
		return fmt.Errorf("Id is empty.\n")
	}

	sentence := []string{
		fmt.Sprintf("%s/remove", i.Path()),
		fmt.Sprintf("=.id=%s", id),
	}

	_, err := c.RunArgs(sentence)
	return err
}
