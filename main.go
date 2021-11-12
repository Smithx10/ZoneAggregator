package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type ZoneAggregator struct {
	IP             string          `json:"ip"`
	UDPPort        int             `json:"udp_port"`
	TCPPort        int             `json:"tcp_port"`
	ZoneAggregates []ZoneAggregate `json:"zone_aggregates"`
}

type ZoneAggregate struct {
	Zone  string `json:"zone"`
	TTL   uint32 `json:"ttl"`
	Peers []Peer `json:"peers"`
}

type Peer struct {
	Address string   `json:"address"`
	Zones   []string `json:"zones"`
}

func NewZoneAggregator() (*ZoneAggregator, error) {

	configPath := os.Getenv("CONFIG_PATH")
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var za *ZoneAggregator
	err = json.Unmarshal(configFile, &za)
	if err != nil {
		return nil, err
	}

	dns.HandleFunc(".", za.RequestHandler)

	tcp := dns.Server{Addr: za.IP + ":" + strconv.Itoa(za.TCPPort), Net: "tcp"}
	udp := dns.Server{Addr: za.IP + ":" + strconv.Itoa(za.UDPPort), Net: "udp"}
	// Spin Up Servers
	go tcp.ListenAndServe()
	go udp.ListenAndServe()

	return za, nil
}

func (za *ZoneAggregator) RequestHandler(w dns.ResponseWriter, r *dns.Msg) {
	fmt.Println(r.Question)
	m := za.doAggregation(r)
	w.WriteMsg(m)
}

func (za *ZoneAggregator) doAggregation(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	var answer []dns.RR
	// See if our query matches any of our aggregate zones
	for _, aggr := range za.ZoneAggregates {
		// ToLower the record... Microsoft CNames come in ".NET"
		// Even if .net is the cname ./facepalm.
		r.Question[0].Name = strings.ToLower(r.Question[0].Name)
		if strings.Contains(r.Question[0].Name, aggr.Zone) {
			for _, peer := range aggr.Peers {
				for _, zone := range peer.Zones {
					q := r.Copy()
					// Convert the names from AggrZone to PeerZone
					newName := strings.ReplaceAll(q.Question[0].Name, aggr.Zone, zone)
					q.Question[0].Name = newName
					c := new(dns.Client)
					in, _, err := c.Exchange(q, peer.Address)
					if err != nil {
						fmt.Printf("Error Received during Query: %s", err.Error())
					}

					if len(in.Answer) > 0 {
						// Convert the names from PeerZone to AggrZone
						for _, a := range in.Answer {
							newName := strings.ReplaceAll(a.Header().Name, zone, aggr.Zone)
							a.Header().Name = newName
							a.Header().Ttl = aggr.TTL
							answer = append(answer, a)
						}
					}
				}
			}
		}
	}
	m.Id = r.Id
	m.SetReply(r)
	m.Answer = answer
	m.Authoritative = true
	return m
}

func main() {
	za, err := NewZoneAggregator()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("ZoneAggregator is Running on: %s, TCP: %s, UDP: %s\n", za.IP, strconv.Itoa(za.TCPPort), strconv.Itoa(za.UDPPort))

	// Don't Exit Main
	for {
		time.Sleep(5 * time.Minute)
	}
}
