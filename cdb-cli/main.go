package main

import (
//    "bytes"
    "fmt"
    "io"
    "io/ioutil"
//    "mime/multipart"
    "net/http"
    "os"
	"flag"
    "path/filepath"
    "strings"
)

var verbose = false

func push(local_path string, remote_path string) {
    if ! strings.HasPrefix(remote_path,"/") {
        fatal("ERROR: remote path must be absolute",1)
    }

	f, err := os.Open(local_path)
	if err != nil {
		fatal("ERROR: cannot open file '"+local_path+"'",1)
	}
	defer f.Close()

    url := "http://192.168.81.1:8080/file"+remote_path

    fmt.Println("REMOVE ME: url =",url)

    client := &http.Client{}
    req,err := http.NewRequest("PUT",url,f)
	res, err := client.Do(req)
	if err != nil {
		fatal("ERROR: "+err.Error(),1)
	} else {
		defer res.Body.Close()
		_, err := ioutil.ReadAll(res.Body)
		if err!= nil {
			fatal("ERROR: "+err.Error(),1)
		}
	}
}

func pull(remote_path string, local_path string) {
    fmt.Println("REMOVE ME: remote_path="+remote_path+" local_path="+local_path)

    if ! strings.HasPrefix(remote_path,"/") {
        fatal("ERROR: remote_path path must be absolute",1)
    }

	f, err := os.Create(local_path)
	if err != nil {
		fatal("ERROR: cannot write file '"+local_path+"'",1)
	}
	defer f.Close()

    url := "http://192.168.81.1:8080/file"+remote_path


	res, err := http.Get(url)
	if err != nil {
		fatal("ERROR: "+err.Error(),1)
	} else {
		defer res.Body.Close()

        _, err := io.Copy(f,res.Body)
		if err!= nil {
			fatal("ERROR: "+err.Error(),1)
		}
	}
}

func fatal(msg string, code int) {
		flag.Usage()
        fmt.Println(msg+"\n")
        os.Exit(code)
}

func usage() {
	fmt.Println("")
	fmt.Printf("USAGE: %s [options] COMMAND\n", filepath.Base(os.Args[0]))
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  info                        Display information about device")
	fmt.Println("  push <local> <remote>       Upload file to device")
	fmt.Println("  pull <remote> <local>       Download file from device")
	fmt.Println("  help                        Print this message")
	fmt.Println("")
	fmt.Printf("Run '%s COMMAND --help' for more information on the command\n", filepath.Base(os.Args[0]))
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -v	Verbose execution")
	fmt.Println("")
}

func main() {
    //TODO: switch to urfave/cli for cmdline handling
	flag.Usage=usage
	flag.BoolVar(&verbose, "v", false, "Verbose execution")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
	    fatal("ERROR: no Command Specified",1)
    }

    switch args[0] {
    case "help":
        flag.Usage()
    case "push":
        if len(args)!=3 {
            fatal("ERROR: invalid number of arguments",1)
        } else {
            push(args[1],args[2])
        }
    case "pull":
        if len(args)!=3 {
            fatal("ERROR: invalid number of arguments",1)
        } else {
            pull(args[1],args[2])
        }
    default:
        fatal("ERROR: unknown command '"+args[0]+"'",1)
    }
}
