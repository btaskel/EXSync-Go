package protocol

//func Server(address string) (listener *quic.Listener, err error) {
//	listener, err = quic.ListenAddr(address, &tls.Config{InsecureSkipVerify: true}, nil)
//	if err != nil {
//		return nil, err
//	}
//	return listener, nil
//}
//
//func generateTLSConfig() *tls.Config {
//	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
//	if err != nil {
//		panic(err)
//	}
//
//	config := &tls.Config{Certificates: []tls.Certificate{cert}}
//	config.NextProtos = append(config.NextProtos, quic.Version1Draft29)
//
//	return config
//}
