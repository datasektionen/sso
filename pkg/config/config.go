package config

import (
	"net/url"
	"os"
	"strconv"
)

type Cfg struct {
	KTHOIDCIssuerURL    *url.URL
	KTHOIDCClientID     string
	KTHOIDCClientSecret string
	KTHOIDCRPOrigin     *url.URL
	Origin              *url.URL
	Port                int
	DatabaseURL         *url.URL
}

var Config Cfg

func init() {
	Config = Cfg{
		KTHOIDCIssuerURL:    getURL("KTH_ISSUER_URL"),
		KTHOIDCClientID:     os.Getenv("KTH_CLIENT_ID"),
		KTHOIDCClientSecret: os.Getenv("KTH_CLIENT_SECRET"),
		KTHOIDCRPOrigin:     getOrigin("KTH_RP_ORIGIN"),
		Origin:              getOrigin("ORIGIN"),
		Port:                getInt("PORT", 3000),
		DatabaseURL:         getURL("DATABASE_URL"),
	}
}

func getURL(envName string) *url.URL {
	u, err := url.Parse(os.Getenv(envName))
	if err != nil {
		panic(err)
	}
	return u
}

func getOrigin(envName string) *url.URL {
	u := getURL(envName)
	withoutSchemeHost := *u
	withoutSchemeHost.Scheme = ""
	withoutSchemeHost.Host = ""
	if withoutSchemeHost != (url.URL{}) {
		panic("Only scheme and host (=hostname and port) may be set on $" + envName)
	}
	return u
}

func getInt(envName string, defaultValue int) int {
	s, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
