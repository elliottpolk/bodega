package config

import "fmt"

type Server struct {
	RpcPort   string
	HttpPort  string
	HttpsPort string
	TlsCert   string
	TlsKey    string
}

type Db struct {
	Addr     string
	Port     string
	Name     string
	User     string
	Password string
}

type Composition struct {
	Server *Server
	Db     *Db
}

func (db *Db) ConnString() string {
	uri := "mongodb://"
	if len(db.User) > 0 && len(db.Password) > 0 {
		uri = fmt.Sprintf("%s%s:%s@", uri, db.User, db.Password)
	}

	if addr := db.Addr; len(addr) > 0 {
		uri = fmt.Sprintf("%s%s", uri, addr)
	}

	if port := db.Port; len(port) > 0 {
		uri = fmt.Sprintf("%s:%s", uri, port)
	}

	return uri
}
