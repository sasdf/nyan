package main

import (
    "fmt"
    "os"
    "math"
    "time"
    "io"
    "strings"
    "net"
    _ "hash/crc32"
)


const bufSize = 1<<20
const minPrintInterval = 0.1

const usageStr = `
Usage: nyan <command> [args...]

Receive file:
    nyan recv <file> <port> [host]

    Receive data using TCP from host:port and save to <file>.
    If host argument is omitted, it will listen on port <port>.
    If <file> is "-", data will be written to stdout.

Send file:
    nyan send <file> <port> [host]

    Send <file> to using TCP to host:port.
    If host argument is omitted, it will listen on port <port>.
    If <file> is "-", data will be read from stdin.
`

type Pipe interface {
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    Close() error
}

func fatal(err error) {
    fmt.Fprintln(os.Stderr, "[!] Fatal error:", err)
    os.Exit(1)
}

func usage() {
    fmt.Fprintln(os.Stderr, strings.TrimSpace(usageStr))
    os.Exit(1)
}

func connect(port string, host string) net.Conn {
    addr := net.JoinHostPort(host, port)

    if host == "" {
        ln, err := net.Listen("tcp", addr)
        if err != nil { fatal(err) }
        defer ln.Close()

        fmt.Fprintln(os.Stderr, "[*] Listening on", addr)
        conn, err := ln.Accept()
        if err != nil { fatal(err) }
        fmt.Fprintln(os.Stderr, "[+] Connection from", conn.RemoteAddr())
        return conn
    } else {
        conn, err := net.Dial("tcp", addr)
        if err != nil { fatal(err) }
        fmt.Fprintln(os.Stderr, "[+] Connected to", conn.RemoteAddr())
        return conn
    }

}

func open(path string, write bool) (file *os.File, err error) {
    if path == "-" && write { return os.Stdout, nil }
    if path == "-"          { return os.Stdin, nil  }
    if write { return os.Create(path) }
    return os.Open(path)
}

func fmtsi(x float64) string {
    for _, u := range " KMGTP" {
        if x < 1000 { return fmt.Sprintf("%6.2f%c", x, u) }
        x /= 1024
    }
    return fmt.Sprintf("%6.2f%s", x, "E")
}

func fmttime(x_ float64) string {
    x := int64(math.Round(x_))
    return fmt.Sprintf("%02d:%02d:%02d", x/3600, x/60%60, x%60)
}

var lastPrint *time.Time = nil
func progress(cur_ int64, total_ int64, dur_ time.Duration, force bool) {
    now := time.Now()
    if !force && lastPrint != nil && now.Sub(*lastPrint).Seconds() < minPrintInterval {
        return
    }
    lastPrint = &now

    cur, total, dur := float64(cur_), float64(total_), dur_.Seconds()
    if dur < 0.1 { dur = 0.1 }

    speed := cur / dur

    if total_ == 0 {
        fmt.Fprintf(
            os.Stderr,
            "\r[o] %sB [%s] (%sB/s)",
            fmtsi(cur),
            fmttime(dur),
            fmtsi(speed),
            )
    } else {
        ratio := cur / total
        eta := dur / ratio
        fmt.Fprintf(
            os.Stderr,
            "\r[o] %6.2f%% - %sB / %sB [%s / %s] (%sB/s)",
            ratio * 100.0,
            fmtsi(cur), fmtsi(total),
            fmttime(dur), fmttime(eta),
            fmtsi(speed),
            )
    }
}

func pipe(src Pipe, dst Pipe, size int64) {
    var buf [bufSize]byte
    var cur int64 = 0
    // crc32State := crc32.NewIEEE()
    start := time.Now()
    for {
        nbytes, err := src.Read(buf[:])
        if err == io.EOF { break }
        if err != nil { fmt.Fprintf(os.Stderr, "\n"); fatal(err) }
        _, err = dst.Write(buf[:nbytes])
        if err != nil { fmt.Fprintf(os.Stderr, "\n"); fatal(err) }
        // _, err = crc32State.Write(buf[:nbytes])
        // if err != nil { fmt.Fprintf(os.Stderr, "\n"); fatal(err) }

        cur += int64(nbytes)
        progress(cur, size, time.Now().Sub(start), false)
    }
    fmt.Fprintf(os.Stderr, "\n")
    // fmt.Fprintf(os.Stderr, "[+] CRC32: %08x\n", crc32State.Sum32())
}

func main() {
    if len(os.Args) < 4 {
        usage()
    }

    cmd := os.Args[1]
    path := os.Args[2]
    port := os.Args[3]
    host := "";  if len(os.Args) > 4 { host = os.Args[4] }

    var file *os.File
    var err error
    var size int64 = 0

    if cmd == "recv" {
        file, err = open(path, true)
    } else if cmd == "send" {
        file, err = open(path, false)
    } else {
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
        usage()
    }
    if err != nil { fatal(err) }
    defer file.Close()

    info, err := file.Stat()
    if err != nil { fatal(err) }
    size = info.Size()

    if size > 0 {
        fmt.Fprintf(os.Stderr, "[+] File Size: %sB\n", fmtsi(float64(size)))
    }

    conn := connect(port, host)
    defer conn.Close()

    if cmd == "send" {
        pipe(file, conn, size)
    } else if cmd == "recv" {
        pipe(conn, file, size)
    } else {
        fatal(fmt.Errorf("WTF: Unknown command: %s", cmd))
    }
}
