package kthldap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/datasektionen/logout/pkg/config"
)

type Person struct {
	KTHID      string `json:"kthid"`
	UGKTHID    string `json:"ug_kthid"`
	FirstName  string `json:"first_name"`
	FamilyName string `json:"family_name"`
}

func Lookup(ctx context.Context, kthid string) (*Person, error) {
	if config.Config.LDAPProxyURL == nil {
		fakeName := strings.ToUpper(kthid[:1]) + kthid[1:]
		return &Person{
			KTHID:      kthid,
			UGKTHID:    "u1" + kthid,
			FirstName:  fakeName,
			FamilyName: fakeName + "sson",
		}, nil
	}

	resp, err := http.Get(config.Config.LDAPProxyURL.String() + "/user?kthid=" + url.QueryEscape(kthid))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected response from ldap proxy %d", resp.StatusCode)
	}
	var p Person
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("Invalid JSON from ldap proxy: %w", err)
	}
	return &p, nil
}
