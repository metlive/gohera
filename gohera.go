package gohera

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var (
	httpHost string
	httpPort int
)

func StartHttpServer() error {

	httpHost = GetString("http.host")
	httpPort = GetInt("http.port")
	if httpPort == 0 {
		handleError(errors.New("http host or port is not valid"))
	}

	fmt.Println("start on:" + "https://" + httpHost + ":" + strconv.Itoa(httpPort))

	handleError(Engine.Run(httpHost + ":" + strconv.Itoa(httpPort)))
	return nil
}

func handleError(err error) {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return
	}

	Error(context.Background(), err, nil)
	panic(err)
}
