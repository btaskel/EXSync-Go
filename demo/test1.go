package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

	//type Person struct {
	//	Name string         `json:"name"`
	//	Age  map[string]any `json:"age"`
	//	ID   struct {
	//		Number string `json:"number"`
	//	} `json:"id"`
	//}
	//p := Person{
	//	Name: "Bt",
	//	Age: map[string]any{
	//		"Bt": 12,
	//		"Askel": map[string]any{
	//			"内置": 15,
	//		},
	//	},
	//}
	//bytes, _ := json.Marshal(p)
	//fmt.Println(string(bytes))

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

	str := "abcDe"
	fmt.Println(strings.ToLower(str))
}
