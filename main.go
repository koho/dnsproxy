package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/thinkgos/go-socks5/ccsocks5"
	"log"
	"net"
	"os"
	"strings"
)

var (
	domainName = flag.String("domain", "", "domain")
	dnsServer = flag.String("server", "8.8.8.8:53", "dns server")
	proxyServer = flag.String("proxy", "", "proxy server")
)

func printAnswer(name string, ans string) {
	printServers()
	fmt.Println()
	fmt.Printf("名称: %s\n", name)
	fmt.Printf("地址: %s\n", ans)
}

func query(domain string) (string, string, error) {
	var con net.Conn
	var err error
	if *proxyServer != "" {
		client := ccsocks5.NewClient(*proxyServer)
		defer client.Close()
		con, err = client.Dial("udp", *dnsServer)
	} else {
		con, err = net.Dial("udp", *dnsServer)
	}
	if err != nil {
		return domain, "", err
	}
	defer con.Close()
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}
	gopacket.SerializeLayers(buffer, options, &layers.DNS{
		ID:           1,
		OpCode:       layers.DNSOpCodeQuery,
		RD:           true,
		QDCount:      1,
		Questions:    []layers.DNSQuestion{{Name: []byte(domain), Type: layers.DNSTypeA, Class: layers.DNSClassIN}},
	})
	dnsPacket := buffer.Bytes()
	_, err = con.Write(dnsPacket)
	if err != nil {
		return domain, "", err
	}
	ans := make([]byte, 2048)
	_, err = con.Read(ans)
	if err != nil {
		return domain, "", err
	}
	packet := gopacket.NewPacket(ans, layers.LayerTypeDNS, gopacket.Default)
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	if dnsLayer != nil {
		dnsAns, _ := dnsLayer.(*layers.DNS)
		if len(dnsAns.Answers) > 0 {
			return string(dnsAns.Answers[0].Name), dnsAns.Answers[0].IP.String(), nil
		} else {
			return domain, "", nil
		}
	}
	return domain, "", nil
}

func printServers() {
	fmt.Printf("服务器: %s\n", *dnsServer)
	fmt.Printf("代理: %s\n", *proxyServer)
}


func main() {
	flag.Parse()
	if !strings.Contains(*dnsServer, ":") {
		*dnsServer = *dnsServer + ":53"
	}
	if *proxyServer != "" && !strings.Contains(*proxyServer, ":") {
		*proxyServer = *proxyServer + ":7890"
	}
	fmt.Println("DNS 查询")
	fmt.Println()
	if *domainName != "" {
		name, addr, err := query(*domainName)
		if err != nil {
			log.Fatal(err)
		}
		printAnswer(name, addr)
		return
	}
	printServers()
	for {
		fmt.Println()
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		cmd = strings.TrimSpace(cmd)
		command := strings.Split(cmd, " ")
		switch command[0] {
		case "server":
			*dnsServer = command[1]
			if !strings.Contains(*dnsServer, ":") {
				*dnsServer = *dnsServer + ":53"
			}
			printServers()
		case "proxy":
			*proxyServer = command[1]
			if !strings.Contains(*proxyServer, ":") {
				*proxyServer = *proxyServer + ":7890"
			}
			printServers()
		case "exit":
			return
		default:
			name, addr, err := query(command[0])
			if err != nil {
				log.Println(err)
				println()
				continue
			}
			printAnswer(name, addr)
		}
	}
}
