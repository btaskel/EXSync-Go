package socket

//// SendFile 发送当前文件实例到远程
////
//// startSize : 起始文件大小
////
//// endSize : 终止文件大小
//func (s *Session) SendFile(f os.File, startSize, endSize int64) (err error) {
//	readSize := endSize - startSize
//	_, err = f.Seek(readSize, 0)
//	if err != nil {
//		return err
//	}
//
//	f.Read()
//
//	if s.aesGCM != nil {
//
//	}
//}
