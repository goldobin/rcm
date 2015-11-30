package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestSaveLoadClusterConf(t *testing.T) {

	cases := []ClusterConf{
		ClusterConf{
			ListenHost:  "127.0.0.1",
			ListenPorts: []int{7501, 7502, 7503, 7504, 7505, 7506},
			Persistence: true,
		},
		ClusterConf{
			ListenHost:  "loaclhost",
			ListenPorts: []int{7501, 7502},
			Persistence: false,
		},
	}

	tmpdir, err := ioutil.TempDir("", "rcm_cluster_conf_test")

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpdir)

	for _, conf := range cases {

		fname := tmpdir + "/" + randStringRunes(6) + ".yml"

		err := SaveClusterConf(fname, &conf)

		if err != nil {
			t.Fatal(err)
		}

		loadedConf, err := LoadClusterConf(fname)

		if err != nil {
			t.Fatal(err)
		}

		if reflect.DeepEqual(loadedConf, conf) {
			t.Errorf("Expected %v but got %v", conf, *loadedConf)
		}
	}
}

func TestSaveNodeConf(t *testing.T) {
	cases := []RedisNodeConf{
		RedisNodeConf{
			ListenIp:    "127.0.0.1",
			ListenPort:  6379,
			Persistence: true,
			DataDir:     "/tmp",
			LogFile:     "/tmp/redis.log",
		},
		RedisNodeConf{
			ListenPort:  6379,
			Persistence: true,
			DataDir:     "/tmp",
			LogFile:     "/tmp/redis.log",
		},
	}

	tmpdir, err := ioutil.TempDir("", "rcm_cluster_conf_test")

	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpdir)

	for _, conf := range cases {

		fname := tmpdir + "/" + randStringRunes(6) + ".conf"

		SaveRedisConf(fname, &conf)

		data, err := ioutil.ReadFile(fname)

		if err != nil {
			t.Fatal(err)
		}

		dataStr := string(data)

		checkConfigParam := func(paramName string, expected string) {

			re := redisConfRegExps[paramName]

			if re == nil {
				panic(fmt.Sprintf("Ther is no pattern to extract value for '%s' parameter form redis config", paramName))
			}

			matches := re.FindStringSubmatch(dataStr)

			if expected == "" {
				if len(matches) != 0 {
					t.Errorf("'%s' record SHOULD NOT be present in the redis.conf file", paramName)
				}
			} else {
				if len(matches) != 2 {
					t.Errorf("'%s' record SHOULD be present in the redis.conf file", paramName)
				} else {
					listenHost := matches[1]
					if matches[1] != expected {
						t.Errorf("Expected %v but got %v for '%s' parameter", expected, listenHost, paramName)
					}
				}
			}
		}

		var expectedListenPortStr string

		if conf.ListenPort > 0 {
			expectedListenPortStr = strconv.Itoa(conf.ListenPort)
		} else {
			expectedListenPortStr = ""
		}

		checkConfigParam("bind", conf.ListenIp)
		checkConfigParam("port", expectedListenPortStr)

		checkConfigParam("dir", conf.DataDir)
	}
}

// Supporting code

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var redisConfRegExps = map[string]*regexp.Regexp{
	"bind": regexp.MustCompile(`bind\s*([0-9\w\._-]+)\s*\n`),
	"port": regexp.MustCompile(`port\s*([0-9]+)\s*\n`),
	"dir":  regexp.MustCompile(`dir\s*([0-9\w/\._-]+)\s*\n`),
}
