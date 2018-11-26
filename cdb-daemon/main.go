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
    "syscall"
    "strconv"
//    "path"
    "encoding/json"
)

var executable_path = "bla"

/// TODO: determine NAND layout
var mount_point="/rootfs"
var mtd_device=4
var ubi_vol="/dev/ubi0_0"
var ubi_dev="/dev/ubi0"

func run_cmd(name string, args ...string) error {
    log.Println("RUNNING: "+name+" "+strings.Join(args," "))
    cmd := exec.Command(name,args...)
    stderr,err := cmd.StderrPipe()
    if err != nil {
        log.Fatal(err)
        return err
    }
    if err:=cmd.Start(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("%s\n", slurp)
        log.Printf(err.Error())
        return err
    }
    if err:=cmd.Wait(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("stderr=%s\n", slurp)
        log.Fatal(err.Error())
        return err
    }
    return nil
}

func ubi_detach(mtd_device int) error {
    return run_cmd("ubidetach","-m"+strconv.Itoa(mtd_device))
}


func ubi_attach(mtd_device int) error {
    return run_cmd("ubiattach","-m"+strconv.Itoa(mtd_device))
}

func create_ubi() error {
    // Write ubi.cfg
    ubi_cfg := []byte(`
[rootfs]
mode=ubi
vol_id=0
vol_size=7168MiB
vol_type=dynamic
vol_name=rootfs
vol_alignment=1
`)
    if err := ioutil.WriteFile("/tmp/ubi.cfg", ubi_cfg, 0644); err != nil {
            log.Printf("create_ubi(): error writing /tmp/ubi.cfg");
            return err
    }

    // Run ubinize
    if err := run_cmd("ubinize","-o","/tmp/ubi.bin","-p","0x400000","-m","0x4000","-s","0x4000","-M","dist3","/tmp/ubi.cfg"); err != nil {
        return err
    }

     // Run flash_erase
    if err := run_cmd("flash_erase","/dev/mtd4","0","2044"); err != nil {
        return err
    }

    // Run nand_write
    if err := run_cmd("nandwrite","-m","-p","/dev/mtd4","/tmp/ubi.bin"); err != nil {
        return err
    }

    return nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    log.Printf(err.Error())
    return false
}

func mounted(mountpoint string) (bool, error) {
	mntpoint, err := os.Stat(mountpoint)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	parent, err := os.Stat(filepath.Join(mountpoint, ".."))
	if err != nil {
		return false, err
	}
	mntpointSt := mntpoint.Sys().(*syscall.Stat_t)
	parentSt := parent.Sys().(*syscall.Stat_t)
	return mntpointSt.Dev != parentSt.Dev, nil
}

func run(w http.ResponseWriter, r *http.Request) {

    cmd_string := r.FormValue("cmd")
    args := strings.Split(cmd_string," ")

    log.Println("cmd="+cmd_string)
    log.Println("args="+strings.Join(args," "))

    /// Do not use the '-v' parameter with tar unless it is read out...
    /// Also: We need gnu tar!
    cmd := exec.Command(args[0],args[1:]...)
    stdout,err := cmd.StdoutPipe()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    /// if tar -v is used this has to be read out!!!!!
    stderr,err := cmd.StderrPipe()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    if err:=cmd.Start(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("%s\n", slurp)
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    nBytes, nChunks := int64(0), int64(0)
    reader := bufio.NewReader(stdout)

    io.Copy(w,reader)

//    buf := make([]byte, 0, 4*1024*1024)
//    for {
//        log.Println("reading...")
//        n, err := reader.Read(buf[:cap(buf)])
//        buf = buf[:n]
//        if n == 0 {
//            if err == nil {
//                log.Println("got zero bytes --> continue")
//                continue
//            }
//            if err == io.EOF {
//                break
//            }
//            slurp, _ := ioutil.ReadAll(stderr)
//            log.Printf("%s\n", slurp)
//            log.Println(err.Error())
//			w.WriteHeader(http.StatusConflict)
//            return
//        }
//        nChunks++
//        nBytes += int64(len(buf))
//        log.Println("got: ",nBytes)
//
//        if err != nil && err != io.EOF {
//            slurp, _ := ioutil.ReadAll(stderr)
//            log.Printf("%s\n", slurp)
//            log.Println(err.Error())
//			w.WriteHeader(http.StatusConflict)
//            return
//        }
//
//        nWritten, werr := w.Write(buf)
//        if werr != nil || nWritten != len(buf) {
//            log.Printf("nWritten = %d\nlen(buf) = %d", nWritten, len(buf))
//            log.Println(werr)
//			w.WriteHeader(http.StatusConflict)
//            return
//        }
//
////        log.Println("got: [",string(buf[:]),"]")
//    }
    if err:=cmd.Wait(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Println("hellau")
        log.Printf("stderr=%s\n", slurp)
        log.Println(err.Error())
        slurp2, _ := ioutil.ReadAll(stdout)
        log.Printf("%s\n", slurp2)
        w.WriteHeader(http.StatusConflict)
        return
    }
    log.Println("Bytes:", nBytes, "Chunks:", nChunks)
}

func tar_gz_rootfs(w http.ResponseWriter, r *http.Request) {
    _tar_rootfs(w,r,[]string{"/bin/tar","cz","-C","/rootfs","--exclude=dev/*","--exclude=/var/cache/apt/*","."})
}

func tar_rootfs(w http.ResponseWriter, r *http.Request) {
    _tar_rootfs(w,r,[]string{"/bin/tar","c","-C","/rootfs","--exclude=dev/*","--exclude=/var/cache/apt/*","."})
}

func untar_rootfs(w http.ResponseWriter, r *http.Request) {
    _untar_rootfs(w,r,[]string{"/bin/tar","x","-C","/rootfs"})
}

func _tar_rootfs(w http.ResponseWriter, r *http.Request, tar_cmd []string) {

    if ! exists(ubi_dev) {
        log.Println("Attach UBI volume")
        if err:=ubi_attach(mtd_device); err!=nil {
            log.Printf(err.Error())
            w.WriteHeader(http.StatusConflict)
            return
        }
    }

    if ! exists(mount_point) {
        log.Println("Create mount point")
        if err:=os.Mkdir(mount_point,0666); err!=nil {
            log.Printf(err.Error())
            w.WriteHeader(http.StatusConflict)
            return
        }
    }

	if is_mounted,err:=mounted(mount_point); err != nil {
        log.Printf(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
	} else if !is_mounted {
		log.Println("Mounting ubifs")
		if err:=syscall.Mount(ubi_vol,mount_point,"ubifs",syscall.MS_RDONLY,""); err!=nil {
			log.Printf(err.Error())
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

    /// Do not use the '-v' parameter with tar unless it is read out...
    /// Also: We need gnu tar!
    cmd := exec.Command(tar_cmd[0],tar_cmd[1:]...)
    stdout,err := cmd.StdoutPipe()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    /// if tar -v is used this has to be read out!!!!!
    stderr,err := cmd.StderrPipe()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    if err:=cmd.Start(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("%s\n", slurp)
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    nBytes, nChunks := int64(0), int64(0)
    reader := bufio.NewReader(stdout)
    buf := make([]byte, 0, 4*1024*1024)
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
            log.Println(err.Error())
			w.WriteHeader(http.StatusConflict)
            return
        }
        nChunks++
        nBytes += int64(len(buf))
        log.Println("got: ",nBytes)

        if err != nil && err != io.EOF {
            slurp, _ := ioutil.ReadAll(stderr)
            log.Printf("%s\n", slurp)
            log.Println(err.Error())
			w.WriteHeader(http.StatusConflict)
            return
        }

        nWritten, werr := w.Write(buf)
        if werr != nil || nWritten != len(buf) {
            log.Printf("nWritten = %d\nlen(buf) = %d", nWritten, len(buf))
            log.Println(werr)
			w.WriteHeader(http.StatusConflict)
            return
        }

//        log.Println("got: [",string(buf[:]),"]")
    }
    if err:=cmd.Wait(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Println("hellau")
        log.Printf("stderr=%s\n", slurp)
        log.Println(err.Error())
        slurp2, _ := ioutil.ReadAll(stdout)
        log.Printf("%s\n", slurp2)
        w.WriteHeader(http.StatusConflict)
        return
    }
    log.Println("Bytes:", nBytes, "Chunks:", nChunks)
}

func _untar_rootfs(w http.ResponseWriter, r *http.Request, tar_cmd []string) {

    if r.Method != "PUT"  {
        msg := "ERROR: "+r.URL.Path + " does only accept PUT requests sorry\n\n"
        w.WriteHeader(http.StatusConflict)
        w.Write([]byte(msg))
        return
    }

    if ! exists(mount_point) {
        log.Println("Create mount point")
        if err:=os.Mkdir(mount_point,0666); err!=nil {
            log.Printf(err.Error())
            w.WriteHeader(http.StatusConflict)
            return
        }
    }

	if is_mounted,err:=mounted(mount_point); err != nil && is_mounted {
		log.Println("Unmounting ubifs")
		if err:=syscall.Unmount(mount_point,0); err!=nil {
			log.Printf(err.Error())
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

    if exists(ubi_dev) {
        log.Println("Detach UBI volume")
        if err:=ubi_detach(mtd_device); err!=nil {
            log.Printf(err.Error())
            w.WriteHeader(http.StatusConflict)
            return
        }
    }

    if err:=create_ubi(); err !=nil {
            log.Printf(err.Error())
            w.WriteHeader(http.StatusConflict)
            return
     }

    log.Println("Attaching ubi")
    if err:=ubi_attach(mtd_device); err!=nil {
        log.Printf(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

     log.Println("Mounting ubifs")
     if err:=syscall.Mount(ubi_vol,mount_point,"ubifs",syscall.MS_SYNC,""); err!=nil {
         log.Printf(err.Error())
         w.WriteHeader(http.StatusConflict)
         return
     }

    /// Do not use the '-v' parameter with tar unless it is read out...
    /// Also: We need gnu tar!
    log.Println("creating stdin pipe")
    cmd := exec.Command(tar_cmd[0],tar_cmd[1:]...)
    stdin,err := cmd.StdinPipe()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    log.Println("creating stderr pipe")
    /// if tar -v is used this has to be read out!!!!!
    stderr,err := cmd.StderrPipe()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    log.Println("exec tar")
    if err:=cmd.Start(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("%s\n", slurp)
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    nBytes, nChunks := int64(0), int64(0)
    writer := bufio.NewWriter(stdin)

    log.Println("start streaming...")

    if _, err := io.Copy(writer,r.Body); err != nil {
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
    }

    writer.Flush()
    stdin.Close()

    if err:=cmd.Wait(); err != nil {
        slurp, _ := ioutil.ReadAll(stderr)
        log.Printf("stderr=%s\n", slurp)
        log.Println(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }
    log.Println("Bytes:", nBytes, "Chunks:", nChunks)

    log.Println("Unmounting ubifs")
    if err:=syscall.Unmount(mount_point,0); err!=nil {
        log.Printf(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }
    log.Println("Detach UBI volume")
    if err:=ubi_detach(mtd_device); err!=nil {
        log.Printf(err.Error())
        w.WriteHeader(http.StatusConflict)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func info(w http.ResponseWriter, r *http.Request) {
    root:="/sys/class/mtd"

    var mtds []*MTD
    filepath.Walk(root, func (p string, info os.FileInfo, err error) error {
        if info.IsDir() && strings.Compare(p,root)!=0 {
            return filepath.SkipDir
        }

        if match,_ :=regexp.MatchString(".*mtd[0-9]+$",p); match==true {
            log.Println("MATCH:",p)

            mtd := FromSysFs(p)

            log.Println("mtd",mtd)
            mtds = append(mtds,mtd)
        }

        return err
    })
    json.NewEncoder(w).Encode(mtds)
}

func file(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    dest_path := "/" + vars["path"]
    log.Println("r.URL.Path =", string(r.URL.Path), " dest_path:",dest_path," METHOD =",r.Method)
    log.Println("executable_path =",executable_path)

    var dest_mode os.FileMode
    if dest_info,err := os.Lstat(dest_path); err!=nil {
        dest_mode = 0666
    } else {
        dest_mode = dest_info.Mode()
    }

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

            if strings.Compare(executable_path,dest_path)==0 {
                //overwriting ourself here, some extra caution necesary
                f, err := os.OpenFile(dest_path+".__new", os.O_WRONLY|os.O_CREATE, dest_mode)
                if err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusConflict)
                    return
                }
                defer f.Close()
                io.Copy(f, r.Body)
                if err:=os.Rename(f.Name(),dest_path); err!=nil {
                    log.Println(err)
                }
                //TODO: trigger restart
            } else {
                f, err := os.OpenFile(dest_path, os.O_WRONLY|os.O_CREATE, dest_mode)
                if err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusConflict)
                    return
                }
                defer f.Close()
                io.Copy(f, r.Body)
            }
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
    }
}


func main() {
    var err error

    log.Println("CDBD says hello! "+os.Args[0]+"\n\n")
    var e error
    //executable_path, err = os.Executable()
    executable_path, err = os.Readlink("/proc/self/exe")
    if e != nil {
        panic(err)
    }

    r:=mux.NewRouter()
    r.HandleFunc("/backup", tar_rootfs )
    r.HandleFunc("/restore", untar_rootfs )
    r.HandleFunc("/zbackup", tar_gz_rootfs )
    r.HandleFunc("/info", info )
    r.HandleFunc("/run", run )
    r.HandleFunc("/file/{path:.*}", file )

    if err:=http.ListenAndServe(":8080",r); err!=nil {
        log.Fatal("ListenAndServe:",err)
    }

}
