package routeros

import (
	"testing"
	"time"
)

type Platform struct {
	Name string `routeros:"platform"`
	Arch string `routeros:"architecture-name"`
}

type Resource struct {
	Platform string `routeros:"platform"`
}

func (q Resource) Path() string {
	return "/system/resource/"
}

func (q Resource) Where() string {
	return ""
}

func (q Resource) GetId() string {
	return ""
}

func TestFetchSingleRow(t *testing.T) {

	c, err := DialTimeout(*routerosAddress, *routerosUsername, *routerosPassword, time.Second)

	if err != nil {
		t.Fatalf("Connection error. Error: %s", err)
	}

	ret, err := c.RunArgs([]string{"/system/resource/print", "=.proplist=platform,architecture-name"})
	if err != nil {
		t.Fatalf("Get resource command returns error. Error: %s", err)
	}
	platform := Platform{}
	ret.Fetch(&platform)

	if platform.Name == "" {
		t.Fatalf("Name should be not empty. Got: %s", platform.Name)
	}

	if platform.Arch == "" {
		t.Fatalf("Arch should be not empty. Got: %s", platform.Arch)
	}

	if platform.Name != "MikroTik" {
		t.Fatalf("Name should have MikroTik string. Got: %s", platform.Name)
	}

	defer c.Close()
}

func TestFetchInterface(t *testing.T) {

	c, err := DialTimeout(*routerosAddress, *routerosUsername, *routerosPassword, time.Second)

	if err != nil {
		t.Fatalf("Connection error. Error: %s", err)
	}
	r := Resource{}
	err = c.Print(&r)

	if err != nil {
		t.Fatalf("Print commant error. Error: %s", err)
	}

	if r.Platform != "MikroTik" {
		t.Fatalf("Platform should have MikroTik string. Got: %s", r.Platform)
	}
	defer c.Close()
}
