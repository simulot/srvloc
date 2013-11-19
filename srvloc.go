// srvloc project srvloc.go
package srvloc

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"time"
)

type svrlocQuery struct {
	header  srvlocHeader
	payload srvlocQueryPayload
}

type srvlocHeader struct {
	version       byte
	function      byte
	length        int16
	flag          byte
	dialect       byte
	language      string
	encoding      int16
	transactionID int16
}

type srvlocQueryPayload struct {
	previousResponseList string
	serviceURL           string
	scopList             string
	attributeList        string
}

type srvlocResponse struct {
	header  srvlocHeader
	payload srvlocResponsePayload
}

type srvlocResponsePayload struct {
	errorCode     uint16
	attributeList map[string]string
}

// Place a string in the buffer
func writeString(buf *bytes.Buffer, s string) {
	for i := range s {
		binary.Write(buf, binary.BigEndian, s[i])
	}
}

// Read a string having length from the buffer
func readString(buf *bytes.Buffer, length int) string {
	return string(buf.Next(length))
}

// Read string length and then string from the buffer
func readLenString(buf *bytes.Buffer) string {
	var length int16
	binary.Read(buf, binary.BigEndian, &length)
	return readString(buf, int(length))
}

// Write string length and string in the buffer
func writeLenString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.BigEndian, uint16(len(s)))
	writeString(buf, s)
}

// Write wrvlocHeader structure into the buffer with respect of integers' Endianness
func (p *srvlocHeader) write(buf *bytes.Buffer) {
	binary.Write(buf, binary.BigEndian, p.version)
	binary.Write(buf, binary.BigEndian, p.function)
	binary.Write(buf, binary.BigEndian, p.length)
	binary.Write(buf, binary.BigEndian, p.flag)
	binary.Write(buf, binary.BigEndian, p.dialect)
	writeString(buf, p.language)
	binary.Write(buf, binary.BigEndian, p.encoding)
	binary.Write(buf, binary.BigEndian, p.transactionID)
}

// Read rvlocHeader structure from the buffer with respect of integers' Endianness
func (p *srvlocHeader) read(buf *bytes.Buffer) {
	binary.Read(buf, binary.BigEndian, &p.version)
	binary.Read(buf, binary.BigEndian, &p.function)
	binary.Read(buf, binary.BigEndian, &p.length)
	binary.Read(buf, binary.BigEndian, &p.flag)
	binary.Read(buf, binary.BigEndian, &p.dialect)
	p.language = readString(buf, 2)
	binary.Read(buf, binary.BigEndian, &p.encoding)
	binary.Read(buf, binary.BigEndian, &p.transactionID)
}

// Write svrlocQuery structure into the buffer with respect of integers' Endianness
func (p *svrlocQuery) write() (buf bytes.Buffer) {
	payload := new(bytes.Buffer)

	writeLenString(payload, p.payload.previousResponseList)
	writeLenString(payload, p.payload.serviceURL)
	writeLenString(payload, p.payload.scopList)
	writeLenString(payload, p.payload.attributeList)
	p.header.length = int16(payload.Len()) + 12 // 12: size of the header
	p.header.write(&buf)
	buf.Write(payload.Bytes())
	return
}

// Read srvlocResponse structure from the buffer with respect of integers' Endianness
func (p *srvlocResponse) read(buf *bytes.Buffer) {
	p.header.read(buf)
	binary.Read(buf, binary.BigEndian, &p.payload.errorCode)
	s := strings.Split(readLenString(buf), ")")
	if len(s) == 0 {
		return
	}
	p.payload.attributeList = make(map[string]string)

	for _, v := range s {
		if len(v) == 0 {
			continue
		}
		v = v[1:] // Skip "("
		pair := strings.Split(v, "=")
		if strings.Count(pair[1], ";") == 0 {
			p.payload.attributeList[pair[0]] = pair[1]
		} else {
			s := strings.Split(pair[1], ";")
			for _, v := range s {
				if len(v) > 0 {
					pair := strings.Split(v, ":")
					p.payload.attributeList[pair[0]] = pair[1]
				}
			}
		}

	}
}

// Manage sequences for queries
func newSequence() func() int16 {
	var Sequence int16 = 0
	return func() (s int16) {
		s = Sequence
		Sequence++
		return
	}
}

// Create an srvlocQuery frame for the given service
func newsrvlocQuery(service string) *svrlocQuery {
	return &svrlocQuery{
		header: srvlocHeader{
			version:  1, //byte
			function: 6, //byte
			//Length:        0,     //int16
			flag:     0,    //byte
			dialect:  0,    //byte
			language: "en", //[2]byte
			encoding: 3,    //int16
			//transactionID: 0, //int16
		},
		payload: srvlocQueryPayload{
			previousResponseList: "",      //string
			serviceURL:           service, //string
			scopList:             "",      //string
			attributeList:        "",      //string
		},
	}
}

// Sequence will be continued along program lifespan
var transctionSequence = newSequence()

func sendsrvlocQuery(service string) (*srvlocResponse, error) {

	Qframe := newsrvlocQuery(service)
	Rframe := new(srvlocResponse)
	bigBuffer := make([]byte, 1500)
	var RIP *net.UDPAddr

	multiCastAddr, err := net.ResolveUDPAddr("udp", "224.0.1.60:427")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	Qframe.header.transactionID = transctionSequence()
	buf := Qframe.write()
	_, err = conn.WriteTo(buf.Bytes(), multiCastAddr)
	if err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, RIP, err = conn.ReadFromUDP(bigBuffer)
	buf2 := bytes.NewBuffer(bigBuffer)
	_ = RIP
	if err != nil {
		return nil, err
	}

	Rframe.read(buf2)
	return Rframe, nil
}

const (
	HPStatusProbing = iota
	HPStatusOK
	HPStatusUnreachable
)

type HPDevice struct {
	IPAddress    string
	URL          string
	Model        string
	SerialNumber string
	HostName     string
	Attributes   map[string]string
	Status       int
	LastSeen     time.Time
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

/* Send SRVLOC query on the network and listen HP printers
TODO: Manage several printers on the network
*/
func ProbeHPPrinter() (*HPDevice, error) {

	f, err := sendsrvlocQuery("service:x-hpnp-discover:")
	if err != nil {
		return nil, err
	}

	dev := new(HPDevice)
	dev.LastSeen = time.Now()
	dev.Model = f.payload.attributeList["MDL"]
	dev.SerialNumber = f.payload.attributeList["SN"]
	dev.IPAddress = f.payload.attributeList["x-hp-ip"]
	dev.URL = f.payload.attributeList["x-hp-ip"]
	dev.HostName = f.payload.attributeList["x-hp-hn"]
	dev.Status = HPStatusOK
	dev.Attributes = f.payload.attributeList
	return dev, nil

}
