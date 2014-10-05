package main 

import (
	"testing"
	"fmt"
	"reflect"
)


func TestGetDevices(t *testing.T) {
	devices, err := getDevices()
	if len(devices) == 0 || devices == nil || err != nil{
		t.Errorf("Didn't found any devices, failed")
	}
}

func BenchmarkGetDevices(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		devices, err :=  getDevices()
		if devices == nil || err != nil {
			fmt.Println("Failed in devices")
		}
	}
}

func TestGetHandle(t *testing.T) {
	devices, err := getDevices()
	if err != nil {
		t.Errorf("Failed to initialize")
	}
	handle := getHandle(devices[0])
	typeHandle := fmt.Sprintf("%v", reflect.TypeOf(handle))
	if handle == nil || typeHandle != "*pcap.Pcap" {
		t.Errorf("No handle found")
	} 
}

func BenchmarkGetHandle(b *testing.B) {
	b.StopTimer()
	devices, err :=  getDevices()
	if devices == nil || err != nil {
		fmt.Println("Failed in devices")
	}
	b.StartTimer()
	for n := 0; n <= b.N; n++ {
		handle := getHandle(devices[0])
		if handle == nil {
			fmt.Println("No handle")
		}
	}
}