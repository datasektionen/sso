package config

import (
	"net/url"
	"os"
	"strconv"
)

type Cfg struct {
	KTHOIDCIssuerURL      *url.URL
	KTHOIDCClientID       string
	KTHOIDCClientSecret   string
	KTHOIDCRPOrigin       *url.URL
	Origin                *url.URL
	Port                  int
	DatabaseURL           *url.URL
	Dev                   bool
	LDAPProxyURL          *url.URL
	OIDCProviderIssuerURL *url.URL
	OIDCProviderKey       string
	PlsURL                *url.URL
	SpamURL               *url.URL
	SpamAPIKey            string
}

var Config Cfg

func init() {
	Config = Cfg{
		KTHOIDCIssuerURL:      getURL("KTH_ISSUER_URL", false),
		KTHOIDCClientID:       os.Getenv("KTH_CLIENT_ID"),
		KTHOIDCClientSecret:   os.Getenv("KTH_CLIENT_SECRET"),
		KTHOIDCRPOrigin:       getOrigin("KTH_RP_ORIGIN", false),
		Origin:                getOrigin("ORIGIN", false),
		Port:                  getInt("PORT", 7000),
		DatabaseURL:           getURL("DATABASE_URL", false),
		Dev:                   os.Getenv("DEV") == "true",
		LDAPProxyURL:          getURL("LDAP_PROXY_URL", true),
		OIDCProviderIssuerURL: getURL("OIDC_PROVIDER_ISSUER_URL", false),
		OIDCProviderKey:       os.Getenv("OIDC_PROVIDER_KEY"),
		PlsURL:                getURL("PLS_URL", true),
		SpamURL:               getURL("SPAM_URL", true),
		SpamAPIKey:            os.Getenv("SPAM_API_KEY"),
	}
}

func getURL(envName string, optional bool) *url.URL {
	envValue, ok := os.LookupEnv(envName)
	if !ok {
		if !optional {
			panic("Missing $" + envName)
		}
		return nil
	}
	u, err := url.Parse(envValue)
	if err != nil {
		panic(err)
	}
	return u
}

func getOrigin(envName string, optional bool) *url.URL {
	u := getURL(envName, optional)
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
