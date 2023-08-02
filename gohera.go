package gohera

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

	fmt.Println("start on:" + "http://" + httpHost + ":" + strconv.Itoa(httpPort))
	fmt.Printf("服务启动，运行模式：%v，版本号：%s，进程号：%d", GetEnv, "1.0.0", os.Getpid())

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
