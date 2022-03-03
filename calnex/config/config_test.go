/*
Copyright (c) Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/facebook/time/calnex/api"
	"github.com/go-ini/ini"
	"github.com/stretchr/testify/require"
)

func TestChSet(t *testing.T) {
	testConfig := `[measure]
ch0\used=Yes
ch1\used=Yes
ch2\used=Yes
ch3\used=Yes
ch4\used=Yes
ch5\used=Yes
`

	expectedConfig := `[measure]
ch0\used=No
ch1\used=No
ch2\used=No
ch3\used=No
ch4\used=No
ch5\used=No
`
	c := config{}

	f, err := ini.Load([]byte(testConfig))
	require.NoError(t, err)

	s := f.Section("measure")
	c.chSet(s, api.ChannelA, api.ChannelF, "%s\\used", api.NO)
	require.True(t, c.changed)

	buf, err := api.ToBuffer(f)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, buf.String())
}

func TestBaseConfig(t *testing.T) {
	testConfig := `[measure]
continuous=Off
meas_time=10 minutes
tie_mode=TIE
ch8\used=No
`

	expectedConfig := `[measure]
continuous=On
meas_time=1 days 1 hours
tie_mode=TIE + 1 PPS TE
ch8\used=Yes
`

	c := config{}

	f, err := ini.Load([]byte(testConfig))
	require.NoError(t, err)

	s := f.Section("measure")

	c.baseConfig(s)
	require.True(t, c.changed)

	buf, err := api.ToBuffer(f)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, buf.String())
}

func TestNicConfig(t *testing.T) {
	testConfig := `[measure]
ch6\used=Yes
ch6\synce_enabled=On
ch6\protocol_enabled=On
ch6\ptp_synce\ptp\dscp=42
ch6\ptp_synce\ethernet\dhcp_v4=Enabled
ch6\ptp_synce\ethernet\dhcp_v6=Disabled
ch6\ptp_synce\ethernet\gateway=192.168.4.1
ch6\ptp_synce\ethernet\gateway_v6=2000::000a
ch6\ptp_synce\ethernet\ip_address=192.168.4.200
ch6\ptp_synce\ethernet\ipv6_address=2000::000b
ch6\ptp_synce\ethernet\mask=255.255.255.0
ch6\ptp_synce\ethernet\mask_v6=32
ch6\ptp_synce\ntp\protocol_level=UDP/IPv4
ch6\ptp_synce\ntp\server_ip_ipv6=
ch6\ptp_synce\mode\probe_type=
ch7\used=Yes
ch7\synce_enabled=On
ch7\protocol_enabled=On
ch7\ptp_synce\ptp\dscp=42
ch7\ptp_synce\ethernet\dhcp_v4=Disabled
ch7\ptp_synce\ethernet\dhcp_v6=Static
ch7\ptp_synce\ethernet\gateway=192.168.5.1
ch7\ptp_synce\ethernet\gateway_v6=2000::000a
ch7\ptp_synce\ethernet\ip_address=192.168.5.200
ch7\ptp_synce\ethernet\ipv6_address=2000::000b
ch7\ptp_synce\ethernet\mask=255.255.255.0
`

	expectedConfig := `[measure]
ch6\used=Yes
ch6\synce_enabled=Off
ch6\protocol_enabled=On
ch6\ptp_synce\ptp\dscp=0
ch6\ptp_synce\ethernet\dhcp_v4=Disabled
ch6\ptp_synce\ethernet\dhcp_v6=Static
ch6\ptp_synce\ethernet\gateway=fd00:3226:310a::a
ch6\ptp_synce\ethernet\gateway_v6=fd00:3226:310a::a
ch6\ptp_synce\ethernet\ip_address=fd00:3226:310a::1
ch6\ptp_synce\ethernet\ipv6_address=fd00:3226:310a::1
ch6\ptp_synce\ethernet\mask=64
ch6\ptp_synce\ethernet\mask_v6=64
ch6\ptp_synce\ntp\protocol_level=UDP/IPv6
ch6\ptp_synce\ntp\server_ip_ipv6=::1
ch6\ptp_synce\mode\probe_type=NTP client
ch7\used=No
ch7\synce_enabled=Off
ch7\protocol_enabled=Off
ch7\ptp_synce\ptp\dscp=0
ch7\ptp_synce\ethernet\dhcp_v4=Disabled
ch7\ptp_synce\ethernet\dhcp_v6=Disabled
ch7\ptp_synce\ethernet\gateway=192.168.5.1
ch7\ptp_synce\ethernet\gateway_v6=2000::000a
ch7\ptp_synce\ethernet\ip_address=192.168.5.200
ch7\ptp_synce\ethernet\ipv6_address=2000::000b
ch7\ptp_synce\ethernet\mask=255.255.255.0
`

	c := config{}

	f, err := ini.Load([]byte(testConfig))
	require.NoError(t, err)

	s := f.Section("measure")

	n := &NetworkConfig{
		Eth1: net.ParseIP("fd00:3226:310a::1"),
		Gw1:  net.ParseIP("fd00:3226:310a::a"),
		Eth2: net.ParseIP("fd00:3226:310a::2"),
		Gw2:  net.ParseIP("fd00:3226:310a::a"),
	}

	c.nicConfig(s, n)
	require.True(t, c.changed)

	buf, err := api.ToBuffer(f)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, buf.String())
}

func TestMeasureConfig(t *testing.T) {
	testConfig := `[measure]
ch0\used=No
ch1\used=Yes
ch2\used=No
ch3\used=Yes
ch4\used=No
ch5\used=Yes
ch6\used=No
ch7\used=Yes
ch8\used=No
ch9\used=No
ch10\used=No
ch11\used=No
ch12\used=No
ch13\used=No
ch14\used=No
ch15\used=No
ch16\used=No
ch17\used=No
ch18\used=No
ch19\used=No
ch20\used=No
ch21\used=No
ch22\used=No
ch23\used=No
ch24\used=No
ch25\used=No
ch26\used=No
ch27\used=No
ch28\used=No
ch29\used=No
ch30\used=No
ch31\used=No
ch32\used=No
ch33\used=No
ch34\used=No
ch35\used=No
ch36\used=No
ch37\used=No
ch38\used=No
ch6\protocol_enabled=Off
ch7\protocol_enabled=Off
ch9\protocol_enabled=Off
ch10\protocol_enabled=Off
ch11\protocol_enabled=Off
ch12\protocol_enabled=Off
ch13\protocol_enabled=Off
ch14\protocol_enabled=Off
ch15\protocol_enabled=Off
ch16\protocol_enabled=Off
ch17\protocol_enabled=Off
ch18\protocol_enabled=Off
ch19\protocol_enabled=Off
ch20\protocol_enabled=Off
ch21\protocol_enabled=Off
ch22\protocol_enabled=Off
ch23\protocol_enabled=Off
ch24\protocol_enabled=Off
ch25\protocol_enabled=Off
ch26\protocol_enabled=Off
ch27\protocol_enabled=Off
ch28\protocol_enabled=Off
ch29\protocol_enabled=Off
ch30\protocol_enabled=Off
ch31\protocol_enabled=Off
ch32\protocol_enabled=Off
ch33\protocol_enabled=Off
ch34\protocol_enabled=Off
ch35\protocol_enabled=Off
ch36\protocol_enabled=Off
ch37\protocol_enabled=Off
ch38\protocol_enabled=Off
ch9\ptp_synce\mode\probe_type=NTP
ch9\ptp_synce\ntp\server_ip=10.32.1.168
ch9\ptp_synce\ntp\server_ip_ipv6=2000::000a
ch9\ptp_synce\physical_packet_channel=Channel 2
ch9\ptp_synce\ntp\normalize_delays=On
ch9\ptp_synce\ntp\protocol_level=UDP/IPv4
ch9\ptp_synce\ntp\poll_log_interval=1 packet/1 s
ch30\ptp_synce\mode\probe_type=PTP
ch30\ptp_synce\ptp\master_ip=10.32.1.168
ch30\ptp_synce\ptp\master_ip_ipv6=2000::000a
ch30\ptp_synce\physical_packet_channel=Channel 2
ch30\ptp_synce\ptp\protocol_level=UDP/IPv4
ch30\ptp_synce\ptp\log_announce_int=1 packet/1 s
ch30\ptp_synce\ptp\log_delay_req_int=1 packet/1 s
ch30\ptp_synce\ptp\log_sync_int=1 packet/1 s
ch30\ptp_synce\ptp\stack_mode=Multicast
ch30\ptp_synce\ptp\domain=0
`

	expectedConfig := `[measure]
ch0\used=No
ch1\used=No
ch2\used=No
ch3\used=No
ch4\used=No
ch5\used=No
ch6\used=No
ch7\used=Yes
ch8\used=No
ch9\used=Yes
ch10\used=No
ch11\used=No
ch12\used=No
ch13\used=No
ch14\used=No
ch15\used=No
ch16\used=No
ch17\used=No
ch18\used=No
ch19\used=No
ch20\used=No
ch21\used=No
ch22\used=No
ch23\used=No
ch24\used=No
ch25\used=No
ch26\used=No
ch27\used=No
ch28\used=No
ch29\used=No
ch30\used=Yes
ch31\used=No
ch32\used=No
ch33\used=No
ch34\used=No
ch35\used=No
ch36\used=No
ch37\used=No
ch38\used=No
ch6\protocol_enabled=Off
ch7\protocol_enabled=Off
ch9\protocol_enabled=On
ch10\protocol_enabled=Off
ch11\protocol_enabled=Off
ch12\protocol_enabled=Off
ch13\protocol_enabled=Off
ch14\protocol_enabled=Off
ch15\protocol_enabled=Off
ch16\protocol_enabled=Off
ch17\protocol_enabled=Off
ch18\protocol_enabled=Off
ch19\protocol_enabled=Off
ch20\protocol_enabled=Off
ch21\protocol_enabled=Off
ch22\protocol_enabled=Off
ch23\protocol_enabled=Off
ch24\protocol_enabled=Off
ch25\protocol_enabled=Off
ch26\protocol_enabled=Off
ch27\protocol_enabled=Off
ch28\protocol_enabled=Off
ch29\protocol_enabled=Off
ch30\protocol_enabled=On
ch31\protocol_enabled=Off
ch32\protocol_enabled=Off
ch33\protocol_enabled=Off
ch34\protocol_enabled=Off
ch35\protocol_enabled=Off
ch36\protocol_enabled=Off
ch37\protocol_enabled=Off
ch38\protocol_enabled=Off
ch9\ptp_synce\mode\probe_type=NTP
ch9\ptp_synce\ntp\server_ip=fd00:3226:301b::3f
ch9\ptp_synce\ntp\server_ip_ipv6=fd00:3226:301b::3f
ch9\ptp_synce\physical_packet_channel=Channel 1
ch9\ptp_synce\ntp\normalize_delays=Off
ch9\ptp_synce\ntp\protocol_level=UDP/IPv6
ch9\ptp_synce\ntp\poll_log_interval=1 packet/16 s
ch30\ptp_synce\mode\probe_type=PTP
ch30\ptp_synce\ptp\master_ip=fd00:3016:3109:face:0:1:0
ch30\ptp_synce\ptp\master_ip_ipv6=fd00:3016:3109:face:0:1:0
ch30\ptp_synce\physical_packet_channel=Channel 1
ch30\ptp_synce\ptp\protocol_level=UDP/IPv6
ch30\ptp_synce\ptp\log_announce_int=1 packet/16 s
ch30\ptp_synce\ptp\log_delay_req_int=1 packet/16 s
ch30\ptp_synce\ptp\log_sync_int=1 packet/16 s
ch30\ptp_synce\ptp\stack_mode=Unicast
ch30\ptp_synce\ptp\domain=0
`

	c := config{}

	f, err := ini.Load([]byte(testConfig))
	require.NoError(t, err)

	s := f.Section("measure")

	mc := map[api.Channel]MeasureConfig{
		api.ChannelVP1: {
			Target: "fd00:3226:301b::3f",
			Probe:  api.ProbeNTP,
		},
		api.ChannelVP22: {
			Target: "fd00:3016:3109:face:0:1:0",
			Probe:  api.ProbePTP,
		},
	}

	c.measureConfig(s, CalnexConfig(mc))
	require.True(t, c.changed)

	buf, err := api.ToBuffer(f)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, buf.String())
}

func TestConfig(t *testing.T) {
	expectedConfig := `[measure]
continuous=On
meas_time=1 days 1 hours
tie_mode=TIE + 1 PPS TE
ch0\used=No
ch1\used=No
ch2\used=No
ch3\used=No
ch4\used=No
ch5\used=No
ch6\used=Yes
ch7\used=No
ch8\used=Yes
ch9\used=Yes
ch10\used=No
ch11\used=No
ch12\used=No
ch13\used=No
ch14\used=No
ch15\used=No
ch16\used=No
ch17\used=No
ch18\used=No
ch19\used=No
ch20\used=No
ch21\used=No
ch22\used=No
ch23\used=No
ch24\used=No
ch25\used=No
ch26\used=No
ch27\used=No
ch28\used=No
ch29\used=No
ch30\used=Yes
ch31\used=No
ch32\used=No
ch33\used=No
ch34\used=No
ch35\used=No
ch36\used=No
ch37\used=No
ch38\used=No
ch6\protocol_enabled=On
ch7\protocol_enabled=Off
ch9\protocol_enabled=On
ch10\protocol_enabled=Off
ch11\protocol_enabled=Off
ch12\protocol_enabled=Off
ch13\protocol_enabled=Off
ch14\protocol_enabled=Off
ch15\protocol_enabled=Off
ch16\protocol_enabled=Off
ch17\protocol_enabled=Off
ch18\protocol_enabled=Off
ch19\protocol_enabled=Off
ch20\protocol_enabled=Off
ch21\protocol_enabled=Off
ch22\protocol_enabled=Off
ch23\protocol_enabled=Off
ch24\protocol_enabled=Off
ch25\protocol_enabled=Off
ch26\protocol_enabled=Off
ch27\protocol_enabled=Off
ch28\protocol_enabled=Off
ch29\protocol_enabled=Off
ch30\protocol_enabled=On
ch31\protocol_enabled=Off
ch32\protocol_enabled=Off
ch33\protocol_enabled=Off
ch34\protocol_enabled=Off
ch35\protocol_enabled=Off
ch36\protocol_enabled=Off
ch37\protocol_enabled=Off
ch38\protocol_enabled=Off
ch6\synce_enabled=Off
ch6\ptp_synce\ptp\dscp=0
ch6\ptp_synce\ethernet\dhcp_v4=Disabled
ch6\ptp_synce\ethernet\dhcp_v6=Static
ch6\ptp_synce\ethernet\gateway=fd00:3226:310a::a
ch6\ptp_synce\ethernet\gateway_v6=fd00:3226:310a::a
ch6\ptp_synce\ethernet\ip_address=fd00:3226:310a::1
ch6\ptp_synce\ethernet\ipv6_address=fd00:3226:310a::1
ch6\ptp_synce\ethernet\mask=64
ch6\ptp_synce\ethernet\mask_v6=64
ch6\ptp_synce\ntp\protocol_level=UDP/IPv6
ch6\ptp_synce\ntp\server_ip_ipv6=::1
ch6\ptp_synce\mode\probe_type=NTP client
ch7\synce_enabled=Off
ch7\ptp_synce\ptp\dscp=0
ch7\ptp_synce\ethernet\dhcp_v4=Disabled
ch7\ptp_synce\ethernet\dhcp_v6=Disabled
ch9\ptp_synce\mode\probe_type=NTP
ch9\ptp_synce\ntp\server_ip=fd00:3226:301b::3f
ch9\ptp_synce\ntp\server_ip_ipv6=fd00:3226:301b::3f
ch9\ptp_synce\physical_packet_channel=Channel 1
ch9\ptp_synce\ntp\normalize_delays=Off
ch9\ptp_synce\ntp\protocol_level=UDP/IPv6
ch9\ptp_synce\ntp\poll_log_interval=1 packet/16 s
ch30\ptp_synce\mode\probe_type=PTP
ch30\ptp_synce\ptp\master_ip=fd00:3016:3109:face:0:1:0
ch30\ptp_synce\ptp\master_ip_ipv6=fd00:3016:3109:face:0:1:0
ch30\ptp_synce\physical_packet_channel=Channel 1
ch30\ptp_synce\ptp\protocol_level=UDP/IPv6
ch30\ptp_synce\ptp\log_announce_int=1 packet/16 s
ch30\ptp_synce\ptp\log_delay_req_int=1 packet/16 s
ch30\ptp_synce\ptp\log_sync_int=1 packet/16 s
ch30\ptp_synce\ptp\stack_mode=Unicast
ch30\ptp_synce\ptp\domain=0
`
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request) {
		if strings.Contains(r.URL.Path, "getsettings") {
			// FetchSettings
			fmt.Fprintln(w, "[measure]\nch0\\used=No\nch6\\used=Yes\nch9\\used=Yes\nch22\\used=Yes")
		} else if strings.Contains(r.URL.Path, "getstatus") {
			// FetchStatus
			fmt.Fprintln(w, "{\n\"referenceReady\": \"true\",\n\"modulesReady\": \"true\",\n\"measurementActive\": \"true\"\n}")
		} else if strings.Contains(r.URL.Path, "stopmeasurement") {
			// StopMeasure
			fmt.Fprintln(w, "{\n\"result\": \"true\"\n}")
		} else if strings.Contains(r.URL.Path, "setsettings") {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			// Config comes back shuffled every time
			require.ElementsMatch(t, strings.Split(expectedConfig, "\n"), strings.Split(string(b), "\n"))
			// PushSettings
			fmt.Fprintln(w, "{\n\"result\": \"true\"\n}")
		} else if strings.Contains(r.URL.Path, "startmeasurement") {
			// StartMeasure
			fmt.Fprintln(w, "{\n\"result\": \"true\"\n}")
		}
	}))
	defer ts.Close()

	parsed, _ := url.Parse(ts.URL)
	calnexAPI := api.NewAPI(parsed.Host, true)
	calnexAPI.Client = ts.Client()

	n := &NetworkConfig{
		Eth1: net.ParseIP("fd00:3226:310a::1"),
		Gw1:  net.ParseIP("fd00:3226:310a::a"),
		Eth2: net.ParseIP("fd00:3226:310a::2"),
		Gw2:  net.ParseIP("fd00:3226:310a::a"),
	}

	mc := map[api.Channel]MeasureConfig{
		api.ChannelVP1: {
			Target: "fd00:3226:301b::3f",
			Probe:  api.ProbeNTP,
		},
		api.ChannelVP22: {
			Target: "fd00:3016:3109:face:0:1:0",
			Probe:  api.ProbePTP,
		},
	}

	err := Config(parsed.Host, true, n, CalnexConfig(mc), true)
	require.NoError(t, err)
}

func TestConfigFail(t *testing.T) {
	n := &NetworkConfig{}
	mc := map[api.Channel]MeasureConfig{}

	err := Config("localhost", true, n, CalnexConfig(mc), true)
	require.Error(t, err)
}

func TestJSONExport(t *testing.T) {
	expected := `{"30":{"target":"fd00:3016:3109:face:0:1:0","probe":0},"9":{"target":"fd00:3226:301b::3f","probe":2}}`
	mc := map[api.Channel]MeasureConfig{
		api.ChannelVP1: {
			Target: "fd00:3226:301b::3f",
			Probe:  api.ProbeNTP,
		},
		api.ChannelVP22: {
			Target: "fd00:3016:3109:face:0:1:0",
			Probe:  api.ProbePTP,
		},
	}

	jsonData, err := json.Marshal(mc)
	require.NoError(t, err)
	require.Equal(t, expected, string(jsonData))
}
