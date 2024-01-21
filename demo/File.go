package main

//func main() {
//	file, err := os.Open(".\\demo\\test1.go")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer file.Close()
//
//	h := xxhash.New()
//	buf := make([]byte, 4096) // 4KB blocks
//	for {
//		n, err := file.Read(buf)
//		if err != nil && err != io.EOF {
//			log.Fatal(err)
//		}
//		if n == 0 {
//			break
//		}
//
//		if _, err := h.Write(buf[:n]); err != nil {
//			log.Fatal(err)
//		}
//	}
//
//	fmt.Printf("xxHash128: %x\n", h.Sum(nil))
//}
