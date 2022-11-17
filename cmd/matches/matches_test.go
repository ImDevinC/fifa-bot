package main

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestHandleRequest(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	HandleRequest(context.TODO())
}
