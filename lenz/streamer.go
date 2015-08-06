package lenz

import (
	"math"

	"github.com/HunanTV/eru-agent/defines"
	"github.com/HunanTV/eru-agent/logs"
)

func Streamer(route *defines.Route, logstream chan *defines.Log, stdout bool) {
	var upstreams map[string]*UpStream = map[string]*UpStream{}
	var types map[string]struct{}
	var count int64 = 0
	if route.Source != nil {
		types = make(map[string]struct{})
		for _, t := range route.Source.Types {
			types[t] = struct{}{}
		}
	}
	defer func() {
		for _, remote := range upstreams {
			remote.Flush()
			for _, log := range remote.Tail() {
				logs.Info("Streamer can't send to remote", log)
			}
			remote.Close()
		}
	}()
	for logline := range logstream {
		if types != nil {
			if _, ok := types[logline.Type]; !ok {
				continue
			}
		}
		logline.Tag = route.Target.AppendTag
		logline.Count = count
		switch stdout {
		case true:
			logs.Info("Debug Output", logline)
		default:
			for offset := 0; offset < route.Backends.Len(); offset++ {
				addr, err := route.Backends.Get(logline.Name, offset)
				if err != nil {
					logs.Info("Get backend failed", err, logline.Name, logline.Data)
					break
				}
				if _, ok := upstreams[addr]; !ok {
					if ups, err := NewUpStream(addr); err != nil || ups == nil {
						route.Backends.Remove(addr)
						continue
					} else {
						upstreams[addr] = ups
					}
				}
				if err := upstreams[addr].WriteData(logline); err != nil {
					upstreams[addr].Close()
					for _, log := range upstreams[addr].Tail() {
						logstream <- log
					}
					delete(upstreams, addr)
					continue
				}
				//logs.Debug("Lenz Send", logline.Name, logline.EntryPoint, logline.ID, "to", addr)
				break
			}
		}
		if count == math.MaxInt64 {
			count = 0
		} else {
			count++
		}
	}
}
