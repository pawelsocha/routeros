package routeros

import (
	"testing"
	"time"
)

type Platform struct {
	Name string `routeros:"platform"`
	Arch string `routeros:"architecture-name"`
}

func TestFetchSingleRow(t *testing.T) {

	c, err := DialTimeout(*routerosAddress, *routerosUsername, *routerosPassword, time.Second)

	if err != nil {
		t.Fatalf("Connection error. Error: %s", err)
	}

	ret, err := c.RunArgs([]string{"/system/resource/print", "=.proplist=platform,architecture-name"})
	if err != nil {
		t.Fatalf("Ger resource command returns error. Error: %s", err)
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
