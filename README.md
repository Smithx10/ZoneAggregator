
### Zone Aggregator is used to aggregate Multiple DNS Zones on a DNS Server into One.

### Examples:

You have the following 4 Records.
 * dns-test.us-east-1.example.com. 
 * dns-test.us-east-2.example.com.
 * dns-test.us-east-3.example.com.
 * dns-test.us-east-4.example.com.

But you want all  4 of them to be resolvable under 1 record, under us-east.example.com.
 * dns-test.us-east.example.com.


Run the ZoneAggregator Binaray with the following config.
 * "zone" is the Aggregated Zone of all the zones provided.
 * "address" is the address of the DNS server who resolves these domains.
 * "ttl" is the ttl that will be returned with the transformed records.

```
{
  "ip": "127.0.0.1",
  "udp_port": 9000,
  "tcp_port": 9000,
  "zone_aggregates": [
    {
        "zone": "us-east.example.com.",
        "ttl": 0,
        "peers": [
          {
            "address": "127.0.0.1:53",
            "zones": [
             "us-east-1.example.com.",
             "us-east-2.example.com.",
             "us-east-3.example.com.",
             "us-east-4.example.com."
            ]
          }
        ]
    }
  ]
}
```

```
dig dns-test.us-east.example.com. +short 
10.20.30.40
10.20.31.40
10.20.32.40
10.20.33.40
```


I usually run this binary on my Nameserver, and use SubZone Delegation.
http://www.zytrax.com/books/dns/ch9/delegate.html
