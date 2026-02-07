package main

import (
	"fmt"
	"net/http"

	"github.com/Algatux/yagaw"
	"github.com/pho3b/tiny-logger/logs/log_level"
)

func main() {
	yagaw.Log = yagaw.InitLogger(log_level.DebugLvlName)
	server := yagaw.NewServer("", 8080)

	router := server.GetRouter()

	router.RegisterRoute(yagaw.GET, "/test/{a}", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(rw, "Welcome to our custom HTTP server!")
	})

	yagaw.Log.Debug(router.RegisteredRoutes())

	server.Run()
}
