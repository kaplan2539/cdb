package main

import (
    "bufio"
    "io"
    "io/ioutil"
    "path/filepath"
    "log"
    "os"
    "os/exec"
    "net/http"
    "gopkg.in/gorilla/mux.v1"
    "strings"
    "regexp"
)

func tar_rootfs(w http.ResponseWriter, r *http.Request) {
    /// Do not use the '-v' parameter with tar unless it is read out...
    cmd := exec.Command("tar","cz","rootfs")
    stdout,err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }

    /// if tar -v is used this has to be read out!!!!!
    stderr,err := cmd.StderrPipe()
    if err != nil {
        log.Fatal(err)
    }

    if err:=cmd.Start(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("%s\n", slurp)
        log.Fatal(err)
    }

    nBytes, nChunks := int64(0), int64(0)
    reader := bufio.NewReader(stdout)
    buf := make([]byte, 0, 4*1024)
    for {
        log.Println("reading...")
        n, err := reader.Read(buf[:cap(buf)])
        buf = buf[:n]
        if n == 0 {
            if err == nil {
                log.Println("got zero bytes --> continue")
                continue
            }
            if err == io.EOF {
                break
            }
            slurp, _ := ioutil.ReadAll(stderr)
            log.Printf("%s\n", slurp)
            log.Fatal(err)
        }
        nChunks++
        nBytes += int64(len(buf))
        log.Println("got: ",nBytes)

        // process buf
        if err != nil && err != io.EOF {
            slurp, _ := ioutil.ReadAll(stderr)
            log.Printf("%s\n", slurp)
            log.Fatal(err)
        }

        nWritten, werr := w.Write(buf)
        if werr != nil || nWritten != len(buf) {
            log.Fatal(werr)
            log.Printf("nWritten = %d\nlen(buf) = %d", nWritten, len(buf))
        }

        // log.Println("got: [",string(buf[:]),"]")
    }
    if err:=cmd.Wait(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("stderr=%s\n", slurp)
        log.Fatal(err)
    }
    log.Println("Bytes:", nBytes, "Chunks:", nChunks)
}

func info(w http.ResponseWriter, r *http.Request) {
    root:="/dev/"

    filepath.Walk(root, func (path string, info os.FileInfo, err error) error {
        if info.IsDir() && strings.Compare(path,root)!=0 {
            return filepath.SkipDir
        }

        if match,_ :=regexp.MatchString(".*mtd[0-9]+$",path); match==true {
//        if match,_ :=filepath.Match("*",path); match==true {
            log.Println("MATCH:",path)
//        } else {
//            log.Println(path)
        }

        return err
    })
}

func file(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    dest_path := "/" + vars["path"]
    log.Println("r.URL.Path =", string(r.URL.Path), " dest_path:",dest_path," METHOD =",r.Method)

    switch r.Method {
        case "GET":
            log.Println("GET HANDLER")
                f, err := os.Open(dest_path)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusConflict)
                return
            }
            defer f.Close()
            io.Copy(w, f)
        case "PUT":
            log.Println("PUT HANDLER")
            f, err := os.OpenFile(dest_path, os.O_WRONLY|os.O_CREATE, 0666)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusConflict)
                return
            }
            defer f.Close()
            io.Copy(f, r.Body)
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
    }
}


func main() {

    log.Println("CDBD says hello\n\n")

    r:=mux.NewRouter()
    r.HandleFunc("/backup", tar_rootfs )
    r.HandleFunc("/info", info )
    r.HandleFunc("/file/{path:.*}", file )

    if err:=http.ListenAndServe(":8080",r); err!=nil {
        log.Fatal("ListenAndServe:",err)
    }

}
