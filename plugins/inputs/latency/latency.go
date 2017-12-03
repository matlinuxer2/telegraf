// +build !windows

package latency

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type TimeRec struct {
	t_ns int64
	name string
}

type TimeRecs struct {
	recs []TimeRec
}

func (history *TimeRecs) record(name string) {
	t_ns := time.Now().UnixNano()
	item := TimeRec{t_ns, name}
	history.recs = append(history.recs, item)

	return
}

func (history *TimeRecs) show() {
	for i, curr := range history.recs {
		if i < 1 {
			continue
		}
		prev := history.recs[i-1]

		t_delta := float64(curr.t_ns-prev.t_ns) / 1000000 // in ms

		fmt.Printf("[%3d] %12.3f ms | [ %25s -> %-25s ] \n", i, float64(t_delta), prev.name, curr.name)
	}

	return
}

func (history *TimeRecs) calc() []float64 {
	result := []float64{}

	indeces := []int{2, 4}

	for _, idx := range indeces {
		curr := history.recs[idx]
		prev := history.recs[idx-1]
		t_delta := float64(curr.t_ns-prev.t_ns) / 1000000 // in ms
		result = append(result, t_delta)
	}

	return result
}

type Latency struct {
	Urls []string
}

func (_ *Latency) Description() string {
	return "Get tcp latency"
}

func (_ *Latency) SampleConfig() string {
	sampleConfig := `
  ## latency
  urls = ["192.168.1.1"] # required
`
	return sampleConfig
}

func (p *Latency) Gather(acc telegraf.Accumulator) error {
	history := TimeRecs{}
	history.record("init")

	history.record("connecting")
	raddr, _ := net.ResolveTCPAddr("tcp", "117.18.232.133:80")
	ipconn, _ := net.DialTCP("tcp", nil, raddr)

	history.record("sending")
	_, _ = ipconn.Write([]byte("GET / HTTP/1.1\r\n\r\n"))

	history.record("receiving")
	_, _ = bufio.NewReader(ipconn).ReadString('\n')

	history.record("closing")
	ipconn.Close()
	history.record("exit")

    ret3 := history.calc()

	fields := map[string]interface{}{"result_code": 0}
	tags := map[string]string{"name": "latency"}
	fields["connect"] = ret3[0]
	fields["response"] = ret3[1]
	acc.AddFields("latency", fields, tags)

	return nil
}

func init() {
	inputs.Add("latency", func() telegraf.Input { return &Latency{} })
}
