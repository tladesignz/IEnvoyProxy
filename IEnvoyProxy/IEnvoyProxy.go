package IEnvoyProxy

import (
	"errors"
	"io/fs"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"fmt"
	hysteria2 "github.com/apernet/hysteria/app/cmd"
	v2ray "github.com/v2fly/v2ray-core/envoy"
	"gitlab.com/stevenmcdonald/tubesocks"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/cmd/lyrebird"
	snowflakeclient "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client"
)

var meekPort = 47000

// MeekPort - Port where Lyrebird will provide its Meek service.
// Only use this after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func MeekPort() int {
	return meekPort
}

// This functionality is disabled, but values are required. Values are ignored
var obfs2Port = 47100
var obfs3Port = 47200
var scramblesuitPort = 47400

var obfs4Port = 47300

// Obfs4Port - Port where Lyrebird will provide its Obfs4 service.
// Only use this property after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func Obfs4Port() int {
	return obfs4Port
}

var obfs4TubeSocksPort = 47350

// Obfs4TubeSocksPort - Port where TubeSocks will listen to forward to Lyrebird's Obfs4 service.
// Only use this property after calling StartObfs4! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func Obfs4TubeSocksPort() int {
	return obfs4TubeSocksPort
}

var meekTubeSocksPort = 47360

// MeekTubeSocksPort - Port where TubeSocks will listen to forward to Lyrebird's Meek service.
// Only use this property after calling StartMeek! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func MeekTubeSocksPort() int {
	return meekTubeSocksPort
}

var webtunnelPort = 47500

// WebtunnelPort - Port where Lyrebird will provide its Webtunnel service.
// Only use this property after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func WebtunnelPort() int {
	return webtunnelPort
}

var v2raySrtpPort = 47600
var v2rayWechatPort = 47700
var v2rayWsPort = 47800
var snowflakePort = 47900

//goland:noinspection GoUnusedExportedFunction
func V2raySrtpPort() int {
	return v2raySrtpPort
}

//goland:noinspection GoUnusedExportedFunction
func V2rayWechatPort() int {
	return v2rayWechatPort
}

//goland:noinspection GoUnusedExportedFunction
func V2rayWsPort() int {
	return v2rayWsPort
}

var hysteria2Port = 48000

//goland:noinspection GoUnusedExportedFunction
func Hysteria2Port() int {
	return hysteria2Port
}

// SnowflakePort - Port where Snowflake will provide its service.
// Only use this property after calling StartSnowflake! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func SnowflakePort() int {
	return snowflakePort
}

var lyrebirdRunning = false
var v2rayWsRunning = false
var v2raySrtpRunning = false
var v2rayWechatRunning = false
var snowflakeRunning = false
var hysteria2Running = false

// StateLocation - Sets TOR_PT_STATE_LOCATION
var StateLocation string

/// Lyrebird (forked from obfs4proxy)

// LyrebirdLogFile - The log file name used by Lyrebird.
//
// The Lyrebird log file can be found at `filepath.Join(StateLocation, LyrebirdLogFile())`.
//
//goland:noinspection GoUnusedExportedFunction
func LyrebirdLogFile() string {
	return lyrebird.LyrebirdLogFile
}

// StartLyrebird - Start Lyrebird.
//
// This will test, if the default ports are available. If not, it will increment them until there is.
// Only use the port properties after calling this, they might have been changed!
//
// @param logLevel Log level (ERROR/WARN/INFO/DEBUG). Defaults to ERROR if empty string.
//
// @param enableLogging Log to TOR_PT_STATE_LOCATION/lyrebird.log.
//
// @param unsafeLogging Disable the address scrubber.
//
// @return Port number where Lyrebird will listen on for Obfs4(!), if no error happens during start up.
//
//	If you need the other ports, check MeekPort, Obfs2Port, Obfs3Port, ScramblesuitPort and WebtunnelPort properties!
//
//goland:noinspection GoUnusedExportedFunction
func StartLyrebird(logLevel string, enableLogging, unsafeLogging bool) int {
	if lyrebirdRunning {
		return obfs4Port
	}

	lyrebirdRunning = true

	// we disable everything but obfs4 and meek_lite in TOR_PT_CLIENT_TRANSPORTS
	// so their settings are ignored

	meekPort = findPort(meekPort)
	obfs4Port = findPort(obfs4Port)
	webtunnelPort = findPort(webtunnelPort)

	fixEnv()

	go lyrebird.Start(&meekPort, &obfs2Port, &obfs3Port, &obfs4Port, &scramblesuitPort, &webtunnelPort, &logLevel, &enableLogging, &unsafeLogging)

	return obfs4Port
}

////////
// XXX
// This is probably not the ideal way to do things, but it's expedient.
// We've been unable to configure cronet to use a socks proxy that requires
// auth info, tubesocks bridges that gap by running a second socks proxy.
// It would probably be better to patch the Lyrebird code to take the auth
// info as a parameter to StartObfs4/StartMeek() for us, but that requires more
// invasive changes. Todo maybe?

//goland:noinspection GoUnusedExportedFunction
func StartObfs4(user, password, logLevel string, enableLogging, unsafeLogging bool) int {
	if !lyrebirdRunning {
		StartLyrebird(logLevel, enableLogging, unsafeLogging)
	}

	obfs4TubeSocksPort = findPort(obfs4TubeSocksPort)
	var obfs4Url = "127.0.0.1:" + strconv.Itoa(obfs4Port)

	go tubesocks.Start(user, password, obfs4Url, obfs4TubeSocksPort)

	return obfs4TubeSocksPort
}

//goland:noinspection GoUnusedExportedFunction
func StartMeek(user, password, logLevel string, enableLogging, unsafeLogging bool) int {
	if !lyrebirdRunning {
		StartLyrebird(logLevel, enableLogging, unsafeLogging)
	}

	meekTubeSocksPort = findPort(meekTubeSocksPort)
	var meekUrl = "127.0.0.1:" + strconv.Itoa(meekPort)

	go tubesocks.Start(user, password, meekUrl, meekTubeSocksPort)

	return meekTubeSocksPort
}

// StopLyrebird - Stop Lyrebird.
//
//goland:noinspection GoUnusedExportedFunction
func StopLyrebird() {
	if !lyrebirdRunning {
		return
	}

	go lyrebird.Stop()

	lyrebirdRunning = false
}

/// V2Ray

// StartV2RayWs - Start V2Ray client for websocket transport
//
// @param serverAddress - Hostname of WS web server proxy
//
// @oaram serverPort - Port of the WS listener (probably 443)
//
// @param wsPath - path the websocket
//
// @param id - v2ray UUID for auth
//
//goland:noinspection GoUnusedExportedFunction
func StartV2RayWs(serverAddress, serverPort, wsPath, id string) int {
	if v2rayWsRunning {
		return v2rayWsPort
	}

	v2rayWsPort = findPort(v2rayWsPort)
	clientPort := strconv.Itoa(v2rayWsPort)

	v2rayWsRunning = true

	go v2ray.StartWs(&clientPort, &serverAddress, &serverPort, &wsPath, &id)

	return v2rayWsPort
}

//goland:noinspection GoUnusedExportedFunction
func StopV2RayWs() {
	if !v2rayWsRunning {
		return
	}

	go v2ray.StopWs()

	v2rayWsRunning = false
}

//goland:noinspection GoUnusedExportedFunction
func StartV2raySrtp(serverAddress, serverPort, id string) int {
	log.Println("Starting V2Ray SRTP")
	if v2raySrtpRunning {
		log.Printf("V2Ray SRTP already running on %d", v2raySrtpPort)
		return v2raySrtpPort
	}

	v2raySrtpPort = findPort(v2raySrtpPort)
	clientPort := strconv.Itoa(v2raySrtpPort)

	v2raySrtpRunning = true

	go v2ray.StartSrtp(&clientPort, &serverAddress, &serverPort, &id)
	log.Printf("V2Ray SRTP started on %d", v2raySrtpPort)

	return v2raySrtpPort
}

//goland:noinspection GoUnusedExportedFunction
func StopV2RaySrtp() {
	if !v2raySrtpRunning {
		return
	}

	go v2ray.StopSrtp()

	v2raySrtpRunning = false
}

//goland:noinspection GoUnusedExportedFunction
func StartV2RayWechat(serverAddress, serverPort, id string) int {
	log.Println("Starting V2Ray WeChat")
	if v2rayWechatRunning {
		log.Printf("V2Ray WeChat already running on %d", v2rayWechatPort)
		return v2rayWechatPort
	}

	v2rayWechatPort = findPort(v2rayWechatPort)
	clientPort := strconv.Itoa(v2rayWechatPort)

	v2rayWechatRunning = true

	go v2ray.StartWechat(&clientPort, &serverAddress, &serverPort, &id)
	log.Printf("V2Ray WeChat started on %d", v2rayWechatPort)

	return v2rayWechatPort
}

//goland:noinspection GoUnusedExportedFunction
func StopV2RayWechat() {
	if !v2rayWechatRunning {
		return
	}

	go v2ray.StopWechat()

	v2rayWechatRunning = false
}

/// Hysteria2

// StartHysteria2 - Start the Hysteria2 client.
//
// @param server A Hysteria2 server URL https://v2.hysteria.network/docs/developers/URI-Scheme/
//
// @return Port number where Hysteria2 will listen on, if no error happens during start up.
//
//goland:noinspection GoUnusedExportedFunction
func StartHysteria2(server string) int {
	if hysteria2Running {
		return hysteria2Port
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not get home dir: %s\n", err)

		return 0
	}

	err = os.MkdirAll(fmt.Sprintf("%s/.hysteria", home), 0755)
	if err != nil {
		log.Printf("Could not create home dir: %s\n", err)

		return 0
	}

	hysteria2Port = findPort(hysteria2Port)

	err = os.WriteFile(fmt.Sprintf("%s/.hysteria/config", home),
		[]byte(fmt.Sprintf("server: %s\n\nsocks5:\n  listen: 127.0.0.1:%d\n", server, hysteria2Port)), 0644)
	if err != nil {
		log.Printf("Could not write config file: %s\n", err)

		return 0
	}

	hysteria2Running = true

	go hysteria2.Start()

	// Need to sleep a little here, to give Hysteria2 a chance to start,
	// before we return the port. Otherwise, Hysteria2 wouldn't be listening
	// on that configured SOCKS5 port, yet and connections would fail.
	time.Sleep(time.Second)

	return hysteria2Port
}

//goland:noinspection GoUnusedExportedFunction
func StopHysteria2() {
	if !hysteria2Running {
		return
	}

	go hysteria2.Stop()

	home, err := os.UserHomeDir()

	if err == nil {
		_ = os.Remove(fmt.Sprintf("%s/.hysteria/config", home))
	}

	hysteria2Running = false
}

/// Snowflake

// StartSnowflake - Start the Snowflake client.
//
// @param ice Comma-separated list of ICE servers.
//
// @param url URL of signaling broker.
//
// @param fronts Comma-separated list of front domains.
//
// @param ampCache OPTIONAL. URL of AMP cache to use as a proxy for signaling.
//
//	Only needed when you want to do the rendezvous over AMP instead of a domain fronted server.
//
// @param sqsQueueURL OPTIONAL. URL of SQS Queue to use as a proxy for signaling.
//
// @param sqsCredsStr OPTIONAL. Credentials to access SQS Queue
//
// @param logFile Name of log file. OPTIONAL. Defaults to no log.
//
// @param logToStateDir Resolve the log file relative to Tor's PT state dir.
//
// @param keepLocalAddresses Keep local LAN address ICE candidates.
//
// @param unsafeLogging Prevent logs from being scrubbed.
//
// @param maxPeers Capacity for number of multiplexed WebRTC peers. DEFAULTs to 1 if less than that.
//
// @return Port number where Snowflake will listen on, if no error happens during start up.
//
//goland:noinspection GoUnusedExportedFunction
func StartSnowflake(ice, url, fronts, ampCache, sqsQueueURL, sqsCredsStr, logFile string,
	logToStateDir, keepLocalAddresses, unsafeLogging bool,
	maxPeers int) int {

	if snowflakeRunning {
		return snowflakePort
	}

	snowflakeRunning = true

	for !IsPortAvailable(snowflakePort) {
		snowflakePort++
	}

	fixEnv()

	go snowflakeclient.Start(&snowflakePort, &ice, &url, &fronts, &ampCache, &sqsQueueURL, &sqsCredsStr,
		&logFile, &logToStateDir, &keepLocalAddresses, &unsafeLogging, &maxPeers)

	return snowflakePort
}

// StopSnowflake - Stop the Snowflake client.
//
//goland:noinspection GoUnusedExportedFunction
func StopSnowflake() {
	if !snowflakeRunning {
		return
	}

	go snowflakeclient.Stop()

	snowflakeRunning = false
}

// SnowflakeClientConnected - Interface to use when clients connect
// to the snowflake proxy. For use with StartSnowflakeProxy
type SnowflakeClientConnected interface {
	// Connected - callback method to handle snowflake proxy client connections.
	Connected()
}

///////////////////
// Helper functions

// Hack: Set some environment variables that are either
// required, or values that we want. Have to do this here, since we can only
// launch this in a thread and the manipulation of environment variables
// from within an iOS app won't end up in goptlib properly.
//
// Note: This might be called multiple times when using different functions here,
// but that doesn't necessarily mean, that the values set are independent each
// time this is called. It's still the ENVIRONMENT, we're changing here, so there might
// be race conditions.
func fixEnv() {
	info, err := os.Stat(StateLocation)

	// If dir does not exist, try to create it.
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(StateLocation, 0700)

		if err == nil {
			info, err = os.Stat(StateLocation)
		}
	}

	// If it is not a dir, panic.
	if err == nil && !info.IsDir() {
		err = fs.ErrInvalid
	}

	// Create a file within dir to test writability.
	if err == nil {
		tempFile := StateLocation + "/.iptproxy-writetest"
		var file *os.File
		file, err = os.Create(tempFile)

		// Remove the test file again.
		if err == nil {
			_ = file.Close()

			err = os.Remove(tempFile)
		}
	}

	if err != nil {
		panic("Error with StateLocation directory \"" + StateLocation + "\":\n" +
			"  " + err.Error() + "\n" +
			"  StateLocation needs to be set to a writable directory.\n" +
			"  Use an app-private directory to avoid information leaks.\n" +
			"  Use a non-temporary directory to allow reuse of potentially stored state.")
	}

	_ = os.Setenv("TOR_PT_CLIENT_TRANSPORTS", "meek_lite,obfs4,webtunnel,snowflake")
	_ = os.Setenv("TOR_PT_MANAGED_TRANSPORT_VER", "1")
	_ = os.Setenv("TOR_PT_STATE_LOCATION", StateLocation)
}

func findPort(port int) int {
	temp := port
	for !IsPortAvailable(temp) {
		temp++
	}
	return temp
}

// IsPortAvailable - Checks to see if a given port is not in use.
//
// @param port The port to check.
func IsPortAvailable(port int) bool {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)

	if err != nil {
		return true
	}

	_ = conn.Close()

	return false
}
