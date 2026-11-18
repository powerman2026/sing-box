package group

import (
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/interrupt"
	"github.com/sagernet/sing-box/common/urltest"
	N "github.com/sagernet/sing/common/network"
)

type testOutbound struct {
	adapter.Outbound
	tag string
}

func (o *testOutbound) Tag() string { return o.tag }

func (o *testOutbound) Network() []string { return []string{N.NetworkTCP, N.NetworkUDP} }

func TestConcurrentURLTestSelection(t *testing.T) {
	first := &testOutbound{tag: "first"}
	second := &testOutbound{tag: "second"}
	group := &URLTestGroup{
		outbounds:      []adapter.Outbound{first, second},
		history:        urltest.NewHistoryStorage(),
		interruptGroup: interrupt.NewGroup(),
	}
	out := &URLTest{group: group}
	group.history.StoreURLTestHistory(first.Tag(), &adapter.URLTestHistory{Delay: 10})
	group.performUpdateCheck()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			out.Selected(N.NetworkTCP)
			out.Selected(N.NetworkUDP)
			out.References()
		}
	}()
	group.history.StoreURLTestHistory(second.Tag(), &adapter.URLTestHistory{Delay: 1})
	group.performUpdateCheck()
	<-done

	if out.Selected(N.NetworkTCP) != second || out.Selected(N.NetworkUDP) != second {
		t.Fatal("expected the faster outbound to be selected for both networks")
	}
}
