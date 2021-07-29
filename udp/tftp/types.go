package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

// TFTP limits datagram packets to 516 bytes or fewer to avoid fragmentation.
const (
    DatagramSize = 516 // max supported size
    BlockSize = DatagramSize - 4 // 4-byte header, 2 of them is an operation code 
)

type OpCode uint64

const (
    OpRRQ OpCode = iota + 1 // Read request
    _ // no WRQ support
    OpData
    OpAck
    OpErr
)

type ErrCode uint16

const (
    ErrUnknown ErrCode = iota
    ErrNotFound
    ErrAccessViolation
    ErrDiskFull
    ErrIllegalOp
    ErrUnknownID
    ErrFileExists
    ErrNoUser
)


type ReadReq struct {
    Filename string
    Mode string
}

// Although not used by our server, a client would make use of this method
func (q ReadReq) MarshalBinary() ([]byte, error) {
    mode := "octet"
    if q.Mode != "" {
        mode = q.Mode
    }

    // operation code + filename + 0 byte + mode + 0 byte
    cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

    b := new(bytes.Buffer)
    b.Grow(cap)

    err := binary.Write(b, binary.BigEndian, OpRRQ) // write operation code
    if err != nil {
        return nil, err
    }
    _, err = b.WriteString(q.Filename) // write filename
    if err != nil {
        return nil, err
    }

    err = b.WriteByte(0) 
    if err != nil { return nil, err }

    _, err = b.WriteString(mode)
    if err != nil {
        return nil, err
    }

    return b.Bytes(), nil
}

func (q *ReadReq) UnmarshalBinary(p []byte) error {
    r := bytes.NewBuffer(p)
    var code OpCode
    err := binary.Read(r, binary.BigEndian, &code) // read operation code
    if err != nil {
        return err
    }
    if code != OpRRQ {
        return errors.New("invalid RRQ")
    }
    q.Filename, err = r.ReadString(0) // read filename
    if err != nil {
        return errors.New("invalid RRQ")
    }
    q.Filename = strings.TrimRight(q.Filename, "\x00") // remove the 0-byte
    if len(q.Filename) == 0 {
        return errors.New("invalid RRQ")
    }
    q.Mode, err = r.ReadString(0) // read mode
    if err != nil {
        return errors.New("invalid RRQ")
    }
    q.Mode = strings.TrimRight(q.Mode, "\x00") // remove the 0-byte
    if len(q.Mode) == 0 {
        return errors.New("invalid RRQ")
    }
    actual := strings.ToLower(q.Mode) // enforce octet mode
    if actual != "octet" {
        return errors.New("only binary transfers supported")
    }
    return nil
}


type Data struct {
    Block uint16
    Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
    b := new(bytes.Buffer)
    b.Grow(DatagramSize)
    
    d.Block++

    err := binary.Write(b, binary.BigEndian, OpData)
    if err != nil {
        return nil, err
    }
    
    err = binary.Write(b,binary.BigEndian, d.Block) 
    if err != nil { return nil, err }

    _, err = io.CopyN(b, d.Payload, BlockSize)
    if err != nil && err != io.EOF {
        return nil, err 
    }

    return b.Bytes(), nil
}