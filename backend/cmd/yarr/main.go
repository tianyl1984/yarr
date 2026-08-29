package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nkanaev/yarr/src/server"
	"github.com/nkanaev/yarr/src/storage"
)

var OptList = make([]string, 0)

func opt(envVar, defaultValue string) string {
	OptList = append(OptList, envVar)
	value := os.Getenv(envVar)
	if value != "" {
		return value
	}
	return defaultValue
}

func randomSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatal("Failed to generate session secret: ", err)
	}
	return hex.EncodeToString(buf)
}

func main() {

	var addr, db, authurl, authsecret string

	flag.CommandLine.SetOutput(os.Stdout)

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(out, "\nThe environmental variables, if present, will be used to provide\nthe default values for the params above:")
		fmt.Fprintln(out, " ", strings.Join(OptList, ", "))
		fmt.Fprintln(out, "\nEnvironment-only settings (no flag):")
		fmt.Fprintln(out, "  YARR_PROXY\n\toutbound proxy, applied only to feeds with use_proxy enabled")
		fmt.Fprintln(out, "  YARR_BROWSERLESS\n\tbrowserless endpoint for scraping JS-rendered pages")
		fmt.Fprintln(out, "  TZ\n\ttime zone deciding when the daily 04:00 feed refresh fires")
	}

	flag.StringVar(&addr, "addr", opt("YARR_ADDR", "127.0.0.1:7070"), "address to run server on")
	flag.StringVar(&authurl, "auth-url", opt("YARR_AUTH_URL", ""), "base `url` of the cf-worker-auth SSO service (enables login when set)")
	flag.StringVar(&authsecret, "auth-secret", opt("YARR_AUTH_SECRET", ""), "secret used to sign session cookies (random per start if unset)")
	flag.StringVar(&db, "db", opt("YARR_DB", ""), "mysql connection string")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	if db == "" {
		log.Fatal("Failed to get db config")
	}

	store, err := storage.New(db)
	if err != nil {
		log.Fatal("Failed to initialise database: ", err)
	}

	srv := server.NewServer(store, addr)

	if authurl != "" {
		srv.AuthURL = strings.TrimRight(authurl, "/")
		if authsecret == "" {
			authsecret = randomSecret()
			log.Print("YARR_AUTH_SECRET not set: using a random session secret (sessions reset on restart)")
		}
		srv.AuthSecret = authsecret
	}

	log.Printf("starting server at %s", srv.GetAddr())
	srv.Start()
}
