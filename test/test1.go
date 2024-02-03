package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
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

}

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
