package main

import (
	"syscall/js"
	"testing"
)

func TestGreet(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "World", want: "Hello, World!"},
		{name: "Go", want: "Hello, Go!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := greet(js.Null(), []js.Value{
				js.ValueOf(tt.name),
			})
			jsResult := result.(js.Value)
			if got := jsResult.String(); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{a: 1, b: 2, want: 3},
		{a: -1, b: 1, want: 0},
	}
	for _, tt := range tests {
		result := add(js.Null(), []js.Value{
			js.ValueOf(tt.a),
			js.ValueOf(tt.b),
		})
		jsResult := result.(js.Value)
		if got := jsResult.Int(); got != tt.want {
			t.Errorf("got %v, want %v", got, tt.want)
		}
	}
}

func TestGetInfo(t *testing.T) {
	result := getInfo(js.Null(), []js.Value{
		js.ValueOf("test-app"),
	})
	jsResult := result.(js.Value)
	if got := jsResult.Get("name").String(); got != "test-app" {
		t.Errorf("got name %v", got)
	}
	if got := jsResult.Get("active").Bool(); !got {
		t.Errorf("expected active to be true")
	}
}
