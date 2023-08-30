package exporter

import (
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

func (e *Exporter) extractClientListMetrics(ch chan<- prometheus.Metric, c redis.Conn) {
	reply, err := redis.String(doRedisCmd(c, "CLIENT", "LIST"))
	if err != nil {
		log.Errorf("CLIENT LIST err: %s", err)
		return
	}
	label := []string{"id", "ip", "port", "age", "idle"}
	for _, clientInfo := range strings.Split(reply, "\n") {
		if matched, _ := regexp.MatchString(`^id=\d+ addr=\d+`, clientInfo); !matched {
			continue
		} else {
			r := make(map[string]string)
			for _, kvPart := range strings.Split(clientInfo, " ") {
				vPart := strings.Split(kvPart, "=")
				if len(vPart) != 2 {
					log.Debugf("Invalid format for client list string, got: %s", kvPart)
					break
				}
				switch vPart[0] {
				case "id":
					r["id"] = vPart[1]
				case "addr":
					hostPortString := strings.Split(vPart[1], ":")
					if len(hostPortString) != 2 {
						break
					}
					r["ip"] = hostPortString[0]
					r["port"] = hostPortString[1]
				case "idle":
					r["idle"] = vPart[1]
				case "age":
					r["age"] = vPart[1]
				}
			}

			if len(r) != 0 {
				e.metricDescriptions["connected_clients_list_idle"] = newMetricDescr(
					e.options.Namespace,
					"connected_clients_list_idle",
					"ccc",
					label)

				res, err := strconv.ParseFloat(r["idle"], 64)
				if err != nil {
					continue
				}

				e.registerConstMetricGauge(
					ch,
					"connected_clients_list_idle",
					res,
					r["id"], r["ip"], r["port"], r["idle"], r["age"])
			}
		}
	}
}
