package main

import (
	"io"
	"log"

	"github.com/song940/tuntap-go/tuntap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	IPv4 = 4
	IPv6 = 6
)

func main() {
	config := tuntap.Config{
		DeviceType: tuntap.TUN,
	}
	config.Name = "utun9"
	ifce, err := tuntap.New(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Interface Name: %s\n", ifce.Name())

	buf := make([]byte, 1500)
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		go process(buf[:n], ifce)
	}
}

func process(p []byte, w io.WriteCloser) {
	version := p[0] >> 4
	switch version {
	case IPv4:
		hdr, _ := ipv4.ParseHeader(p)
		handleIPv4(w, hdr, p[hdr.Len:])
	case 6:
		hdr, _ := ipv6.ParseHeader(p)
		handleIPv6(w, hdr, p[40:])
	default:
		log.Println("Received non-IP packet")
	}
}

func handleIPv4(w io.WriteCloser, header *ipv4.Header, payload []byte) {
	log.Printf("Received IPv4 packet from %v to %v", header.Src, header.Dst)
	switch header.Protocol {
	case 1:
		handleICMPPacket(payload, w)
	case 6:
		log.Println("Received TCP packet")
	case 17:
		log.Println("Received UDP packet")
	default:
		log.Printf("Received packet with unknown protocol: %d", header.Protocol)
	}
}

func handleIPv6(w io.WriteCloser, header *ipv6.Header, payload []byte) {
	log.Printf("Received IPv6 packet from %v to %v", header.Src, header.Dst)
}

func handleICMPPacket(b []byte, w io.WriteCloser) {
	msg, _ := icmp.ParseMessage(1, b)
	switch msg.Type {
	case ipv4.ICMPTypeEcho:
		echo := msg.Body.(*icmp.Echo)
		log.Println("Received ICMP packet", echo.ID, echo.Seq, echo.Data)
		// reply := icmp.Message{
		// 	Type: ipv4.ICMPTypeEchoReply,
		// 	Code: 0,
		// 	Body: msg.Body,
		// }
		// replyBytes, _ := reply.Marshal(nil)
		// w.Write(replyBytes)
	case ipv4.ICMPTypeEchoReply:
		echo := msg.Body.(*icmp.Echo)
		log.Println("Received ICMP packet", echo.ID, echo.Seq, echo.Data)
	default:
		log.Println("ICMP:", msg.Type, msg.Code)
	}
}
