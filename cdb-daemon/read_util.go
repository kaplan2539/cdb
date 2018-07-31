package main

import(
//    "path/filepath"
//    "path"
    "strings"
    "io/ioutil"
    "strconv"
)

func read_string(path string) (string, error) {
    var dat []byte
    var err error

    if dat, err = ioutil.ReadFile(path); err==nil {
        return strings.TrimSuffix(string(dat),"\n"),err
    }
    return "",err
}

func read_uint64(path string) (uint64, error) {
    var dat []byte
    var err error

    if dat, err = ioutil.ReadFile(path); err==nil {
        return strconv.ParseUint(strings.TrimSuffix(string(dat),"\n"),10,64)
    }
    return 0,err
}


