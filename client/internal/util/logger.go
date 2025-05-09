package util

import (
	"errors"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func LogIPv4Packet(data []byte) {
	packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		log.Printf("Not an IPv4 packet")
		return
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		log.Printf("Failed to cast to IPv4 layer")
		return
	}

	switch {
	case packet.Layer(layers.LayerTypeTCP) != nil:
		if err := logTCPPacket(packet.Layer(layers.LayerTypeTCP), ip); err != nil {
			log.Print(err)
		}
	case packet.Layer(layers.LayerTypeUDP) != nil:
		if err := logUDPPacket(packet.Layer(layers.LayerTypeUDP), ip); err != nil {
			log.Print(err)
		}
	default:
		log.Printf("strange %s", ip.Protocol)
	}
}

func logTCPPacket(tcpLayer gopacket.Layer, ip *layers.IPv4) error {
	tcp, ok := tcpLayer.(*layers.TCP)
	if !ok {
		return errors.New("Failed to cast to TCP layer")
	}

	log.Printf("TCP Packet: %s:%d -> %s:%d", ip.SrcIP, tcp.SrcPort, ip.DstIP, tcp.DstPort)
	return nil
}

func logUDPPacket(udpLayer gopacket.Layer, ip *layers.IPv4) error {
	udp, ok := udpLayer.(*layers.UDP)
	if !ok {
		return errors.New("Failed to cast to UDP layer")
	}

	log.Printf("UDP Packet: %s:%d -> %s:%d", ip.SrcIP, udp.SrcPort, ip.DstIP, udp.DstPort)
	return nil
}
