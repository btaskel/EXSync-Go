package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Log struct {
		LogLevel string `json:"loglevel"`
	} `json:"log"`
	Server struct {
		Addr struct {
			ID       string `json:"id"`
			IP       string `json:"ip"`
			Port     int    `json:"port"`
			Password string `json:"password"`
		} `json:"addr"`
		Setting struct {
			Encode     string `json:"encode"`
			IOBalance  bool   `json:"iobalance"`
			Encryption string `json:"encryption"`
		} `json:"setting"`
		Scan struct {
			Enabled bool     `json:"enabled"`
			Type    string   `json:"type"`
			Max     int      `json:"max"`
			Devices []string `json:"devices"`
		} `json:"scan"`
		Plugin struct {
			Enabled   bool     `json:"enabled"`
			Blacklist []string `json:"blacklist"`
		} `json:"plugin"`
		Proxy struct {
			Enabled  bool   `json:"enabled"`
			Hostname string `json:"hostname"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"proxy"`
	} `json:"server"`
	Userdata []struct {
		Spacename string   `json:"spacename"`
		Path      string   `json:"path"`
		Interval  int      `json:"interval"`
		Autostart bool     `json:"autostart"`
		Active    bool     `json:"active"`
		Devices   []string `json:"devices"`
	} `json:"userdata"`
	Version float64 `json:"version"`
}

func createJson() {
	// 创建一个Config实例，并设置一些默认值
	config := Config{
		Log: struct {
			LogLevel string `json:"loglevel"`
		}{
			LogLevel: "info",
		},
		Server: struct {
			Addr struct {
				ID       string `json:"id"`
				IP       string `json:"ip"`
				Port     int    `json:"port"`
				Password string `json:"password"`
			} `json:"addr"`
			Setting struct {
				Encode     string `json:"encode"`
				IOBalance  bool   `json:"iobalance"`
				Encryption string `json:"encryption"`
			} `json:"setting"`
			Scan struct {
				Enabled bool     `json:"enabled"`
				Type    string   `json:"type"`
				Max     int      `json:"max"`
				Devices []string `json:"devices"`
			} `json:"scan"`
			Plugin struct {
				Enabled   bool     `json:"enabled"`
				Blacklist []string `json:"blacklist"`
			} `json:"plugin"`
			Proxy struct {
				Enabled  bool   `json:"enabled"`
				Hostname string `json:"hostname"`
				Port     int    `json:"port"`
				Username string `json:"username"`
				Password string `json:"password"`
			} `json:"proxy"`
		}{
			Addr: struct {
				ID       string `json:"id"`
				IP       string `json:"ip"`
				Port     int    `json:"port"`
				Password string `json:"password"`
			}{
				ID:       "defaultID", // 你可以替换为你的函数getRandomString(8)
				IP:       "127.0.0.1",
				Port:     5002,
				Password: "defaultPassword", // 你可以替换为你的函数getRandomString(10)
			},
			Setting: struct {
				Encode     string `json:"encode"`
				IOBalance  bool   `json:"iobalance"`
				Encryption string `json:"encryption"`
			}{
				Encode:     "utf-8",
				IOBalance:  false,
				Encryption: "AES_ECB",
			},
			Scan: struct {
				Enabled bool     `json:"enabled"`
				Type    string   `json:"type"`
				Max     int      `json:"max"`
				Devices []string `json:"devices"`
			}{
				Enabled: true,
				Type:    "lan",
				Max:     5,
				Devices: []string{"127.0.0.1:5001"},
			},
			Plugin: struct {
				Enabled   bool     `json:"enabled"`
				Blacklist []string `json:"blacklist"`
			}{
				Enabled:   true,
				Blacklist: []string{},
			},
			Proxy: struct {
				Enabled  bool   `json:"enabled"`
				Hostname string `json:"hostname"`
				Port     int    `json:"port"`
				Username string `json:"username"`
				Password string `json:"password"`
			}{
				Enabled:  false,
				Hostname: "localhost",
				Port:     0,
				Username: "",
				Password: "",
			},
		},
		Userdata: []struct {
			Spacename string   `json:"spacename"`
			Path      string   `json:"path"`
			Interval  int      `json:"interval"`
			Autostart bool     `json:"autostart"`
			Active    bool     `json:"active"`
			Devices   []string `json:"devices"`
		}{
			{
				Spacename: "",
				Path:      "",
				Interval:  30,
				Autostart: true,
				Active:    true,
				Devices:   []string{""},
			},
		},
		Version: 0.01,
	}

	// 创建一个文件
	file, err := os.Create("config.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// 创建一个json编码器并将Config实例编码到文件中
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		fmt.Println("Error encoding config to JSON:", err)
	}
}

func test1(name string) {
	name = "wwww"
	fmt.Println(&name)
}
func test2(name *string) {
	fmt.Println(&name)
}

type test3 struct {
	dict sync.Map
}

func main() {
	//dc := map[string]interface{}{}
	//dc["1"] = 5
	//dc["2"] = "w"
	////fmt.Println(dc["5"] == nil)
	//delete(dc, "5")

	//tic1 := time.Now()
	//time.Sleep(time.Second * 1)
	//tic2 := time.Now()
	//fmt.Println(tic2.Unix() > tic1.Unix())

	//fmt.Println(int32(time.Now().Unix()))

	//var buf []byte
	//buf = append(buf, 255)
	//fmt.Println(buf)

	//buf := make([]byte, 8)
	//binary.BigEndian.PutUint64(buf, uint64(time.Now().Unix()))
	//x := binary.BigEndian.Uint64(buf)
	//fmt.Println(x)
	//name := "abc"
	//test1(name)
	//test2(&name)
	//fmt.Println(name)

	//data := map[string]interface{}{
	//	"command": "data", // 或者 "comm"
	//	"type":    "file",
	//	"method":  "get",
	//	"data": map[string]int{
	//		"a": 1,
	//		// ....
	//	},
	//}
	//
	//marshal, err := json.Marshal(data)
	//if err != nil {
	//	return
	//}
	//fmt.Println(string(marshal))

	//t := test3{}
	//t.dict.Store("5", [][]byte{{5, 6, 7}})
	//t.dict.Store("5", [][]byte{{5, 6, 7}})
	//
	//// 追加
	//i, _ := t.dict.Load("5")
	//t.dict, ok := append(i, []byte{1,2,3}...)

	//t.dict.Store("5", [][]byte{{5, 6, 7}})
	//
	//// 追加
	//i, _ := t.dict.Load("5")
	//if iSlice, ok := i.([][]byte); ok {
	//	iSlice = append(iSlice, []byte{1, 2, 3})
	//	t.dict.Store("5", iSlice)
	//}
	//
	//fmt.Println(t.dict.Load("5"))
	//ls := [][]byte{{1}, {2, 2, 3, 4}}
	////fmt.Println(ls[1:])
	//a := ls[1:]
	//fmt.Println(a)

	//slice := []byte("123awd")
	//slice2 := []byte("测")
	//fmt.Println(slice, slice2)
	//result := append(slice, slice2...)
	//fmt.Println(result)

	//command := map[string]any{
	//	"Name": "Bt",
	//	"Age": map[string]any{
	//		"Bt": 12,
	//		"Askel": map[string]any{
	//			"内置": 15,
	//		},
	//	},
	//}
	//var personConfig *Person

	//type Person struct {
	//	Name string         `json:"name"`
	//	Age  map[string]any `json:"age"`
	//	ID   struct {
	//		Number string `json:"number"`
	//	} `json:"id"`
	//}
	//
	//p := Person{
	//	Name: "Bt",
	//	Age: map[string]any{
	//		"Bt": 12,
	//		"Askel": map[string]any{
	//			"内置": 15,
	//		},
	//	},
	//}
	//
	//marshal, err := jsoniter.Marshal(p)
	//if err != nil {
	//	return
	//}
	//fmt.Println(string(marshal))
	//err = jsoniter.Unmarshal(marshal, &p) // 将json转换为Go对象
	//if err != nil {
	//	fmt.Println(err) // 处理错误
	//}
	//fmt.Println(p.Name, p.Age, p.ID.Number) // 输出Go对象的属性

	//bytes, _ := json.Marshal(p)
	//fmt.Println(string(bytes))
	//fmt.Println("________")
	//
	//var personConfig map[string]any
	//err := json.Unmarshal(bytes, &personConfig)
	//if err != nil {
	//	return
	//}
	//fmt.Println(personConfig["name"].(string))
	//file, err := os.ReadFile(".\\demo\\config.json")
	//if err != nil {
	//	fmt.Println("退出了0")
	//	return
	//}
	//
	//var config Person
	//err = json.Unmarshal(file, &config)
	//if err != nil {
	//	fmt.Println("退出了")
	//	return
	//}
	//
	//fmt.Println(config.ID == "")
	////askel := config["age"].(map[string]any)
	////fmt.Println(askel["Askel"])

	//str := "abcDe"
	//fmt.Println(strings.ToLower(str))
	//str := "abcdefw0123456789"
	//fmt.Println(string([]byte(str)[8:]))
	//var i = 10
	//{
	//	var i = 5
	//	fmt.Println(i)
	//}
	//fmt.Println(i)
	//var flag bool
	//fmt.Println(flag)

	//a := map[string]struct{}{"a": {}, "b": {}, "c": {}}
	//for k := range a {
	//	fmt.Printf(k)
	//}
	//fmt.Println(runtime.NumCPU())
	//type Server struct {
	//	mergeSocket map[string]map[string]bool
	//	flag        bool
	//}
	//server := Server{mergeSocket: map[string]map[string]bool{}}
	//server.mergeSocket["123"] = map[string]bool{
	//	"a": true,
	//}
	//server.mergeSocket["123"] = map[string]bool{
	//	"a": false,
	//	"b": true,
	//}
	//k, ok := server.mergeSocket["test"]["abce"]
	//
	//fmt.Println(k, ok)
	//fmt.Println(server.mergeSocket["123"])
	//fmt.Println(server.mergeSocket)
	//fmt.Println(server.flag)
	//t := []byte{2, 3, 5}
	//fmt.Println(t)

	//dc := map[string]any{
	//	"a": "b",
	//	"b": "c",
	//}
	//fmt.Println(len(dc))
	//comm := Command{
	//	Command: "a",
	//	Type:    "s",
	//	Method:  "5",
	//	Data: map[string]interface{}{
	//		"abc": 555,
	//	},
	//}
	//comm2 := Reply{Data: map[string]interface{}{
	//	"sss": 125,
	//}}
	//YourFunction(comm2)
	//s := map[string]any{
	//	"data": map[string]string{
	//		"a": "sss",
	//	},
	//}
	//commandJson, err := json.Marshal(s)
	//if err != nil {
	//	return
	//}
	//var d map[string]any
	//err = json.Unmarshal(commandJson, &d)
	//if err != nil {
	//	return
	//}
	//m, ok := d["datas"].(map[string]any)
	//fmt.Println(ok, m)

	//o := option2{
	//	Age: 5,
	//}
	//Run(o)

	//fmt.Println(&h)
	//p := &h
	//fmt.Println(&p)
	//fmt.Println(p.Age)
	//t := test("Bt", 15, true)
	////println(t.Name)
	//p := Person{H: t}
	//p.pr()
	//fmt.Println(p.H.Age)
	//fmt.Println(t.Age)
	//a := 1
	//b := a
	//fmt.Println(a)
	//fmt.Println(b)
	//type Inner struct {
	//	Field int
	//}
	//
	//type Outer struct {
	//	Inner
	//}
	//
	//o := Outer{
	//	InnerStruct: Inner{
	//		Field: 42,
	//	},
	//}
	//fmt.Println(o.InnerStruct.Field) // 输出：42
	//dc := map[string]string{
	//	"a": "A",
	//}
	//value, ok := dc["a"]
	//fmt.Println(ok)
	//fmt.Println(value)
	//fmt.Println(time.Now().UnixMilli())
	//fileInfo, err := os.Stat("test\\test1.go")
	//fmt.Println(err) //
	//fmt.Println(fileInfo.Size())
	//type User struct {
	//	Name string
	//	Age  uint
	//}
	//
	////var user User
	//user := User{}
	//if user == nil {
	//	fmt.Println("")
	//}

	//slice := []int{1, 2, 3, 4}
	//slice := make([]int, 8, 8)
	////slice[0] = 1
	////slice[1] = 1
	//slice = append(slice, 1)
	//slice = append(slice, 1)
	////slice = append(slice, 1)
	////slice = append(slice, 1)
	//fmt.Println(cap(slice))
	//fmt.Println(len(slice))
	//fmt.Println(slice)
	//slice := []byte{2, 3, 9, 5}
	//fmt.Println(string(slice[2]))
	//type Human struct {
	//	Name int
	//	Age  int
	//}
	//
	////h := new(Human)
	//h := &Human{}
	//fmt.Println(h.Age)

	//var h Human
	//fmt.Println(h.Age)
	//fmt.Println(h)
	//var human = Human{}
	//fmt.Println(human.Age)
	//h := Human{
	//	Name: 1,
	//	Age:  19,
	//}
	//fmt.Printf("%p\n", &h)
	//fmt.Println(&h.Name)
	//fmt.Println(&h.Age)
	//i := 5
	//var p *int
	//p = &i
	//fmt.Println(*p)

	//	t := testStruct{Name: "Bt"}
	//	//fmt.Println(&t)
	//	fmt.Println(t.Name)
	//	t.Test()
	//	fmt.Println(t.Name)
	//
	//}
	//
	//type testStruct struct {
	//	Name string
	//}
	//
	//func (t testStruct) Test() {
	//	t.Name = "None"
	//	fmt.Println(t.Name)
	//a := 5
	//fmt.Println(a / 2)
	//r := 5 % 2
	//fmt.Println(r)
	//f, err := os.OpenFile("test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0667)
	//buf := []byte{1, 55, 255}
	//n, err := f.Write(buf)
	//if err != nil {
	//	return
	//}
	//fmt.Println(n)
	//fileStat, err := os.Stat("tsest.txt")
	//if err != nil && os.IsNotExist(err) {
	//	fmt.Println(err)
	//}
	//fmt.Println(fileStat.Size())
	//type Human struct {
	//	Name string
	//	Age  uint
	//}
	//fmt.Println(unsafe.Sizeof(Human{}))
	//test(1)
	//fmt.Println(time.Now().UnixMilli())
	//_, b := time.Now().Zone()
	//fmt.Println(b / 3600)
	//
	//m_b := time.Now().UnixMilli() // 8
	//
	//m_a := 1707387183486 - 18000000 // 3

	//t := time.Now()
	//fmt.Println(t.Year(), t.Month(), t.Day())

	//type Human struct {
	//}
	//human := &Human{}
	//fmt.Println(human)
	//if human == nil {
	//	fmt.Println("空")
	//} else {
	//	fmt.Println("非空")
	//}
	//a := 5000
	//b := 3600
	//fmt.Println(a / b)
	//var i *int
	//fmt.Println(i)
	//type void struct {
	//}
	//fmt.Println(unsafe.Sizeof(make(map[string]void)))
	//dc := make(map[string]chan string)
	//dc["abc"] = make(chan string)
	//delete(dc, "abc")
	//fmt.Println(dc)
	//wd, err := os.Getwd()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(wd)
	//
	//var input string
	//fmt.Scanln(&input)
	//
	//mapL := make(map[string]int)
	//editMap(mapL)
	//fmt.Println(mapL)
	//mapL := Maps{Map: make(map[string]int)}
	//mapL.editMap()
	//fmt.Println(mapL)
	//map1 := Maps{Map: make(map[string]int)}
	//map1.editMap()
	//map2 := Maps2{make(map[string]int)}
	//map2.editMap(map1.Map)
	//fmt.Println(map1.Map)
	//fmt.Println(os.Getwd())
	//f, err := os.Open("test\\test1.go")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	//seek, err := f.Seek(1000, 1)
	//if err != nil {
	//	return
	//}
	//fmt.Println(seek)
	//buf := make([]byte, 4096)
	//n, err := f.Read(buf)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(buf[:n]))
	//m := map[string]any{}
	//stat, ok := m["stat"].(string)
	//if ok {
	//	fmt.Println("ok")
	//	fmt.Println(stat)
	//}
	//_ = NewT()
	//time.Sleep(3 * time.Second)
	//fmt.Println(t.Name)

	//fmt.Println(time.Now().Unix())
	//
	//m := map[string]any{}
	//m["a"] = "asss"
	//
	//hash, _ := m["b"].(string)
	//fmt.Println(hash)
	//writeFile()
	//type file struct {
	//	Name string
	//	Age  int
	//}
	//a := file{
	//	//Name: "ss",
	//	Age: 5,
	//}
	//fmt.Println(len(a.Name))
	//fmt.Println(time.Now().UnixMilli())
	//fileStat, err := os.Stat(".\\test\\test2.go")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//t := fileStat.ModTime()
	//editTime := t.Unix()
	//
	//fmt.Println(editTime)
	//
	//Bt := time.Unix(editTime+5, 0)
	//err = os.Chtimes(".\\test\\test2.go", Bt, Bt)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(Bt.Unix())

	//t := TimeFunc(1708339242, 8, 9)
	//fmt.Println(t)
	//read()
	//s := &ser{}
	//go s.sT()
	//s = nil
	//time.Sleep(5 * time.Second)
	//a := func() {
	//	go process()
	//}
	//a()
	//time.Sleep(1 * time.Second)
	//fmt.Println("end")
	//time.Sleep(4 * time.Second)
	//d := map[string]any{
	//	"aa": map[string]any{
	//		"bb": 12,
	//	},
	//}
	//for s, v := range d {
	//	fmt.Println(s, v.(map[string]any)["bb"].(int))
	//}
	//type file struct {
	//	FileName string
	//	Path     string
	//}
	//f := file{
	//	FileName: "a",
	//	Path:     "b",
	//}
	//for k := range f {
	//	fmt.Println(k)
	//}

	//f, err := os.OpenFile(".\\test\\test.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0667)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	////f.Write()
	//f.
	//os.Create()
	//file, err := os.OpenFile(".\\test\\test.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	//if err != nil {
	//	return
	//}
	//err = file.Truncate(1024 * 1024)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//file.Write([]byte{50, 50})
	////buf := make([]byte, 512)
	////file.Read(buf)
	//currentPosition, err := file.Seek(0, 1)
	//fmt.Println(currentPosition)
	//
	//file.Close()
	//
	//f, err := os.OpenFile(".\\test\\test.txt", os.O_TRUNC|os.O_WRONLY, 0666)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//f.Write([]byte{51, 51})
	//f.Write([]byte{52, 52})
	//f.Close()
	////err = os.Truncate(".\\test\\test.txt", 1024*1024)
	////fmt.Println(err)
	//r := strings.NewReader("abcde")
	//buf := make([]byte, 4)
	//if _, err := io.ReadFull(r, buf); err != nil {
	//	log.Fatal(err)
	//}
	//bufio.fmt.Println(buf)
	//addrs, err := net.InterfaceAddrs()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	////fmt.Println(addrs)
	//for _, addr := range addrs {
	//	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
	//		if ipnet.IP.To4() != nil {
	//			a := strings.Split(ipnet.IP.String(), ".")
	//			fmt.Println(a)
	//
	//		}
	//	}
	//}
	//fmt.Println(^uint64(0) >> 1)
	//fmt.Println(255 * 3)
	//leng := []byte{15, 255} // 1111 1111 0111 1111
	//num := int32(binary.BigEndian.Uint16(leng))
	//fmt.Println(num)
	////
	//intToByte := 28
	//b := make([]byte, 2)
	//binary.BigEndian.PutUint16(b, uint16(intToByte))
	//fmt.Println(b)
	//
	//fmt.Println(5 >> 2)
	//
	////str := "abcdefss"
	////fmt.Println(str[:6])
	//timer := time.NewTicker(1 * time.Second)
	//fmt.Println(timer)
	//ctx := context.Background()
	//fmt.Println(<-ctx.Done())
	//ctx2, ctx2Cancel := context.WithCancel(ctx)
	//ctx2Cancel()
	//fmt.Println(<-ctx2.Done())
	//ctx3 := context.WithValue(ctx2, "a", "b")
	//fmt.Println(<-ctx3.Done())

	//wg := sync.WaitGroup{}
	//wg.Add(2)
	//t1(&wg)
	//t2(&wg)
	//wg.Wait()

	//var wait = sync.WaitGroup{}
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//wait.Add(1)
	//go func() {
	//	GetFile(ctx, &wait)
	//	fmt.Println(1)
	//	wait.Done()
	//}()
	//fmt.Println("ss")
	//time.Sleep(1 * time.Second)
	//cancel()
	//wait.Wait()
	//buf := make([]byte, 255)
	//addBuf(buf)
	//fmt.Println(buf)

	fmt.Println(walk(".\\test"))
}

func walk(walkPath string) ([]string, error) {
	var filePaths []string
	files, err := os.ReadDir(walkPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			//fmt.Println(file.Name())
			//fmt.Println(filepath.Join(walkPath, file.Name()))
			subFiles, err := walk(filepath.Join(walkPath, file.Name()))
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			filePaths = append(filePaths, subFiles...)
		} else {
			filePaths = append(filePaths, filepath.Join(walkPath, file.Name()))
		}
	}
	return filePaths, nil
}

func addBuf(b []byte) {
	b = append(b, 25)
}

func GetFile(ctx context.Context, wait *sync.WaitGroup) {
	defer ctx.Done()

	go func() {
		select {
		case <-ctx.Done():
			wait.Done()
			return
		}
	}()

	fmt.Println("start")
	time.Sleep(2 * time.Second)
	fmt.Println("over")

}

func t1(wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("exit t1")
	}()
	go func() {
		fmt.Println(1)
		wg.Done()
		return
	}()
	fmt.Println()
}
func t2(wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("exit t2")
	}()
	go func() {
		fmt.Println(3)
		wg.Done()
		return
	}()
}

func minKey(m map[string]int) string {
	minV := int(^uint(0) >> 1) // 最大的int值
	minK := ""
	for k, v := range m {
		if v < minV {
			minV = v
			minK = k
		}
	}
	return minK
}

func process() {
	fmt.Println("456")
	time.Sleep(3 * time.Second)
	fmt.Println("123")
}

type ser struct {
}

func (s *ser) sT() {
	fmt.Println("1")
	time.Sleep(3 * time.Second)
	fmt.Println("2")
}

func read() {
	info := os.Getenv(".\\")
	fmt.Println(info)
}

//	func TimeFunc(t, offset1, offset2 int64) int64 {
//		return time.Unix(t-(offset2*3600-offset1*3600), 0).Unix()
//	}
func TimeFunc(t, offset1, offset2 int64) int64 {
	return time.Unix(t-(9*3600-8*3600), 0).Unix()
}

func writeFile() {
	var t int64 = 1708339242
	unix := time.Unix(t, 0)

	path := "test\\test.txt"
	data := []byte{10, 50, 100, 255}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0776)
	if err != nil {
		fmt.Println(err)
		return
	}
	var i int
	for {
		if i == 4 {
			return
		}
		buf := []byte{data[i]}
		_, err = f.Write(buf)
		if err != nil {
			return
		}
		err = os.Chtimes(path, unix, unix)
		if err != nil {
			fmt.Println(err)
		}
		i++
	}
}

//func server() {
//	listener, err := net.Listen("tcp")
//	if err != nil {
//		return
//	}
//	conn, err := listener.Accept()
//	if err != nil {
//		return
//	}
//}

//func threading() {
//	time.Sleep(2 * time.Second)
//	fmt.Println("over")
//}
//
//type Test struct {
//	Name int
//}
//
//func NewT() *Test {
//	go threading()
//	go threading()
//	return &Test{}
//}

//
//type Maps struct {
//	Map map[string]int
//}
//
//func (m *Maps) editMap() {
//	m.Map["a"] = 5
//}
//
//type Maps2 struct {
//	Map map[string]int
//}
//
//func (m *Maps2) editMap(mapL *map[string]int) {
//	mapL["b"] = 10
//}

//
//func test(i int) {
//	if i == 1 {
//		defer fmt.Println(1)
//	} else if i == 2 {
//		defer fmt.Println(2)
//	}
//}

//func test(name string, age int, gender bool) human {
//	h := human{
//		Name:   name,
//		Age:    age,
//		Gender: gender,
//	}
//	return h
//}
//
//type human struct {
//	Name   string
//	Age    int
//	Gender bool
//}
//
//type Person struct {
//	H human
//}
//
//func (p *Person) pr() {
//	p.H.Age = 50
//}

//
//// Run 批量接受指定接口范围内的对象，并对对象本身方法进行处理
//func Run(data optioner) {
//	fmt.Println(data)
//}
//
//type optioner interface {
//	Print()
//	//print()
//}
//
//type option struct {
//	Name string
//}
//
//func (o option) Print() {
//	fmt.Println(o.Name)
//}
//
//type option2 struct {
//	Age int
//}
//
//func (o option2) Print() {
//	fmt.Println(o.Age)
//}

//type option struct {
//	Name string
//	Age  int
//	Data map[string]any
//}
//
//type option2 struct {
//	Data map[string]string
//}

//type animal interface {
//	Run()
//}
//
//type Cat struct {
//}
//
//func (Cat) Run() {
//	fmt.Println("跑步了")
//}

//type Command struct {
//	Command string                 `json:"command"`
//	Type    string                 `json:"type"`
//	Method  string                 `json:"method"`
//	Data    map[string]interface{} `json:"data"`
//}
//
//func (c Command) GetData() map[string]interface{} {
//	return c.Data
//}
//
//type Reply struct {
//	Data map[string]interface{} `json:"data"`
//}
//
//func (r Reply) GetData() map[string]interface{} {
//	return r.Data
//}
//
//type DataHolder interface {
//	GetData() map[string]interface{}
//}
//
//func YourFunction(data DataHolder) {
//	// Your implementation here
//	fmt.Println(data.GetData())
//}
