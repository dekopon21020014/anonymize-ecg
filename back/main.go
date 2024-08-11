package main

import (
	"github.com/dekopon21020014/anonymize-mfer/controller"
	// "back/controller"
)

func main() {
	router := controller.GetRouter()
	router.Run()
}

// ここから下は以前のメイン関数
// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"os"
// )

// func main() {
// 	var path string
// 	var err  error

// 	if len(os.Args) < 2 { // コマンドライン引数なしならとりあえずECG01を対象にする(開発用)
// 		path = "../sample-data/sample-data.MWF"
// 	} else {
// 		path = os.Args[1]
// 	}

// 	mfer := newMfer()
// 	anonymized_data, err := loadMfer(path)
// 	if err != nil {
// 		log.Fatal(err)
// 		m, _ := json.MarshalIndent(mfer, "", "    ")
// 		fmt.Println(string(m))
// 		return
// 	}

// 	file, err := os.Create("output.mwf")
//     if err != nil {
//         fmt.Println("Error creating file:", err)
//         return
//     }
//     defer file.Close()

// 	_, err = file.Write(anonymized_data)
//     if err != nil {
//         fmt.Println("Error writing to file:", err)
//         return
//     }

// 	//m, _ := json.MarshalIndent(mfer, "", "    ")
// 	//fmt.Println(string(m))
// 	// fmt.Printf("%+v", mfer)
// }
