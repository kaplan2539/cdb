package main

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
