package main

import(
    "path"
)

type MTD struct {
    Path        string
    Dev         string
    Type        string
    Name        string
    Offset      uint64
    Size        uint64
    EraseSize   uint64
    OobSize     uint64
    SubPageSize uint64
    WriteSize   uint64
}

func FromSysFs(sysfs_path string) *MTD {
    mtd:= new(MTD)

    mtd.Path="/dev/"+path.Base(sysfs_path)
    mtd.Dev,_ = read_string(sysfs_path+"/dev")
    mtd.Type,_ = read_string(sysfs_path+"/type")
    mtd.Name,_ = read_string(sysfs_path+"/name")
    mtd.Offset,_ = read_uint64(sysfs_path+"/offset")
    mtd.Size,_ = read_uint64(sysfs_path+"/size")
    mtd.EraseSize,_ = read_uint64(sysfs_path+"/erasesize")
    mtd.OobSize,_ = read_uint64(sysfs_path+"/oobsize")
    mtd.SubPageSize,_ = read_uint64(sysfs_path+"/subpagesize")
    mtd.WriteSize,_ = read_uint64(sysfs_path+"/writesize")

    return mtd
}
