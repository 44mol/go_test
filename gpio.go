package main

import (
        "fmt"
        "os"
        "reflect"
        "syscall"
        "unsafe"
        "time"
        "encoding/json"
        "io/ioutil"
        "net/http"
        "net/url"
)

const (
        baseAddr = 0x3F000000
        gpioBaseAddr = baseAddr + 0x200000
        blockSize = 4096
)

var (
        mem8    []uint8
        gpioRegister []uint32
        IncomingUrl string = ""
)

type Slack struct {
        Text        string `json:"text"`
        Username    string `json:"username"`
        Icon_emoji  string `json:"icon_emoji"`
        Icon_url    string `json:"icon_url"`
        Channel     string `json:"channel"`
}


func allocateRegister() (err error) {
    // /dev/memをfile変数に割り当てる
    file, err := os.OpenFile(
            "/dev/mem",             // 開くファイルの指定
            os.O_RDWR|os.O_SYNC,    // アクセス権、読み書き＋同期モード
            0)

    if err != nil {
        return
    }

    defer file.Close()

    mem8, err = syscall.Mmap(
        int(file.Fd()),
        gpioBaseAddr,
        blockSize,
        syscall.PROT_READ|syscall.PROT_WRITE,
        syscall.MAP_SHARED)

    if err != nil {
        return
    }

    header := *(*reflect.SliceHeader)(unsafe.Pointer(&mem8))
    header.Len /= 4
    header.Cap /= 4

    gpioRegister = *(*[]uint32)(unsafe.Pointer(&header))

    return nil
}

func main() {
    err := allocateRegister()
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("%032b\n", &gpioRegister[0])

    gpioRegister[0] = (gpioRegister[0] &^ (7 << 12))

    isPushSwitch := true
    isAtWork := false

    for {
        count := 0
        for i := 0; i < 3; i++ {
            if gpioRegister[13] & (1 << 4) == 0 {
                count += 1
                fmt.Println("ON")
            } else {
                fmt.Println("OFF")
            }
            time.Sleep(time.Second)

        }
        if isPushSwitch == false && count == 3 && isAtWork {
            isPushSwitch = true
            post("投稿テスト:退社")
        } else if isPushSwitch == true && count == 0 {
            isPushSwitch = false
            post("投稿テスト:出社")
            isAtWork = true
        }
    }
}

func post(text string) {
    params, _ := json.Marshal(Slack{
        text,
        "name",
        "",
        "",
        ""})

    resp, _ := http.PostForm(
        IncomingUrl,
        url.Values{"payload": {string(params)}},
    )

    body, _ := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()

    println(string(body))
}

