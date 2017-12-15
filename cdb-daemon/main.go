package main

import (
    "bufio"
    "io"
    "io/ioutil"
    "log"
    "os/exec"
    "net/http"
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
}

func main() {

    http.HandleFunc("/backup", tar_rootfs )
    http.HandleFunc("/info", info )

    if err:=http.ListenAndServe(":8080",nil); err!=nil {
            log.Fatal("ListenAndServe:",err)
    }
}
